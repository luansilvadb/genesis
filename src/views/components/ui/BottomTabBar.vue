<!-- src/views/components/ui/BottomTabBar.vue -->
<script setup lang="ts">
import { Home, Wallet, Plus } from 'lucide-vue-next'

export type Tab = 'hoje' | 'pessoal'

interface Props {
  modelValue: Tab
  isMonthClosed?: boolean
  isReadOnly?: boolean
}

defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:modelValue', tab: Tab): void
  (e: 'click-fab'): void
}>()

</script>

<template>
  <div class="fixed left-4 right-4 bottom-[calc(1rem+env(safe-area-inset-bottom,0px))] z-40 bg-white border border-stone/20 rounded-pill shadow-premium max-w-[500px] mx-auto pointer-events-auto">
    <nav class="w-full h-18 sm:h-20 flex items-center px-2">
      <!-- Lado Esquerdo: Tab Casa -->
      <div class="flex-1 flex justify-center items-center h-full">
        <button
          class="flex-1 min-w-[48px] min-h-[48px] flex flex-col items-center justify-center relative group outline-none cursor-pointer border-none bg-transparent rounded-2xl transition-all duration-300 ease-jelly active:scale-92"
          :class="[
            modelValue === 'hoje' ? 'text-ember' : 'text-graphite/85 hover:text-charcoal'
          ]"
          aria-label="Casa"
          :aria-selected="modelValue === 'hoje'"
          @click="emit('update:modelValue', 'hoje')"
        >
          <!-- Jelly Active Indicator -->
          <div
            class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-12 h-12 rounded-full bg-ember/[0.06] transition-all duration-500 ease-jelly"
            :class="modelValue === 'hoje' ? 'scale-100 opacity-100' : 'scale-0 opacity-0'"
          />
          <div class="relative flex flex-col items-center justify-center gap-1 transition-transform duration-300">
            <Home
              class="w-5.5 h-5.5 transition-all duration-500 ease-jelly"
              :class="modelValue === 'hoje' ? 'stroke-[2.5px] drop-shadow-[0_2px_8px_rgba(255,62,0,0.2)] scale-110' : 'stroke-[1.8px] scale-100'"
            />
            <span class="text-[11px] font-bold leading-none text-center">Casa</span>
          </div>
        </button>
      </div>

      <!-- Centro: FAB (Botão de Adicionar) -->
      <div class="flex justify-center items-center px-2 shrink-0">
        <button
          :disabled="isMonthClosed || isReadOnly"
          class="w-13 h-13 sm:w-14 sm:h-14 rounded-full bg-ember text-white flex items-center justify-center border-none transition-all duration-500 ease-jelly shadow-[0_12px_32px_-8px_rgba(255,62,0,0.5)] hover:bg-ember/90 hover:scale-105 active:scale-90 disabled:opacity-40 disabled:grayscale disabled:cursor-not-allowed cursor-pointer group"
          aria-label="Adicionar novo gasto"
          data-testid="novo-lancamento-fab"
          @click="emit('click-fab')"
        >
          <Plus class="w-6 h-6 stroke-[3px] group-hover:rotate-90 transition-transform duration-500 ease-jelly" />
        </button>
      </div>

      <!-- Lado Direito: Tab Pessoal -->
      <div class="flex-1 flex justify-center items-center h-full">
        <button
          class="flex-1 min-w-[48px] min-h-[48px] flex flex-col items-center justify-center relative group outline-none cursor-pointer border-none bg-transparent rounded-2xl transition-all duration-300 ease-jelly active:scale-92"
          :class="[
            modelValue === 'pessoal' ? 'text-ember' : 'text-graphite/85 hover:text-charcoal'
          ]"
          aria-label="Pessoal"
          :aria-selected="modelValue === 'pessoal'"
          @click="emit('update:modelValue', 'pessoal')"
        >
          <!-- Jelly Active Indicator -->
          <div
            class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-12 h-12 rounded-full bg-ember/[0.06] transition-all duration-500 ease-jelly"
            :class="modelValue === 'pessoal' ? 'scale-100 opacity-100' : 'scale-0 opacity-0'"
          />
          <div class="relative flex flex-col items-center justify-center gap-1 transition-transform duration-300">
            <Wallet
              class="w-5.5 h-5.5 transition-all duration-500 ease-jelly"
              :class="modelValue === 'pessoal' ? 'stroke-[2.5px] drop-shadow-[0_2px_8px_rgba(255,62,0,0.2)] scale-110' : 'stroke-[1.8px] scale-100'"
            />
            <span class="text-[11px] font-bold leading-none text-center">Pessoal</span>
          </div>
        </button>
      </div>
    </nav>
  </div>
</template>
