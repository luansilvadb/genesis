<script setup lang="ts">
import { computed } from 'vue'
import { useToast } from '../../../composables/useToast'
import { X, CheckCircle2, AlertCircle, Info } from 'lucide-vue-next'

const { visible, message, type, hide } = useToast()

const config = computed(() => {
  switch (type.value) {
    case 'success':
      return {
        icon: CheckCircle2,
        colorClass: 'text-meadow',
        bgClass: 'bg-meadow/10'
      }
    case 'error':
      return {
        icon: AlertCircle,
        colorClass: 'text-coral',
        bgClass: 'bg-coral/10'
      }
    default:
      return {
        icon: Info,
        colorClass: 'text-sky',
        bgClass: 'bg-sky/10'
      }
  }
})
</script>

<template>
  <Transition name="toast-slide">
    <div 
      v-if="visible" 
      class="fixed top-5 left-1/2 z-[9999] w-[90%] max-w-[420px] bg-card border-none rounded-card-lg p-4 flex items-center gap-4 shadow-lg pointer-events-auto custom-toast"
      role="alert"
    >
      <!-- Borda Inset Simulatada para profundidade tátil -->
      <div class="absolute inset-0 rounded-card-lg shadow-subtle pointer-events-none" />

      <!-- Ícone Dinâmico -->
      <div
        class="flex items-center justify-center w-10 h-10 rounded-xl shrink-0 transition-colors duration-500"
        :class="config.bgClass"
      >
        <component
          :is="config.icon"
          class="w-5 h-5 animate-in zoom-in-50 duration-500"
          :class="config.colorClass"
          stroke-width="3"
        />
      </div>

      <div class="flex-1 text-sm font-bold text-charcoal leading-tight tracking-tight">
        {{ message }}
      </div>

      <button 
        class="w-8 h-8 rounded-full flex items-center justify-center text-ash hover:text-charcoal hover:bg-stone transition-all cursor-pointer border-none bg-transparent" 
        aria-label="Fechar notificação"
        @click="hide"
      >
        <X class="w-4 h-4" />
      </button>
    </div>
  </Transition>
</template>

<style scoped>
.custom-toast {
  translate: -50% 0;
}
.toast-slide-enter-active {
  transition: all 0.5s var(--ease-spring);
}
.toast-slide-leave-active {
  transition: all 0.3s ease;
}
.toast-slide-enter-from {
  translate: -50% -120px;
  opacity: 0;
  scale: 0.9;
}
.toast-slide-leave-to {
  translate: -50% -120px;
  opacity: 0;
}
</style>
