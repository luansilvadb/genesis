<!-- src/views/components/ledger/dashboard/DashboardHeader.vue -->
<script setup lang="ts">
import { useTemplateRef, onMounted, onUnmounted } from 'vue'
import { Calendar, User, House } from 'lucide-vue-next'
import IllustrationMascot from '../../ui/IllustrationMascot.vue'
import AppBar from '../../ui/AppBar.vue'
import type { TenantSummary } from '../../../../models/services/TenantSessionService'

defineProps<{
  isAuthed: boolean
  activeTenantObj: TenantSummary | null
}>()

const emit = defineEmits<{
  (e: 'open-periodo'): void
  (e: 'openSettings'): void
}>()

// ─── DOM Refs (Direct DOM Mutation Pattern) ───────────────────────────────────
const appBarRef = useTemplateRef<InstanceType<typeof AppBar>>('appBarRef')
const leftBtnRef = useTemplateRef<HTMLElement>('leftBtnRef')
const leftLabelRef = useTemplateRef<HTMLElement>('leftLabelRef')
const rightBtnRef = useTemplateRef<HTMLElement>('rightBtnRef')
const rightLabelRef = useTemplateRef<HTMLElement>('rightLabelRef')  
const centerRef = useTemplateRef<HTMLElement>('centerRef')
const mascotRef = useTemplateRef<HTMLElement>('mascotRef')
const tenantNameRef = useTemplateRef<HTMLElement>('tenantNameRef')

// ─── Flutter-Faithful SliverAppBar Constants & State ──────────────────────────
const EXPANDED_HEIGHT = 120
const COLLAPSED_HEIGHT = 52
const MAX_SHRINK_OFFSET = EXPANDED_HEIGHT - COLLAPSED_HEIGHT // 68

// Variáveis de estado locais — plain let, NUNCA Vue ref() reativo
let shrinkOffset = 0 // [0, MAX_SHRINK_OFFSET]
let rafId: number | null = null

/**
 * commitStyles — aplica as mutações de estilo diretamente sobre os elementos DOM.
 * Sem acoplar à reatividade do Vue e livre de transições CSS conflitantes.
 */
function commitStyles(t: number): void {
  const header = appBarRef.value?.headerEl
  const parallax = appBarRef.value?.parallaxEl
  if (!header) return

  try {
    const pad = parseFloat(getComputedStyle(header).getPropertyValue('--parent-pad')) || 24
  const translateY = (t * MAX_SHRINK_OFFSET) / 2
  const translateX = Math.max(0, pad - 16) * t

  // ── AppBar <header> ────────────────────────────────────────────────────────
  header.style.backgroundColor = t > 0.05
    ? `rgba(251, 250, 249, ${Math.min(0.98, 0.98 * t)})`
    : 'transparent'
  header.style.boxShadow = t > 0.6
    ? `0 ${6 * t * t}px ${24 * t}px -4px rgba(67,70,69,${0.08 * t}), 0 0 1px rgba(18,18,18,${0.1 * t})`
    : 'none'
  header.style.borderBottom = `1px solid rgba(242, 240, 237, ${Math.max(0, (t - 0.8) * 10)})`

  // ── Parallax Layer ─────────────────────────────────────────────────────────
  if (parallax) {
    parallax.style.opacity = String(1 - t)
    parallax.style.transform = `translateY(${t * 24}px)`
  }

  // ── Branding Central ───────────────────────────────────────────────────────
  if (centerRef.value) {
    centerRef.value.style.transform = `translateY(${translateY}px) scale(${1 - 0.12 * t})`
  }

  // ── Mascote (outer wrapper — transform exclusivo do scroll) ────────────────
  if (mascotRef.value) {
    mascotRef.value.style.top = `${-14 + 18 * t}px`
    mascotRef.value.style.right = `${-12 + 12 * t}px`
    mascotRef.value.style.transform = `scale(${0.95 - 0.2 * t}) rotate(${4 - 4 * t}deg)`
  }

  // ── Tenant Name ────────────────────────────────────────────────────────────
  if (tenantNameRef.value) {
    tenantNameRef.value.style.opacity = String(Math.max(0, 1 - 2.5 * t))
  }

  // ── Botão Esquerdo ─────────────────────────────────────────────────────────
  if (leftBtnRef.value) {
    leftBtnRef.value.style.transform = `translateY(${translateY}px) translateX(${-translateX}px) scale(${1 - 0.05 * t})`
    leftBtnRef.value.style.backgroundColor = `rgba(242, 240, 237, ${0.4 + 0.1 * t})`
    leftBtnRef.value.style.boxShadow = t > 0.8 ? 'var(--shadow-subtle)' : 'none'
  }
  if (leftLabelRef.value) {
    leftLabelRef.value.style.transform = `scale(${1 - 0.1 * t})`
  }

  // ── Botão Direito ──────────────────────────────────────────────────────────
  if (rightBtnRef.value) {
    rightBtnRef.value.style.transform = `translateY(${translateY}px) translateX(${translateX}px) scale(${1 - 0.05 * t})`
    rightBtnRef.value.style.backgroundColor = `rgba(242, 240, 237, ${0.4 + 0.1 * t})`
    rightBtnRef.value.style.boxShadow = t > 0.8 ? 'var(--shadow-subtle)' : 'none'
  }
  if (rightLabelRef.value) {
    rightLabelRef.value.style.transform = `scale(${1 - 0.1 * t})`
  }
  } catch {
    // AppBar's internal DOM structure changed — skip animation frame.
    // Next scroll event will re-attempt with fresh refs.
  }
}

