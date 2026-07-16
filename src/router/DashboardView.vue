<script setup lang="ts">
import { inject } from 'vue'
import DashboardSaldos from '../views/screens/DashboardSaldos.vue'
import NovoLancamentoWizard from '../views/screens/NovoLancamentoWizard.vue'
import ConfiguracoesMembros from '../views/screens/ConfiguracoesMembros.vue'
import BottomSheet from '../views/components/ui/BottomSheet.vue'
import Drawer from '../views/components/ui/Drawer.vue'
import BottomTabBar, { type Tab } from '../views/components/ui/BottomTabBar.vue'
import { useToast } from '../composables/useToast'
import { ref, type Ref } from 'vue'
import type { Membro } from '../models/entities/Membro'
import type { Cartao } from '../models/entities/Cartao'
import type { Fatura } from '../models/entities/Fatura'

interface AppState {
  isAuthed: Ref<boolean>
  hasTenant: Ref<boolean>
  isLoading: Ref<boolean>
  membros: Ref<Membro[]>
  ativos: Ref<Membro[]>
  cartoes: Ref<Cartao[]>
  faturasAbertas: Ref<Fatura[]>
  faturasFechadas: Ref<Fatura[]>
  isReadOnly: Ref<boolean>
  isLancarGastoBloqueado: Ref<boolean>
  sincronizarDados: () => Promise<void>
  recarregarMembros: () => Promise<void>
}

defineEmits(['logout'])

const state = inject<AppState>('appState')!

const toast = useToast()

const currentView = ref<'dashboard' | 'wizard' | 'settings'>('dashboard')
const activeTab = ref<Tab>('hoje')
const isMonthClosed = ref(false)

const handlePeriodoStatusChanged = (fechado: boolean) => {
  isMonthClosed.value = fechado
}

const handleTabChange = (tab: Tab) => {
  if (tab === 'perfil') {
    currentView.value = 'settings'
  } else {
    activeTab.value = tab
    currentView.value = 'dashboard'
  }
}

const handleFabClick = () => {
  if (isMonthClosed.value) {
    toast.show('Este mês está encerrado. Reabra o período para fazer novos lançamentos.', 'error')
    return
  }
  if (state.isLancarGastoBloqueado.value) {
    toast.show('O administrador desativou a permissão de lançar despesas para o seu papel.', 'error')
    return
  }
  currentView.value = 'wizard'
}

const handleSalvarTransacao = () => {
  // Fecha o wizard imediatamente — o socket 'gastos_alterados' dispara
  // a sincronização em background (~300ms), sem bloquear a UI.
  currentView.value = 'dashboard'
}
</script>

<template>
  <div class="max-w-[75rem] mx-auto px-4 md:px-6 pt-2 md:pt-4 pb-20 md:pb-24 relative">
    <main class="relative z-10">
      <DashboardSaldos
        :membros="state.membros.value"
        :faturasAbertas="state.faturasAbertas.value"
        :faturasFechadas="state.faturasFechadas.value"
        :cartoes="state.cartoes.value"
        :is-loading="state.isLoading.value"
        :active-tab="activeTab"
        :is-read-only="state.isReadOnly.value"
        @openSettings="currentView = 'settings'"
        @periodoStatusChanged="handlePeriodoStatusChanged"
        @navigate-home="activeTab = 'hoje'"
        @navigate-pessoal="activeTab = 'pessoal'"
      />
    </main>

    <BottomSheet
      :model-value="currentView === 'wizard'"
      @update:model-value="(val: boolean) => { if (!val) currentView = 'dashboard' }"
      width-class="md:w-[560px]"
      max-height="95dvh"
      content-class="p-0 h-full"
    >
      <NovoLancamentoWizard
        v-if="currentView === 'wizard'"
        :membros="state.ativos.value"
        :is-private="activeTab === 'pessoal'"
        @salvar="handleSalvarTransacao"
        @cancelar="currentView = 'dashboard'"
      />
    </BottomSheet>

    <Drawer
      :model-value="currentView === 'settings'"
      @update:model-value="(val: boolean) => { if (!val) currentView = 'dashboard' }"
      width-class="md:max-w-[480px]"
      content-class="p-0 h-full"
    >
      <ConfiguracoesMembros
        @voltar="currentView = 'dashboard'"
        @logout="$emit('logout')"
      />
    </Drawer>

    <BottomTabBar
      :model-value="activeTab"
      :is-month-closed="isMonthClosed"
      :is-read-only="state.isLancarGastoBloqueado.value"
      @update:model-value="handleTabChange"
      @click-fab="handleFabClick"
    />
  </div>
</template>
