<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useMembros } from '../../viewmodels/useMembros'
import { useCasasMultitenant } from '../../viewmodels/useCasasMultitenant'
import Button from '../components/ui/Button.vue'
import PerfilUsuarioTab from '../components/settings/PerfilUsuarioTab.vue'
import GestaoAcessoTab from '../components/settings/GestaoAcessoTab.vue'
import ConfiguracoesCasaTab from '../components/settings/ConfiguracoesCasaTab.vue'
import AuditLogsTab from '../components/settings/AuditLogsTab.vue'
import TenantSwitcherModal from '../components/ui/TenantSwitcherModal.vue'

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
  <div class="h-full flex flex-col bg-canvas overflow-hidden">
    <!-- Header -->
    <div v-if="!isModoFoco" class="shrink-0 px-6 pt-6 pb-2 sm:px-8 sm:pt-8 sm:pb-3">
      <h2 class="text-display text-2xl xs:text-3xl sm:text-4xl text-charcoal leading-[1.15]">Perfil <span class="text-ember">dos</span> Usuários</h2>
      <p class="text-[11px] sm:text-xs text-ash font-bold mt-1.5 uppercase tracking-[0.15em]">Configure quem mora aqui e como cada um contribui</p>
    </div>

    <!-- Navegação de Abas — responsiva: carrossel no mobile, grid wrap no sm+ -->
    <div v-if="!isModoFoco" class="shrink-0 relative mb-5">
      <!-- Fade indicators para carrossel mobile -->
      <div class="absolute left-0 top-0 bottom-0 w-6 bg-gradient-to-r from-canvas to-transparent z-20 pointer-events-none sm:hidden rounded-l-2xl" />
      <div class="absolute right-0 top-0 bottom-0 w-6 bg-gradient-to-r from-transparent to-canvas z-20 pointer-events-none sm:hidden rounded-r-2xl" />

      <!-- Mobile: carrossel horizontal com scroll -->
      <div class="flex sm:hidden px-4 overflow-x-auto no-scrollbar scroll-smooth gap-1.5">
        <button
          v-for="tab in visibleTabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="shrink-0 px-4 py-2.5 min-h-[44px] rounded-pill font-bold text-xs uppercase tracking-wider cursor-pointer border border-stone/20 transition-all duration-400 ease-jelly select-none active:scale-95"
          :class="activeTab === tab.key ? 'bg-ember text-white border-ember shadow-[0_2px_10px_rgba(255,62,0,0.2)]' : 'bg-white text-ash hover:text-charcoal hover:border-stone/40'"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- sm+: grid flex-wrap de pills -->
      <div class="hidden sm:flex flex-wrap justify-center gap-2 px-4 sm:px-6">
        <button
          v-for="tab in visibleTabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          class="px-5 py-2.5 min-h-[44px] rounded-pill font-bold text-xs sm:text-sm uppercase tracking-wider cursor-pointer border-none transition-all duration-500 ease-jelly select-none active:scale-95"
          :class="activeTab === tab.key ? 'bg-white text-ember shadow-[0_2px_12px_rgba(255,62,0,0.12)] ring-1 ring-stone/30' : 'bg-white/60 text-ash hover:text-charcoal hover:bg-white'"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- Conteúdo das Abas -->
    <div class="flex-1 overflow-y-auto px-4 sm:px-8 pb-6 sm:pb-8 custom-scrollbar">
      <div class="max-w-2xl mx-auto py-3 sm:py-4">
        
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

    <!-- Footer -->
    <div v-if="!isModoFoco" class="shrink-0 px-4 py-4 sm:px-8 sm:py-5 border-t border-stone/30 bg-white">
      <Button variant="secondary" class="w-full" @click="emit('voltar')">Fechar</Button>
    </div>
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
