<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  descricao: string
  wizFlow: 'expense' | 'loan' | null
}

const props = defineProps<Props>()
const emit = defineEmits(['update:descricao'])

const internalDescricao = computed({
  get: () => props.descricao,
  set: (val) => emit('update:descricao', val)
})

const quickChips = computed(() => props.wizFlow === 'loan' 
  ? ['Empréstimo', 'Luz dele', 'Uber compartilhado', 'Supermercado']
  : ['Mercado', 'Ifood', 'Luz', 'Internet', 'Água', 'Limpeza']
)
</script>

<template>
  <div class="space-y-5">
    <div class="rounded-card bg-parchment p-4 shadow-subtle">
      <label
        for="wizard-description-input"
        class="block text-[10px] font-bold text-graphite uppercase tracking-widest mb-2"
      >O que foi comprado?</label>
      <input
        id="wizard-description-input"
        v-model="internalDescricao"
        type="text"
        class="w-full bg-transparent border-none outline-none text-[23px] font-bold text-charcoal tracking-tight placeholder:text-ash"
        placeholder="Ex: Supermercado do mês"
        autofocus
      >
    </div>
    <div
      class="flex gap-2 flex-wrap"
      role="group"
      aria-label="Sugestões rápidas"
    >
      <button
        v-for="chip in quickChips"
        :key="chip"
        class="px-3.5 py-2 rounded-full bg-stone hover:bg-ash/20 text-[11px] font-bold text-graphite transition-colors border-none cursor-pointer uppercase tracking-wider"
        @click="internalDescricao = chip"
      >
        {{ chip }}
      </button>
    </div>
  </div>
</template>
