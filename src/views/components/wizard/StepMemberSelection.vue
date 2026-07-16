<script setup lang="ts">
import { ref } from 'vue'
import MembroAvatar from '../ui/MembroAvatar.vue'
import { Plus } from 'lucide-vue-next'

interface Member {
  id: string
  nome: string
}

interface Props {
  membros: Member[]
  currentState: string
  selectedId: string | null
  compradorSelecionadoId?: string | null
  isPrivate?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits(['select', 'adicionar-externo'])

import { computed } from 'vue'
import { useMembros } from '../../../viewmodels/useMembros'

const { currentMembro } = useMembros()

const filteredMembers = computed(() => {
  let list = props.membros
  if (props.currentState === 'BORROWER_SELECTION') {
    list = list.filter(m => m.id !== props.compradorSelecionadoId)
  }
  const currId = currentMembro.value?.id
  if (props.isPrivate && currId) {
    list = list.filter(m => m.id !== currId)
  }
  return list
})

const handleSelect = (id: string) => {
  emit('select', id)
}

const mostrarInputExterno = ref(false)
const nomeExterno = ref('')

const handleAdicionarExterno = () => {
  if (nomeExterno.value.trim()) {
    emit('adicionar-externo', nomeExterno.value.trim())
    nomeExterno.value = ''
    mostrarInputExterno.value = false
  }
}
</script>

<template>
  <div class="space-y-4">
    <div 
      class="grid grid-cols-2 gap-3"
      role="listbox"
      :aria-label="currentState === 'BORROWER_SELECTION' ? 'Selecionar quem pegou emprestado' : 'Selecionar quem pagou'"
    >
      <button
        v-for="m in filteredMembers"
        :key="m.id"
        role="option"
        :aria-selected="selectedId === m.id"
        class="group flex flex-col items-center gap-3 p-4 rounded-card bg-parchment hover:bg-stone transition-all duration-300 border-none cursor-pointer"
        @click="handleSelect(m.id)"
      >
        <MembroAvatar 
          :nome="m.nome.replace(' (Externo)', '')" 
          size="md" 
          :variant="selectedId === m.id ? 'ember' : 'sky'" 
        />
        <span class="font-bold text-[11px] text-charcoal uppercase tracking-wider truncate max-w-full px-1">{{ m.nome }}</span>
      </button>
    </div>

    <!-- Botão/Input para Adicionar Externo -->
    <div
      v-if="isPrivate"
      class="pt-4 border-t border-stone"
    >
      <div v-if="!mostrarInputExterno">
        <button 
          type="button"
          class="w-full py-3.5 rounded-xl border-2 border-dashed border-stone hover:border-ember/40 text-xs font-bold text-ash hover:text-ember transition-colors flex items-center justify-center gap-2 cursor-pointer bg-transparent"
          @click="mostrarInputExterno = true"
        >
          <Plus class="w-4 h-4" />
          + Pessoa Externa
        </button>
      </div>
      <div
        v-else
        class="flex gap-2 items-center animate-in fade-in duration-200"
      >
        <input 
          v-model="nomeExterno"
          type="text" 
          placeholder="Nome da pessoa externa"
          class="flex-1 px-4 py-3.5 rounded-xl border border-stone bg-canvas text-xs font-bold text-charcoal focus:outline-none focus:border-ember"
          @keyup.enter="handleAdicionarExterno"
        >
        <button 
          type="button"
          class="h-[46px] px-4 rounded-xl bg-midnight text-white text-[10px] uppercase font-bold tracking-wider cursor-pointer border-none"
          @click="handleAdicionarExterno"
        >
          Adicionar
        </button>
        <button 
          type="button"
          class="h-[46px] px-3 rounded-xl bg-stone text-charcoal text-[10px] uppercase font-bold tracking-wider cursor-pointer border-none"
          @click="mostrarInputExterno = false"
        >
          Cancelar
        </button>
      </div>
    </div>
  </div>
</template>
