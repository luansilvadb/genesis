<!-- src/views/components/ui/TenantSwitcherModal.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Home, Check, Plus, Key, ChevronRight, ChevronLeft, ArrowRight, Loader2, Copy } from 'lucide-vue-next'
import { logger } from '../../../shared/utils/logger'
import { tenantSessionService } from '../../../shared/container'
import type { TenantSummary } from '../../../models/services/TenantSessionService'

const emit = defineEmits<{
  'casa-selecionada': []
}>()

const tenants = ref<TenantSummary[]>([])
const activeTenantId = ref<string | null>(null)

// Estados de controle interno
const modo = ref<'lista' | 'criar' | 'entrar'>('lista')
const nomeCasa = ref('')
const codigoConvite = ref('')
const loading = ref(false)
const errorMsg = ref('')
const casaCriada = ref<TenantSummary | null>(null)
const copiedCode = ref<string | null>(null)

const copiarCodigo = async (codigo: string) => {
  try {
    await navigator.clipboard.writeText(codigo)
    copiedCode.value = codigo
    setTimeout(() => {
      if (copiedCode.value === codigo) {
        copiedCode.value = null
      }
    }, 2000)
  } catch (err) {
    logger.error('Falha ao copiar:', err)
  }
}

onMounted(() => {
  carregarCasas()
})

const carregarCasas = () => {
  tenants.value = tenantSessionService.getTenants()
  activeTenantId.value = tenantSessionService.getActiveTenantId()
}

const selecionarCasa = (id: string) => {
  if (id === activeTenantId.value) {
    emit('casa-selecionada') // apenas fecha o modal
    return
  }
  tenantSessionService.setActiveTenant(id)
  window.dispatchEvent(new CustomEvent('divi:tenant-changed'))
  emit('casa-selecionada')
}

const irParaCriar = () => {
  modo.value = 'criar'
  errorMsg.value = ''
  casaCriada.value = null
  nomeCasa.value = ''
}

const irParaEntrar = () => {
  modo.value = 'entrar'
  errorMsg.value = ''
  codigoConvite.value = ''
}

const voltar = () => {
  modo.value = 'lista'
  errorMsg.value = ''
  casaCriada.value = null
  nomeCasa.value = ''
  codigoConvite.value = ''
  carregarCasas()
}

const criarCasa = async () => {
  if (!nomeCasa.value.trim()) {
    errorMsg.value = 'Dê um nome para a sua casa'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    const tenant = await tenantSessionService.criarCasa(nomeCasa.value.trim())
    casaCriada.value = tenant
  } catch (err: unknown) {
    logger.error(err instanceof Error ? err.message : String(err))
    errorMsg.value = err instanceof Error ? err.message : 'Não foi possível criar a casa'
  } finally {
    loading.value = false
  }
}

const entrarCasa = async () => {
  if (!codigoConvite.value.trim()) {
    errorMsg.value = 'Digite o código de convite'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    await tenantSessionService.entrarCasa(codigoConvite.value.trim())
    window.dispatchEvent(new CustomEvent('divi:tenant-changed'))
    emit('casa-selecionada')
  } catch (err: unknown) {
    logger.error(err instanceof Error ? err.message : String(err))
    errorMsg.value = err instanceof Error ? err.message : 'Código inválido ou casa não encontrada'
  } finally {
    loading.value = false
  }
}

const selecionarNovaCasaCriada = () => {
  window.dispatchEvent(new CustomEvent('divi:tenant-changed'))
  emit('casa-selecionada')
}
</script>

