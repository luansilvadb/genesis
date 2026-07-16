<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useLoginViewModel } from '../../viewmodels/useLoginViewModel'
import { tenantSessionService } from '../../shared/container'
import IllustrationMascot from '../components/ui/IllustrationMascot.vue'
import { useAsync } from '../../composables/useAsync'
import MembroAvatar from '../components/ui/MembroAvatar.vue'
import { logger } from '../../shared/utils/logger'
import type { InvitePreview } from '../../models/services/TenantSessionService'

const emit = defineEmits<{
  (e: 'auth-success'): void
  (e: 'forgot-password'): void
}>()

const {
  email,
  nome,
  password,
  inviteCode,
  membroId,
  errorMsg,
  handleLogin,
  handleRegister,
  handleGoogleLogin
} = useLoginViewModel()

const isRegisterMode = ref(false)
const { loading, errorMsg: errorMsgAsync, run } = useAsync()
const housePreview = ref<InvitePreview | null>(null)
const selectedMembroNome = ref('')
const showPassword = ref(false)
const googleLoginActive = ref(false)
const googleInicializado = ref(false)

const inicializarGoogleSignIn = () => {
  if (googleInicializado.value) return
  logger.info('Inicializando Google Sign-In. Origem:', window.location.origin)
  const google = (window as unknown as { google: { accounts: { id: { initialize: (config: Record<string, unknown>) => void; renderButton: (element: HTMLElement, options: Record<string, unknown>) => void } } } }).google
  if (typeof google === 'undefined' || !google.accounts) {
    logger.error('Google SDK não foi carregado corretamente.')
    return
  }

  const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID as string
  if (!clientId) {
    logger.error('VITE_GOOGLE_CLIENT_ID não configurado')
    return
  }

  google.accounts.id.initialize({
    client_id: clientId,
    callback: async (response: any) => {
      googleLoginActive.value = true
      errorMsg.value = ''
      try {
        const result = await handleGoogleLogin(response.credential)
        if (result) {
          emit('auth-success')
        }
        // Se falhar, o errorMsg já foi definido dentro de handleGoogleLogin.
        // Não sobrescrevemos com errorMsgAsync vazio, preservando a mensagem do viewmodel.
      } catch (err) {
        // Se handleGoogleLogin lançar exceção, capturamos e exibimos
        errorMsg.value = 'Ocorreu um erro inesperado ao autenticar com o Google.'
        logger.error('Erro no callback Google:', err)
      } finally {
        googleLoginActive.value = false
      }
    },
    auto_select: false,
    itp_support: false,
  })

  const btnContainer = document.getElementById('google-signin-btn')
  if (btnContainer) {
    google.accounts.id.renderButton(btnContainer, {
      type: 'standard',
      theme: 'outline',
      size: 'large',
      text: 'continue_with',
      shape: 'pill',
      logo_alignment: 'left',
      width: btnContainer.clientWidth || 320
    })
  }
  googleInicializado.value = true
}

onMounted(async () => {
  // Carrega SDK do Google
  if (!document.getElementById('google-jssdk')) {
    const script = document.createElement('script')
    script.id = 'google-jssdk'
    script.src = 'https://accounts.google.com/gsi/client'
    script.async = true
    script.defer = true
    script.onload = () => inicializarGoogleSignIn()
    document.head.appendChild(script)
  } else {
    setTimeout(() => inicializarGoogleSignIn(), 100)
  }

  const params = new URLSearchParams(window.location.search)
  const code = params.get('invite')
  if (code) {
    inviteCode.value = code
    isRegisterMode.value = true
    try {
      housePreview.value = await tenantSessionService.getInvitePreview(code)
    } catch (e) {
      logger.error('Falha ao carregar preview do convite:', e)
    }
  }
})

const selectMembro = (membro: InvitePreview['membrosDisponiveis'][number]) => {
  membroId.value = membro.id
  selectedMembroNome.value = membro.nome
  nome.value = membro.nome // Pré-preenche o nome
  isRegisterMode.value = true
}

const onSubmit = async () => {
  const success = await run(
    () => isRegisterMode.value ? handleRegister() : handleLogin(),
    'Ocorreu um erro inesperado'
  )

  if (success) {
    emit('auth-success')
  } else if (errorMsgAsync.value) {
    errorMsg.value = errorMsgAsync.value
  }
}
</script>

