<script setup lang="ts">
import { ref, computed } from 'vue'
import { User, Plus, ArrowLeft } from 'lucide-vue-next'
import { useMembros } from '../../../viewmodels/useMembros'
import { useToast } from '../../../composables/useToast'
import { Membro, type MembroRole } from '../../../models/entities/Membro'
import { mensagemErro } from '../../../shared/utils/mensagemErro'
import Button from '../ui/Button.vue'
import MembroListItem from '../ledger/membros/MembroListItem.vue'
import MembroFormBottomSheet from '../ledger/membros/MembroFormBottomSheet.vue'
import { aplicarMascaraBRLText, formatarCentavosParaBRL } from '../../../shared/utils/formatarMoeda'

defineProps<{
  activeTenantId: string | null
}>()

const emit = defineEmits(['focus-change'])

const { 
  membros, 
  adicionarMembro, 
  desativarMembro, 
  ativarMembro, 
  atualizarRoleMembro,
  atualizarRendaMembro,
  currentMembro 
} = useMembros()

const toast = useToast()

const variants: ('ember' | 'sky' | 'sunburst' | 'flamingo' | 'meadow')[] = ['ember', 'sky', 'sunburst', 'flamingo', 'meadow']

type FormExpose = { resetForm: () => void }
interface NovoMembroDados { nome: string; email: string; password: string; rendaCentavos?: number }

const membroFormRef = ref<FormExpose | null>(null)
const novoMembroFormAberto = ref(false)
const mostrarBottomSheet = ref(false)
const membroSelecionado = ref<Membro | null>(null)
const roleSelecionada = ref<MembroRole>('MORADOR')
const ativoSelecionado = ref(true)
const rendaSelecionadaText = ref('')
const salvando = ref(false)

const podeGerenciarMoradores = computed(() => {
  return currentMembro.value?.role === 'ADMIN'
})

const podeEditarRole = computed(() => {
  if (!membroSelecionado.value) return false
  if (currentMembro.value?.role !== 'ADMIN') return false
  if (membroSelecionado.value.userId === currentMembro.value?.userId) return false
  return true
})

const podeEditarStatus = computed(() => {
  if (!membroSelecionado.value) return false
  if (currentMembro.value?.role !== 'ADMIN') return false
  if (membroSelecionado.value.userId === currentMembro.value?.userId) return false
  const outrosAdminsAtivos = membros.value.filter(m => m.role === 'ADMIN' && m.ativo && m.id !== membroSelecionado.value?.id)
  if (membroSelecionado.value.role === 'ADMIN' && outrosAdminsAtivos.length === 0) return false
  return true
})

const abrirNovoMembroForm = () => {
  membroFormRef.value?.resetForm()
  novoMembroFormAberto.value = true
  emit('focus-change', true)
}

const handleAdicionarMembro = async (dados: NovoMembroDados) => {
  try {
    await adicionarMembro(dados.nome, dados.email, dados.password, dados.rendaCentavos)
    toast.show('Membro adicionado com sucesso', 'success')
    novoMembroFormAberto.value = false
    emit('focus-change', false)
  } catch (error: unknown) {
    toast.show(mensagemErro(error, 'Erro ao adicionar membro'), 'error')
  }
}

const handleRendaInput = (e: Event) => {
  const target = e.target as HTMLInputElement
  rendaSelecionadaText.value = aplicarMascaraBRLText(target.value)
}

const abrirEdicaoMembro = (membro: Membro) => {
  if (!podeGerenciarMoradores.value) return
  membroSelecionado.value = membro
  roleSelecionada.value = membro.role
  ativoSelecionado.value = membro.ativo
  rendaSelecionadaText.value = membro.rendaCentavos 
    ? formatarCentavosParaBRL(membro.rendaCentavos, false)
    : ''
  mostrarBottomSheet.value = true
  emit('focus-change', true)
}

const handleSalvarEdicao = async () => {
  if (!membroSelecionado.value) return
  salvando.value = true
  try {
    const novaRole = roleSelecionada.value
    
    const roleAlterada = novaRole !== membroSelecionado.value.role
    if (roleAlterada) {
      await atualizarRoleMembro(membroSelecionado.value.id, novaRole)
    }
    
    if (ativoSelecionado.value !== membroSelecionado.value.ativo) {
      if (ativoSelecionado.value) await ativarMembro(membroSelecionado.value.id)
      else await desativarMembro(membroSelecionado.value.id)
    }

    let novaRendaCentavos: number | undefined = undefined
    if (rendaSelecionadaText.value) {
      const cleanValue = rendaSelecionadaText.value.replace(/\./g, '').replace(',', '.')
      const floatVal = parseFloat(cleanValue)
      if (!isNaN(floatVal)) {
        novaRendaCentavos = Math.round(floatVal * 100)
      }
    }

    if (novaRendaCentavos !== membroSelecionado.value.rendaCentavos) {
      await atualizarRendaMembro(membroSelecionado.value.id, novaRendaCentavos)
    }
    
    toast.show('Alterações salvas com sucesso', 'success')
    mostrarBottomSheet.value = false
    emit('focus-change', false)
  } catch (error: unknown) {
    toast.show(mensagemErro(error, 'Erro ao salvar alterações'), 'error')
  } finally {
    salvando.value = false
  }
}

