<script setup lang="ts">
import SkeletonBlock from './SkeletonBlock.vue'

interface Props {
  variant?: 'hoje'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'hoje'
})
</script>

<template>
  <section class="skeleton-theme" aria-busy="true" aria-live="polite" data-testid="skeleton-mimic">
    <div class="pointer-events-none select-none space-y-12">
      <!-- Header Simplificado -->
      <header class="flex items-center justify-between pb-6 pt-3 mb-8 border-b border-stone/50">
        <div class="space-y-2">
          <SkeletonBlock shape="text" tone="soft" width="3rem" height="0.625rem" />
          <SkeletonBlock shape="text" tone="strong" width="7rem" height="1.5rem" />
        </div>
        <SkeletonBlock width="2.75rem" height="2.75rem" radius="0.75rem" />
      </header>

      <!-- Blocos de Conteúdo Essenciais -->
      <div class="space-y-8">
        <div v-for="i in 3" :key="i" class="p-4 rounded-2xl border border-stone bg-canvas flex items-center gap-4">
          <SkeletonBlock shape="circle" width="3rem" height="3rem" />
          <div class="flex-1 space-y-2">
            <SkeletonBlock shape="text" tone="strong" width="60%" height="0.875rem" />
            <SkeletonBlock shape="text" tone="soft" width="40%" height="0.625rem" />
          </div>
          <SkeletonBlock shape="text" tone="strong" width="4rem" height="1.25rem" />
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.skeleton-theme {
  --skeleton-base: rgba(242, 240, 237, 0.88);
  --skeleton-soft: rgba(242, 240, 237, 0.56);
  --skeleton-strong: rgba(226, 223, 219, 0.96);
}
</style>
