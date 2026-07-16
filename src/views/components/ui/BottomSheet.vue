<template>
  <Teleport to="body">
    <!-- Backdrop overlay - fixed, doesn't move with bottom sheet -->
    <Transition name="fade">
      <div
        v-if="modelValue"
        ref="backdropEl"
        class="fixed inset-0 z-[998] bg-black/40 transition-opacity duration-300"
        @click="onBackdropClick"
      />
    </Transition>

    <!-- Wrapper centradora do BottomSheet para alinhamento horizontal no desktop -->
    <Transition
      name="slide-up"
      @after-leave="onTransitionLeave"
    >
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[999] flex justify-center items-end p-0 pointer-events-none"
      >
        <div
          ref="sheetEl"
          role="dialog"
          aria-modal="true"
          :aria-label="title || 'Diálogo'"
          class="pointer-events-auto relative flex flex-col bg-canvas border-t border-x border-stone/30 shadow-lg text-graphite
                 rounded-t-[32px] max-h-[90dvh] w-full max-w-full min-w-0 overflow-hidden"
          :class="widthClass"
          :style="{ maxHeight, minHeight }"
          style="touch-action: none"
          @touchstart="onTouchStart"
          @touchmove="onTouchMove"
          @touchend="onTouchEnd"
          @mousedown="onMouseDown"
        >
          <!-- Drag handle -->
          <div class="flex justify-center pt-3 pb-2 cursor-grab active:cursor-grabbing shrink-0">
            <div class="h-1.5 w-12 rounded-full bg-stone" />
          </div>

          <!-- Header -->
          <div
            v-if="title || $slots.header || $slots.title || subtitle || $slots.subtitle"
            class="flex items-start justify-between px-6 pt-2 pb-6 shrink-0"
          >
            <div class="flex-1 min-w-0 pr-4">
              <slot name="header">
                <slot name="title">
                  <h2
                    v-if="title"
                    class="text-3xl font-display text-charcoal leading-[1.1] tracking-tight"
                  >
                    {{ title }}
                  </h2>
                </slot>
                <slot name="subtitle">
                  <p
                    v-if="subtitle"
                    class="text-sm text-graphite font-medium mt-2 leading-relaxed opacity-80"
                  >
                    {{ subtitle }}
                  </p>
                </slot>
              </slot>
            </div>
            <button
              v-if="showClose"
              class="w-12 h-12 rounded-full bg-stone/50 flex items-center justify-center text-ash transition-all hover:bg-stone hover:text-charcoal cursor-pointer border-none focus-visible:ring-2 focus-visible:ring-ember focus-visible:outline-none shrink-0 -mt-1"
              aria-label="Fechar"
              @click="close"
            >
              <svg
                class="w-6 h-6"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2.5"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <!-- Divider -->
          <div
            v-if="(title || $slots.header || $slots.title) && showDivider"
            class="h-px bg-stone/60 mx-6 shrink-0"
          />

          <!-- Content -->
          <div
            ref="contentEl"
            :class="['overflow-y-auto flex-1 custom-scrollbar', contentClass]"
            @scroll.passive="onContentScroll"
          >
            <slot />
          </div>

          <!-- Footer -->
          <div
            v-if="$slots.footer"
            class="p-6 pt-4 border-t border-stone shrink-0 bg-white shadow-[0_-4px_12px_rgba(0,0,0,0.03)]"
          >
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onUnmounted, useTemplateRef, nextTick } from 'vue'
import { useBottomSheetState } from '../../../viewmodels/useBottomSheetState'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
  subtitle: { type: String, default: '' },
  showClose: { type: Boolean, default: true },
  showDivider: { type: Boolean, default: true },
  maxHeight: { type: String, default: '90dvh' },
  minHeight: { type: String, default: undefined },
  widthClass: { type: String, default: 'md:w-[480px]' },
  contentClass: { type: String, default: 'px-6 pb-8' }
})

const emit = defineEmits(['update:modelValue'])

const close = () => emit('update:modelValue', false)

let mountTime = 0
const onBackdropClick = () => {
  if (Date.now() - mountTime < 300) return
  close()
}

const { registerOpen, registerClose, isAnyBottomSheetOpen } = useBottomSheetState()