const cancelarEdicao = () => {
  mostrarBottomSheet.value = false
  emit('focus-change', false)
}

const cancelarNovoMembro = () => {
  novoMembroFormAberto.value = false
  emit('focus-change', false)
}
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-300">
    <!-- Modo Foco para Novo Morador -->
    <div
      v-if="novoMembroFormAberto"
      class="bg-white border border-stone/30 rounded-2xl shadow-subtle overflow-hidden"
    >
      <MembroFormBottomSheet 
        ref="membroFormRef"
        :active-tenant-id="activeTenantId"
        @salvar="handleAdicionarMembro"
        @cancelar="cancelarNovoMembro"
      />
    </div>
    <!-- Modo Foco para Editar Morador -->
    <div
      v-else-if="mostrarBottomSheet"
      class="bg-white border border-stone/30 rounded-2xl shadow-subtle overflow-hidden animate-in fade-in slide-in-from-right-3 duration-300"
    >
      <div class="px-6 pt-6 pb-2 flex items-center gap-3">
        <button 
          type="button"
          class="w-10 h-10 rounded-full bg-white border border-stone/60 text-charcoal flex items-center justify-center cursor-pointer shadow-sm hover:scale-105 hover:text-ember hover:border-ash/50 active:scale-95 transition-all duration-300 ease-out focus:outline-none" 
          aria-label="Voltar"
          @click="cancelarEdicao"
        >
          <ArrowLeft class="w-5 h-5" />
        </button>
        <div>
          <h3 class="text-heading-sm text-charcoal flex items-center gap-2">
            Editar <span class="text-ember">Morador</span>
          </h3>
          <p class="text-xs text-ash font-medium mt-0.5 uppercase tracking-wider">
            Gerencie {{ membroSelecionado?.nome }}
          </p>
        </div>
      </div>

      <div class="p-6 space-y-6">
        <!-- Papel na Casa (Role sistêmica) -->
        <div class="space-y-2">
          <label class="text-xs font-bold uppercase tracking-widest text-graphite block ml-1">Papel na Casa</label>
          <select
            v-model="roleSelecionada"
            :disabled="!podeEditarRole"
            class="w-full p-3.5 rounded-2xl border border-stone bg-white outline-none font-bold text-charcoal"
          >
            <option value="ADMIN">
              Administrador
            </option>
            <option value="MORADOR">
              Morador
            </option>
            <option value="VISUALIZADOR">
              Visualizador
            </option>
          </select>
          <!-- Descrição contextual por role -->
          <p
            v-if="roleSelecionada === 'VISUALIZADOR'"
            class="text-xs text-ash ml-1"
          >
            Exemplos: filho dependente, ex-parceiro em transição, contador externo.
          </p>
        </div>

        <div class="space-y-2">
          <label class="text-xs font-bold uppercase tracking-widest text-graphite block ml-1">Renda Mensal (R$)</label>
          <input
            v-model="rendaSelecionadaText"
            type="text"
            placeholder="Ex: 3.500,00"
            class="w-full p-3.5 rounded-2xl border border-stone bg-white outline-none font-bold text-charcoal focus:border-ember transition-all text-sm"
            @input="handleRendaInput"
          >
        </div>
        <div class="flex items-center justify-between p-3.5 bg-parchment border border-stone rounded-2xl">
          <span class="text-xs font-bold text-charcoal">Morador Ativo</span>
          <button
            :disabled="!podeEditarStatus"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors"
            :class="ativoSelecionado ? 'bg-meadow' : 'bg-stone'"
            @click="ativoSelecionado = !ativoSelecionado"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="ativoSelecionado ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <div class="flex gap-2.5 pt-2">
          <Button
            variant="secondary"
            class="flex-1 font-bold uppercase tracking-widest text-xs h-12"
            @click="cancelarEdicao"
          >
            Cancelar
          </Button>
          <Button
            variant="primary"
            :disabled="salvando"
            :loading="salvando"
            class="flex-1 font-bold uppercase tracking-widest text-xs h-12"
            @click="handleSalvarEdicao"
          >
            Salvar
          </Button>
        </div>
      </div>
    </div>
    <div
      v-else
      class="space-y-6"
    >
      <div class="bg-white border border-stone/30 rounded-2xl shadow-subtle overflow-hidden">
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-heading-sm text-charcoal flex items-center gap-2">
            <User class="w-5 h-5 text-ember" />
            Quem mora aqui
          </h3>
          <p class="text-xs text-ash font-medium mt-1 uppercase tracking-wider">
            Lista de membros ativos e suspensos
          </p>
        </div>

        <div class="p-4 space-y-4">
          <div
            v-if="podeGerenciarMoradores"
            class="flex justify-center pb-2"
          >
            <Button
              variant="secondary"
              class="w-full"
              @click="abrirNovoMembroForm"
            >
              <Plus class="w-5 h-5 mr-2" />
              Novo Morador
            </Button>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-2">
            <MembroListItem 
              v-for="(membro, idx) in membros" 
              :key="membro.id"
              :membro="membro"
              :variant="variants[idx % variants.length]"
              :clickable="podeGerenciarMoradores"
              @click="abrirEdicaoMembro(membro)"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
