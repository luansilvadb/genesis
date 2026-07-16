import { Gasto } from '../entities/Gasto'
import { Dinheiro } from '../entities/Dinheiro'
import { valorParcelaAtual } from '../entities/ParcelaCalculator'

interface ItemExtratoPessoal {
  id: string
  descricao: string
  valorPago: Dinheiro
  valorConsumido: Dinheiro
  valorLiquido: Dinheiro
  saldoAcumulado: Dinheiro
  isLoan: boolean
  isSettlement: boolean
  compradorId: string
  borrowerId: string | null
  createdAt: Date
}

interface SaldoPessoalPendencia {
  id: string
  nome: string
  saldo: Dinheiro
}

// ── Helpers privados ──

/** Retorna quanto o membro pagou e consumiu em um gasto privado. */
function calcularValoresPessoais(g: Gasto, membroId: string): { pagos: number; consumidos: number } {
  if (g.isLoan) {
    const vp = valorParcelaAtual(g.valorTotal, g.installments, g.totalInstallments).centavos
    return {
      pagos: g.compradorId === membroId ? vp : 0,
      consumidos: g.borrowerId === membroId ? vp : 0,
    }
  }

  if (g.isSettlement && g.settlementDetails) {
    const vp = valorParcelaAtual(g.valorTotal, g.installments, g.totalInstallments).centavos
    return {
      pagos: g.settlementDetails.fromMemberId === membroId ? vp : 0,
      consumidos: g.settlementDetails.toMemberId === membroId ? vp : 0,
    }
  }

  // Gasto comum privado
  const pagadorId = (g.method === 'card' && g.cardOwner) ? g.cardOwner : g.compradorId
  let totalDebitos = 0
  let consumidos = 0
  g.divisoes.forEach(div => {
    const vd = valorParcelaAtual(div.valor, g.installments, g.totalInstallments)
    if (vd.centavos <= 0) return
    totalDebitos += vd.centavos
    if (div.membroId === membroId) consumidos += vd.centavos
  })
  return {
    pagos: pagadorId === membroId ? totalDebitos : 0,
    consumidos,
  }
}

/** Ajusta saldos de pendências para um gasto privado (loan ou settlement). */
function acumularSaldosPessoais(g: Gasto, membroId: string, saldos: Record<string, number>): void {
  if (g.isLoan) {
    const vp = valorParcelaAtual(g.valorTotal, g.installments, g.totalInstallments)
    if (vp.centavos <= 0) return
    const v = vp.centavos
    if (g.compradorId === membroId && g.borrowerId && g.borrowerId !== membroId)
      saldos[g.borrowerId] = (saldos[g.borrowerId] || 0) + v
    if (g.compradorId !== membroId && g.borrowerId === membroId)
      saldos[g.compradorId] = (saldos[g.compradorId] || 0) - v
    return
  }

  if (g.isSettlement && g.settlementDetails) {
    const vp = valorParcelaAtual(g.valorTotal, g.installments, g.totalInstallments)
    if (vp.centavos <= 0) return
    const v = vp.centavos
    const { fromMemberId, toMemberId } = g.settlementDetails
    if (fromMemberId === membroId && toMemberId !== membroId)
      saldos[toMemberId] = (saldos[toMemberId] || 0) + v
    if (fromMemberId !== membroId && toMemberId === membroId)
      saldos[fromMemberId] = (saldos[fromMemberId] || 0) - v
  }
}

// ── API pública ──

export class ExtratoPessoalService {
  /**
   * Calcula o extrato de movimentações pessoais (gastos privados do próprio membro).
   */
  static obterExtratoPessoal(membroId: string, gastos: Gasto[]): ItemExtratoPessoal[] {
    const gastosPrivados = gastos.filter(g => {
      if (!g.isPrivate) return false
      const envolvidoNasDivisoes = g.divisoes.some(d => d.membroId === membroId)
      return g.compradorId === membroId || g.cardOwner === membroId || g.borrowerId === membroId || envolvidoNasDivisoes
    })

    const ordenados = [...gastosPrivados].sort((a, b) => a.createdAt.getTime() - b.createdAt.getTime())
    let saldoAcumulado = Dinheiro.deCentavos(0)

    return ordenados.map(g => {
      const { pagos, consumidos } = calcularValoresPessoais(g, membroId)

      const valorPago = Dinheiro.deCentavos(pagos)
      const valorConsumido = Dinheiro.deCentavos(consumidos)
      const valorLiquido = valorPago.subtrair(valorConsumido)
      saldoAcumulado = saldoAcumulado.somar(valorLiquido)

      return {
        id: g.id, descricao: g.descricao,
        valorPago, valorConsumido, valorLiquido, saldoAcumulado,
        isLoan: g.isLoan, isSettlement: g.isSettlement,
        compradorId: g.compradorId, borrowerId: g.borrowerId,
        createdAt: g.createdAt,
      }
    })
  }

  /**
   * Calcula saldos liquidados com outras pessoas (externas ou da casa).
   */
  static calcularSaldosPessoais(membroId: string, gastos: Gasto[]): SaldoPessoalPendencia[] {
    const saldosCentavos: Record<string, number> = {}
    
    for (const g of gastos) {
      if (!g.isPrivate) continue
      acumularSaldosPessoais(g, membroId, saldosCentavos)
    }

    const lista: SaldoPessoalPendencia[] = []
    for (const key in saldosCentavos) {
      const s = saldosCentavos[key]
      if (s === 0) continue
      lista.push({
        id: key,
        nome: key.startsWith('externo:') ? key.substring(8) : '',
        saldo: Dinheiro.deCentavos(s),
      })
    }
    return lista
  }
}
