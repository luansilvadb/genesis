<script setup lang="ts">
import { ref, nextTick, computed } from 'vue'
import { Edit2, Check, X } from 'lucide-vue-next'
import { useMembros } from '../../../viewmodels/useMembros'
import { useToast } from '../../../composables/useToast'
import { mensagemErro } from '../../../shared/utils/mensagemErro'
import MembroAvatar from '../ui/MembroAvatar.vue'
import Button from '../ui/Button.vue'
import ConfiguracoesCartoes from '../ledger/ConfiguracoesCartoes.vue'
import { aplicarMascaraBRLText, formatarCentavosParaBRL } from '../../../shared/utils/formatarMoeda'

const props = defineProps<{
  isModoFoco: boolean
}>()

const emit = defineEmits(['logout', 'focus-change'])

const { currentMembro, atualizarNomeMembro, atualizarRendaMembro, tenantPermissions } = useMembros()

const podeAlterarNome = computed(() => {
  const role = currentMembro.value?.role
  if (!role || role === 'ADMIN') return true
  const perms = tenantPermissions.value[role]
  const defaultAllow = role === 'MORADOR'
  return perms ? perms.ALLOW_ALTERAR_NOME : defaultAllow
})

const podeAlterarRenda = computed(() => {
  const role = currentMembro.value?.role
  if (!role || role === 'ADMIN') return true
  const perms = tenantPermissions.value[role]
  const defaultAllow = role === 'MORADOR'
  return perms ? perms.ALLOW_ALTERAR_RENDA : defaultAllow
})
const toast = useToast()

const editandoNome = ref(false)
const nomeEditado = ref('')
const salvandoNome = ref(false)
const inputNomeRef = ref<HTMLInputElement | null>(null)

const iniciarEdicaoNome = () => {
  nomeEditado.value = currentMembro.value?.nome || ''
  editandoNome.value = true
  nextTick(() => {
    inputNomeRef.value?.focus()
  })
}

const cancelarEdicaoNome = () => {
  editandoNome.value = false
  nomeEditado.value = ''
}

const handleSalvarNome = async () => {
  if (!currentMembro.value) return
  
  const nomeLimpo = nomeEditado.value.trim()
  if (!nomeLimpo || nomeLimpo.length < 2) {
    toast.show('O nome deve ter pelo menos 2 caracteres', 'error')
    return
  }
  
  salvandoNome.value = true
  try {
    await atualizarNomeMembro(currentMembro.value.id, nomeLimpo)
    toast.show('Nome atualizado com sucesso', 'success')
    editandoNome.value = false
  } catch (error: unknown) {
    toast.show(mensagemErro(error, 'Erro ao salvar nome'), 'error')
  } finally {
    salvandoNome.value = false
  }
}

const handleLogout = () => {
  emit('logout')
}

const handleCartaoFocusChange = (active: boolean) => {
  emit('focus-change', active)
}

const editandoRenda = ref(false)
const rendaEditadaText = ref('')
const salvandoRenda = ref(false)
const inputRendaRef = ref<HTMLInputElement | null>(null)

const iniciarEdicaoRenda = () => {
  rendaEditadaText.value = currentMembro.value?.rendaCentavos 
    ? formatarCentavosParaBRL(currentMembro.value.rendaCentavos, false)
    : ''
  editandoRenda.value = true
  nextTick(() => {
    inputRendaRef.value?.focus()
  })
}

const cancelarEdicaoRenda = () => {
  editandoRenda.value = false
  rendaEditadaText.value = ''
}

const handleSalvarRenda = async () => {
  if (!currentMembro.value) return
  
  salvandoRenda.value = true
  try {
    let novaRendaCentavos: number | undefined = undefined
    if (rendaEditadaText.value) {
      const cleanValue = rendaEditadaText.value.replace(/\./g, '').replace(',', '.')
      const floatVal = parseFloat(cleanValue)
      if (!isNaN(floatVal)) {
        novaRendaCentavos = Math.round(floatVal * 100)
      }
    }
    await atualizarRendaMembro(currentMembro.value.id, novaRendaCentavos)
    toast.show('Renda atualizada com sucesso', 'success')
    editandoRenda.value = false
  } catch (error: unknown) {
    toast.show(mensagemErro(error, 'Erro ao salvar renda'), 'error')
  } finally {
    salvandoRenda.value = false
  }
}

const handleRendaInput = (e: Event) => {
  const target = e.target as HTMLInputElement
  rendaEditadaText.value = aplicarMascaraBRLText(target.value)
}
</script>

