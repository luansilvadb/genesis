import type { IGastoRepository } from '../repositories/IGastoRepository'
import type { IFaturaRepository } from '../repositories/IFaturaRepository'
import type { ICartaoRepository } from '../repositories/ICartaoRepository'
import { Gasto, type PaymentMethod, type SplitMode } from '../entities/Gasto'
import { Dinheiro } from '../entities/Dinheiro'
import { DivisaoDeGasto } from '../entities/DivisaoDeGasto'
import type { Fatura } from '../entities/Fatura'
import { LancamentoService } from './LancamentoService'
import { resolverCartao, type CartaoResolvido } from './CartaoResolver'

export interface LancarGastoInput {
  flow: 'expense' | 'loan' | 'settlement'
  paymentMethod: PaymentMethod
  compradorId: string
  valor: number
  descricao: string
  divisoes: DivisaoDeGasto[]
  installments: number
  cardOwnerId: string | null
  borrowerId: string | null
  periodo: { mes: number; ano: number }
  isPrivate?: boolean
  splitMode: SplitMode
  settlementDetails?: {
    fromMemberId: string
    toMemberId: string
    method: 'pix' | 'cash'
  }
}

type AtualizarGastoDados = {
  descricao: string
  valorTotal: Dinheiro
  compradorId: string
  method: 'pix' | 'card'
  cardOwner: string | null
  divisoes: DivisaoDeGasto[]
  installments: number
}

export class GastoService {
  constructor(
    private gastoRepo: IGastoRepository,
    private faturaRepo: IFaturaRepository,
    private cartaoRepo: ICartaoRepository,
    private lancamentoService: LancamentoService = new LancamentoService(gastoRepo, faturaRepo, cartaoRepo)
  ) {}

  async lancarGastoOuEmprestimo(dados: LancarGastoInput): Promise<void> {
    await this.lancamentoService.lancarGastoOuEmprestimo(dados)
  }

  async excluirGasto(id: string): Promise<void> {
    await this.gastoRepo.excluir(id)
  }

  async lancarGastoContaFixa(dados: {
    faturaId: string
    conta: { id: string; name: string }
    valorCentavos: number
    compradorId: string
    participantes: string[]
  }): Promise<void> {
    await this.lancamentoService.lancarGastoContaFixa(dados)
  }

  async atualizarGastoCompleto(gastoId: string, dados: AtualizarGastoDados): Promise<void> {
    const original = await this.gastoRepo.buscarPorId(gastoId)
    if (!original) throw new Error('Gasto não encontrado')

    const todosCartoes = (await this.cartaoRepo.listarTodos()) || []
    const cartaoResolvido = resolverCartao(
      dados.method,
      dados.cardOwner,
      dados.compradorId,
      todosCartoes
    )

    if (original.grupoParcelasId) {
      await this.atualizarGrupoParcelas(original, dados, cartaoResolvido)
      return
    }

    if (dados.method === 'card' && dados.installments > 1) {
      await this.relancarGasto(original, [original.id], dados)
      return
    }

    await this.atualizarGastoIndividual(original, dados, cartaoResolvido)
  }

  private async atualizarGrupoParcelas(
    original: Gasto,
    dados: AtualizarGastoDados,
    cartaoResolvido: CartaoResolvido
  ): Promise<void> {
    const gastosDoGrupo = (await this.gastoRepo.listarTodos())
      .filter(g => g.grupoParcelasId === original.grupoParcelasId)

    const estruturaMudou = original.totalInstallments !== dados.installments || original.method !== dados.method
    if (estruturaMudou) {
      const primeiraParcela = gastosDoGrupo.reduce(
        (anterior, atual) => atual.installments > anterior.installments ? atual : anterior,
        gastosDoGrupo[0] || original
      )
      await this.relancarGasto(primeiraParcela, gastosDoGrupo.map(g => g.id), dados)
      return
    }

    await this.salvarParcelasAtualizadas(gastosDoGrupo, dados, cartaoResolvido)
  }

