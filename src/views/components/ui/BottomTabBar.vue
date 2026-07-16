<!-- src/views/components/ui/BottomTabBar.vue -->
<script setup lang="ts">
import { ref } from 'vue'
import { User, Plus, Calendar } from 'lucide-vue-next'
import MembroAvatar from './MembroAvatar.vue'

export type Tab = 'hoje' | 'pessoal' | 'perfil'

interface Props {
  modelValue: Tab
  isMonthClosed?: boolean
  isReadOnly?: boolean
  currentYear?: string | number
  currentMonthName?: string
  faturaSelecionadaFechada?: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:modelValue', tab: Tab): void
  (e: 'click-fab'): void
  (e: 'open-periodo'): void
}>()

const tabs = [
  { id: 'perfil', label: 'Ajustes', icon: User },
] as const

const userName = ref(localStorage.getItem('divi_username') || 'Ajustes')
</script>

<template>
  <div class="fixed left-4 right-4 bottom-[calc(1rem+env(safe-area-inset-bottom,0px))] z-40 bg-white border border-stone/20 rounded-pill shadow-premium max-w-[500px] mx-auto pointer-events-auto">
    <nav class="w-full h-18 sm:h-20 flex items-center px-2">
      <!-- Lado Esquerdo: Seletor de Período -->
      <div class="flex-1 flex justify-around items-center h-full">
        <!-- Botão de Período (Data) -->
        <button
          class="flex-1 min-w-[48px] min-h-[48px] flex flex-col items-center justify-center relative group outline-none cursor-pointer border-none bg-transparent rounded-2xl transition-all duration-300 ease-jelly active:scale-92 text-graphite/85 hover:text-charcoal"
          aria-label="Selecionar período"
          @click="emit('open-periodo')"
        >
          <div class="relative flex flex-col items-center justify-center gap-1 transition-transform duration-300">
            <Calendar class="w-5.5 h-5.5 stroke-[1.8px] transition-all duration-500 ease-jelly group-hover:stroke-[2.5px] group-hover:scale-110" />
            <span class="text-[11px] font-bold leading-none text-center">Período</span>
          </div>
        </button>
      </div>

      <!-- Centro: FAB (Botão de Adicionar) -->
      <div class="flex justify-center items-center px-2 shrink-0">
        <button
          @click="emit('click-fab')"
          :disabled="isMonthClosed || isReadOnly"
          class="w-13 h-13 sm:w-14 sm:h-14 rounded-full bg-ember text-white flex items-center justify-center border-none transition-all duration-500 ease-jelly shadow-[0_12px_32px_-8px_rgba(255,62,0,0.5)] hover:bg-ember/90 hover:scale-105 active:scale-90 disabled:opacity-40 disabled:grayscale disabled:cursor-not-allowed cursor-pointer group"
          aria-label="Adicionar novo gasto"
          data-testid="novo-lancamento-fab"
        >
          <Plus class="w-6 h-6 stroke-[3px] group-hover:rotate-90 transition-transform duration-500 ease-jelly" />
        </button>
      </div>

      <!-- Lado Direito: Acertos e Ajustes -->
      <div class="flex-1 flex justify-around items-center h-full">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="emit('update:modelValue', tab.id)"
          class="flex-1 min-w-[48px] min-h-[48px] flex flex-col items-center justify-center relative group outline-none cursor-pointer border-none bg-transparent rounded-2xl transition-all duration-300 ease-jelly active:scale-92"
          :class="[
            modelValue === tab.id ? 'text-ember' : 'text-graphite/85 hover:text-charcoal'
          ]"
          :aria-label="tab.label"
          :aria-selected="modelValue === tab.id"
        >
          <!-- Jelly Active Indicator -->
          <div
            class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-12 h-12 rounded-full bg-ember/[0.06] transition-all duration-500 ease-jelly"
            :class="modelValue === tab.id ? 'scale-100 opacity-100' : 'scale-0 opacity-0'"
          />

          <div class="relative flex flex-col items-center justify-center gap-1 transition-transform duration-300">
            <MembroAvatar
              v-if="tab.id === 'perfil'"
              :nome="userName"
              size="xs"
              variant="ember"
              class="!border-none !shadow-none transition-all duration-500 ease-jelly"
              :class="modelValue === 'perfil' ? 'scale-115' : 'grayscale opacity-60 contrast-75'"
            />
            <component
              v-else
              :is="tab.icon"
              class="w-5.5 h-5.5 transition-all duration-500 ease-jelly"
              :class="modelValue === tab.id ? 'stroke-[2.5px] drop-shadow-[0_2px_8px_rgba(255,62,0,0.2)] scale-110' : 'stroke-[1.8px] scale-100'"
            />
            <span class="text-[11px] font-bold leading-none text-center">
              {{ tab.label }}
            </span>
          </div>
        </button>
      </div>
    </nav>
  </div>
</template>