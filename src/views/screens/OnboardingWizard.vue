<script setup lang="ts">
import { ref } from 'vue'
import { useOnboardingViewModel } from '../../viewmodels/useOnboardingViewModel'
import { Plus, ArrowRight, Check, ChevronLeft, CreditCard, Trash2 } from 'lucide-vue-next'

const emit = defineEmits<{
  'concluido': []
  'cancelar': []
}>()

const vm = useOnboardingViewModel()
const {
  loading,
  errorMsg,
  etapaWizard,
  nomeCasa,
  contasSugeridas,
  cartoesCadastro,
  casaCriada,
  avancarPasso,
  adicionarContaCustomizada,
  adicionarCartaoLista,
  removerCartaoLista,
  finalizarWizard
} = vm

const novaContaNome = ref('')
const novaContaIcon = ref('💰')
const novaContaValor = ref('')
const mostrarFormNovaConta = ref(false)

function handleAdicionarContaCustomizada() {
  if (!novaContaNome.value.trim()) return
  adicionarContaCustomizada(novaContaNome.value, novaContaIcon.value, novaContaValor.value)
  novaContaNome.value = ''
  novaContaIcon.value = '💰'
  novaContaValor.value = ''
  mostrarFormNovaConta.value = false
}

const novoCartaoNome = ref('')
const novoCartaoDia = ref(10)
const mostrarFormNovoCartao = ref(false)

function handleAdicionarCartaoLista() {
  if (!novoCartaoNome.value.trim()) return
  adicionarCartaoLista(novoCartaoNome.value, novoCartaoDia.value)
  novoCartaoNome.value = ''
  novoCartaoDia.value = 10
  mostrarFormNovoCartao.value = false
}

function handleVoltar() {
  vm.voltar(() => emit('cancelar'))
}
</script>

