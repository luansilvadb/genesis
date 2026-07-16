<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ArrowLeft } from 'lucide-vue-next'
import { useMembros } from '../../viewmodels/useMembros'
import { useCasasMultitenant } from '../../viewmodels/useCasasMultitenant'
import Button from '../components/ui/Button.vue'
import PerfilUsuarioTab from '../components/settings/PerfilUsuarioTab.vue'
import GestaoAcessoTab from '../components/settings/GestaoAcessoTab.vue'
import ConfiguracoesCasaTab from '../components/settings/ConfiguracoesCasaTab.vue'
import AuditLogsTab from '../components/settings/AuditLogsTab.vue'
import TenantSwitcherModal from '../components/ui/TenantSwitcherModal.vue'
import SettingsSidebar from '../components/settings/SettingsSidebar.vue'

const emit = defineEmits(['voltar', 'logout'])

const { carregar, currentMembro, tenantPermissions } = useMembros()
const { activeTenantId } = useCasasMultitenant()

const activeTab = ref<'perfil' | 'acesso' | 'casa' | 'casas' | 'logs'>('perfil')
const isModoFoco = ref(false)

const isAdmin = computed(() => currentMembro.value?.role === 'ADMIN')

const isVerLogsPermitido = computed(() => {
  const role = currentMembro.value?.role
  if (!role || role === 'ADMIN') return true
  const perms = tenantPermissions.value[role]
  return perms ? perms.ALLOW_VER_AUDIT_LOGS !== false : true
})

type TabKey = 'perfil' | 'acesso' | 'casa' | 'casas' | 'logs'
interface TabDef { key: TabKey; label: string; show: boolean }

const visibleTabs = computed<TabDef[]>(() => {
  const tabs: TabDef[] = [
    { key: 'perfil', label: 'Meu Perfil', show: true },
    { key: 'acesso', label: 'Acessos', show: true },
    { key: 'casas', label: 'Casas', show: true },
    { key: 'casa', label: 'Permissões', show: isAdmin.value },
    { key: 'logs', label: 'Logs', show: isVerLogsPermitido.value },
  ]
  return tabs.filter(t => t.show)
})

const handleFocusChange = (active: boolean) => {
  isModoFoco.value = active
}

const handleLogout = () => {
  emit('logout')
}

onMounted(async () => {
  await carregar()
})
</script>

<template>
  <div class="h-full flex bg-canvas overflow-hidden">
    <!-- ===== Área de Conteúdo ===== -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <!-- ===== MOBILE HEADER (< lg) ===== -->
      <div
        v-if="!isModoFoco"
        class="lg:hidden shrink-0 px-4 pt-5 pb-3"
      >
        <div class="flex items-center gap-3">
          <button
            type="button"
            class="w-10 h-10 rounded-full bg-white border border-stone/60 text-charcoal flex items-center justify-center cursor-pointer shadow-sm hover:scale-105 hover:text-ember hover:border-ash/50 active:scale-95 transition-all duration-300 ease-out shrink-0"
            aria-label="Voltar ao Dashboard"
            @click="emit('voltar')"
          >
            <ArrowLeft class="w-5 h-5" />
          </button>
          <div class="min-w-0">
            <h2 class="text-heading-sm text-charcoal leading-tight truncate">
              Perfil <span class="text-ember">dos</span> Usuários
            </h2>
            <p class="text-[10px] sm:text-[11px] text-ash font-bold mt-0.5 uppercase tracking-[0.15em]">
              Configure quem mora aqui
            </p>
          </div>
        </div>
      </div>

      <!-- ===== MOBILE TABS (< lg): carrossel horizontal ===== -->
      <div
        v-if="!isModoFoco"
        class="lg:hidden shrink-0 mb-3"
      >
        <div class="flex px-3 overflow-x-auto no-scrollbar scroll-smooth gap-1.5">
          <button
            v-for="tab in visibleTabs"
            :key="tab.key"
            type="button"
            class="shrink-0 px-4 py-2.5 min-h-[44px] rounded-pill font-bold text-xs uppercase tracking-wider cursor-pointer border transition-all duration-300 ease-spring select-none active:scale-95"
            :class="activeTab === tab.key
              ? 'bg-ember text-white border-ember'
              : 'bg-transparent text-ash border-stone/20 hover:text-charcoal hover:border-stone/40 hover:bg-white/60'"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <!-- ===== CONTEÚDO PRINCIPAL (scrollável) ===== -->
      <div class="flex-1 overflow-y-auto custom-scrollbar">
        <div
          class="py-2 sm:py-4 lg:py-6"
          :class="isModoFoco
            ? 'px-4 sm:px-6 max-w-2xl mx-auto'
            : 'px-4 sm:px-6 lg:px-8 xl:px-10 w-full'"
        >
          <PerfilUsuarioTab
            v-if="activeTab === 'perfil'"
            :is-modo-foco="isModoFoco"
            @logout="handleLogout"
            @focus-change="handleFocusChange"
          />

          <GestaoAcessoTab
            v-else-if="activeTab === 'acesso'"
            :active-tenant-id="activeTenantId"
            @focus-change="handleFocusChange"
          />

          <ConfiguracoesCasaTab
            v-else-if="activeTab === 'casa'"
            :active-tenant-id="activeTenantId"
            @focus-change="handleFocusChange"
          />

          <TenantSwitcherModal
            v-else-if="activeTab === 'casas'"
            @casa-selecionada="emit('voltar')"
          />

          <AuditLogsTab
            v-else-if="activeTab === 'logs'"
            :active-tenant-id="activeTenantId"
            @focus-change="handleFocusChange"
          />
        </div>
      </div>

      <!-- ===== MOBILE FOOTER (< lg) ===== -->
      <div
        v-if="!isModoFoco"
        class="lg:hidden shrink-0 px-4 py-4 border-t border-stone/30 bg-white"
      >
        <Button
          variant="secondary"
          class="w-full h-12 text-xs font-bold uppercase tracking-widest"
          @click="emit('voltar')"
        >
          Voltar ao Dashboard
        </Button>
      </div>
    </div>

    <!-- ===== DESKTOP: Sidebar lateral direita (lg+) ===== -->
    <SettingsSidebar
      v-if="!isModoFoco"
      :active-tab="activeTab"
      :visible-tabs="visibleTabs"
      :current-membro="currentMembro"
      @select-tab="activeTab = $event"
      @voltar="emit('voltar')"
      @logout="handleLogout"
    />
  </div>
</template>

<style scoped>
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