/**
 * applyStyles — implementa a rolagem linear simples (Pinned SliverAppBar).
 */
function applyStyles(): void {
  const currentScrollY = window.scrollY
  shrinkOffset = Math.max(0, Math.min(MAX_SHRINK_OFFSET, currentScrollY))
  const t = shrinkOffset / MAX_SHRINK_OFFSET
  commitStyles(t)
}

function handleScroll(): void {
  if (rafId !== null) cancelAnimationFrame(rafId)
  rafId = requestAnimationFrame(applyStyles)
}

onMounted(() => {
  shrinkOffset = Math.min(window.scrollY, MAX_SHRINK_OFFSET)
  commitStyles(shrinkOffset / MAX_SHRINK_OFFSET)

  window.addEventListener('scroll', handleScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
  if (rafId !== null) cancelAnimationFrame(rafId)
})
</script>

<template>
  <!--
    DashboardHeader — Zero-Jitter SliverAppBar (Direction-Aware State Machine)
    
    Todos os estilos driven por scroll são aplicados via el.style dentro do RAF (commitStyles).
    O template NÃO usa :style para propriedades interpoladas pelo scroll.
    CSS transitions nessas propriedades estão removidas — exceto one-shot snap transition.
  -->
  <AppBar
    ref="appBarRef"
    class="mb-4"
  >
    <!-- Slot Esquerdo: Botão Período -->
    <template #left>
      <button
        ref="leftBtnRef"
        class="flex items-center gap-2.5 text-left group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ember focus-visible:ring-offset-4 rounded-2xl px-3 py-1.5 cursor-pointer active:scale-95 origin-left border border-stone/20"
        aria-label="Selecionar período"
        @click="emit('open-periodo')"
      >
        <div class="w-8 h-8 rounded-full bg-ember/10 flex items-center justify-center group-hover:bg-ember/20 group-hover:text-ember transition-colors duration-300">
          <Calendar
            class="w-4 h-4 text-ember group-hover:scale-110 transition-transform duration-500 ease-jelly"
            aria-hidden="true"
          />
        </div>
        <div
          ref="leftLabelRef"
          class="flex flex-col left-label-stack"
        >
          <span class="text-[7.5px] font-bold uppercase tracking-[0.2em] mb-0.5 text-ash/60 group-hover:text-ember transition-colors duration-300">Seletor</span>
          <span class="text-base font-bold tracking-tight leading-none text-charcoal group-hover:text-ember transition-colors duration-300">Período</span>
        </div>
      </button>
    </template>

    <!-- Slot Central: Branding DIVI. + Mascote Guardião -->
    <template #center>
      <div
        v-if="isAuthed && activeTenantObj"
        ref="centerRef"
        class="flex flex-col items-center justify-center select-none relative"
      >
        <!--
          mascotRef (OUTER WRAPPER):
          Recebe top, right, transform via commitStyles — NUNCA via CSS animation.
          O wobble está isolado no INNER WRAPPER abaixo (Safeguard #8).
        -->
        <div
          ref="mascotRef"
          class="absolute z-0 opacity-80 pointer-events-none mascot-outer"
          aria-hidden="true"
        >
          <!-- INNER WRAPPER: animate-wobble isolado — não conflita com transform do outer -->
          <div class="animate-wobble">
            <IllustrationMascot
              variant="ember"
              :size="24"
              mood="happy"
            />
          </div>
        </div>

        <div
          ref="tenantNameRef"
          class="relative mb-1.5 text-[8.5px] font-bold uppercase tracking-[0.2em] whitespace-nowrap flex items-center gap-2 justify-center text-ember"
        >
          <House
            class="w-2 h-2 text-ember/40"
            aria-hidden="true"
          />
          <span>{{ activeTenantObj.name }}</span>
        </div>

        <h1 class="font-display text-3xl font-bold text-charcoal tracking-[-0.04em] leading-none relative z-10">
          DIVI<span class="text-ember">.</span>
        </h1>
      </div>

      <!-- Estado não autenticado -->
      <div
        v-else
        ref="centerRef"
        class="flex flex-col items-center justify-center select-none relative px-4"
      >
        <div
          ref="mascotRef"
          class="absolute z-0 opacity-80 pointer-events-none mascot-outer"
          aria-hidden="true"
        >
          <div class="animate-wobble">
            <IllustrationMascot
              variant="ember"
              :size="24"
              mood="happy"
            />
          </div>
        </div>
        <span
          ref="tenantNameRef"
          class="text-[7.5px] font-bold text-ash/60 uppercase tracking-[0.25em] block leading-none mb-1.5 relative z-10"
        >Finanças Residenciais</span>
        <h1 class="font-display text-3xl font-bold text-charcoal tracking-[-0.04em] leading-none relative z-10">
          DIVI<span class="text-ember">.</span>
        </h1>
      </div>
    </template>

    <!-- Slot Direito: Ajustes -->
    <template #right>
      <button
        ref="rightBtnRef"
        class="flex items-center gap-2.5 text-right group focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ember focus-visible:ring-offset-4 rounded-2xl px-3 py-1.5 cursor-pointer active:scale-95 origin-right border border-stone/20"
        aria-label="Abrir ajustes"
        @click="emit('openSettings')"
      >
        <div
          ref="rightLabelRef"
          class="flex flex-col text-right right-label-stack"
        >
          <span class="text-[7.5px] font-bold uppercase tracking-[0.2em] mb-0.5 text-ash/60 group-hover:text-ember whitespace-nowrap transition-colors duration-300">Conta</span>
          <span class="text-xs font-bold text-charcoal leading-none whitespace-nowrap">Ajustes</span>
        </div>
        <div class="w-8 h-8 rounded-full bg-white/40 flex items-center justify-center group-hover:bg-ember/10 group-hover:text-ember transition-colors duration-300">
          <User
            class="w-4 h-4 group-hover:scale-110 transition-transform duration-500 ease-jelly"
            aria-hidden="true"
          />
        </div>
      </button>
    </template>
  </AppBar>
</template>

<style scoped>
/*
 * Norm #5 — CSS/JS Transition Separation:
 * Propriedades mutadas pelo RAF loop (transform, background-color, box-shadow,
 * opacity, height, margin, padding, width, border) NÃO devem ter transition aqui.
 * Exceção única: one-shot transition aplicada programaticamente no snap (commitStyles).
 * CSS transition só é permitida em hover/focus em props NÃO driven por scroll.
 */

.left-label-stack {
  transform-origin: left center;
}

.right-label-stack {
  transform-origin: right center;
}

/*
 * mascot-outer: outer wrapper com transform exclusivamente owned pelo RAF loop.
 * Safeguard #8: NUNCA adicionar CSS animation em transform neste elemento.
 * top e right são sobrescritos pelo commitStyles via el.style.
 */
.mascot-outer {
  position: absolute;
  top: -14px;  /* valor inicial (t=0, expandido) */
  right: -12px;
}
</style>