const getScrollbarWidth = () => {
  return window.innerWidth - document.documentElement.clientWidth
}

const lockScroll = () => {
  registerOpen()
  const scrollbarWidth = getScrollbarWidth()
  document.body.style.overflow = 'hidden'
  if (scrollbarWidth > 0) {
    document.documentElement.style.setProperty('--scrollbar-compensate', `${scrollbarWidth}px`)
  }
}

const unlockScroll = () => {
  if (!isAnyBottomSheetOpen.value) {
    document.body.style.overflow = ''
    document.documentElement.style.setProperty('--scrollbar-compensate', '0px')
  }
}

watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    mountTime = Date.now()
    lockScroll()
    nextTick(() => {
      resetDragStyles(false)
    })
  } else {
    registerClose()
  }
}, { immediate: true })

const onTransitionLeave = () => {
  unlockScroll()
}

// Refs do template
const sheetEl = useTemplateRef<HTMLElement>('sheetEl')
const backdropEl = useTemplateRef<HTMLElement>('backdropEl')
const contentEl = useTemplateRef<HTMLElement>('contentEl')

// Variáveis locais de controle de gesto (não reativas para máximo desempenho e suavidade)
let touchStartY = 0
let currentTranslateY = 0
let isDragging = false
let ignoreGesture = false
let lastScrollTop = 0
let rafId: number | null = null
let sheetHeight = 0

// Atualiza o cache de scroll de forma passiva
const onContentScroll = (e: Event) => {
  lastScrollTop = (e.target as HTMLElement).scrollTop
}

// Função auxiliar para aplicar estilos diretamente acelerados por GPU
const applyDragStyles = (delta: number) => {
  if (sheetEl.value) {
    sheetEl.value.style.transform = `translateY(${delta}px)`
  }
  if (backdropEl.value && sheetHeight > 0) {
    const opacity = Math.max(0, Math.min(1, 1 - (delta / sheetHeight)))
    backdropEl.value.style.opacity = opacity.toString()
  }
}

// Agenda a atualização visual usando requestAnimationFrame
const scheduleApplyStyles = (delta: number) => {
  if (rafId !== null) {
    cancelAnimationFrame(rafId)
  }
  rafId = requestAnimationFrame(() => {
    applyDragStyles(delta)
    rafId = null
  })
}

// Limpa as propriedades de transição inline para resposta tátil imediata
const clearTransitions = () => {
  if (sheetEl.value) {
    sheetEl.value.style.transition = 'none'
  }
  if (backdropEl.value) {
    backdropEl.value.style.transition = 'none'
  }
}

// Restaura os estilos padrão do BottomSheet
const resetDragStyles = (withTransition: boolean) => {
  if (sheetEl.value) {
    sheetEl.value.style.transition = withTransition ? 'transform 0.4s var(--ease-spring)' : ''
    sheetEl.value.style.transform = ''
  }
  if (backdropEl.value) {
    backdropEl.value.style.transition = withTransition ? 'opacity 0.4s ease-out' : ''
    backdropEl.value.style.opacity = ''
  }
  
  if (withTransition) {
    setTimeout(() => {
      if (props.modelValue) {
        if (sheetEl.value) {
          sheetEl.value.style.transition = ''
          sheetEl.value.style.transform = ''
        }
        if (backdropEl.value) {
          backdropEl.value.style.transition = ''
          backdropEl.value.style.opacity = ''
        }
      }
    }, 400)
  }
}

// Animação de fechamento controlada imperativamente
const animateCloseAndEmit = () => {
  if (sheetEl.value) {
    sheetEl.value.style.transition = 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
    sheetEl.value.style.transform = 'translateY(100%)'
  }
  if (backdropEl.value) {
    backdropEl.value.style.transition = 'opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
    backdropEl.value.style.opacity = '0'
  }
  setTimeout(() => {
    close()
  }, 300)
}

onUnmounted(() => {
  if (rafId !== null) {
    cancelAnimationFrame(rafId)
  }
  if (props.modelValue) {
    registerClose()
    document.body.style.overflow = ''
    document.documentElement.style.setProperty('--scrollbar-compensate', '0px')
  }
})

