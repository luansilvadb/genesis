<script setup lang="ts">
import { computed } from 'vue'
import { Loader2 } from 'lucide-vue-next'

interface Props {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost' | 'inverted'
  size?: 'default' | 'sm' | 'lg' | 'icon'
  class?: string
  loading?: boolean
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'default',
  class: '',
  loading: false,
  disabled: false
})

const classes = computed(() => {
  return [
    'relative inline-flex items-center justify-center whitespace-nowrap transition-all duration-300 ease-spring focus-visible:outline-none disabled:pointer-events-none disabled:opacity-40 active:scale-90 cursor-pointer border-none select-none',
    props.loading && 'transition-none',
    props.size === 'default' && 'h-12 px-6 text-sm font-bold uppercase tracking-widest',
    props.size === 'sm' && 'h-9 px-4 text-[10px] font-bold uppercase tracking-wider',
    props.size === 'lg' && 'h-16 px-10 text-base font-bold uppercase tracking-[0.15em]',
    props.size === 'icon' && 'h-11 w-11 text-sm font-bold',
    props.variant === 'primary' && 'bg-midnight text-white hover:bg-charcoal rounded-pill shadow-md hover:shadow-lg',
    props.variant === 'secondary' && 'bg-stone text-charcoal hover:bg-ash/20 rounded-pill',
    props.variant === 'outline' && 'bg-transparent text-charcoal border border-stone hover:bg-stone/50 rounded-xl font-bold px-8 py-3',
    props.variant === 'ghost' && 'text-ember hover:opacity-80 font-bold p-2 bg-transparent',
    props.variant === 'inverted' && 'bg-white text-midnight hover:bg-parchment rounded-pill font-bold shadow-subtle',
    props.class
  ].filter(Boolean).join(' ')
})
</script>

<template>
  <button
    type="button"
    :class="classes"
    :disabled="disabled || loading"
    :aria-busy="loading"
  >
    <Loader2
      v-if="loading"
      class="absolute w-5 h-5 animate-spin text-current"
    />
    <span
      class="inline-flex items-center gap-1.5"
      :class="{ 'invisible': loading }"
    >
      <slot />
    </span>
  </button>
</template>
