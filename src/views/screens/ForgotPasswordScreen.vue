<script setup lang="ts">
import { ref } from 'vue'
import { tenantSessionService } from '../../shared/container'
import IllustrationMascot from '../components/ui/IllustrationMascot.vue'
import { useAsync } from '../../composables/useAsync'

const emit = defineEmits(['back'])
const email = ref('')
const message = ref('')
const { loading, errorMsg, run } = useAsync()

const onSubmit = async () => {
  message.value = ''
  
  const success = await run(
    () => tenantSessionService.forgotPassword(email.value),
    'Falha ao solicitar recuperação. Tente novamente mais tarde.'
  )

  if (success) {
    message.value = 'Se o e-mail estiver cadastrado, você receberá um link de recuperação em instantes.'
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
          Recuperar Senha<span class="text-ember">.</span>
        </h1>
        <p class="text-body text-graphite mt-2">
          Informe seu e-mail para receber um link de acesso.
        </p>
      </div>

      <Transition
        name="fade-slide"
        mode="out-in"
      >
        <div
          v-if="message"
          class="text-center space-y-6"
        >
          <div class="bg-sky/10 text-sky px-4 py-4 rounded-card text-sm font-semibold">
            ✅ {{ message }}
          </div>
          <button
            class="w-full bg-midnight hover:bg-charcoal text-white font-semibold py-4 px-6 rounded-pill text-sm tracking-widest uppercase transition-all duration-300 cursor-pointer border-none"
            @click="emit('back')"
          >
            Voltar ao Login
          </button>
        </div>

        <form
          v-else
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
              placeholder="seu@email.com"
              class="w-full bg-canvas border border-stone rounded-card px-4 py-3.5 text-body text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all duration-200"
            >
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
            <span>Enviar Link</span>
          </button>
        </form>
      </Transition>

      <div
        v-if="!message"
        class="mt-8 pt-6 border-t border-stone text-center"
      >
        <button
          type="button"
          class="text-caption font-semibold text-graphite hover:text-charcoal uppercase tracking-widest bg-transparent border-none cursor-pointer transition-colors duration-200"
          @click="emit('back')"
        >
          Voltar para o Login
        </button>
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
