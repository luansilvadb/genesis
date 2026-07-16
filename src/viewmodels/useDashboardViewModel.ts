import { computed } from 'vue'
import { Fatura } from '../models/entities/Fatura'
import { Cartao } from '../models/entities/Cartao'
import { useCartoesEFaturas } from './useCartoesEFaturas'
import { useContasFixas } from './useContasFixas'
import { useDashboardUIState } from './useDashboardUIState'
import { useDashboardPeriodos } from './useDashboardPeriodos'
import { useDashboardNetting } from './useDashboardNetting'
import { valorParcelaAtual } from '../models/entities/ParcelaCalculator'
import { formatarMesAno, NOMES_MESES } from '../shared/utils/meses'
import { useToast } from '../composables/useToast'
import { gastoPertenceAoPeriodo } from '../shared/utils/gastoPeriodo'
import { gastoService, faturaService, auditLogRepository } from '../shared/container'
import { Dinheiro } from '../models/entities/Dinheiro'
import { DivisaoDeGasto } from '../models/entities/DivisaoDeGasto'
import type { Gasto } from '../models/entities/Gasto'
import type { ContaFixa } from '../models/entities/ContaFixa'

interface ConfirmarAjusteGastoInput {
  descricao: string
  valorTotal: Dinheiro
  compradorId: string
  method: 'pix' | 'card'
  cardOwner: string | null
  divisoes: DivisaoDeGasto[]
  installments: number
}

interface ConfirmarLancarBillInput {
  valorCentavos: number
  compradorId: string
  splitIds: string[]
}

export interface DashboardProps { membros: { id: string; nome: string }[]; faturasAbertas: Fatura[]; faturasFechadas: Fatura[]; cartoes: Cartao[] }

const isGastoValidoParaSoma = (gasto: Gasto) => !gasto.id.startsWith('audit-settlement-') && (!gasto.isSettlement || gasto.descricao.includes('Saldo Inicial'));
const isGastoComum = (gasto: Gasto) => !gasto.isSettlement;

/** Retorna o mês/ano seguinte para abertura de novas faturas. */
const proximoPeriodo = (mes: number, ano: number): { mes: number; ano: number } =>
  mes === 12 ? { mes: 1, ano: ano + 1 } : { mes: mes + 1, ano };

/**
 * Estorna um lançamento, tratando separadamente acertos (settlements)
 * que precisam de reabertura de fatura em vez de exclusão direta.
 */
async function estornarLancamento(
  item: Gasto,
  toast: ReturnType<typeof useToast>,
  gastoSvc: typeof gastoService,
  cartoesEFaturas: ReturnType<typeof useCartoesEFaturas>,
): Promise<void> {
  // Settlements não podem ser excluídos diretamente — reabre a fatura.
  if (item.id.startsWith('audit-settlement-')) {
    if (item.faturaId) {
      await cartoesEFaturas.reabrirFatura(item.faturaId)
      toast.show('Período reaberto para estorno do acerto.', 'success')
      return
    }
    toast.show('Acertos não possuem fatura associada. Exclua o lançamento diretamente ou reabra o período manualmente.', 'error')
    return
  }

  await gastoSvc.excluirGasto(item.id)
  await cartoesEFaturas.inicializar()
  toast.show('Estornado', 'success')
}

