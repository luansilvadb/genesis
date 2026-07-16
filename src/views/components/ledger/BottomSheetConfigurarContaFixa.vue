<script setup lang="ts">
import { ref, watch } from 'vue'
import type { ContaFixa } from '../../../models/entities/ContaFixa'
import BottomSheet from '../ui/BottomSheet.vue'
import Button from '../ui/Button.vue'
import MembroAvatar from '../ui/MembroAvatar.vue'
import { Check, ArrowLeft, Smile, ChevronRight } from 'lucide-vue-next'
import { formatarBRL, aplicarMascaraBRLText } from '../../../shared/utils/formatarMoeda'

const props = defineProps<{
  visible: boolean
  bill: ContaFixa | null
  membros: { id: string; nome: string }[]
  loading?: boolean
}>()

const emit = defineEmits(['save', 'delete', 'cancel'])

const name = ref('')
const icon = ref('💡')
const fixedValue = ref<number | null>(null)
const fixedValueFormatado = ref('')
const defaultSplit = ref<string[]>([])

const modoSelecaoIcone = ref(false)
const customIconInput = ref('')
const showingCustomInput = ref(false)

const allEmojis = [
  '💰', '💳', '💸', '🪙', '💵', '🧾', '🏦', '📈',
  '🏠', '🔑', '🔌', '💧', '🌐', '🧼', '🧹', '🛋️', '📦', '🪴', '🛠️', '🪵',
  '🛒', '🍔', '🍕', '☕', '🍎', '🍣', '🍻', '🥤', '🍿', '🍳', '🥚', '🥐', '🎂',
  '🚗', '⛽', '✈️', '🚇', '🚲', '🛵', '🎫', '🗺️', '⛵', '🚢', '🚁', '🧭',
  '🎮', '⚽', '🎵', '🏋️', '🐾', '🐶', '🐱', '🌻', '📚', '🎬', '🏆', '🎙️', '🧸',
  '💊', '🏥', '🦷', '🧴', '🧼', '👓', '🩺', '🧘', '🎁', '🎈',
  '⚡', '⚙️', '💿', '♻️', '🎯', '✉️', '📢', '🔒'
]

watch(() => props.bill, (newBill) => {
  modoSelecaoIcone.value = false
  customIconInput.value = ''
  showingCustomInput.value = false
  if (newBill) {
    name.value = newBill.name
    icon.value = newBill.icon
    const v = newBill.fixedValueCentavos !== null && newBill.fixedValueCentavos !== undefined ? newBill.fixedValueCentavos / 100 : null
    fixedValue.value = v
    fixedValueFormatado.value = v !== null && v > 0 ? formatarBRL(v, false) : ''
    
    const validSplitIds = (newBill.defaultSplit || []).filter(s =>
      props.membros.some(m => m.id === s.membroId)
    )
    if (validSplitIds.length > 0) {
      defaultSplit.value = validSplitIds.map(s => s.membroId)
    } else {
      defaultSplit.value = props.membros.map(m => m.id)
    }
  } else {
    name.value = ''
    icon.value = '💡'
    fixedValue.value = null
    fixedValueFormatado.value = ''
    defaultSplit.value = props.membros.map(m => m.id)
  }
}, { immediate: true })

watch(() => props.visible, (newVisible) => {
  if (!newVisible) {
    modoSelecaoIcone.value = false
    customIconInput.value = ''
    showingCustomInput.value = false
  }
})

const selecionarIcone = (e: string) => {
  icon.value = e
  modoSelecaoIcone.value = false
}

const confirmarCustomIcon = () => {
  if (customIconInput.value.trim()) {
    icon.value = customIconInput.value.trim()
    modoSelecaoIcone.value = false
    showingCustomInput.value = false
  }
}

const toggleSplit = (id: string) => {
  if (defaultSplit.value.includes(id)) {
    if (defaultSplit.value.length > 1) {
      defaultSplit.value = defaultSplit.value.filter(sid => sid !== id)
    }
  } else {
    defaultSplit.value.push(id)
  }
}

const handleFixedValueInput = (e: Event) => {
  const target = e.target as HTMLInputElement
  const mascarado = aplicarMascaraBRLText(target.value)
  fixedValueFormatado.value = mascarado
  fixedValue.value = mascarado === '' ? null : parseFloat(mascarado.replace(/\./g, '').replace(',', '.'))
}

