<script setup lang="ts">
import { ShoppingCart, Handshake } from 'lucide-vue-next'

interface Props {
  wizFlow: 'expense' | 'loan' | 'loan_given' | 'loan_taken' | null
  isPrivate?: boolean
}

defineProps<Props>()
const emit = defineEmits(['select'])
</script>

<template>
  <div
    class="grid gap-3"
    role="listbox"
    aria-label="Opções de fluxo"
  >
    <template v-if="!isPrivate">
      <!-- Gasto Compartilhado (Casa) -->
      <button
        role="option"
        :aria-selected="wizFlow === 'expense'"
        class="group w-full flex items-center gap-3 p-4 rounded-card bg-parchment hover:bg-stone transition-all text-left border border-transparent cursor-pointer"
        :class="{ '!border-midnight bg-white shadow-subtle': wizFlow === 'expense' }"
        @click="emit('select', 'expense')"
      >
        <div class="w-10 h-10 rounded-full bg-white shadow-subtle text-graphite flex items-center justify-center shrink-0 border border-stone/10">
          <ShoppingCart
            class="w-5 h-5 text-midnight"
            aria-hidden="true"
          />
        </div>
        <div class="min-w-0">
          <strong class="block text-[15px] font-bold text-charcoal tracking-tight">Despesa Compartilhada</strong>
          <span class="text-xs text-graphite font-semibold">Conta ou compra que será dividida com a casa</span>
        </div>
      </button>

      <!-- Empréstimo (Casa) -->
      <button
        role="option"
        :aria-selected="wizFlow === 'loan'"
        class="group w-full flex items-center gap-3 p-4 rounded-card bg-parchment hover:bg-stone transition-all text-left border border-transparent cursor-pointer"
        :class="{ '!border-midnight bg-white shadow-subtle': wizFlow === 'loan' }"
        @click="emit('select', 'loan')"
      >
        <div class="w-10 h-10 rounded-full bg-white shadow-subtle text-graphite flex items-center justify-center shrink-0 border border-stone/10">
          <Handshake
            class="w-5 h-5 text-midnight"
            aria-hidden="true"
          />
        </div>
        <div class="min-w-0">
          <strong class="block text-[15px] font-bold text-charcoal tracking-tight">Repasse Direto</strong>
          <span class="text-xs text-graphite font-semibold">Transferência de dinheiro apenas entre moradores</span>
        </div>
      </button>
    </template>

    <template v-else>
      <!-- Opções Exclusivas Pessoal -->
      <!-- A Receber -->
      <button
        role="option"
        :aria-selected="wizFlow === 'loan_given'"
        class="group w-full flex items-center gap-3 p-4 rounded-card bg-parchment hover:bg-stone transition-all text-left border border-transparent cursor-pointer"
        :class="{ '!border-midnight bg-white shadow-subtle': wizFlow === 'loan_given' }"
        @click="emit('select', 'loan_given')"
      >
        <div class="w-10 h-10 rounded-full bg-meadow/10 shadow-subtle text-meadow flex items-center justify-center shrink-0 border border-meadow/20">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          ><path d="m5 12 7-7 7 7" /><path d="M12 19V5" /></svg>
        </div>
        <div class="min-w-0">
          <strong class="block text-[15px] font-bold text-meadow tracking-tight">A Receber</strong>
          <span class="text-xs text-graphite font-semibold">Alguém lhe deve dinheiro (ex: emprestei, vendi algo)</span>
        </div>
      </button>

      <!-- A Pagar -->
      <button
        role="option"
        :aria-selected="wizFlow === 'loan_taken'"
        class="group w-full flex items-center gap-3 p-4 rounded-card bg-parchment hover:bg-stone transition-all text-left border border-transparent cursor-pointer"
        :class="{ '!border-midnight bg-white shadow-subtle': wizFlow === 'loan_taken' }"
        @click="emit('select', 'loan_taken')"
      >
        <div class="w-10 h-10 rounded-full bg-coral/10 shadow-subtle text-coral flex items-center justify-center shrink-0 border border-coral/20">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          ><path d="M12 5v14" /><path d="m19 12-7 7-7-7" /></svg>
        </div>
        <div class="min-w-0">
          <strong class="block text-[15px] font-bold text-coral tracking-tight">A Pagar</strong>
          <span class="text-xs text-graphite font-semibold">Você deve dinheiro a alguém (ex: prestação, peguei)</span>
        </div>
      </button>
    </template>
  </div>
</template>
