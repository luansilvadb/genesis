<script setup lang="ts">
interface Props {
  variant?: 'default'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default'
})

const variantClasses: Record<string, string> = {
  default: 'rounded-2xl bg-card shadow-subtle p-6 sm:p-8 border border-stone/20',
}
</script>

<template>
  <div :class="['transition-all duration-500 ease-spring hover:border-ember/20 group', variantClasses[props.variant]]">
    <slot />
  </div>
</template>