// Proteção para não iniciar arraste em botões, campos de entrada ou marcadores no-drag
const shouldStartDrag = (target: HTMLElement): boolean => {
  if (
    target.closest('button') ||
    target.closest('input') ||
    target.closest('select') ||
    target.closest('textarea') ||
    target.closest('a') ||
    target.closest('[role="button"]') ||
    target.closest('.no-drag')
  ) {
    return false
  }
  return true
}

// Handlers de Toque Mobile
const onTouchStart = (e: TouchEvent) => {
  const target = e.target as HTMLElement
  if (!shouldStartDrag(target)) {
    ignoreGesture = true
    isDragging = false
    return
  }

  ignoreGesture = false
  touchStartY = e.touches[0].clientY
  sheetHeight = sheetEl.value?.clientHeight || 0
  
  lastScrollTop = contentEl.value?.scrollTop || 0
  isDragging = lastScrollTop <= 0
  
  clearTransitions()
}

const onTouchMove = (e: TouchEvent) => {
  if (ignoreGesture) return

  const clientY = e.touches[0].clientY
  const delta = clientY - touchStartY

  if (isDragging) {
    if (delta > 0) {
      if (e.cancelable) e.preventDefault()
      currentTranslateY = delta
      scheduleApplyStyles(delta)
    } else if (currentTranslateY !== 0) {
      if (e.cancelable) e.preventDefault()
      currentTranslateY = 0
      scheduleApplyStyles(0)
    }
  } else {
    // Transição contínua "scroll-to-drag"
    if (delta > 0 && lastScrollTop <= 0) {
      isDragging = true
      touchStartY = clientY // redefine para o ponto de toque atual para suavizar a transição
      currentTranslateY = 0
      if (e.cancelable) e.preventDefault()
      clearTransitions()
      scheduleApplyStyles(0)
    }
  }
}

const onTouchEnd = (e: TouchEvent) => {
  if (ignoreGesture) return

  if (rafId !== null) {
    cancelAnimationFrame(rafId)
    rafId = null
  }

  if (!isDragging) return
  isDragging = false

  const clientY = e.changedTouches[0].clientY
  const finalDelta = clientY - touchStartY
  const threshold = Math.max(100, sheetHeight * 0.25)

  if (finalDelta > threshold) {
    animateCloseAndEmit()
  } else {
    resetDragStyles(true)
  }
}

// Handlers de Mouse Desktop
const onMouseDown = (e: MouseEvent) => {
  const target = e.target as HTMLElement
  if (!shouldStartDrag(target)) {
    ignoreGesture = true
    return
  }

  ignoreGesture = false
  touchStartY = e.clientY
  sheetHeight = sheetEl.value?.clientHeight || 0
  
  lastScrollTop = contentEl.value?.scrollTop || 0
  isDragging = lastScrollTop <= 0
  
  clearTransitions()

  const onMouseMove = (ev: MouseEvent) => {
    if (ignoreGesture) return

    const clientY = ev.clientY
    const delta = clientY - touchStartY

    if (isDragging) {
      if (delta > 0) {
        currentTranslateY = delta
        scheduleApplyStyles(delta)
      } else if (currentTranslateY !== 0) {
        currentTranslateY = 0
        scheduleApplyStyles(0)
      }
    } else {
      if (delta > 0 && lastScrollTop <= 0) {
        isDragging = true
        touchStartY = clientY
        currentTranslateY = 0
        clearTransitions()
        scheduleApplyStyles(0)
      }
    }
  }

  const onMouseUp = (ev: MouseEvent) => {
    if (rafId !== null) {
      cancelAnimationFrame(rafId)
      rafId = null
    }

    if (!ignoreGesture && isDragging) {
      isDragging = false
      const finalDelta = ev.clientY - touchStartY
      const threshold = Math.max(100, sheetHeight * 0.25)
      
      if (finalDelta > threshold) {
        animateCloseAndEmit()
      } else {
        resetDragStyles(true)
      }
    }

    window.removeEventListener('mousemove', onMouseMove)
    window.removeEventListener('mouseup', onMouseUp)
  }

  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', onMouseUp)
}
</script>

<style scoped>
/* Transição limpa padrão de slide-up e slide-down */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.5s var(--ease-spring);
}
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}

/* Fade transition for backdrop */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease-out;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

</style>
