<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { Plus, Minus } from 'lucide-vue-next'
import { formatarBRL, aplicarMascaraBRLText } from '../../../shared/utils/formatarMoeda'

interface Props {
  valor: number
  installments: number
  wizFlow: 'expense' | 'loan' | null
  wizPayment: 'pix' | 'card' | null
}

const props = defineProps<Props>()
const emit = defineEmits(['update:valor', 'update:installments'])

const valorFormatado = ref('')

onMounted(() => {
  if (props.valor > 0) {
    valorFormatado.value = formatarBRL(props.valor, false)
  }
})

watch(() => props.valor, (newVal) => {
  if (newVal === 0 && valorFormatado.value !== '') {
    valorFormatado.value = ''
  } else if (newVal > 0) {
    const numericCurrent = parseFloat(valorFormatado.value.replace(/\./g, '').replace(',', '.'))
    if (numericCurrent !== newVal) {
      valorFormatado.value = formatarBRL(newVal, false)
    }
  }
})

const handleInput = (e: Event) => {
  const target = e.target as HTMLInputElement
  const maskedValue = aplicarMascaraBRLText(target.value)
  valorFormatado.value = maskedValue
  
  if (maskedValue === '') {
    emit('update:valor', 0)
  } else {
    const cleanValue = maskedValue.replace(/\./g, '').replace(',', '.')
    emit('update:valor', parseFloat(cleanValue))
  }
}

const internalInstallments = computed({
  get: () => props.installments,
  set: (val) => emit('update:installments', val)
})

const infoParcelamento = computed(() => {
  if (props.installments <= 1) return 'À vista'
  const parcela = Number(props.valor) / props.installments
  return `${props.installments}x de ${formatarBRL(parcela)}`
})
</script>

<template>
  <div class="space-y-5">
    <div class="rounded-card bg-parchment p-5 shadow-subtle transition-all duration-300">
      <label
        for="wizard-value-input"
        class="block text-[10px] font-bold text-graphite uppercase tracking-widest mb-2"
      >Valor total do lançamento</label>
      <div class="flex items-center gap-2">
        <span
          class="text-[23px] font-bold text-charcoal tracking-tight"
          aria-hidden="true"
        >R$</span>
        <input
          id="wizard-value-input"
          :value="valorFormatado"
          type="text"
          inputmode="numeric"
          class="[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none w-full bg-transparent border-none outline-none text-[40px] leading-none font-bold text-midnight tracking-tighter placeholder:text-ash"
          placeholder="0,00"
          autofocus
          @input="handleInput"
        >
      </div>
    </div>

    <div
      v-if="wizFlow === 'loan' || wizPayment === 'card'"
      class="rounded-card bg-white shadow-subtle p-4 space-y-3"
    >
      <span class="block text-[10px] font-bold text-graphite uppercase tracking-widest">Opções de Parcelamento</span>
      <div class="flex items-center justify-between gap-3">
        <button 
          type="button" 
          class="w-10 h-10 rounded-full bg-stone hover:opacity-80 flex items-center justify-center border-none cursor-pointer transition-opacity" 
          aria-label="Diminuir parcelas"
          @click="internalInstallments = Math.max(1, internalInstallments - 1)"
        >
          <Minus
            class="w-4 h-4"
            aria-hidden="true"
          />
        </button>
        <div
          class="text-center"
          aria-live="polite"
        >
          <span class="text-[23px] font-bold text-charcoal tracking-tight">{{ internalInstallments }}x</span>
          <p class="text-xs font-semibold text-graphite">
            {{ infoParcelamento }}
          </p>
        </div>
        <button 
          type="button" 
          class="w-10 h-10 rounded-full bg-stone hover:opacity-80 flex items-center justify-center border-none cursor-pointer transition-opacity" 
          aria-label="Aumentar parcelas"
          @click="internalInstallments = Math.max(1, internalInstallments + 1)"
        >
          <Plus
            class="w-4 h-4"
            aria-hidden="true"
          />
        </button>
      </div>
    </div>
  </div>
</template>