  private async relancarGasto(original: Gasto, idsParaExcluir: string[], dados: AtualizarGastoDados): Promise<void> {
    const periodo = original.faturaId
      ? (await this.faturaRepo.buscarPorId(original.faturaId))?.periodo
      : { mes: original.createdAt.getMonth() + 1, ano: original.createdAt.getFullYear() }

    if (!periodo) throw new Error(`Fatura ou período original não encontrado para o gasto ${original.id}`)

    // Create replacement gasto(s) FIRST. If this fails, nothing is lost.
    await this.lancarGastoOuEmprestimo({
      flow: original.isSettlement ? 'settlement' : original.isLoan ? 'loan' : 'expense',
      paymentMethod: dados.method,
      compradorId: dados.compradorId,
      valor: dados.valorTotal.centavos / 100,
      descricao: dados.descricao,
      divisoes: dados.divisoes,
      installments: dados.installments,
      cardOwnerId: dados.cardOwner,
      borrowerId: original.borrowerId,
      periodo: periodo,
      isPrivate: original.isPrivate,
      splitMode: original.splitMode,
      settlementDetails: original.settlementDetails
        ? {
            fromMemberId: original.settlementDetails.fromMemberId,
            toMemberId: original.settlementDetails.toMemberId,
            method: original.settlementDetails.method,
          }
        : undefined
    })

    // Delete originals AFTER successful recreation.
    // If this fails, duplicates exist — the caller should sync to detect them.
    try {
      if (idsParaExcluir.length === 1) await this.gastoRepo.excluir(idsParaExcluir[0])
      else await this.gastoRepo.excluirMuitos(idsParaExcluir)
    } catch (deleteErr) {
      throw new Error(
        `Gasto recriado com sucesso, mas falha ao excluir ${idsParaExcluir.length} gasto(s) original(is): ${(deleteErr as Error).message}. ` +
        `Recarregue a página para evitar duplicatas. IDs: ${idsParaExcluir.join(', ')}`
      )
    }
  }

  private async salvarParcelasAtualizadas(
    gastos: Gasto[],
    dados: AtualizarGastoDados,
    cartaoResolvido: CartaoResolvido
  ): Promise<void> {
    const faturasParaSalvar: Fatura[] = []
    const gastosParaSalvar: Gasto[] = []
    const faturasPersistidas = await this.faturaRepo.listarTodas()

    for (const gasto of gastos) {
      const faturaAtual = gasto.faturaId ? (faturasPersistidas.find(f => f.id === gasto.faturaId) ?? null) : null
      let faturaId = gasto.faturaId
      if (faturaAtual && cartaoResolvido.cartaoId) {
        const novaFatura = await this.lancamentoService.obterOuCriarFaturaMemoria(
          cartaoResolvido.cartaoId,
          faturaAtual.periodo.mes,
          faturaAtual.periodo.ano,
          cartaoResolvido.responsavelFaturaId,
          faturasParaSalvar,
          faturasPersistidas
        )
        faturaId = novaFatura.id
      }
      gastosParaSalvar.push(this.criarGastoAtualizado(
        gasto,
        dados,
        faturaId,
        cartaoResolvido.cardOwner,
        gasto.installments,
        gasto.totalInstallments
      ))
    }

    if (faturasParaSalvar.length > 0) await this.faturaRepo.salvarMuitas(faturasParaSalvar)

    // Resolve any composite fatura IDs (cartaoId-mes-ano) to real backend-generated
    // UUIDs. Build the lookup map from already-loaded faturas plus newly-saved ones
    // to avoid a redundant full listarTodas() call.
    const todasFaturas = faturasPersistidas.concat(faturasParaSalvar)
    const idRealPorComposto = new Map<string, string>()
    for (const f of todasFaturas) {
      const chave = `${f.cartaoId}-${f.periodo.mes}-${f.periodo.ano}`
      idRealPorComposto.set(chave, f.id)
    }
    // Composite IDs have the format: <cartao-uuid>-<mes>-<ano> (3 segments, last two are numbers).
    const compositoPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}-\d{1,2}-\d{4}$/
    for (const g of gastosParaSalvar) {
      if (g.faturaId && compositoPattern.test(g.faturaId)) {
        const realId = idRealPorComposto.get(g.faturaId)
        if (realId) {
          ;(g as { faturaId: string | null }).faturaId = realId
        }
      }
    }

