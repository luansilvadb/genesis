<!-- src/views/screens/DashboardSaldos.vue -->
<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Tab } from '../components/ui/BottomTabBar.vue'
import { useDashboardViewModel } from '../../viewmodels/useDashboardViewModel'
import { useCasasMultitenant } from '../../viewmodels/useCasasMultitenant'
import { useMembros } from '../../viewmodels/useMembros'
import type { Fatura } from '../../models/entities/Fatura'
import type { Cartao } from '../../models/entities/Cartao'
import ContasFixasPanel from '../components/ledger/ContasFixasPanel.vue'
import ActivityFeed from '../components/ledger/ActivityFeed.vue'
import DashboardHeader from '../components/ledger/dashboard/DashboardHeader.vue'
import UnifiedBalancePanel from '../components/ledger/dashboard/UnifiedBalancePanel.vue'
import NettingPanel from '../components/ledger/dashboard/NettingPanel.vue'


import DashboardModalsManager from './DashboardModalsManager.vue'
import IllustrationMascot from '../components/ui/IllustrationMascot.vue'
import SkeletonMimic from '../components/ui/SkeletonMimic.vue'

import PersonalBalancePanel from '../components/ledger/dashboard/PersonalBalancePanel.vue'

interface Props {
  membros: { id: string; nome: string; ativo?: boolean }[]
  faturasAbertas: Fatura[]
  faturasFechadas: Fatura[]
  cartoes: Cartao[]
  activeTab?: Tab
  isLoading?: boolean
  isReadOnly?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits(['periodoStatusChanged', 'navigate-home', 'navigate-pessoal'])

const vm = useDashboardViewModel(props, emit)

const {
  faturaSelecionadaFechada,
  saldosUnificadosAtivos,
  nettingTransferencias,
  membrosVisiveis,
  contasFixas,
  gastosFaturaSelecionada,
  getMembroNome,
  currentMonthName,
  currentYear,
  abrirLancarBill,
  abrirConfigurarBill,
  abrirNovoBill,
  abrirAjustarGasto,
  abrirConfirmacaoEstornoGasto,
  estornarContaFixa,
  totalLancamentosPeriodoSelecionado,
  gastosPrivadosFiltrados
} = vm

const {
  isAuthed,
  activeTenantId,
  casas,
  isCreating,
  isEntering,
  form,
  copiedCode,
  activeTenantObj,
  selecionarCasa,
  criarNovaCasa,
  entrarPorCodigo,
  copyInviteCode,
  handleLogoutClick
} = useCasasMultitenant()

const { tenantPermissions, currentMembro } = useMembros()

const obterBloqueioFlag = (flagName: 'ALLOW_LANCAR_GASTO' | 'ALLOW_GERENCIAR_CARTOES' | 'ALLOW_GERENCIAR_CONTAS_FIXAS' | 'ALLOW_REGISTRAR_NETTING' | 'ALLOW_VER_AUDIT_LOGS' | 'ALLOW_FECHAR_PERIODO') => {
  if (props.isReadOnly) return true
  const role = currentMembro.value?.role
  if (!role || role === 'ADMIN') return false
  const perms = tenantPermissions.value[role]
  const defaultAllow = role === 'MORADOR'
  const allowed = perms ? perms[flagName] : defaultAllow
  return !allowed
}

const isLancarGastoBloqueado = computed(() => obterBloqueioFlag('ALLOW_LANCAR_GASTO'))
const isGerenciarContasFixasBloqueado = computed(() => obterBloqueioFlag('ALLOW_GERENCIAR_CONTAS_FIXAS'))
const isRegistrarNettingBloqueado = computed(() => obterBloqueioFlag('ALLOW_REGISTRAR_NETTING'))
const isVerAuditLogsBloqueado = computed(() => obterBloqueioFlag('ALLOW_VER_AUDIT_LOGS'))
const isFecharPeriodoBloqueado = computed(() => obterBloqueioFlag('ALLOW_FECHAR_PERIODO'))

const abrirAuditLogs = () => {
  if (isVerAuditLogsBloqueado.value) return
  vm.abrirAuditLogs()
}

const isHoje = computed(() => !props.activeTab || props.activeTab === 'hoje')
const isPessoal = computed(() => props.activeTab === 'pessoal')
const membrosAtivos = computed(() => props.membros.filter(m => m.ativo !== false))

const transitionName = ref('tab-slide-right')
const tabOrder: Tab[] = ['hoje', 'pessoal']

watch(() => props.activeTab, (newTab, oldTab) => {
  const newIndex = tabOrder.indexOf(newTab || 'hoje')
  const oldIndex = tabOrder.indexOf(oldTab || 'hoje')

  if (newIndex > oldIndex) {
    transitionName.value = 'tab-slide-right'
  } else if (newIndex < oldIndex) {
    transitionName.value = 'tab-slide-left'
  }
})

defineExpose({
  periodoSelecionado: vm.periodoSelecionado,
  currentMonthName,
  currentYear,
  faturaSelecionadaFechada,
  abrirHistorico: () => vm.abrirModal('historico'),
})
</script>

<template>
  <div class="space-y-12">
    <SkeletonMimic
      v-if="props.isLoading"
      :variant="'hoje'"
      key="skeleton"
      data-testid="skeleton-mimic"
    />
    <div v-else key="content" class="space-y-12">
      <DashboardHeader
        :current-year="currentYear"
        :current-month-name="currentMonthName"
        :fatura-selecionada-fechada="faturaSelecionadaFechada"
        :is-authed="isAuthed"
        :active-tenant-obj="activeTenantObj"
        :pode-ver-logs="!isVerAuditLogsBloqueado"
        @open-historico="vm.abrirModal('historico')"
        @open-casas="vm.abrirModal('casas')"
        @open-audit-logs="abrirAuditLogs"
        @navigate-home="emit('navigate-home')"
        @navigate-pessoal="emit('navigate-pessoal')"
      />