const salvar = () => {
  emit('save', {
    id: props.bill?.id || `rec_custom_${Date.now()}`,
    name: name.value,
    icon: icon.value,
    fixedValueCentavos: fixedValue.value && fixedValue.value > 0 ? Math.round(fixedValue.value * 100) : null,
    defaultSplit: defaultSplit.value.map(id => ({ membroId: id, valorCentavos: 0 }))
  })
}
</script>

<template>
  <BottomSheet 
    :model-value="visible" 
    subtitle="Gerencie modelos de gastos recorrentes." 
    max-height="90dvh"
    @update:model-value="val => { if (!val) $emit('cancel') }"
  >
    <template #title>
      <h3
        v-if="!modoSelecaoIcone"
        class="text-3xl font-display text-charcoal leading-tight"
      >
        Configurar <span class="text-ember">Conta Fixa</span>
      </h3>
      <div
        v-else
        class="flex items-center gap-3"
      >
        <button 
          class="w-10 h-10 rounded-full bg-stone hover:bg-stone/80 text-charcoal flex items-center justify-center cursor-pointer transition-all border-none focus:outline-none" 
          @click="modoSelecaoIcone = false"
        >
          <ArrowLeft class="w-5 h-5" />
        </button>
        <h3 class="text-2xl font-display text-charcoal leading-tight">
          Selecione o <span class="text-ember">Ícone</span>
        </h3>
      </div>
    </template>

    <!-- Estado do Formulário Principal -->
    <div
      v-if="!modoSelecaoIcone"
      class="space-y-6 pt-2"
    >
      <!-- Nome -->
      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase tracking-widest text-graphite ml-1">Nome do Talão / Categoria</label>
        <input 
          v-model="name" 
          type="text" 
          class="w-full px-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal focus:border-ember transition-all text-sm" 
          placeholder="Ex: Aluguel, Internet..."
        >
      </div>

      <!-- Card do Emoji Representativo (Clicável) -->
      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase tracking-widest text-graphite ml-1">Emoji Representativo</label>
        <button 
          class="flex items-center gap-3 p-3.5 w-full rounded-2xl border border-stone bg-canvas hover:bg-stone/50 transition-all text-left group cursor-pointer"
          @click="modoSelecaoIcone = true"
        >
          <div class="w-12 h-12 rounded-xl bg-white border border-stone flex items-center justify-center text-2xl shadow-subtle group-hover:scale-105 transition-transform shrink-0">
            {{ icon }}
          </div>
          <div class="flex-grow">
            <span class="text-[10px] font-bold uppercase tracking-widest text-graphite block">Emoji Selecionado</span>
            <span class="text-xs text-ash font-medium mt-0.5 block">Clique para alterar o ícone</span>
          </div>
          <ChevronRight class="w-5 h-5 text-ash" />
        </button>
      </div>

      <!-- Valor Sugerido -->
      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase tracking-widest text-graphite ml-1">Valor Sugerido (Opcional)</label>
        <div class="relative">
          <span class="absolute left-4 top-1/2 -translate-y-1/2 text-graphite text-sm font-bold">R$</span>
          <input 
            :value="fixedValueFormatado"
            type="text"
            inputmode="numeric"
            class="w-full pl-10 pr-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal focus:border-ember transition-all text-sm"
            placeholder="0,00" 
            @input="handleFixedValueInput" 
          >
        </div>
      </div>

      <!-- Divisão Padrão -->
      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase tracking-widest text-graphite ml-1">Quem divide por padrão?</label>
        <div class="grid grid-cols-3 gap-2">
          <button 
            v-for="m in membros" 
            :key="m.id"
            class="group relative py-3 rounded-xl font-bold text-[11px] uppercase tracking-wider transition-all duration-300 border-none cursor-pointer flex flex-col items-center gap-2"
            :class="defaultSplit.includes(m.id) ? 'bg-white shadow-subtle scale-[1.02] text-charcoal' : 'bg-stone/50 text-graphite opacity-60 hover:opacity-100'"
            @click="toggleSplit(m.id)"
          >
            <MembroAvatar
              :nome="m.nome"
              size="sm"
              :variant="defaultSplit.includes(m.id) ? 'meadow' : 'sky'"
            />
            <span class="truncate max-w-full px-1">{{ m.nome }}</span>
            <div
              v-if="defaultSplit.includes(m.id)"
              class="absolute top-1.5 right-1.5 animate-in zoom-in-50 duration-300"
            >
              <Check class="w-3.5 h-3.5 text-meadow" />
            </div>
          </button>
        </div>
      </div>
    </div>

    <!-- Estado de Seleção de Ícone -->
    <div
      v-else
      class="space-y-6 pt-2 animate-in fade-in slide-in-from-right-3 duration-350"
    >
      <!-- Opção: Emoji Personalizado (Teclado Livre) -->
      <div class="space-y-3">
        <button 
          class="w-full flex items-center justify-between p-4 rounded-xl border border-stone bg-white hover:bg-stone/30 transition-all text-left cursor-pointer group shadow-sm border-none"
          @click="showingCustomInput = !showingCustomInput"
        >
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-stone/50 flex items-center justify-center shrink-0">
              <Smile class="w-4 h-4 text-ember" />
            </div>
            <div>
              <span class="text-xs font-bold text-charcoal block">Emoji Personalizado</span>
              <p class="text-[10px] text-ash font-medium mt-0.5">
                Use o teclado ou cole o caractere que desejar
              </p>
            </div>
          </div>
          <ChevronRight
            class="w-4 h-4 text-ash transition-transform duration-300"
            :class="{ 'rotate-90': showingCustomInput }"
          />
        </button>

        <!-- Campo de Input Customizado Expansível -->
        <div
          v-if="showingCustomInput"
          class="p-4 rounded-xl border border-stone bg-stone/20 space-y-3 animate-in fade-in slide-in-from-top-2 duration-250"
        >
          <div class="flex flex-col sm:flex-row gap-2">
            <input 
              v-model="customIconInput" 
              type="text" 
              placeholder="Cole ou digite um emoji..."
              class="w-full sm:flex-grow px-3.5 py-3 rounded-lg border border-stone bg-white outline-none font-bold text-charcoal focus:border-ember text-sm shadow-subtle"
              maxlength="4"
              @keyup.enter="confirmarCustomIcon"
            >
            <Button 
              variant="primary" 
              class="w-full sm:w-auto px-4 text-[10px] uppercase font-bold tracking-widest h-[46px] shrink-0" 
              :disabled="!customIconInput.trim()"
              @click="confirmarCustomIcon"
            >
              Confirmar
            </Button>
          </div>
          <p class="text-[9px] text-graphite font-bold uppercase tracking-wider ml-1">
            Dica: Limite de até 4 caracteres para caber no layout (ex: "🚀", "R$", "C6").
          </p>
        </div>
      </div>

      <!-- Grade de Ícones Curados -->
      <div class="space-y-2.5">
        <label class="block text-[10px] font-bold uppercase tracking-widest text-graphite ml-1">Coleção de Ícones</label>
        
        <div class="grid grid-cols-6 sm:grid-cols-8 gap-2 p-3 bg-stone/20 rounded-2xl border border-stone max-h-[340px] overflow-y-auto custom-scrollbar">
          <button 
            v-for="e in allEmojis" 
            :key="e"
            class="text-2xl w-12 h-12 flex items-center justify-center rounded-xl bg-white border border-stone/30 hover:bg-stone/50 active:scale-90 transition-all cursor-pointer shadow-sm"
            :class="icon === e ? 'bg-ember/10 border-ember scale-110 shadow-subtle ring-2 ring-ember/20' : ''"
            @click="selecionarIcone(e)"
          >
            {{ e }}
          </button>
        </div>
      </div>
    </div>

    <!-- Rodapé (Apenas na Tela Principal) -->
    <template
      v-if="!modoSelecaoIcone"
      #footer
    >
      <div class="flex flex-col gap-3">
        <div class="flex gap-3">
          <Button
            variant="secondary"
            class="flex-1 font-bold uppercase tracking-widest text-[10px] h-12"
            @click="$emit('cancel')"
          >
            Cancelar
          </Button>
          <Button
            variant="primary"
            class="flex-[2] font-bold uppercase tracking-widest text-[10px] h-12"
            :disabled="!name || props.loading"
            :loading="props.loading"
            @click="salvar"
          >
            Salvar Configuração
          </Button>
        </div>
        <button 
          v-if="bill"
          class="w-full py-2 text-[10px] font-bold uppercase tracking-widest text-coral hover:bg-coral/5 rounded-lg transition-all border-none bg-transparent cursor-pointer" 
          @click="$emit('delete', bill)"
        >
          Excluir Modelo de Conta
        </button>
      </div>
    </template>
  </BottomSheet>
</template>

<style scoped>
</style>