    if (gastosParaSalvar.length > 0) {
      const erros: { id: string; message: string }[] = []
      for (const g of gastosParaSalvar) {
        try {
          await this.gastoRepo.atualizar(g.id, g)
        } catch (e) {
          erros.push({ id: g.id, message: (e as Error).message })
        }
      }
      if (erros.length > 0) {
        const detalhes = erros.map(e => `${e.id}: ${e.message}`).join('; ')
        throw new Error(`Falha ao atualizar ${erros.length} de ${gastosParaSalvar.length} parcelas: ${detalhes}`)
      }
    }
  }

  private async atualizarGastoIndividual(
    original: Gasto,
    dados: AtualizarGastoDados,
    cartaoResolvido: CartaoResolvido
  ): Promise<void> {
    let faturaId: string | null = original.faturaId

    if (cartaoResolvido.cartaoId) {
      const periodo = original.faturaId
        ? (await this.faturaRepo.buscarPorId(original.faturaId))?.periodo
        : { mes: original.createdAt.getMonth() + 1, ano: original.createdAt.getFullYear() }

      if (periodo) {
        const novaFatura = await this.faturaRepo.assegurarObterOuCriarFatura(
          cartaoResolvido.cartaoId,
          periodo.mes,
          periodo.ano,
          cartaoResolvido.responsavelFaturaId
        )
        faturaId = novaFatura.id
      }
    } else {
      faturaId = null
    }

    const atualizado = this.criarGastoAtualizado(
      original,
      dados,
      faturaId,
      cartaoResolvido.cardOwner,
      dados.installments,
      dados.installments
    )
    await this.gastoRepo.atualizar(atualizado.id, atualizado)
  }

  private criarGastoAtualizado(
    original: Gasto,
    dados: AtualizarGastoDados,
    faturaId: string | null,
    cardOwner: string | null,
    installments: number,
    totalInstallments: number
  ): Gasto {
    return new Gasto({
      id: original.id,
      faturaId,
      descricao: dados.descricao,
      valorTotal: dados.valorTotal,
      compradorId: dados.compradorId,
      divisoes: dados.divisoes,
      method: dados.method,
      cardOwner,
      installments,
      totalInstallments,
      grupoParcelasId: original.grupoParcelasId,
      isLoan: original.isLoan,
      borrowerId: original.borrowerId,
      recurringBillId: original.recurringBillId,
      isSettlement: original.isSettlement,
      settlementDetails: original.settlementDetails,
      isPrivate: original.isPrivate,
      splitMode: original.splitMode
    })
  }

  async removerAssociacaoContaFixa(contaFixaId: string): Promise<void> {
    const gastosAssociados = (await this.gastoRepo.listarTodos()).filter(g => g.recurringBillId === contaFixaId)
    if (gastosAssociados.length === 0) return

    // Use PUT (atualizar) for existing gastos — salvarMuitos uses POST /batch
    // which would create duplicates instead of updating.
    await Promise.all(gastosAssociados.map(g =>
      this.gastoRepo.atualizar(g.id, this.desassociarContaFixa(g))
    ))
  }

  private desassociarContaFixa(g: Gasto): Gasto {
    return new Gasto({
      id: g.id,
      faturaId: g.faturaId,
      descricao: g.descricao,
      valorTotal: g.valorTotal,
      compradorId: g.compradorId,
      divisoes: g.divisoes,
      installments: g.installments,
      totalInstallments: g.totalInstallments,
      isLoan: g.isLoan,
      borrowerId: g.borrowerId,
      recurringBillId: null,
      isSettlement: g.isSettlement,
      settlementDetails: g.settlementDetails,
      method: g.method,
      cardOwner: g.cardOwner,
      grupoParcelasId: g.grupoParcelasId,
      isPrivate: g.isPrivate,
      splitMode: g.splitMode
    })
  }
}
