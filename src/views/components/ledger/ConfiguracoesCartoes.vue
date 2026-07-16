<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useCartoesEFaturas } from '../../../viewmodels/useCartoesEFaturas'
import { useMembros } from '../../../viewmodels/useMembros'
import Button from '../ui/Button.vue'
import { Trash2, CreditCard, Calendar, ChevronDown, Plus, ArrowLeft } from 'lucide-vue-next'
import { useCasasMultitenant } from '../../../viewmodels/useCasasMultitenant'
import { useToast } from '../../../composables/useToast'
import { mensagemErro } from '../../../shared/utils/mensagemErro'
import IllustrationMascot from '../ui/IllustrationMascot.vue'
import { obterCorCartao } from '../../../shared/utils/obterCorCartao'

const emit = defineEmits<{
  (e: 'focus-change', active: boolean): void
}>()

const { activeTenantId } = useCasasMultitenant()
const { currentMembro, tenantPermissions } = useMembros()
const { cartoes, adicionarCartao, excluirCartao } = useCartoesEFaturas()
const toast = useToast()

const nome = ref('')
const diaFechamento = ref<number>(10)

const formularioAberto = ref(false)

const podeGerenciarCartoes = computed(() => {
  const role = currentMembro.value?.role
  if (!role || role === 'ADMIN') return true
  const perms = tenantPermissions.value[role]
  const defaultAllow = role === 'MORADOR'
  return perms ? perms.ALLOW_GERENCIAR_CARTOES : defaultAllow
})

watch(formularioAberto, (val) => {
  emit('focus-change', val)
})

const meusCartoes = computed(() => {
  if (!currentMembro.value) return []
  return cartoes.value.filter(c => c.responsavelPadraoId === currentMembro.value?.id)
})

const isSubmitting = ref(false)

const adicionarCard = async () => {
  const ownerId = currentMembro.value?.id
  if (!nome.value || !ownerId) return
  isSubmitting.value = true
  try {
    await adicionarCartao(nome.value, diaFechamento.value, ownerId)
    nome.value = ''
    diaFechamento.value = 10
    formularioAberto.value = false
  } catch (error: unknown) {
    toast.show(mensagemErro(error, 'Erro ao cadastrar cartão'), 'error')
  } finally {
    isSubmitting.value = false
  }
}

const handleExcluir = async (id: string) => {
  try {
    await excluirCartao(id)
  } catch (error: unknown) {
    toast.show(mensagemErro(error, 'Erro ao excluir cartão'), 'error')
  }
}
</script>

