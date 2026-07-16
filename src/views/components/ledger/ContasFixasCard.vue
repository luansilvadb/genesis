<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import type { ContaFixa } from '../../../models/entities/ContaFixa'
import { Check } from 'lucide-vue-next'
import type { Gasto } from '../../../models/entities/Gasto'
import { formatarCentavosParaBRL } from '../../../shared/utils/formatarMoeda'

const props = defineProps<{
  bill: ContaFixa
  gasto?: Gasto
  obterNomeMembro: (id: string) => string
  isMonthClosed: boolean
  isReadOnly?: boolean
}>()

const emit = defineEmits<{
  (e: 'lancar', bill: ContaFixa): void
  (e: 'estornar', bill: ContaFixa): void
  (e: 'configurar', bill: ContaFixa): void
}>()

const cardRef = ref<HTMLElement | null>(null)

interface RippleState {
  active: boolean
  x: number
  y: number
  radius: number
  opacity: number
  scale: number
  type: 'tap' | 'long'
}

const ripple = ref<RippleState>({
  active: false,
  x: 0,
  y: 0,
  radius: 0,
  opacity: 0,
  scale: 0,
  type: 'tap'
})

let startTime = 0
let maxRadiusGlobal = 0
let isHolding = false
let hasTriggered = false
let animationFrameId: number | null = null

const DURATION_LONG = 800 // ms para segurar completo (long press)

const triggerAction = () => {
  emit('configurar', props.bill)
}

const triggerTapAction = () => {
  if (props.gasto) {
    emit('estornar', props.bill)
  } else {
    emit('lancar', props.bill)
  }
}

const cancelInteraction = () => {
  isHolding = false
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId)
    animationFrameId = null
  }
  
  if (ripple.value.active && ripple.value.type === 'long') {
    const fadeTick = () => {
      if (ripple.value.opacity > 0) {
        ripple.value.opacity -= 0.05
        requestAnimationFrame(fadeTick)
      } else {
        ripple.value.active = false
      }
    }
    requestAnimationFrame(fadeTick)
  }
}

const onPointerDown = (e: PointerEvent) => {
  if (props.isReadOnly || props.isMonthClosed) return
  const card = cardRef.value
  if (!card) return

  const rect = card.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top

  isHolding = true
  hasTriggered = false
  startTime = performance.now()

  const dx = Math.max(x, rect.width - x)
  const dy = Math.max(y, rect.height - y)
  maxRadiusGlobal = Math.sqrt(dx * dx + dy * dy)

  ripple.value.active = true
  ripple.value.type = 'long'
  ripple.value.x = x
  ripple.value.y = y
  ripple.value.radius = maxRadiusGlobal
  ripple.value.scale = 0
  ripple.value.opacity = 0.35

  if (animationFrameId) cancelAnimationFrame(animationFrameId)
  
  const tick = (now: number) => {
    if (!ripple.value.active) return

    if (ripple.value.type === 'long') {
      if (!isHolding) return
      
      const elapsed = now - startTime
      const progress = Math.min(elapsed / DURATION_LONG, 1)
      const easeProgress = 1 - Math.pow(1 - progress, 3) // ease-out cubic
      ripple.value.scale = easeProgress

      if (progress >= 1) {
        hasTriggered = true
        triggerAction()
        cancelInteraction()
        return
      }
    }

    animationFrameId = requestAnimationFrame(tick)
  }

  animationFrameId = requestAnimationFrame(tick)
}

const onPointerUp = () => {
  if (props.isReadOnly || props.isMonthClosed) return
  if (!isHolding || hasTriggered) {
    isHolding = false
    return
  }

  hasTriggered = true
  ripple.value.type = 'tap'

  triggerTapAction()

  requestAnimationFrame(() => {
    ripple.value.scale = 1
    ripple.value.opacity = 0
  })

  setTimeout(() => {
    if (ripple.value.type === 'tap') {
      ripple.value.active = false
      hasTriggered = false
    }
  }, 300)

  isHolding = false
  if (animationFrameId) {
    cancelAnimationFrame(animationFrameId)
    animationFrameId = null
  }
}

const onPointerLeave = () => {
  cancelInteraction()
}

onUnmounted(() => {
  if (animationFrameId) cancelAnimationFrame(animationFrameId)
})
</script>

<template>
  <div 
    ref="cardRef"
    class="relative overflow-hidden group flex items-center justify-between p-4 rounded-xl border transition-all duration-300 select-none"
    :class="[
      gasto ? 'bg-meadow/5 border-meadow/20' : 'bg-canvas border-stone',
      { 'cursor-pointer hover:border-ember/30': !props.isReadOnly && !props.isMonthClosed }
    ]"
    :data-testid="`conta-fixa-card-${bill.id}`"
    @pointerdown="onPointerDown"
    @pointerup="onPointerUp"
    @pointerleave="onPointerLeave"
    @pointercancel="onPointerLeave"
  >
    <div 
      v-if="ripple.active"
      class="absolute rounded-full pointer-events-none"
      :class="[
        gasto ? 'bg-coral/25' : 'bg-ember/20',
        ripple.type === 'tap' ? 'ripple-transition' : ''
      ]"
      :style="{
        left: ripple.x + 'px',
        top: ripple.y + 'px',
        width: ripple.radius * 2 + 'px',
        height: ripple.radius * 2 + 'px',
        transform: `translate(-50%, -50%) scale(${ripple.scale})`,
        opacity: ripple.opacity
      }"
    />

    <div class="flex items-center gap-4 min-w-0 flex-1 pointer-events-none">
      <div class="w-10 h-10 rounded-lg bg-white border border-stone flex items-center justify-center text-xl shadow-subtle group-hover:scale-110 transition-transform duration-500">
        {{ bill.icon }}
      </div>
      <div class="min-w-0 flex-1">
        <span class="font-bold text-sm block text-charcoal truncate tracking-tight">{{ bill.name }}</span>
        <div
          v-if="gasto"
          class="flex items-center mt-0.5"
        >
          <span class="text-[10px] text-meadow font-semibold uppercase tracking-wider">
            {{ formatarCentavosParaBRL(gasto.valorTotal.centavos) }} por {{ obterNomeMembro(gasto.compradorId) }}
          </span>
        </div>
        <div
          v-else
          class="flex items-center mt-0.5"
        >
          <span class="text-[10px] text-graphite font-semibold uppercase tracking-widest opacity-60">Aguardando lançamento</span>
        </div>
      </div>
    </div>
    
    <div
      v-if="gasto"
      class="shrink-0 pointer-events-none"
    >
      <div class="w-6 h-6 rounded-full bg-meadow flex items-center justify-center shadow-sm animate-in zoom-in-50 duration-500">
        <Check
          class="w-3.5 h-3.5 text-white"
          stroke-width="4"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.ripple-transition {
  transition: transform 300ms cubic-bezier(0.1, 0.8, 0.3, 1), opacity 250ms ease-out;
}
</style>
