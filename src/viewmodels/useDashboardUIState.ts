import { ref } from 'vue'
import type { Gasto } from '../models/entities/Gasto'
import type { ContaFixa } from '../models/entities/ContaFixa'
import type { AuditLogDto } from '../models/repositories/http/HttpAuditLogRepository'
import type { TransferenciaNetting } from '../models/services/NettingService'

export function useDashboardUIState() {
  const modalStack = ref<string[]>([])
  const abrirModal = (n: string) => !modalStack.value.includes(n) && modalStack.value.push(n)
  const fecharModal = (n: string) => modalStack.value = modalStack.value.filter(x => x !== n)
  const isModalNoTopo = (n: string) => modalStack.value[modalStack.value.length - 1] === n

  const gastoParaAjustar = ref<Gasto | null>(null)
  const billSelecionada = ref<ContaFixa | null>(null)
  const isSubmittingPix = ref(false)
  const isSubmittingAjusteGasto = ref(false)
  const isSubmittingLancarBill = ref(false)
  const isSubmittingSalvarTemplate = ref(false)
  const itemParaEstornar = ref<Gasto | ContaFixa | null>(null)
  const itemTypeParaEstornar = ref('')
  const nettingTarget = ref<TransferenciaNetting | null>(null)
  
  const auditLogs = ref<AuditLogDto[]>([])
  const isLogsLoading = ref(false)

  return {
    modalStack, abrirModal, fecharModal, isModalNoTopo,
    gastoParaAjustar, billSelecionada,
    isSubmittingPix, isSubmittingAjusteGasto, isSubmittingLancarBill, isSubmittingSalvarTemplate,
    itemParaEstornar, itemTypeParaEstornar,
    nettingTarget,
    auditLogs, isLogsLoading,
    abrirConfirmacaoEstornoGasto: (g: Gasto) => { itemParaEstornar.value = g; itemTypeParaEstornar.value = 'Lançamento'; abrirModal('confirmacao-estorno') },
    abrirConfirmacaoEstornoBill: (b: ContaFixa) => { itemParaEstornar.value = b; itemTypeParaEstornar.value = 'Conta Fixa'; abrirModal('confirmacao-estorno') },
    abrirLancarBill: (b: ContaFixa) => { billSelecionada.value = b; abrirModal('lancar-conta-fixa') },
    abrirConfigurarBill: (b: ContaFixa) => { billSelecionada.value = b; abrirModal('configurar-conta-fixa') },
    abrirNovoBill: () => { billSelecionada.value = null; abrirModal('configurar-conta-fixa') },
    abrirAjustarGasto: (g: Gasto) => { gastoParaAjustar.value = g; abrirModal('ajustar-gasto') },
    abrirBottomSheetNetting: (t: TransferenciaNetting) => { nettingTarget.value = t; abrirModal('acerto-netting') },
    abrirNovoPeriodoBottomSheet: () => {
      abrirModal('novo-periodo')
    }
  }
}
