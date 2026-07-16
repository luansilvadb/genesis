<script setup lang="ts">
import {
  User,
  Users,
  Home,
  Shield,
  FileText,
  ArrowLeft,
  LogOut
} from 'lucide-vue-next'
import type { LucideIcon } from 'lucide-vue-next'
import MembroAvatar from '../ui/MembroAvatar.vue'
import type { Membro } from '../../../models/entities/Membro'

type TabKey = 'perfil' | 'acesso' | 'casa' | 'casas' | 'logs'

interface SidebarTab {
  key: TabKey
  label: string
  show: boolean
}

defineProps<{
  activeTab: TabKey
  visibleTabs: SidebarTab[]
  currentMembro: Membro | null | undefined
}>()

const emit = defineEmits<{
  (e: 'select-tab', key: TabKey): void
  (e: 'voltar'): void
  (e: 'logout'): void
}>()

const iconMap: Record<TabKey, LucideIcon> = {
  perfil: User,
  acesso: Users,
  casa: Shield,
  casas: Home,
  logs: FileText,
}
</script>

<template>
  <aside class="hidden lg:flex flex-col w-[272px] shrink-0 h-full bg-white/80 backdrop-blur-sm border-l border-stone/30 overflow-y-auto">
    <!-- Cabeçalho do Perfil -->
    <div class="px-5 pt-8 pb-5 border-b border-stone/20">
      <div class="flex items-center gap-3.5">
        <MembroAvatar
          v-if="currentMembro"
          :nome="currentMembro.nome"
          variant="ember"
          size="md"
        />
        <div class="min-w-0 flex-1">
          <h3 class="text-sm font-bold text-charcoal tracking-tight truncate leading-tight">
            {{ currentMembro?.nome || 'Usuário' }}
          </h3>
          <p class="text-[10px] font-bold text-ash uppercase tracking-wider mt-0.5">
            {{ currentMembro?.role === 'ADMIN' ? 'Administrador' : currentMembro?.role === 'MORADOR' ? 'Morador' : 'Visualizador' }}
          </p>
        </div>
      </div>
    </div>

    <!-- Navegação -->
    <nav class="flex-1 px-3 py-4 space-y-0.5">
      <button
        v-for="tab in visibleTabs"
        :key="tab.key"
        type="button"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-left border-none cursor-pointer transition-all duration-200 select-none group"
        :class="activeTab === tab.key
          ? 'bg-ember/8 text-ember font-bold'
          : 'text-graphite hover:bg-stone/60 hover:text-charcoal font-medium'"
        @click="emit('select-tab', tab.key)"
      >
        <component
          :is="iconMap[tab.key]"
          :class="activeTab === tab.key ? 'text-ember' : 'text-ash group-hover:text-graphite'"
          class="w-[18px] h-[18px] shrink-0 transition-colors duration-200"
          :stroke-width="activeTab === tab.key ? 2.5 : 2"
        />
        <span class="text-xs leading-none">{{ tab.label }}</span>
        <div
          v-if="activeTab === tab.key"
          class="ml-auto w-1.5 h-1.5 rounded-full bg-ember"
        />
      </button>
    </nav>

    <!-- Footer Actions -->
    <div class="px-3 py-4 border-t border-stone/20 space-y-1.5">
      <button
        type="button"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-left border-none cursor-pointer transition-all duration-200 text-graphite hover:bg-coral/6 hover:text-coral font-medium group"
        @click="emit('logout')"
      >
        <LogOut
          class="w-[18px] h-[18px] text-ash group-hover:text-coral transition-colors duration-200"
          :stroke-width="2"
        />
        <span class="text-xs leading-none">Sair da conta</span>
      </button>
      <button
        type="button"
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl text-left border-none cursor-pointer transition-all duration-200 text-graphite hover:bg-stone/60 hover:text-charcoal font-medium group"
        @click="emit('voltar')"
      >
        <ArrowLeft
          class="w-[18px] h-[18px] text-ash group-hover:text-charcoal transition-colors duration-200"
          :stroke-width="2"
        />
        <span class="text-xs leading-none">Voltar ao Dashboard</span>
      </button>
    </div>
  </aside>
</template>
