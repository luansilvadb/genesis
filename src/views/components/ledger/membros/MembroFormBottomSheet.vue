<script setup lang="ts">
import { ref } from 'vue'
import Button from '../../ui/Button.vue'
import MembroAvatar from '../../ui/MembroAvatar.vue'
import { ArrowLeft } from 'lucide-vue-next'
import { aplicarMascaraBRLText } from '../../../../shared/utils/formatarMoeda'

interface Props {
  activeTenantId: string | null
}

defineProps<Props>()
const emit = defineEmits(['salvar', 'cancelar'])

const novoNome = ref('')
const novoEmail = ref('')
const novoPassword = ref('')
const novaRendaText = ref('')

const resetForm = () => {
  novoNome.value = ''
  novoEmail.value = ''
  novoPassword.value = ''
  novaRendaText.value = ''
}

const handleRendaInput = (e: Event) => {
  const target = e.target as HTMLInputElement
  novaRendaText.value = aplicarMascaraBRLText(target.value)
}

const handleAdicionar = () => {
  let rendaCentavos: number | undefined = undefined
  if (novaRendaText.value) {
    const cleanValue = novaRendaText.value.replace(/\./g, '').replace(',', '.')
    const floatVal = parseFloat(cleanValue)
    if (!isNaN(floatVal)) {
      rendaCentavos = Math.round(floatVal * 100)
    }
  }

  emit('salvar', {
    nome: novoNome.value,
    email: novoEmail.value,
    password: novoPassword.value,
    rendaCentavos: rendaCentavos
  })
}

defineExpose({ resetForm })
</script>

<script lang="ts">
export default {
  name: 'MembroFormBottomSheet'
}
</script>

<template>
  <div class="animate-in fade-in slide-in-from-right-3 duration-300">
    <!-- Título com Botão Circular de Voltar -->
    <div class="px-6 pt-6 pb-2 flex items-center gap-3">
      <button 
        type="button"
        class="w-10 h-10 rounded-full bg-white border border-stone/60 text-charcoal flex items-center justify-center cursor-pointer shadow-sm hover:scale-105 hover:text-ember hover:border-ash/50 active:scale-95 transition-all duration-300 ease-out focus:outline-none" 
        aria-label="Voltar"
        @click="emit('cancelar')"
      >
        <ArrowLeft class="w-5 h-5" />
      </button>
      <div>
        <h3 class="text-heading-sm text-charcoal flex items-center gap-2">
          Novo <span class="text-ember">Morador</span>
        </h3>
        <p class="text-[10px] text-ash font-medium mt-0.5 uppercase tracking-wider">
          Adicione um novo morador à casa
        </p>
      </div>
    </div>

    <!-- Conteúdo do Formulário -->
    <div class="p-6 space-y-6">
      <div class="flex items-center gap-3 p-3.5 rounded-2xl border border-stone bg-parchment/30">
        <MembroAvatar
          :nome="novoNome.trim() || '?'"
          variant="ember"
          size="md"
        />
        <div class="flex-1 min-w-0">
          <span class="text-sm font-bold text-charcoal leading-none block truncate">{{ novoNome.trim() || 'Nome do morador...' }}</span>
          <p class="text-[10px] text-ash mt-0.5">
            {{ novoEmail.trim() || 'morador@email.com' }}
          </p>
        </div>
        <span class="px-2 py-0.5 rounded-full text-[8px] font-bold bg-stone text-ash uppercase tracking-widest shrink-0">Preview</span>
      </div>

      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Nome Completo</label>
        <div class="relative w-full">
          <input
            v-model="novoNome"
            maxlength="50"
            type="text"
            placeholder="Ex: Luana Oliveira"
            class="w-full px-4 py-3.5 pr-12 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal focus:border-ember transition-all text-sm"
          >
          <span
            v-if="novoNome.length > 0"
            class="absolute right-3 top-1/2 -translate-y-1/2 text-[10px] font-bold text-ash/60"
          >{{ novoNome.length }}/50</span>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-4">
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase text-graphite tracking-widest ml-1 block">E-mail</label>
          <input
            v-model="novoEmail"
            type="email"
            placeholder="exemplo@email.com"
            autocomplete="email"
            class="w-full px-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal focus:border-ember placeholder:text-ash text-sm transition-all"
          >
        </div>
        <div class="space-y-2">
          <label class="text-[10px] font-bold uppercase text-graphite tracking-widest ml-1 block">Senha</label>
          <input
            v-model="novoPassword"
            type="password"
            placeholder="••••••"
            class="w-full px-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal focus:border-ember placeholder:text-ash text-sm transition-all"
          >
        </div>
      </div>

      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Renda Mensal (R$ - Opcional)</label>
        <input
          v-model="novaRendaText"
          type="text"
          placeholder="Ex: 3.500,00"
          class="w-full px-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal focus:border-ember transition-all text-sm"
          @input="handleRendaInput"
        >
      </div>

      <!-- Botões de Ação -->
      <div class="flex gap-2.5 pt-2">
        <Button
          variant="secondary"
          class="flex-1 font-bold uppercase tracking-widest text-[10px] h-12"
          @click="emit('cancelar')"
        >
          Cancelar
        </Button>
        <Button 
          variant="primary" 
          :disabled="!novoNome.trim() || !novoEmail.trim() || !novoPassword.trim() || !activeTenantId" 
          class="flex-1 font-bold uppercase tracking-widest text-[10px] h-12" 
          @click="handleAdicionar"
        >
          Cadastrar
        </Button>
      </div>
    </div>
  </div>
</template>

