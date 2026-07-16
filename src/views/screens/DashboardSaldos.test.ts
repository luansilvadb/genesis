import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { ref } from 'vue'
import DashboardSaldos from './DashboardSaldos.vue'

vi.mock('../../viewmodels/useDashboardViewModel', () => ({
  useDashboardViewModel: () => ({
    faturaSelecionadaFechada: ref(false),
    saldosUnificadosAtivos: ref({}),
    nettingTransferencias: ref([]),
    membrosVisiveis: ref([]),
    contasFixas: ref([]),
    gastosFaturaSelecionada: ref([]),
    getMembroNome: vi.fn(),
    currentMonthName: ref('Junho'),
    currentYear: ref(2026),
    abrirLancarBill: vi.fn(),
    abrirConfigurarBill: vi.fn(),
    abrirNovoBill: vi.fn(),
    abrirAjustarGasto: vi.fn(),
    abrirConfirmacaoEstornoGasto: vi.fn(),
    abrirBottomSheetNetting: vi.fn(),
    abrirNovoPeriodoBottomSheet: vi.fn(),
    estornarContaFixa: vi.fn(),
    totalLancamentosPeriodoSelecionado: ref(1),
    reabrirPeriodoSelecionado: vi.fn(),
    isDropdownAbertosOpen: ref(false),
    periodoSelecionado: ref(null),
    abrirModal: vi.fn()
  })
}))

vi.mock('../../viewmodels/useCasasMultitenant', () => ({
  useCasasMultitenant: () => ({
    isAuthed: ref(true),
    activeTenantId: ref('tenant-1'),
    casas: ref([]),
    showBottomSheetCasas: ref(false),
    form: ref({}),
    copiedCode: ref(false),
    activeTenantObj: ref({ name: 'Casa' }),
    selecionarCasa: vi.fn(),
    criarNovaCasa: vi.fn(),
    entrarPorCodigo: vi.fn(),
    copyInviteCode: vi.fn(),
    handleLogoutClick: vi.fn()
  })
}))

const props = {
  membros: [],
  faturasAbertas: [],
  faturasFechadas: [],
  cartoes: [],
  activeTab: 'hoje' as const,
  isLoading: false
}

describe('DashboardSaldos loading', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  it('renderiza skeleton somente quando recebe espera real', async () => {
    const wrapper = mount(DashboardSaldos, {
      props,
      global: {
        stubs: {
          DashboardHeader: true,
          UnifiedBalancePanel: true,
          NettingPanel: true,
          ContasFixasPanel: true,
          ActivityFeed: true,
          DetalhamentoSaldosCard: true,
          DashboardModalsManager: true,
          IllustrationMascot: true,
          Card: { template: '<div><slot /></div>' },
          Button: { template: '<button><slot /></button>' }
        }
      }
    })

    expect(wrapper.find('[data-testid="skeleton-mimic"]').exists()).toBe(false)

    await wrapper.setProps({ activeTab: 'pessoal' })
    await vi.advanceTimersByTimeAsync(700)

    expect(wrapper.find('[data-testid="skeleton-mimic"]').exists()).toBe(false)

    await wrapper.setProps({ isLoading: true })

    expect(wrapper.find('[data-testid="skeleton-mimic"]').exists()).toBe(true)
  })
})