<template>
  <div class="space-y-6 w-full">
    <!-- Card de Adicionar Novo Cartão -->
    <div class="bg-white border border-stone/30 rounded-2xl shadow-subtle overflow-hidden">
      <!-- Visão da Listagem de Cartões -->
      <div
        v-if="!formularioAberto"
        class="animate-in fade-in duration-300"
      >
        <div class="px-6 pt-6 pb-2">
          <h3 class="text-heading-sm text-charcoal flex items-center gap-2">
            <CreditCard class="w-5 h-5 text-ember" />
            Meus Cartões
          </h3>
          <p class="text-[11px] text-ash font-medium mt-1 uppercase tracking-wider">
            Gestão de cartões pessoais
          </p>
        </div>

        <div class="p-4 space-y-4">
          <div
            v-if="podeGerenciarCartoes"
            class="flex justify-center pb-2"
          >
            <Button 
              variant="secondary"
              class="w-full"
              @click="formularioAberto = true"
            >
              <Plus class="w-5 h-5" />
              Novo Cartão
            </Button>
          </div>

          <!-- Lista de Cartões -->
          <div class="space-y-2">
            <div
              v-for="c in meusCartoes"
              :key="c.id"
              class="p-4 flex justify-between items-center bg-canvas/50 border border-stone/50 rounded-2xl hover:border-ember/40 transition-all duration-500 group"
            >
              <div class="flex items-center gap-4 min-w-0">
                <div 
                  class="w-10 h-10 rounded-xl shadow-subtle flex items-center justify-center shrink-0 border transition-all duration-500 group-hover:scale-110 group-hover:rotate-3"
                  :style="{ 
                    backgroundColor: obterCorCartao(c.nome) + '10', 
                    borderColor: obterCorCartao(c.nome) + '20' 
                  }"
                >
                  <CreditCard
                    class="w-5 h-5"
                    :style="{ color: obterCorCartao(c.nome) }"
                  />
                </div>
                <div class="min-w-0">
                  <span class="font-bold text-charcoal text-sm block truncate">{{ c.nome }}</span>
                  <p class="text-caption text-[9px] mt-1 text-ash">
                    Fecha dia {{ c.diaFechamento }}
                  </p>
                </div>
              </div>

              <button
                v-if="podeGerenciarCartoes"
                class="bg-coral/5 text-coral hover:bg-coral hover:text-white border-none rounded-full h-10 w-10 flex items-center justify-center transition-all duration-500 active:scale-90 cursor-pointer shrink-0"
                aria-label="Excluir cartão"
                @click="handleExcluir(c.id)"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>

            <div
              v-if="meusCartoes.length === 0"
              class="flex flex-col items-center justify-center py-10 px-6 text-center space-y-4 animate-in fade-in zoom-in-95 duration-700"
            >
              <IllustrationMascot
                variant="sunburst"
                :size="70"
                mood="happy"
                class="opacity-40"
              />
              <p class="text-xs text-ash font-bold italic max-w-[180px] leading-relaxed">
                Você ainda não cadastrou nenhum cartão próprio.
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Visão do Formulário Inline -->
      <div
        v-else
        class="animate-in fade-in slide-in-from-right-3 duration-300"
      >
        <div class="px-6 pt-6 pb-2 flex items-center gap-3">
          <button 
            type="button"
            class="w-10 h-10 rounded-full bg-white border border-stone/60 text-charcoal flex items-center justify-center cursor-pointer shadow-sm hover:scale-105 hover:text-ember hover:border-ash/50 active:scale-95 transition-all duration-300 ease-out focus:outline-none" 
            @click="formularioAberto = false"
          >
            <ArrowLeft class="w-5 h-5" />
          </button>
          <div>
            <h3 class="text-heading-sm text-charcoal flex items-center gap-2">
              Novo <span class="text-ember">Cartão</span>
            </h3>
            <p class="text-[10px] text-ash font-medium mt-0.5 uppercase tracking-wider">
              Adicione um novo cartão de crédito
            </p>
          </div>
        </div>

        <div class="p-6 space-y-6">
          <!-- Nome do Cartão -->
          <div class="space-y-2">
            <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Nome do Cartão</label>
            <input
              v-model="nome"
              type="text"
              placeholder="Ex: Nubank, C6, etc."
              class="w-full px-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal focus:border-ember transition-all text-sm"
              @keyup.enter="adicionarCard"
            >
          </div>

          <!-- Dia de Fechamento -->
          <div class="space-y-2">
            <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Dia de Fechamento</label>
            <div class="relative">
              <Calendar class="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-graphite pointer-events-none z-10 opacity-60" />
              <select
                v-model="diaFechamento"
                class="w-full pl-12 pr-10 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-charcoal transition-all text-sm appearance-none cursor-pointer shadow-subtle focus:border-ember focus:ring-4 focus:ring-ember/10"
              >
                <option
                  v-for="d in 31"
                  :key="d"
                  :value="d"
                >
                  Dia {{ d }}
                </option>
              </select>
              <ChevronDown
                class="absolute right-4 top-1/2 -translate-y-1/2 w-4 h-4 text-graphite pointer-events-none"
              />
            </div>
          </div>

          <!-- Aviso sem casa -->
          <div
            v-if="!activeTenantId"
            class="p-3 bg-coral/5 border border-coral/10 rounded-xl"
          >
            <p class="text-[10px] text-coral font-bold text-center uppercase tracking-widest">
              Selecione uma casa primeiro!
            </p>
          </div>

          <!-- Botões de Ação Inline -->
          <div class="flex gap-2.5 pt-2">
            <Button 
              variant="secondary"
              class="flex-1 h-12 text-[10px] font-bold uppercase tracking-widest"
              :disabled="isSubmitting"
              @click="formularioAberto = false"
            >
              Cancelar
            </Button>
            <Button
              variant="primary"
              class="flex-1 h-12 text-[10px] font-bold uppercase tracking-widest"
              :disabled="!nome || !currentMembro || !activeTenantId || isSubmitting"
              :loading="isSubmitting"
              @click="adicionarCard"
            >
              Cadastrar
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