<template>
  <div class="space-y-8 animate-in fade-in duration-300">
    <!-- Card de Perfil Pessoal -->
    <div v-if="!isModoFoco" class="bg-white border border-stone/30 rounded-2xl shadow-subtle p-6 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
      <div class="flex items-center gap-4 flex-1">
        <MembroAvatar 
          v-if="currentMembro" 
          :nome="currentMembro.nome" 
          variant="ember" 
          size="lg" 
        />
        <div class="flex flex-col flex-1 min-w-0 gap-2.5">
          <!-- Bloco de Nome -->
          <div v-if="!editandoNome" class="flex flex-col min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="text-xl font-bold text-charcoal font-sans tracking-tight truncate">{{ currentMembro?.nome }}</h3>
              <button 
                v-if="podeAlterarNome"
                @click="iniciarEdicaoNome" 
                class="p-1 text-ash hover:text-charcoal transition-colors border-none bg-transparent cursor-pointer flex items-center justify-center"
                aria-label="Editar nome"
              >
                <Edit2 class="w-3.5 h-3.5" />
              </button>
            </div>
            <p class="text-xs text-ash font-medium mt-0.5">@{{ currentMembro?.username || currentMembro?.nome?.toLowerCase() }}</p>
          </div>
          <div v-else class="flex flex-col gap-1.5 flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <input 
                v-model="nomeEditado" 
                type="text" 
                :disabled="salvandoNome"
                @keyup.enter="handleSalvarNome"
                @keyup.esc="cancelarEdicaoNome"
                class="flex-1 w-full max-w-[130px] xs:max-w-[170px] sm:max-w-[260px] px-3 py-1.5 rounded-xl border border-stone bg-white text-sm font-bold text-charcoal outline-none focus:border-ember focus:ring-1 focus:ring-ember transition-all"
                placeholder="Seu nome"
                ref="inputNomeRef"
              />
              <button 
                @click="handleSalvarNome" 
                :disabled="salvandoNome"
                class="p-2 bg-meadow/10 hover:bg-meadow/20 text-meadow rounded-xl border-none cursor-pointer transition-colors flex items-center justify-center disabled:opacity-50"
                aria-label="Salvar nome"
              >
                <Check class="w-4 h-4" />
              </button>
              <button 
                @click="cancelarEdicaoNome" 
                :disabled="salvandoNome"
                class="p-2 bg-stone/10 hover:bg-stone/20 text-ash rounded-xl border-none cursor-pointer transition-colors flex items-center justify-center disabled:opacity-50"
                aria-label="Cancelar edição"
              >
                <X class="w-4 h-4" />
              </button>
            </div>
            <p class="text-[10px] sm:text-[11px] text-ash font-medium ml-1">Pressione Enter para salvar, Esc para cancelar</p>
          </div>

          <!-- Bloco de Renda -->
          <div v-if="!editandoRenda" class="flex items-center gap-2">
            <span class="text-xs text-graphite/70 font-semibold uppercase tracking-wider">Renda:</span>
            <span class="text-xs font-bold text-charcoal">
              {{ currentMembro?.rendaCentavos ? formatarCentavosParaBRL(currentMembro.rendaCentavos) : 'Não informada' }}
            </span>
            <button 
              v-if="podeAlterarRenda"
              @click="iniciarEdicaoRenda" 
              class="p-1 text-ash hover:text-charcoal transition-colors border-none bg-transparent cursor-pointer flex items-center justify-center"
              aria-label="Editar renda"
            >
              <Edit2 class="w-3.5 h-3.5" />
            </button>
          </div>
          <div v-else class="flex flex-col gap-1.5 min-w-0">
            <div class="flex items-center gap-2">
              <input 
                v-model="rendaEditadaText" 
                type="text" 
                :disabled="salvandoRenda"
                @keyup.enter="handleSalvarRenda"
                @keyup.esc="cancelarEdicaoRenda"
                class="flex-1 w-full max-w-[110px] xs:max-w-[150px] sm:max-w-[200px] px-3 py-1.5 rounded-xl border border-stone bg-white text-sm font-bold text-charcoal outline-none focus:border-ember focus:ring-1 focus:ring-ember transition-all"
                placeholder="Ex: 3.500,00"
                @input="handleRendaInput"
                ref="inputRendaRef"
              />
              <button 
                @click="handleSalvarRenda" 
                :disabled="salvandoRenda"
                class="p-2 bg-meadow/10 hover:bg-meadow/20 text-meadow rounded-xl border-none cursor-pointer transition-colors flex items-center justify-center disabled:opacity-50"
                aria-label="Salvar renda"
              >
                <Check class="w-4 h-4" />
              </button>
              <button 
                @click="cancelarEdicaoRenda" 
                :disabled="salvandoRenda"
                class="p-2 bg-stone/10 hover:bg-stone/20 text-ash rounded-xl border-none cursor-pointer transition-colors flex items-center justify-center disabled:opacity-50"
                aria-label="Cancelar edição de renda"
              >
                <X class="w-4 h-4" />
              </button>
            </div>
            <p class="text-[10px] sm:text-[11px] text-ash font-medium ml-1">Pressione Enter para salvar, Esc para cancelar</p>
          </div>
        </div>
      </div>
      <Button 
        @click="handleLogout" 
        variant="secondary" 
        class="w-full md:w-auto text-xs font-bold uppercase tracking-widest h-10 px-5 transition-all duration-300 active:scale-95"
      >
        Sair da Conta
      </Button>
    </div>

    <!-- Componente de Cartões do Usuário -->
    <ConfiguracoesCartoes @focus-change="handleCartaoFocusChange" />
  </div>
</template>
