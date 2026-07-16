<script setup lang="ts">
import { computed, ref } from 'vue'
import MembroAvatar from '../ui/MembroAvatar.vue'
import { Check, Plus } from 'lucide-vue-next'
import { obterMembrosSelecionadosSemRenda } from '../../../shared/utils/rateio'
import { formatarBRL } from '../../../shared/utils/formatarMoeda'

interface Member {
  id: string
  nome: string
  rendaCentavos?: number
}

interface Props {
  membros: Member[]
  participantesDivisao: string[]
  compradorSelecionadoId: string
  splitType: 'equal' | 'proportional'
  valorTotal?: number
  isPrivate?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits(['update:participantesDivisao', 'update:splitType', 'adicionar-externo'])

const internalParticipantes = computed({
  get: () => props.participantesDivisao,
  set: (val) => emit('update:participantesDivisao', val)
})

const toggleSplitMember = (id: string) => {
  const current = [...internalParticipantes.value]
  const idx = current.indexOf(id)
  if (idx >= 0) current.splice(idx, 1)
  else current.push(id)
  internalParticipantes.value = current
}

const selecionarTodos = () => {
  internalParticipantes.value = props.membros.map(m => m.id)
}

const selecionarApenasEu = () => {
  internalParticipantes.value = [props.compradorSelecionadoId]
}

const membrosSelecionadosSemRenda = computed(() =>
  obterMembrosSelecionadosSemRenda(props.membros, props.participantesDivisao)
)

const proporcionalDisponivel = computed(() =>
  props.participantesDivisao.length > 0 && membrosSelecionadosSemRenda.value.length === 0
)

const proporcoesMembros = computed(() => {
  if (props.splitType !== 'proportional' || !proporcionalDisponivel.value) {
    return {}
  }
  
  const participantesComRenda = props.participantesDivisao.map(id => {
    const m = props.membros.find(memb => memb.id === id)
    const renda = m?.rendaCentavos && Number(m.rendaCentavos) > 0 ? Number(m.rendaCentavos) : 0
    return { id, renda }
  })

  const somaRendasTotal = participantesComRenda.reduce((acc, p) => acc + p.renda, 0)
  const resultado: { [id: string]: { percent: number; valor?: number } } = {}

  if (somaRendasTotal <= 0) {
    // All selected members have zero or negative income — fall back to equal split.
    const equalPercent = participantesComRenda.length > 0 ? 100 / participantesComRenda.length : 0
    const equalValue = props.valorTotal ? props.valorTotal / (participantesComRenda.length || 1) : undefined
    participantesComRenda.forEach(p => {
      resultado[p.id] = { percent: equalPercent, valor: equalValue }
    })
  } else {
    participantesComRenda.forEach(p => {
      const percent = (p.renda / somaRendasTotal) * 100
      let valorEstimado: number | undefined = undefined
      if (props.valorTotal) {
        valorEstimado = props.valorTotal * (p.renda / somaRendasTotal)
      }

      resultado[p.id] = {
        percent,
        valor: valorEstimado
      }
    })
  }

  return resultado
})

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
  <div class="space-y-4 animate-in fade-in duration-300">
    <!-- Seletor do Tipo de Rateio -->
    <div class="flex justify-between items-center p-3.5 bg-stone/20 rounded-2xl border border-stone/60">
      <span class="text-xs font-bold text-charcoal">Divisão das Contas</span>
      <div class="inline-flex p-1 bg-stone/40 rounded-xl relative whitespace-nowrap">
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg font-bold text-[10px] uppercase tracking-widest cursor-pointer border-none transition-all duration-200"
          :class="[splitType === 'equal' ? 'bg-white text-charcoal shadow-sm' : 'bg-transparent text-ash hover:text-charcoal']"
          @click="emit('update:splitType', 'equal')"
        >
          Igual
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg font-bold text-[10px] uppercase tracking-widest cursor-pointer border-none transition-all duration-200"
          :class="[splitType === 'proportional' ? 'bg-white text-charcoal shadow-sm' : 'bg-transparent text-ash hover:text-charcoal']"
          @click="emit('update:splitType', 'proportional')"
        >
          Proporcional
        </button>
      </div>
    </div>

