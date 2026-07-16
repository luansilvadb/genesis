<script setup lang="ts">
import { ref, watch } from 'vue'
import Button from '../../ui/Button.vue'
import { Wallet, Banknote } from 'lucide-vue-next'
import BottomSheet from '../../ui/BottomSheet.vue'
import { formatarBRL, aplicarMascaraBRLText } from '../../../../shared/utils/formatarMoeda'

interface Props {
  visible: boolean
  fromId?: string
  toId?: string
  fromName?: string
  toName?: string
  suggestedValue: number
  loading?: boolean
  subtitle?: string
}

const props = withDefaults(defineProps<Props>(), {
  fromId: '',
  toId: '',
  fromName: '',
  toName: '',
  subtitle: 'Confirmar a transferência entre moradores para equilibrar os saldos da casa.'
})
const emit = defineEmits(['confirm', 'cancel'])

const valorReal = ref(0)
const valorFormatado = ref('')
const method = ref<'pix' | 'cash'>('pix')
const descricao = ref('')
const metodos = [
  { id: 'pix', nome: 'Pix', icon: Wallet },
  { id: 'cash', nome: 'Dinheiro', icon: Banknote }
] as const

watch(() => props.visible, (isVisible) => {
  if (isVisible) {
    valorReal.value = props.suggestedValue
    valorFormatado.value = props.suggestedValue > 0 ? formatarBRL(props.suggestedValue, false) : ''
    method.value = 'pix'
    descricao.value = `Acerto: ${props.fromName} ➜ ${props.toName}`
  }
}, { immediate: true })

const handleValorInput = (e: Event) => {
  const target = e.target as HTMLInputElement
  const mascarado = aplicarMascaraBRLText(target.value)
  valorFormatado.value = mascarado
  valorReal.value = mascarado === '' ? 0 : parseFloat(mascarado.replace(/\./g, '').replace(',', '.'))
}

const handleConfirmar = () => {
  if (valorReal.value <= 0 || !props.fromId || !props.toId) return
  emit('confirm', {
    from: props.fromId,
    to: props.toId,
    valor: valorReal.value,
    method: method.value,
    descricao: descricao.value
  })
}

defineExpose({
  valorReal
})
</script>

<template>
  <BottomSheet 
    :model-value="visible" 
    :subtitle="props.subtitle"
    @update:model-value="val => { if (!val) emit('cancel') }"
  >
    <template #title>
      <h3 class="text-3xl font-display text-charcoal leading-tight">
        Registrar <span class="text-ember">Acerto</span>
      </h3>
    </template>

    <div class="space-y-6 pt-2">
      <div class="space-y-6">
        <!-- Valor Input -->
        <div class="space-y-2">
          <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Valor do Repasse</label>
          <div class="relative">
            <span class="absolute left-4 top-1/2 -translate-y-1/2 text-graphite text-sm font-bold">R$</span>
            <input 
              :value="valorFormatado"
              type="text"
              inputmode="numeric"
              class="w-full pl-10 pr-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-sm text-charcoal focus:border-ember transition-all"
              placeholder="0,00"
              @input="handleValorInput"
            >
          </div>
        </div>

        <!-- Descrição -->
        <div class="space-y-2">
          <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Descrição</label>
          <input 
            v-model="descricao"
            type="text"
            readonly
            class="w-full px-4 py-3.5 rounded-xl border border-stone bg-stone/30 outline-none font-bold text-sm text-charcoal cursor-default transition-all"
          >
        </div>

        <!-- Meio de Acerto -->
        <div class="space-y-2">
          <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Meio de Baixa</label>
          <div class="grid grid-cols-2 gap-2">
            <button 
              v-for="m in metodos"
              :key="m.id"
              type="button"
              class="flex flex-col items-center gap-2 py-3 rounded-xl transition-all duration-200 cursor-pointer"
              :class="[
                method === m.id 
                  ? 'border-2 border-charcoal bg-white text-charcoal font-bold shadow-sm' 
                  : 'border-2 border-transparent bg-stone text-charcoal hover:bg-ash/20'
              ]"
              @click="method = m.id"
            >
              <component
                :is="m.icon"
                class="w-4 h-4"
              />
              <span class="text-[11px] font-bold uppercase tracking-wider">{{ m.nome }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="grid grid-cols-2 gap-3">
        <Button
          variant="secondary"
          class="font-bold uppercase tracking-widest text-[10px] h-12"
          :disabled="loading"
          @click="emit('cancel')"
        >
          Cancelar
        </Button>
        <Button
          variant="primary"
          class="font-bold uppercase tracking-widest text-[10px] h-12"
          :disabled="valorReal <= 0 || loading"
          @click="handleConfirmar"
        >
          <span
            v-if="loading"
            class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin mr-2"
          />
          {{ loading ? 'Salvando...' : 'Confirmar' }}
        </Button>
      </div>
    </template>
  </BottomSheet>
</template>

<style scoped>
</style>