export const useDashboardViewModel = (
  props: DashboardProps,
  emit: (event: 'periodoStatusChanged', isLocked: boolean) => void
) => {
  const ui = useDashboardUIState()
  const toast = useToast()
  const cartoesEFaturas = useCartoesEFaturas()
  const contasFixas = useContasFixas()
  const periodosState = useDashboardPeriodos(() => props.faturasAbertas, () => props.faturasFechadas, () => props.cartoes, () => props.membros, emit)

  const reinicializarAposErro = async (contexto: string) => {
    try { await cartoesEFaturas.inicializar() } catch {
      console.error(`Falha ao reinicializar após ${contexto}`)
    }
  }
  
  const periodoSelecionado = periodosState.periodoSelecionado
  const todosGastos = cartoesEFaturas.gastos

  const gastosFiltrados = computed(() => todosGastos.value.filter(gasto => !gasto.isPrivate && gastoPertenceAoPeriodo(gasto, periodoSelecionado.value.mes, periodoSelecionado.value.ano, cartoesEFaturas.faturas.value)))
  const gastosPrivadosFiltrados = computed(() => todosGastos.value.filter(gasto => gasto.isPrivate && gastoPertenceAoPeriodo(gasto, periodoSelecionado.value.mes, periodoSelecionado.value.ano, cartoesEFaturas.faturas.value)))
  
  return {
    ...periodosState, 
    ...useDashboardNetting(computed(() => props.membros), gastosFiltrados), 
    ...ui, 
    contasFixas: contasFixas.contasFixas, 
    gastosFaturaSelecionada: gastosFiltrados, 
    gastosPrivadosFiltrados,
    todosGastos,
    
    totalPeriodoSelecionado: computed(() => gastosFiltrados.value.filter(isGastoComum).reduce((soma, gasto) => soma + valorParcelaAtual(gasto.valorTotal, gasto.installments, gasto.totalInstallments).centavos, 0)),
    totalLancamentosPeriodoSelecionado: computed(() => gastosFiltrados.value.filter(isGastoValidoParaSoma).length),
    
    getMembroNome: (id: string | undefined) => id ? props.membros.find(m => m.id === id)?.nome || 'Desconhecido' : 'Sistema', 
    formatarDinheiro: (centavos: number) => centavos / 100, 
    formatarMesAno, 
    showToast: toast.show,
    
    abrirNovoPeriodoBottomSheet: () => ui.abrirNovoPeriodoBottomSheet(),
    
    confirmarBaixaNetting: async (dados: { from: string; to: string; valor: number; method: 'pix' | 'cash'; descricao: string }) => {
      const fatura = periodosState.faturaPixPeriodoSelecionado.value
      if (!fatura) {
        toast.show('Não foi possível encontrar uma fatura para este período. Verifique se há cartões configurados.', 'error')
        return
      }
      
      ui.isSubmittingPix.value = true
      try {
        await gastoService.lancarGastoOuEmprestimo({
          flow: 'settlement',
          paymentMethod: dados.method,
          compradorId: dados.from,
          valor: dados.valor,
          descricao: dados.descricao,
          divisoes: [new DivisaoDeGasto(dados.to, Dinheiro.deReais(dados.valor))],
          installments: 1,
          cardOwnerId: null,
          borrowerId: null,
          periodo: fatura.periodo,
          splitMode: 'custom',
          settlementDetails: {
            fromMemberId: dados.from,
            toMemberId: dados.to,
            method: dados.method,
          }
        })
        
        await cartoesEFaturas.inicializar()
        ui.fecharModal('acerto-netting')
        toast.show('Acerto registrado!', 'success')
      } finally {
        ui.isSubmittingPix.value = false
      }
    },
    
    confirmarAjusteGasto: async (dados: ConfirmarAjusteGastoInput) => {
      ui.isSubmittingAjusteGasto.value = true
      try {
        await cartoesEFaturas.atualizarGasto(ui.gastoParaAjustar.value!.id, dados)
        ui.fecharModal('ajustar-gasto')
        ui.gastoParaAjustar.value = null
      } finally {
        ui.isSubmittingAjusteGasto.value = false
      }
    },
    
    confirmarLancarBill: async (dados: ConfirmarLancarBillInput) => {
      ui.isSubmittingLancarBill.value = true
      try {
        await contasFixas.lancarGastoContaFixa(periodosState.faturaPixPeriodoSelecionado.value!.id, ui.billSelecionada.value!, dados.valorCentavos, dados.compradorId, dados.splitIds)
        ui.fecharModal('lancar-conta-fixa')
        await cartoesEFaturas.inicializar()
      } finally {
        ui.isSubmittingLancarBill.value = false
      }
    },
    
    confirmarSalvarTemplate: async (t: ContaFixa) => {
      ui.isSubmittingSalvarTemplate.value = true
      try {
        await contasFixas.salvarContaFixa(t)
        ui.fecharModal('configurar-conta-fixa')
      } catch (e) {
        console.error('Erro ao salvar template:', e)
        toast.show('Erro ao salvar conta fixa. Tente novamente.', 'error')
      } finally {
        ui.isSubmittingSalvarTemplate.value = false
      }
    },
    
    confirmarNovoPeriodo: async () => {
      const { mes, ano } = periodoSelecionado.value
      const faturasAbertas = periodosState.faturasPeriodoSelecionado.value.filter(f => f.status === 'ABERTA')
      try {
        await Promise.all(faturasAbertas.map(f => faturaService.fecharFatura(f.id, f.responsavelId, new Date())))
        const prox = proximoPeriodo(mes, ano)
        await faturaService.assegurarFaturasAbertas(cartoesEFaturas.cartoes.value, prox.mes, prox.ano)
        await cartoesEFaturas.inicializar()
        ui.fecharModal('novo-periodo')
        toast.show(`Mês de ${NOMES_MESES[mes - 1]} encerrado!`, 'success')
      } catch (e) {
        console.error('Erro ao fechar período:', e)
        await reinicializarAposErro('erro de fechamento de período')
        toast.show('Erro ao encerrar o mês. Algumas faturas podem não ter sido fechadas. Tente novamente.', 'error')
      }
    },
    
    confirmarEstorno: async () => {
      const i = ui.itemParaEstornar.value!
      const t = ui.itemTypeParaEstornar.value
      
      if (t === 'Lançamento') {
        ui.fecharModal('confirmacao-estorno')
        ui.itemParaEstornar.value = null
        await estornarLancamento(i as Gasto, toast, gastoService, cartoesEFaturas)
      } else {
        // Fecha ambos os modais no mesmo tick para evitar flicker visual:
        // o confirmacao-estorno está sobre o configurar-conta-fixa na stack.
        ui.fecharModal('configurar-conta-fixa')
        ui.fecharModal('confirmacao-estorno')
        ui.itemParaEstornar.value = null
        await contasFixas.excluirContaFixa(i.id)
        toast.show('Conta removida', 'success')
      }
    },
    
    estornarContaFixa: (b: ContaFixa) => {
      const gasto = gastosFiltrados.value.find(z => z.recurringBillId === b.id)
      if (!gasto) {
        toast.show('Nenhum gasto vinculado a esta conta fixa no período', 'error')
        return
      }
      ui.abrirConfirmacaoEstornoGasto(gasto)
    },
    

    reabrirPeriodoSelecionado: async () => {
        const { mes } = periodoSelecionado.value
        try {
            await Promise.all(
                periodosState.faturasPeriodoSelecionado.value
                    .filter(f => f.status === 'FECHADA')
                    .map(f => faturaService.reabrirFatura(f.id))
            )
            await cartoesEFaturas.inicializar()
            toast.show(`Mês de ${NOMES_MESES[mes - 1]} reaberto!`, 'success')
        } catch (e) {
            console.error('Erro ao reabrir período:', e)
            await reinicializarAposErro('erro de reabertura de período')
            toast.show('Erro ao reabrir o mês. Algumas faturas podem não ter sido reabertas.', 'error')
        }
    },

    abrirAuditLogs: async () => {
      if (ui.isLogsLoading.value) return
      
      ui.isLogsLoading.value = true
      ui.abrirModal('audit-logs')
      
      try {
        ui.auditLogs.value = await auditLogRepository.listarTodos()
      } catch {
        toast.show('Erro ao carregar histórico de atividades.', 'error')
        ui.auditLogs.value = []
      } finally {
        ui.isLogsLoading.value = false
      }
    }
  }
}

export type DashboardViewModel = ReturnType<typeof useDashboardViewModel>