<template>
  <Transition
    name="fade"
    mode="out-in"
  >
    <!-- PASSO 4: Sucesso (Casa Criada) -->
    <div
      v-if="etapaWizard === 4 && casaCriada"
      key="sucesso"
      class="text-center animate-in zoom-in-95 duration-200"
    >
      <div class="mb-6">
        <div class="w-16 h-16 bg-meadow/10 rounded-full flex items-center justify-center mx-auto mb-3 border border-meadow/20">
          <Check class="w-8 h-8 text-meadow" />
        </div>
        <h2 class="text-xl font-bold text-charcoal tracking-tight">
          Casa pronta! 🏡
        </h2>
        <p class="text-xs text-graphite mt-1">
          A moradia <strong class="text-charcoal font-bold">{{ casaCriada.name }}</strong> está pronta para registrar despesas e fechar o mês.
        </p>
      </div>

      <div class="bg-parchment shadow-subtle rounded-2xl p-5 mb-6">
        <p class="text-[9px] text-graphite mb-1.5 uppercase tracking-widest font-bold">
          Código de convite
        </p>
        <p class="text-2xl font-bold text-ember tracking-[0.2em] font-mono text-center select-all">
          {{ casaCriada.inviteCode }}
        </p>
        <p class="text-[10px] text-ash mt-3 leading-relaxed font-medium">
          Compartilhe este código com os outros moradores <br>para que eles entrem na casa.
        </p>
      </div>

      <button
        class="w-full bg-midnight hover:bg-charcoal text-white font-bold py-3.5 px-6 rounded-pill text-xs tracking-widest uppercase transition-all duration-300 shadow-md flex items-center justify-center gap-2 border-none cursor-pointer active:scale-95"
        @click="$emit('concluido')"
      >
        Ir para o Dashboard
        <ArrowRight class="w-4 h-4" />
      </button>
    </div>

    <!-- PASSO 1: Nome da Casa -->
    <div
      v-else-if="etapaWizard === 1"
      key="passo-nome"
      class="animate-in fade-in slide-in-from-right-2 duration-200"
    >
      <header class="flex items-center gap-4 mb-6">
        <button
          class="w-9 h-9 rounded-full bg-stone hover:bg-ash/20 flex items-center justify-center text-charcoal transition-colors border-none cursor-pointer"
          @click="handleVoltar"
        >
          <ChevronLeft class="w-4.5 h-4.5" />
        </button>
        <div>
          <h2 class="text-lg font-bold text-charcoal tracking-tight leading-none">
            Criar Nova Casa
          </h2>
          <p class="text-[10px] text-graphite font-semibold mt-1">
            Passo 1 de 3: Identidade
          </p>
        </div>
      </header>

      <div class="space-y-5">
        <div class="space-y-1.5">
          <label
            for="nome-casa"
            class="block text-[10px] font-bold text-charcoal uppercase tracking-widest ml-1"
          >
            Nome da Casa
          </label>
          <div class="relative">
            <input
              id="nome-casa"
              v-model="nomeCasa"
              type="text"
              placeholder="Ex: República Central, Casa Luan..."
              maxlength="60"
              autofocus
              class="w-full bg-canvas border border-stone rounded-xl px-4 py-3 text-sm font-bold text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all"
              @keydown.enter="avancarPasso"
            >
            <span
              v-if="nomeCasa.length > 0"
              class="absolute right-3 top-1/2 -translate-y-1/2 text-[9px] font-bold text-ash/60"
            >
              {{ nomeCasa.length }}/60
            </span>
          </div>
        </div>

        <Transition name="fade">
          <div
            v-if="errorMsg"
            role="alert"
            class="bg-coral/10 text-coral text-caption px-4 py-2.5 rounded-card flex items-center gap-2 font-semibold"
          >
            <span>⚠️</span>
            <span>{{ errorMsg }}</span>
          </div>
        </Transition>

        <button
          :disabled="!nomeCasa.trim()"
          class="w-full bg-ember hover:opacity-90 disabled:opacity-50 text-white font-bold py-3.5 px-6 rounded-pill text-xs tracking-widest uppercase transition-all duration-300 shadow-md flex items-center justify-center gap-2 border-none cursor-pointer active:scale-95"
          @click="avancarPasso"
        >
          Continuar
          <ArrowRight class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- PASSO 2: Parametrizar Contas Fixas -->
    <div
      v-else-if="etapaWizard === 2"
      key="passo-contas"
      class="animate-in fade-in slide-in-from-right-2 duration-200"
    >
      <header class="flex items-center gap-4 mb-5">
        <button
          class="w-9 h-9 rounded-full bg-stone hover:bg-ash/20 flex items-center justify-center text-charcoal transition-colors border-none cursor-pointer"
          @click="handleVoltar"
        >
          <ChevronLeft class="w-4.5 h-4.5" />
        </button>
        <div>
          <h2 class="text-lg font-bold text-charcoal tracking-tight leading-none">
            Contas da Casa
          </h2>
          <p class="text-[10px] text-graphite font-semibold mt-1">
            Passo 2 de 3: Contas Fixas
          </p>
        </div>
      </header>

      <div class="space-y-4">
        <p class="text-[11px] text-graphite font-medium leading-normal">
          Selecione as despesas fixas recorrentes da moradia. Deixe o valor em branco caso ele mude todo mês.
        </p>

        <!-- Grid de Contas -->
        <div class="space-y-2 max-h-[200px] overflow-y-auto pr-1">
          <div
            v-for="conta in contasSugeridas"
            :key="conta.id"
            class="flex items-center gap-3 p-3 rounded-2xl border transition-all cursor-pointer"
            :class="conta.selecionada ? 'bg-ember/5 border-ember/30 shadow-sm' : 'bg-parchment border-stone'"
            @click.self="conta.selecionada = !conta.selecionada"
          >
            <input
              v-model="conta.selecionada"
              type="checkbox"
              class="w-4 h-4 text-ember accent-ember cursor-pointer"
            >
            <span class="text-lg select-none">{{ conta.icon }}</span>
            <span
              class="flex-1 text-xs font-bold text-charcoal select-none"
              @click="conta.selecionada = !conta.selecionada"
            >
              {{ conta.name }}
            </span>
            <div
              v-if="conta.selecionada"
              class="w-24 flex items-center bg-canvas border border-stone rounded-lg px-2 py-1"
            >
              <span class="text-[9px] text-ash font-bold mr-1 select-none">R$</span>
              <input
                v-model="conta.valor"
                type="text"
                placeholder="Variável"
                class="w-full bg-transparent border-none text-[10px] font-bold text-charcoal focus:outline-none text-right"
              >
            </div>
          </div>
        </div>

        <!-- Formulário de Conta Customizada -->
        <div class="bg-parchment rounded-2xl border border-stone p-3.5 space-y-2.5">
          <button
            v-if="!mostrarFormNovaConta"
            class="w-full py-2 bg-white hover:bg-stone text-[10px] font-bold uppercase tracking-wider text-graphite rounded-xl border border-stone transition-all cursor-pointer flex items-center justify-center gap-1.5"
            @click="mostrarFormNovaConta = true"
          >
            <Plus class="w-3.5 h-3.5" /> Adicionar Outra Despesa
          </button>

          <div
            v-else
            class="space-y-2"
          >
            <div class="flex gap-2">
              <input
                v-model="novaContaIcon"
                type="text"
                placeholder="Emoji"
                class="w-12 bg-canvas border border-stone rounded-xl text-center py-2 text-sm focus:outline-none"
              >
              <input
                v-model="novaContaNome"
                type="text"
                placeholder="Nome (Ex: Netflix, Faxina...)"
                class="flex-1 bg-canvas border border-stone rounded-xl px-3 py-2 text-xs font-bold text-charcoal focus:outline-none"
              >
            </div>
            <div class="flex gap-2">
              <div class="flex-1 flex items-center bg-canvas border border-stone rounded-xl px-3 py-2">
                <span class="text-[10px] text-ash font-bold mr-1">R$</span>
                <input
                  v-model="novaContaValor"
                  type="text"
                  placeholder="Valor Estimado"
                  class="w-full bg-transparent border-none text-xs font-bold text-charcoal focus:outline-none"
                >
              </div>
              <button
                :disabled="!novaContaNome.trim()"
                class="bg-midnight hover:bg-charcoal text-white text-xs font-bold px-4 rounded-xl border-none cursor-pointer disabled:opacity-50"
                @click="handleAdicionarContaCustomizada"
              >
                Confirmar
              </button>
            </div>
          </div>
        </div>

        <button
          class="w-full bg-ember hover:opacity-90 text-white font-bold py-3.5 px-6 rounded-pill text-xs tracking-widest uppercase transition-all duration-300 shadow-md flex items-center justify-center gap-2 border-none cursor-pointer active:scale-95"
          @click="avancarPasso"
        >
          Continuar
          <ArrowRight class="w-4 h-4" />
        </button>
      </div>
    </div>

    <!-- PASSO 3: Cartões de Crédito -->
    <div
      v-else-if="etapaWizard === 3"
      key="passo-cartoes"
      class="animate-in fade-in slide-in-from-right-2 duration-200"
    >
      <header class="flex items-center gap-4 mb-5">
        <button
          class="w-9 h-9 rounded-full bg-stone hover:bg-ash/20 flex items-center justify-center text-charcoal transition-colors border-none cursor-pointer"
          @click="handleVoltar"
        >
          <ChevronLeft class="w-4.5 h-4.5" />
        </button>
        <div>
          <h2 class="text-lg font-bold text-charcoal tracking-tight leading-none">
            Cartões de Crédito
          </h2>
          <p class="text-[10px] text-graphite font-semibold mt-1">
            Passo 3 de 3: Cartões
          </p>
        </div>
      </header>

      <div class="space-y-4">
        <p class="text-[11px] text-graphite font-medium leading-normal">
          Cadastre os cartões coletivos que a casa usa para compras. Eles ficarão em seu nome para segurança, mas todos poderão lançar neles.
        </p>

        <!-- Lista de Cartões Cadastrados -->
        <div class="space-y-2 max-h-[160px] overflow-y-auto pr-1">
          <div
            v-for="(cartao, index) in cartoesCadastro"
            :key="index"
            class="flex items-center justify-between p-3 bg-parchment rounded-2xl border border-stone shadow-sm"
          >
            <div class="flex items-center gap-2.5">
              <CreditCard class="w-4 h-4 text-ember" />
              <div>
                <p class="text-xs font-bold text-charcoal leading-none">
                  {{ cartao.nome }}
                </p>
                <p class="text-[9px] text-graphite font-semibold mt-1">
                  Fechamento: Todo dia {{ cartao.diaFechamento }}
                </p>
              </div>
            </div>
            <button
              class="p-1.5 hover:text-coral text-ash hover:bg-stone/50 rounded-lg transition-colors bg-transparent border-none cursor-pointer"
              @click="removerCartaoLista(index)"
            >
              <Trash2 class="w-3.5 h-3.5" />
            </button>
          </div>

          <div
            v-if="cartoesCadastro.length === 0"
            class="text-center py-4 bg-canvas border border-dashed border-stone rounded-2xl"
          >
            <CreditCard class="w-6 h-6 text-ash/40 mx-auto mb-1" />
            <p class="text-[10px] text-ash font-bold">
              Nenhum cartão cadastrado
            </p>
          </div>
        </div>

        <!-- Formulário Novo Cartão -->
        <div class="bg-parchment rounded-2xl border border-stone p-3.5 space-y-2.5">
          <button
            v-if="!mostrarFormNovoCartao"
            class="w-full py-2 bg-white hover:bg-stone text-[10px] font-bold uppercase tracking-wider text-graphite rounded-xl border border-stone transition-all cursor-pointer flex items-center justify-center gap-1.5"
            @click="mostrarFormNovoCartao = true"
          >
            <Plus class="w-3.5 h-3.5" /> Adicionar Cartão
          </button>

          <div
            v-else
            class="space-y-2.5 animate-in fade-in duration-200"
          >
            <div class="space-y-1">
              <label class="block text-[9px] font-bold text-graphite uppercase tracking-widest ml-1">Nome do Cartão</label>
              <input
                v-model="novoCartaoNome"
                type="text"
                placeholder="Ex: Nubank Casa, Visa Luan..."
                class="w-full bg-canvas border border-stone rounded-xl px-3 py-2 text-xs font-bold text-charcoal focus:outline-none"
              >
            </div>
            <div class="flex gap-2.5 items-end">
              <div class="flex-1 space-y-1">
                <label class="block text-[9px] font-bold text-graphite uppercase tracking-widest ml-1">Dia do Fechamento</label>
                <input
                  v-model="novoCartaoDia"
                  type="number"
                  min="1"
                  max="31"
                  class="w-full bg-canvas border border-stone rounded-xl px-3 py-2 text-xs font-bold text-charcoal focus:outline-none text-center"
                  @blur="novoCartaoDia = Math.max(1, Math.min(31, Number(novoCartaoDia) || 1))"
                >
              </div>
              <div class="flex gap-1.5">
                <button
                  class="bg-stone text-charcoal text-xs font-bold py-2.5 px-3 rounded-xl border-none cursor-pointer"
                  @click="mostrarFormNovoCartao = false"
                >
                  Cancelar
                </button>
                <button
                  :disabled="!novoCartaoNome.trim() || !novoCartaoDia"
                  class="bg-midnight hover:bg-charcoal text-white text-xs font-bold py-2.5 px-4 rounded-xl border-none cursor-pointer disabled:opacity-50"
                  @click="handleAdicionarCartaoLista"
                >
                  Confirmar
                </button>
              </div>
            </div>
          </div>
        </div>

        <Transition name="fade">
          <div
            v-if="errorMsg"
            role="alert"
            class="bg-coral/10 text-coral text-caption px-4 py-2.5 rounded-card flex items-center gap-2 font-semibold"
          >
            <span>⚠️</span>
            <span>{{ errorMsg }}</span>
          </div>
        </Transition>

        <button
          :disabled="loading"
          class="w-full bg-midnight hover:bg-charcoal text-white font-bold py-3.5 px-6 rounded-pill text-xs tracking-widest uppercase transition-all duration-300 shadow-md flex items-center justify-center gap-2 border-none cursor-pointer active:scale-95"
          @click="finalizarWizard"
        >
          <span
            v-if="loading"
            class="animate-spin inline-block w-4 h-4 border-2 border-white/30 border-t-white rounded-full"
          />
          <Check
            v-else
            class="w-4 h-4"
          />
          Finalizar e Criar Casa
        </button>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