      <!-- Container Estabilizado -->
      <div class="relative overflow-x-hidden -mx-4 px-4 sm:-mx-6 sm:px-6">
        <Transition :name="transitionName" mode="out-in">
          <div v-if="isHoje" key="hoje" class="space-y-12 pb-2">
            <div v-if="totalLancamentosPeriodoSelecionado === 0" class="py-16 flex flex-col items-center justify-center text-center space-y-8 bg-parchment/30 rounded-3xl border-2 border-dashed border-stone/50 mx-1">
              <div class="animate-float">
                <IllustrationMascot variant="sky" :size="140" mood="sleeping" />
              </div>
              <div class="space-y-3 px-6">
                <h3 class="text-3xl font-display text-charcoal leading-tight">Comece pelas<br><span class="text-sky">Despesas</span></h3>
                <p class="text-sm text-graphite max-w-[280px] mx-auto leading-relaxed font-medium">
                  Registre a primeira despesa compartilhada para acompanhar os saldos e preparar o fechamento do mês.
                </p>
              </div>
            </div>

            <section class="space-y-4">
              <UnifiedBalancePanel
                :membros-visiveis="membrosVisiveis"
                :saldos-unificados-ativos="saldosUnificadosAtivos"
              />
            </section>

            <section v-if="nettingTransferencias.length > 0" class="space-y-4">
              <NettingPanel
                :netting-transferencias="nettingTransferencias"
                :fatura-selecionada-fechada="faturaSelecionadaFechada"
                :get-membro-nome="getMembroNome"
                :is-read-only="isRegistrarNettingBloqueado"
                @abrir-netting="vm.abrirBottomSheetNetting"
              />
            </section>

            <section class="space-y-4">
              <ContasFixasPanel
                :contasFixas="contasFixas"
                :gastos="gastosFaturaSelecionada"
                :membros="props.membros"
                :is-month-closed="faturaSelecionadaFechada"
                :is-read-only="isGerenciarContasFixasBloqueado"
                @lancar="abrirLancarBill"
                @configurar="abrirConfigurarBill"
                @novo="abrirNovoBill"
                @estornar="estornarContaFixa"
              />
            </section>

            <section class="space-y-4">
              <ActivityFeed
                :gastos="gastosFaturaSelecionada"
                :membros="props.membros"
                :is-month-closed="faturaSelecionadaFechada"
                :is-read-only="isLancarGastoBloqueado"
                @excluir="abrirConfirmacaoEstornoGasto"
                @ajustar="abrirAjustarGasto"
              />
            </section>
          </div>

          <div v-else-if="isPessoal" key="pessoal" class="space-y-12 pb-2">
            <PersonalBalancePanel
              :membros="props.membros"
              :gastos="gastosPrivadosFiltrados"
            />
          </div>
        </Transition>
      </div>
    </div>

    <DashboardModalsManager
      :vm="vm"
      :membrosAtivos="membrosAtivos"
      :cartoes="props.cartoes"
      :faturasAbertas="props.faturasAbertas"
      :faturasFechadas="props.faturasFechadas"
      :casasMultitenant="{ activeTenantId, casas, form, copiedCode, isCreating, isEntering, selecionarCasa, criarNovaCasa, entrarPorCodigo, copyInviteCode, handleLogoutClick }"
      :isFecharPeriodoBloqueado="isFecharPeriodoBloqueado"
    />
  </div>
</template>
