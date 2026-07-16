<!-- src/views/components/wizard/StepPaymentMethodSelection.vue -->
<script setup lang="ts">
import { Wallet, CreditCard } from 'lucide-vue-next'
import { obterCorCartao } from '../../../shared/utils/obterCorCartao'

interface Props {
  cartoes: any[]
  isCartaoTrancado: (id: string) => boolean
  selectedCardOwnerId?: string | null
  selectedPaymentMethod?: 'pix' | 'card' | null
}

defineProps<Props>()
const emit = defineEmits(['select'])

const selecionarMetodo = (payment: 'pix' | 'card', cardOwner: string | null) => {
  emit('select', { payment, cardOwner })
}
</script>

<template>
  <div
    class="grid gap-3"
    role="listbox"
    aria-label="Opções de pagamento"
  >
    <button
      role="option"
      :aria-selected="selectedPaymentMethod === 'pix'"
      class="group w-full flex items-center gap-3 p-4 rounded-card bg-parchment hover:bg-stone transition-colors text-left border-none cursor-pointer"
      @click="selecionarMetodo('pix', null)"
    >
      <div class="w-10 h-10 rounded-full bg-white shadow-subtle text-graphite flex items-center justify-center shrink-0">
        <Wallet
          class="w-5 h-5"
          aria-hidden="true"
        />
      </div>
      <div class="min-w-0">
        <strong class="block text-[15px] font-bold text-charcoal tracking-tight">PIX ou Dinheiro</strong>
        <span class="text-xs text-graphite font-semibold">Gasto à vista do caixa</span>
      </div>
    </button>

    <button
      v-for="c in cartoes"
      :key="c.id"
      :disabled="isCartaoTrancado(c.id)"
      role="option"
      :aria-selected="selectedCardOwnerId === c.id"
      class="group w-full flex items-center gap-3 p-4 rounded-card bg-parchment hover:bg-stone transition-colors text-left disabled:opacity-40 disabled:cursor-not-allowed border-none cursor-pointer"
      @click="selecionarMetodo('card', c.id)"
    >
      <div 
        class="w-10 h-10 rounded-full shadow-subtle flex items-center justify-center shrink-0 border transition-all duration-300"
        :style="{ 
          backgroundColor: obterCorCartao(c.nome) + '10', 
          borderColor: obterCorCartao(c.nome) + '20' 
        }"
      >
        <CreditCard
          class="w-5 h-5"
          :style="{ color: obterCorCartao(c.nome) }"
          aria-hidden="true"
        />
      </div>
      <div class="min-w-0 flex-1">
        <div class="flex items-center gap-2">
          <strong class="text-[15px] font-bold text-charcoal tracking-tight">Cartão {{ c.nome }}</strong>
          <span
            v-if="isCartaoTrancado(c.id)"
            class="text-[10px] font-bold text-coral bg-coral/10 px-2 py-0.5 rounded border border-coral/20 shrink-0"
          >FECHADA</span>
        </div>
        <span class="text-xs text-graphite font-semibold">Despesa sob fatura</span>
      </div>
    </button>
  </div>
</template>
