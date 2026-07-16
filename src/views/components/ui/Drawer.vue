<template>
  <Teleport to="body">
    <!-- Backdrop overlay -->
    <Transition name="drawer-fade">
      <div
        v-if="modelValue"
        ref="backdropEl"
        class="fixed inset-0 z-[998] bg-black/40 transition-opacity duration-300"
        @click="onBackdropClick"
      />
    </Transition>

    <!-- Drawer panel - slides from right -->
    <Transition
      name="drawer-slide"
      @after-leave="onTransitionLeave"
    >
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[999] flex justify-end pointer-events-none"
      >
        <div
          ref="drawerEl"
          role="dialog"
          aria-modal="true"
          :aria-label="title || 'Painel lateral'"
          class="pointer-events-auto relative flex flex-col bg-canvas border-l border-stone/30 shadow-2xl text-graphite
                 h-full w-full min-w-0 overflow-hidden"
          :class="widthClass"
          @touchstart.passive="onTouchStart"
          @touchmove="onTouchMove"
          @touchend="onTouchEnd"
        >
          <!-- Header -->
          <div
            v-if="title || $slots.header || $slots.title"
            class="flex items-start justify-between px-6 pt-6 pb-4 shrink-0"
          >
            <div class="flex-1 min-w-0 pr-4">
              <slot name="header">
                <slot name="title">
                  <h2
                    v-if="title"
                    class="text-2xl font-display text-charcoal leading-[1.1] tracking-tight"
                  >
                    {{ title }}
                  </h2>
                </slot>
              </slot>
            </div>
            <button
              v-if="showClose"
              class="w-10 h-10 rounded-full bg-stone/50 flex items-center justify-center text-ash transition-all hover:bg-stone hover:text-charcoal cursor-pointer border-none focus-visible:ring-2 focus-visible:ring-ember focus-visible:outline-none shrink-0"
              aria-label="Fechar"
              @click="close"
            >
              <svg
                class="w-5 h-5"
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
          >
            <slot />
          </div>

          <!-- Footer -->
          <div
            v-if="$slots.footer"
            class="p-6 pt-4 border-t border-stone shrink-0 bg-white"
          >
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onUnmounted, useTemplateRef } from 'vue'
import { useBottomSheetState } from '../../../viewmodels/useBottomSheetState'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '' },
  showClose: { type: Boolean, default: true },
  showDivider: { type: Boolean, default: true },
  widthClass: { type: String, default: 'md:max-w-[480px]' },
  contentClass: { type: String, default: 'px-6 pb-8' }
})

const emit = defineEmits(['update:modelValue'])

const close = () => emit('update:modelValue', false)

let mountTime = 0
const onBackdropClick = () => {
  // Evita fechar acidentalmente logo após montar (mobile tap fantasma)
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
  } else {
    registerClose()
  }
}, { immediate: true })

const onTransitionLeave = () => {
  unlockScroll()
}

// Refs
const drawerEl = useTemplateRef<HTMLElement>('drawerEl')
const backdropEl = useTemplateRef<HTMLElement>('backdropEl')

// Swipe-to-dismiss (touch horizontal)
let touchStartX = 0
let currentTranslateX = 0
let isDragging = false
let ignoreGesture = false
let rafId: number | null = null
let drawerWidth = 0

const onTouchStart = (e: TouchEvent) => {
  const target = e.target as HTMLElement
  if (
    target.closest('button') ||
    target.closest('input') ||
    target.closest('select') ||
    target.closest('textarea') ||
    target.closest('a') ||
    target.closest('[role="button"]')
  ) {
    ignoreGesture = true
    isDragging = false
    return
  }

  ignoreGesture = false
  touchStartX = e.touches[0].clientX
  drawerWidth = drawerEl.value?.clientWidth || 0
  isDragging = true

  if (drawerEl.value) drawerEl.value.style.transition = 'none'
  if (backdropEl.value) backdropEl.value.style.transition = 'none'
}

const applyDragStyles = (delta: number) => {
  if (drawerEl.value) {
    drawerEl.value.style.transform = `translateX(${delta}px)`
  }
  if (backdropEl.value && drawerWidth > 0) {
    const opacity = Math.max(0, Math.min(1, 1 - (delta / drawerWidth)))
    backdropEl.value.style.opacity = opacity.toString()
  }
}

const scheduleApplyStyles = (delta: number) => {
  if (rafId !== null) cancelAnimationFrame(rafId)
  rafId = requestAnimationFrame(() => {
    applyDragStyles(delta)
    rafId = null
  })
}

const resetDragStyles = (withTransition: boolean) => {
  if (drawerEl.value) {
    drawerEl.value.style.transition = withTransition ? 'transform 0.4s var(--ease-spring)' : ''
    drawerEl.value.style.transform = ''
  }
  if (backdropEl.value) {
    backdropEl.value.style.transition = withTransition ? 'opacity 0.4s ease-out' : ''
    backdropEl.value.style.opacity = ''
  }

  if (withTransition) {
    setTimeout(() => {
      if (props.modelValue) {
        if (drawerEl.value) {
          drawerEl.value.style.transition = ''
          drawerEl.value.style.transform = ''
        }
        if (backdropEl.value) {
          backdropEl.value.style.transition = ''
          backdropEl.value.style.opacity = ''
        }
      }
    }, 400)
  }
}

const animateCloseAndEmit = () => {
  if (drawerEl.value) {
    drawerEl.value.style.transition = 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
    drawerEl.value.style.transform = 'translateX(100%)'
  }
  if (backdropEl.value) {
    backdropEl.value.style.transition = 'opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
    backdropEl.value.style.opacity = '0'
  }
  setTimeout(() => {
    close()
  }, 300)
}

const onTouchMove = (e: TouchEvent) => {
  if (ignoreGesture || !isDragging) return

  const clientX = e.touches[0].clientX
  const delta = clientX - touchStartX

  // Só permite arrastar para a direita (fechar)
  if (delta > 0) {
    if (e.cancelable) e.preventDefault()
    currentTranslateX = delta
    scheduleApplyStyles(delta)
  } else if (currentTranslateX !== 0) {
    currentTranslateX = 0
    scheduleApplyStyles(0)
  }
}

const onTouchEnd = (e: TouchEvent) => {
  if (ignoreGesture || !isDragging) return
  isDragging = false

  if (rafId !== null) {
    cancelAnimationFrame(rafId)
    rafId = null
  }

  const clientX = e.changedTouches[0].clientX
  const finalDelta = clientX - touchStartX
  const threshold = Math.max(80, drawerWidth * 0.25)

  if (finalDelta > threshold) {
    animateCloseAndEmit()
  } else {
    resetDragStyles(true)
  }
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
</script>

<style scoped>
/* Drawer slide transition: slides in from the right */
.drawer-slide-enter-active,
.drawer-slide-leave-active {
  transition: transform 0.4s var(--ease-spring);
}
.drawer-slide-enter-from,
.drawer-slide-leave-to {
  transform: translateX(100%);
}

/* Fade transition for backdrop */
.drawer-fade-enter-active,
.drawer-fade-leave-active {
  transition: opacity 0.3s ease-out;
}
.drawer-fade-enter-from,
.drawer-fade-leave-to {
  opacity: 0;
}
</style>