<template>
  <div class="min-h-screen bg-canvas flex items-center justify-center px-4 py-12">
    <!-- Card Container -->
    <div class="w-full max-w-[420px] bg-card rounded-2xl shadow-subtle p-8 sm:p-10 transition-all duration-300">
      <!-- Brand Logo / Header -->
      <div class="text-center mb-8 relative">
        <div class="inline-flex justify-center mb-4 transform hover:rotate-12 transition-transform duration-300 pointer-events-none">
          <IllustrationMascot 
            variant="ember" 
            :size="48" 
            :mood="loading ? 'surprised' : (isRegisterMode ? 'excited' : 'happy')" 
            class="animate-wobble" 
          />
        </div>
        <h1 class="text-display text-4xl mb-2">
          DIVI<span class="text-ember">.</span>
        </h1>
        
        <!-- Invite Context -->
        <div
          v-if="housePreview && !membroId"
          class="mt-8 p-6 bg-parchment rounded-card shadow-subtle border-none animate-in fade-in slide-in-from-top-2 duration-200"
        >
          <p class="text-caption font-semibold text-ember uppercase tracking-widest mb-1">
            Você foi convidado para
          </p>
          <h2 class="text-heading text-charcoal mb-6">
            {{ housePreview.name }}
          </h2>
          
          <p class="text-caption font-semibold text-graphite uppercase tracking-widest mb-3">
            Quem é você na casa?
          </p>
          <div class="grid grid-cols-2 gap-3">
            <button 
              v-for="m in housePreview.membrosDisponiveis" 
              :key="m.id"
              :aria-label="'Entrar como ' + m.nome"
              class="flex flex-col items-center gap-3 p-4 rounded-xl bg-card shadow-subtle hover:scale-[1.02] active:scale-95 transition-all duration-300 text-center group border-none cursor-pointer"
              @click="selectMembro(m)"
            >
              <MembroAvatar
                :nome="m.nome"
                size="md"
                variant="sky"
                class="group-hover:scale-110 transition-transform duration-300"
              />
              <span class="text-[11px] font-bold text-charcoal truncate uppercase tracking-widest">{{ m.nome }}</span>
            </button>
          </div>
        </div>

        <div
          v-else-if="housePreview && membroId && membroId !== 'novo'"
          class="mt-8 p-6 bg-parchment rounded-card shadow-subtle border-none animate-in zoom-in-95 duration-200"
        >
          <p class="text-body text-graphite">
            Criando acesso para <br>
            <span class="text-heading text-charcoal">{{ selectedMembroNome }}</span><br>
            <span class="text-caption text-graphite font-semibold uppercase tracking-widest">na casa {{ housePreview.name }}</span>
          </p>
          <button
            class="mt-4 text-caption font-semibold text-ember hover:opacity-80 transition-opacity uppercase tracking-widest bg-transparent border-none cursor-pointer"
            @click="membroId = ''"
          >
            Alterar perfil
          </button>
        </div>

        <p
          v-else
          class="text-body text-graphite max-w-[280px] mx-auto mt-2"
        >
          Organize despesas compartilhadas, feche o mês e entenda os acertos da casa
        </p>
      </div>

      <!-- Form (Show only if not choosing member or if member already selected) -->
      <Transition
        name="fade-slide"
        mode="out-in"
      >
        <form
          v-if="!housePreview || membroId || !housePreview.membrosDisponiveis?.length"
          class="space-y-6 pt-4 border-t border-stone mt-8"
          @submit.prevent="onSubmit"
        >
          <!-- Error Notification -->
          <Transition name="fade">
            <div 
              v-if="errorMsg" 
              role="alert"
              aria-live="assertive"
              class="bg-coral/10 text-coral text-caption px-4 py-3 rounded-card flex items-center gap-2 font-semibold"
            >
              <span aria-hidden="true">⚠️</span>
              <span>{{ errorMsg }}</span>
            </div>
          </Transition>

          <!-- Context Heading for Register -->
          <div
            v-if="isRegisterMode && housePreview"
            class="space-y-1"
          >
            <h3 class="text-caption font-semibold text-graphite uppercase tracking-widest">
              Configurar Acesso
            </h3>
            <p class="text-xs text-graphite/70">
              Escolha um e-mail, nome e senha para entrar.
            </p>
          </div>

          <!-- Nome Input (Apenas Cadastro) -->
          <div
            v-if="isRegisterMode"
            class="space-y-2 fade-in slide-in-from-top-2"
          >
            <label
              for="nome"
              class="block text-caption font-semibold text-charcoal uppercase tracking-widest ml-1"
            >
              Nome de Exibição
            </label>
            <input
              id="nome"
              v-model="nome"
              type="text"
              required
              tabindex="1"
              placeholder="Como quer ser chamado"
              autocomplete="name"
              class="w-full bg-canvas border border-stone rounded-card px-4 py-3.5 text-body text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all duration-200"
            >
          </div>

          <!-- Email Input -->
          <div class="space-y-2">
            <label
              for="email"
              class="block text-caption font-semibold text-charcoal uppercase tracking-widest ml-1"
            >
              E-mail
            </label>
            <input
              id="email"
              v-model="email"
              type="email"
              required
              tabindex="2"
              placeholder="seu@email.com"
              autocomplete="email"
              class="w-full bg-canvas border border-stone rounded-card px-4 py-3.5 text-body text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all duration-200"
            >
          </div>

          <!-- Password Input -->
          <div class="space-y-2">
            <div class="flex items-center justify-between ml-1">
              <label
                for="password"
                class="block text-caption font-semibold text-charcoal uppercase tracking-widest"
              >
                Senha
              </label>
              <button 
                v-if="!isRegisterMode"
                type="button" 
                tabindex="5"
                class="text-[10px] font-semibold text-ember hover:opacity-80 transition-opacity uppercase tracking-widest bg-transparent border-none cursor-pointer focus:outline-none"
                @click="emit('forgot-password')"
              >
                Esqueci a senha
              </button>
            </div>
            <div class="relative">
              <input
                id="password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                required
                tabindex="3"
                placeholder="••••••••"
                :autocomplete="isRegisterMode ? 'new-password' : 'current-password'"
                class="w-full bg-canvas border border-stone rounded-card pl-4 pr-12 py-3.5 text-body text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all duration-200"
              >
              <button
                type="button"
                tabindex="4"
                class="absolute inset-y-0 right-0 pr-4 flex items-center text-graphite hover:text-charcoal focus:outline-none transition-colors border-none bg-transparent cursor-pointer"
                :aria-label="showPassword ? 'Ocultar senha' : 'Mostrar senha'"
                @click="showPassword = !showPassword"
              >
                <svg
                  v-if="showPassword"
                  xmlns="http://www.w3.org/2000/svg"
                  class="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"
                  />
                </svg>
                <svg
                  v-else
                  xmlns="http://www.w3.org/2000/svg"
                  class="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                  />
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.543 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                  />
                </svg>
              </button>
            </div>
          </div>

          <!-- Submit Button (Midnight Pill) -->
          <button
            type="submit"
            :disabled="loading"
            :aria-busy="loading"
            tabindex="6"
            :aria-label="loading ? (isRegisterMode ? 'Criando conta...' : 'Entrando...') : (isRegisterMode ? 'Criar Conta e Entrar' : 'Entrar')"
            class="w-full bg-midnight hover:bg-charcoal text-white font-semibold py-4 px-6 rounded-pill text-sm tracking-widest uppercase transition-all duration-300 active:scale-95 disabled:opacity-50 disabled:scale-100 flex items-center justify-center gap-2 border-none cursor-pointer"
          >
            <span
              v-if="loading"
              class="animate-spin inline-block w-4 h-4 border-2 border-white/30 border-t-white rounded-full"
              aria-hidden="true"
            />
            <span>{{ isRegisterMode ? 'Criar Conta e Entrar' : 'Entrar' }}</span>
          </button>

          <!-- Divisor "OU" -->
          <div class="flex items-center my-4">
            <div class="flex-grow border-t border-stone" />
            <span class="px-3 text-caption text-ash uppercase font-semibold tracking-wider text-[10px]">Ou</span>
            <div class="flex-grow border-t border-stone" />
          </div>

          <!-- Google Sign-In Button Container -->
          <div class="w-full flex flex-col items-center gap-3">
            <div
              v-if="googleLoginActive"
              class="flex items-center gap-2 text-sm text-graphite font-medium py-2"
            >
              <span
                class="animate-spin inline-block w-4 h-4 border-2 border-ember/30 border-t-ember rounded-full"
                aria-hidden="true"
              />
              <span>Conectando com Google...</span>
            </div>
            <div
              v-show="!googleLoginActive"
              id="google-signin-btn"
              class="w-full min-h-[44px]"
            />
          </div>
        </form>
      </Transition>

      <!-- Toggle Mode -->
      <div class="mt-8 pt-6 border-t border-stone text-center">
        <p class="text-caption text-graphite font-semibold">
          {{ isRegisterMode ? 'Já possui uma conta?' : 'Novo no DIVI?' }}
          <button
            type="button"
            tabindex="7"
            class="ml-1 text-ember hover:opacity-80 font-bold focus:outline-none uppercase tracking-widest text-[10px] bg-transparent border-none cursor-pointer"
            @click="isRegisterMode = !isRegisterMode; errorMsg = ''; membroId = ''"
          >
            {{ isRegisterMode ? 'Faça login' : 'Crie sua conta' }}
          </button>
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.fade-slide-enter-active {
  transition: all 0.18s ease-out;
}
.fade-slide-enter-from {
  opacity: 0;
  transform: translateY(4px);
}
</style>
