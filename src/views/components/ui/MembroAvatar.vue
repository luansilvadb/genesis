<script setup lang="ts">
interface Props {
  nome: string
  src?: string
  variant?: 'ember' | 'meadow' | 'sky' | 'sunburst' | 'flamingo'
  size?: 'xs' | 'sm' | 'md' | 'lg'
}

withDefaults(defineProps<Props>(), {
  src: '',
  variant: 'ember',
  size: 'md'
})

const sizeClasses = {
  xs: 'w-5 h-5 sm:w-5.5 sm:h-5.5 text-[9px]',
  sm: 'w-8 h-8 text-[10px]',
  md: 'w-10 h-10 text-sm',
  lg: 'w-14 h-14 text-lg'
}

const variantColors = {
  ember: 'bg-ember/10 text-ember border-ember/20 shadow-ember/5',
  meadow: 'bg-meadow/10 text-meadow border-meadow/20 shadow-meadow/5',
  sky: 'bg-sky/10 text-sky border-sky/20 shadow-sky/5',
  sunburst: 'bg-sunburst/10 text-sunburst border-sunburst/20 shadow-sunburst/5',
  flamingo: 'bg-flamingo/10 text-flamingo border-flamingo/20 shadow-flamingo/5'
}
</script>

<template>
  <div 
    :class="[
      'relative flex items-center justify-center font-bold border transition-all duration-700 ease-spring shadow-subtle group-hover:scale-110 group-hover:rotate-3 blob-shape shrink-0',
      sizeClasses[size],
      variantColors[variant]
    ]"
    aria-hidden="true"
  >
    <!-- Organic Pulse Overlay -->
    <div class="absolute inset-0 bg-current opacity-[0.05] blob-shape animate-pulse" />
    
    <img
      v-if="src"
      :src="src"
      :alt="nome"
      class="absolute inset-0 w-full h-full object-cover blob-shape"
      referrerpolicy="no-referrer"
      loading="lazy"
    >
    <span
      v-else
      class="relative z-10 uppercase tracking-tighter font-display select-none"
    >{{ nome[0] }}</span>
  </div>
</template>

<style scoped>
@keyframes blob-morph {
  0%, 100% { border-radius: 40% 60% 70% 30% / 40% 50% 60% 50%; }
  33% { border-radius: 70% 30% 50% 50% / 30% 30% 70% 70%; }
  66% { border-radius: 50% 50% 30% 70% / 50% 70% 30% 30%; }
}

.blob-shape {
  animation: blob-morph 8s ease-in-out infinite;
  will-change: border-radius;
}
</style>
