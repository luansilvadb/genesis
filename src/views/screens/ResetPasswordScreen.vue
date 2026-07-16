<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { tenantSessionService } from '../../shared/container'
import { validatePassword } from '../../shared/utils/validatePassword'
import IllustrationMascot from '../components/ui/IllustrationMascot.vue'
import { useAsync } from '../../composables/useAsync'

const props = defineProps<{
  token: string
}>()

const emit = defineEmits(['reset-success'])
const router = useRouter()

const newPassword = ref('')
const { loading, errorMsg, run } = useAsync()
const showPassword = ref(false)

// Remove o token da URL imediatamente ao carregar a pagina,
// antes mesmo do usuario interagir. Isso evita que o token
// fique visivel no historico do navegador ou em screenshots.
onMounted(() => {
  if (props.token) {
    router.replace({ query: {} })
  }
})

const onSubmit = async () => {
  const passwordError = validatePassword(newPassword.value)
  if (passwordError) {
    errorMsg.value = passwordError
    return
  }

  const success = await run(
    () => tenantSessionService.resetPassword(props.token, newPassword.value),
    'Erro inesperado'
  )

  if (success) {
    // Remover token da URL
    router.replace({ query: {} })
    emit('reset-success')
  } else if (success === false) {
    errorMsg.value = 'O link é inválido ou expirou. Solicite um novo link de recuperação.'
  }
}
</script>

<template>
  <div class="min-h-screen bg-canvas flex items-center justify-center px-4 py-12">
    <div class="w-full max-w-[420px] bg-card rounded-2xl shadow-subtle p-8 sm:p-10 transition-all duration-300">
      <div class="text-center mb-8 relative">
        <div class="inline-flex justify-center mb-4 transform hover:rotate-12 transition-transform duration-300 pointer-events-none">
          <IllustrationMascot
            variant="ember"
            :size="48"
            mood="happy"
            class="animate-wobble"
          />
        </div>
        <h1 class="text-display text-3xl mb-2">
          Nova Senha<span class="text-ember">.</span>
        </h1>
        <p class="text-body text-graphite mt-2">
          Crie uma nova senha segura para o seu acesso.
        </p>
      </div>

      <form
        class="space-y-6"
        @submit.prevent="onSubmit"
      >
        <Transition name="fade">
          <div
            v-if="errorMsg"
            class="bg-coral/10 text-coral text-caption px-4 py-3 rounded-card flex items-center gap-2 font-semibold"
          >
            ⚠️ {{ errorMsg }}
          </div>
        </Transition>

        <div class="space-y-2">
          <label
            for="password"
            class="block text-caption font-semibold text-charcoal uppercase tracking-widest ml-1"
          >
            Nova Senha
          </label>
          <div class="relative">
            <input
              id="password"
              v-model="newPassword"
              :type="showPassword ? 'text' : 'password'"
              required
              placeholder="••••••••"
              class="w-full bg-canvas border border-stone rounded-card pl-4 pr-12 py-3.5 text-body text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all duration-200"
            >
            <button
              type="button"
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

        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-midnight hover:bg-charcoal text-white font-semibold py-4 px-6 rounded-pill text-sm tracking-widest uppercase transition-all duration-300 flex items-center justify-center gap-2 cursor-pointer border-none"
        >
          <span
            v-if="loading"
            class="animate-spin inline-block w-4 h-4 border-2 border-white/30 border-t-white rounded-full"
          />
          <span>Redefinir Senha</span>
        </button>
      </form>
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
</style>