    <div
      v-if="splitType === 'proportional' && !proporcionalDisponivel"
      role="alert"
      class="p-3.5 rounded-2xl border border-sunburst/30 bg-sunburst/10 text-[11px] text-charcoal font-semibold leading-relaxed"
    >
      Informe uma renda positiva para
      <strong>{{ membrosSelecionadosSemRenda.map(membro => membro.nome).join(', ') }}</strong>
      ou escolha a divisão igual para continuar.
    </div>

    <div
      class="flex gap-2"
      role="group"
      aria-label="Atalhos de divisão"
    >
      <button 
        class="px-3.5 py-2 rounded-full bg-midnight text-white text-[10px] font-bold uppercase tracking-wider border-none cursor-pointer hover:bg-charcoal transition-colors" 
        @click="selecionarTodos"
      >
        Todos
      </button>
      <button 
        class="px-3.5 py-2 rounded-full bg-stone text-charcoal text-[10px] font-bold uppercase tracking-wider border-none cursor-pointer hover:bg-ash/20 transition-colors" 
        @click="selecionarApenasEu"
      >
        Apenas eu
      </button>
    </div>

    <div
      class="grid grid-cols-3 gap-2"
      role="listbox"
      aria-multiselectable="true"
      aria-label="Selecionar membros para dividir"
    >
      <button
        v-for="m in membros"
        :key="m.id"
        role="option"
        :aria-selected="internalParticipantes.includes(m.id)"
        class="group relative flex flex-col items-center gap-2 p-3 rounded-card transition-all duration-300 border-none cursor-pointer"
        :class="[internalParticipantes.includes(m.id) ? 'bg-white shadow-subtle scale-[1.02]' : 'bg-parchment opacity-80']"
        @click="toggleSplitMember(m.id)"
      >
        <MembroAvatar 
          :nome="m.nome.replace(' (Externo)', '')" 
          size="md" 
          :variant="internalParticipantes.includes(m.id) ? 'meadow' : 'sky'" 
        />
        <span class="text-[10px] font-bold text-charcoal uppercase tracking-tight truncate max-w-full px-1">{{ m.nome }}</span>
        <span 
          v-if="splitType === 'proportional' && internalParticipantes.includes(m.id) && proporcoesMembros[m.id]"
          class="text-[9px] font-bold text-ash mt-0.5 leading-none block text-center animate-in fade-in duration-300"
        >
          {{ Math.round(proporcoesMembros[m.id]?.percent ?? 0) }}%
          <span
            v-if="proporcoesMembros[m.id]?.valor !== undefined"
            class="block text-[8px] text-slate-500 font-semibold mt-0.5"
          >
            {{ formatarBRL(proporcoesMembros[m.id]?.valor ?? 0) }}
          </span>
        </span>
        <Check
          v-if="internalParticipantes.includes(m.id)"
          class="absolute top-2 right-2 w-3.5 h-3.5 text-meadow animate-in zoom-in-50 duration-300"
          aria-hidden="true"
        />
      </button>
    </div>

    <!-- Botão/Input para Adicionar Externo -->
    <div
      v-if="isPrivate"
      class="pt-4 border-t border-stone/50"
    >
      <div v-if="!mostrarInputExterno">
        <button 
          type="button"
          class="w-full py-3.5 rounded-xl border-2 border-dashed border-stone hover:border-ember/40 text-xs font-bold text-ash hover:text-ember transition-colors flex items-center justify-center gap-2 cursor-pointer bg-transparent"
          @click="mostrarInputExterno = true"
        >
          <Plus class="w-4 h-4" />
          Dividir com Pessoa Externa
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