<template>
  <div class="p-6">
    <Transition
      name="fade"
      mode="out-in"
    >
      <!-- 1. MODO: LISTA DE CASAS -->
      <div
        v-if="modo === 'lista'"
        key="lista"
      >
        <div class="mb-6">
          <h2 class="text-xl font-bold text-charcoal tracking-tight">
            Minhas Casas
          </h2>
          <p class="text-xs text-graphite font-semibold mt-1">
            Selecione o espaço que deseja visualizar
          </p>
        </div>

        <div class="space-y-3 mb-8 max-h-[300px] overflow-y-auto custom-scrollbar pr-1">
          <div
            v-for="tenant in tenants"
            :key="tenant.id"
            class="w-full flex items-center justify-between p-4 rounded-2xl border-2 transition-all duration-200 cursor-pointer text-left group"
            :class="activeTenantId === tenant.id ? 'border-ember bg-ember/5' : 'border-stone bg-canvas hover:border-ember/30'"
            @click="selecionarCasa(tenant.id)"
          >
            <div class="flex items-center gap-4">
              <div 
                class="w-10 h-10 rounded-xl flex items-center justify-center transition-colors shrink-0"
                :class="activeTenantId === tenant.id ? 'bg-ember text-white' : 'bg-stone text-charcoal group-hover:bg-ember/10 group-hover:text-ember'"
              >
                <Home class="w-5 h-5" />
              </div>
              <div class="flex flex-col">
                <span class="font-bold text-charcoal text-sm leading-snug">{{ tenant.name }}</span>
                <!-- Código de convite e cópia -->
                <div
                  class="flex items-center gap-1.5 mt-1"
                  @click.stop
                >
                  <span class="text-[9px] text-ash font-bold uppercase tracking-wider">Código:</span>
                  <code class="text-[10px] bg-stone/60 px-1.5 py-0.5 rounded text-charcoal font-mono select-all">
                    {{ tenant.inviteCode }}
                  </code>
                  <button 
                    class="p-1 hover:bg-stone/80 rounded transition-colors border-none bg-transparent cursor-pointer flex items-center justify-center" 
                    :title="'Copiar código de ' + tenant.name"
                    @click="copiarCodigo(tenant.inviteCode)"
                  >
                    <Check
                      v-if="copiedCode === tenant.inviteCode"
                      class="w-3.5 h-3.5 text-meadow"
                    />
                    <Copy
                      v-else
                      class="w-3.5 h-3.5 text-ash group-hover:text-charcoal"
                    />
                  </button>
                </div>
              </div>
            </div>
            <Check
              v-if="activeTenantId === tenant.id"
              class="w-5 h-5 text-ember shrink-0"
            />
          </div>
        </div>

        <div class="space-y-3 border-t border-stone pt-6">
          <button
            class="w-full flex items-center gap-4 p-3 rounded-xl hover:bg-stone transition-colors cursor-pointer border-none bg-transparent text-left group"
            @click="irParaCriar"
          >
            <div class="w-8 h-8 rounded-lg bg-white shadow-subtle flex items-center justify-center text-ember group-hover:scale-110 transition-transform">
              <Plus class="w-4 h-4" />
            </div>
            <div class="flex-1">
              <span class="font-bold text-charcoal text-sm">Criar uma casa nova</span>
            </div>
            <ChevronRight class="w-4 h-4 text-ash" />
          </button>
          
          <button
            class="w-full flex items-center gap-4 p-3 rounded-xl hover:bg-stone transition-colors cursor-pointer border-none bg-transparent text-left group"
            @click="irParaEntrar"
          >
            <div class="w-8 h-8 rounded-lg bg-white shadow-subtle flex items-center justify-center text-charcoal group-hover:scale-110 transition-transform">
              <Key class="w-4 h-4" />
            </div>
            <div class="flex-1">
              <span class="font-bold text-charcoal text-sm">Entrar em uma casa</span>
            </div>
            <ChevronRight class="w-4 h-4 text-ash" />
          </button>
        </div>
      </div>

      <!-- 2. MODO: CRIAR CASA -->
      <div
        v-else-if="modo === 'criar'"
        key="criar"
      >
        <!-- Caso de Sucesso após criação -->
        <div
          v-if="casaCriada"
          class="text-center py-4"
        >
          <div class="mb-6">
            <div class="w-16 h-16 bg-meadow/10 rounded-full flex items-center justify-center mx-auto mb-3 border border-meadow/20">
              <Check class="w-8 h-8 text-meadow" />
            </div>
            <h3 class="text-xl font-bold text-charcoal tracking-tight">
              Casa criada! 🏡
            </h3>
            <p class="text-xs text-graphite mt-1">
              <strong class="text-charcoal font-bold">{{ casaCriada.name }}</strong> está pronta.
            </p>
          </div>

          <div class="bg-parchment shadow-subtle rounded-2xl p-5 mb-6">
            <p class="text-[9px] text-graphite mb-1.5 uppercase tracking-widest font-bold">
              Código de convite
            </p>
            <p class="text-2xl font-bold text-ember tracking-[0.2em] font-mono select-all">
              {{ casaCriada.inviteCode }}
            </p>
            <p class="text-[10px] text-ash mt-3 leading-relaxed font-medium">
              Compartilhe este código com as pessoas <br>que vão morar com você.
            </p>
          </div>

          <button
            class="w-full bg-charcoal hover:bg-midnight text-white font-bold py-3.5 px-6 rounded-pill text-xs tracking-widest uppercase transition-all duration-300 shadow-md flex items-center justify-center gap-2 border-none cursor-pointer active:scale-95"
            @click="selecionarNovaCasaCriada"
          >
            Acessar Nova Casa
            <ArrowRight class="w-4 h-4" />
          </button>
        </div>

        <!-- Formulário para Criar -->
        <div v-else>
          <header class="flex items-center gap-4 mb-6">
            <button
              class="w-10 h-10 rounded-full bg-stone hover:bg-ash/20 flex items-center justify-center text-charcoal transition-colors border-none cursor-pointer"
              @click="voltar"
            >
              <ChevronLeft class="w-5 h-5" />
            </button>
            <div>
              <h2 class="text-lg font-bold text-charcoal tracking-tight leading-none">
                Nova Casa
              </h2>
              <p class="text-xs text-graphite font-semibold mt-1">
                Dê um nome para a sua casa
              </p>
            </div>
          </header>

          <div class="space-y-5">
            <div class="space-y-2">
              <label
                for="modal-nome-casa"
                class="block text-[10px] font-bold text-charcoal uppercase tracking-widest ml-1"
              >
                Nome da Casa
              </label>
              <div class="relative">
                <input
                  id="modal-nome-casa"
                  v-model="nomeCasa"
                  type="text"
                  placeholder="Ex: Casa da Família Silva"
                  maxlength="60"
                  autofocus
                  class="w-full bg-canvas border border-stone rounded-xl px-4 py-3 text-sm text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all duration-200"
                  @keydown.enter="criarCasa"
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
                class="bg-coral/10 text-coral text-[10px] px-4 py-2.5 rounded-lg flex items-center gap-2 font-semibold"
              >
                <span>⚠️</span>
                <span>{{ errorMsg }}</span>
              </div>
            </Transition>

            <button
              :disabled="loading || !nomeCasa.trim()"
              class="w-full bg-ember hover:opacity-90 disabled:opacity-50 text-white font-bold py-3.5 px-6 rounded-pill text-xs tracking-widest uppercase transition-all duration-300 shadow-md flex items-center justify-center gap-2 border-none cursor-pointer active:scale-95"
              @click="criarCasa"
            >
              <Loader2
                v-if="loading"
                class="w-4 h-4 animate-spin"
              />
              <Home
                v-else
                class="w-4 h-4"
              />
              Criar Casa
            </button>
          </div>
        </div>
      </div>

      <!-- 3. MODO: ENTRAR EM UMA CASA -->
      <div
        v-else-if="modo === 'entrar'"
        key="entrar"
      >
        <header class="flex items-center gap-4 mb-6">
          <button
            class="w-10 h-10 rounded-full bg-stone hover:bg-ash/20 flex items-center justify-center text-charcoal transition-colors border-none cursor-pointer"
            @click="voltar"
          >
            <ChevronLeft class="w-5 h-5" />
          </button>
          <div>
            <h2 class="text-lg font-bold text-charcoal tracking-tight leading-none">
              Entrar em Casa
            </h2>
            <p class="text-xs text-graphite font-semibold mt-1">
              Insira o código de convite
            </p>
          </div>
        </header>

        <div class="space-y-5">
          <div class="space-y-2">
            <label
              for="modal-codigo-convite"
              class="block text-[10px] font-bold text-charcoal uppercase tracking-widest ml-1"
            >
              Código de Convite
            </label>
            <input
              id="modal-codigo-convite"
              v-model="codigoConvite"
              type="text"
              placeholder="Ex: CASA-AB12C"
              autofocus
              class="w-full bg-canvas border border-stone rounded-xl px-4 py-3 text-base font-bold text-charcoal placeholder:text-ash focus:outline-none focus:border-charcoal font-mono uppercase tracking-[0.15em] transition-all duration-200 text-center"
              @keydown.enter="entrarCasa"
            >
          </div>

          <Transition name="fade">
            <div
              v-if="errorMsg"
              role="alert"
              class="bg-coral/10 text-coral text-[10px] px-4 py-2.5 rounded-lg flex items-center gap-2 font-semibold"
            >
              <span>⚠️</span>
              <span>{{ errorMsg }}</span>
            </div>
          </Transition>

          <button
            :disabled="loading || !codigoConvite.trim()"
            class="w-full bg-charcoal hover:bg-midnight disabled:opacity-50 text-white font-bold py-3.5 px-6 rounded-pill text-xs tracking-widest uppercase transition-all duration-300 shadow-md flex items-center justify-center gap-2 border-none cursor-pointer active:scale-95"
            @click="entrarCasa"
          >
            <Loader2
              v-if="loading"
              class="w-4 h-4 animate-spin"
            />
            <Key
              v-else
              class="w-4 h-4"
            />
            Entrar na Casa
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.fade-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
.fade-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
