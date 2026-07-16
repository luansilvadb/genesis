<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, provide } from 'vue'
import { useRouter } from 'vue-router'
import { useMembros } from './viewmodels/useMembros'
import { useCartoesEFaturas } from './viewmodels/useCartoesEFaturas'
import { useContasFixas } from './viewmodels/useContasFixas'
import { tenantSessionService, socketService } from './shared/container'
import { useToast } from './composables/useToast'
import { logger } from './shared/utils/logger'
import ToastNotification from './views/components/ui/ToastNotification.vue'
import IllustrationMascot from './views/components/ui/IllustrationMascot.vue'

const router = useRouter()
const toast = useToast()

const isInitializing = ref(true)
const isLoading = ref(false)
const hasError = ref(false)
const errorMessage = ref('')

const { ativos, membros: todosMembros, carregar: recarregarMembros, currentMembro, tenantPermissions } = useMembros()
const { cartoes, inicializar: sincronizarDados, faturasAbertas, faturasFechadas } = useCartoesEFaturas()
const { carregarTemplates: inicializarContasFixas } = useContasFixas()

const isAuthed = ref(tenantSessionService.isAuthenticated())
const hasTenant = ref(!!tenantSessionService.getActiveTenantId())

const isReadOnly = computed(() => currentMembro.value?.role === 'VISUALIZADOR')
const isLancarGastoBloqueado = computed(() => {
  const role = currentMembro.value?.role
  if (!role || role === 'ADMIN') return false
  const perms = tenantPermissions.value[role]
  const defaultAllow = role === 'MORADOR'
  const allowed = perms ? (perms.ALLOW_LANCAR_GASTO !== undefined ? perms.ALLOW_LANCAR_GASTO : defaultAllow) : defaultAllow
  return allowed === false
})

provide('appState', {
  isAuthed,
  hasTenant,
  isLoading,
  membros: todosMembros,
  ativos,
  cartoes,
  faturasAbertas,
  faturasFechadas,
  isReadOnly,
  isLancarGastoBloqueado,
  sincronizarDados,
  recarregarMembros,
})

const assegurarDadosIniciais = async () => {
  const tenantId = tenantSessionService.getActiveTenantId()
  if (!tenantId) return

  isLoading.value = true
  try {
    socketService.desconectar()
    await Promise.all([
      recarregarMembros(),
      sincronizarDados(),
      inicializarContasFixas()
    ])
    inicializarSocket(tenantId)
  } catch (error: unknown) {
    logger.error('Erro na inicialização de dados:', error)
  } finally {
    isLoading.value = false
  }
}

const inicializarSocket = (tenantId: string) => {
  socketService.conectar(tenantId)

  let syncDebounceTimer: ReturnType<typeof setTimeout> | null = null
  let syncing = false

  const agendarSync = () => {
    if (syncDebounceTimer) clearTimeout(syncDebounceTimer)
    syncDebounceTimer = setTimeout(async () => {
      if (syncing) return
      syncing = true
      try {
        await sincronizarDados()
      } finally {
        syncing = false
      }
    }, 300)
  }

  socketService.on('gastos_alterados', agendarSync)
  socketService.on('cartoes_alterados', agendarSync)
  socketService.on('faturas_alteradas', agendarSync)

  socketService.on('membros_alterados', async () => await recarregarMembros())
  socketService.on('contas_fixas_alteradas', async () => await inicializarContasFixas())
  socketService.on('permissoes_alteradas', async () => await recarregarMembros())
}

const handleAuthSuccess = async () => {
  isAuthed.value = true
  const tenantId = tenantSessionService.getActiveTenantId()
  hasTenant.value = !!tenantId
  logger.info('Auth success - redirecionando:', { hasTenant: hasTenant.value, tenantId })
  if (hasTenant.value) {
    try {
      await assegurarDadosIniciais()
    } catch (err) {
      logger.error('Erro ao carregar dados iniciais após login:', err)
    }
    await router.push('/dashboard')
  } else {
    await router.push('/select-tenant')
  }
}

const handleCasaSelecionada = async () => {
  hasTenant.value = true
  await assegurarDadosIniciais()
  router.push('/dashboard')
}

const handleLogout = async () => {
  socketService.desconectar()
  await tenantSessionService.logout()
  isAuthed.value = false
  hasTenant.value = false
  router.push('/login')
}

const handleAuthExpired = () => {
  socketService.desconectar()
  isAuthed.value = false
  hasTenant.value = false
  toast.show('Sua sessão expirou. Faça login novamente.', 'error')
  router.push('/login')
}

const recarregarPagina = () => {
  window.location.reload()
}

const handleAppError = (e: Event) => {
  const detail = (e as CustomEvent).detail
  hasError.value = true
  errorMessage.value = detail.message || 'Ocorreu um erro inesperado no aplicativo.'
  logger.error('Erro capturado pelo Error Boundary:', detail)
}

onMounted(async () => {
  window.addEventListener('divi:tenant-changed', handleCasaSelecionada)
  window.addEventListener('divi:auth-expired', handleAuthExpired)
  window.addEventListener('divi:app-error', handleAppError)

  if (isAuthed.value) {
    try {
      isLoading.value = true
      await tenantSessionService.inicializarSessao()
      isAuthed.value = tenantSessionService.isAuthenticated()
      if (!isAuthed.value) return
      hasTenant.value = !!tenantSessionService.getActiveTenantId()
      if (hasTenant.value) {
        await assegurarDadosIniciais()
      }
    } catch (error: unknown) {
      logger.error('Erro na inicialização da sessão:', error)
    } finally {
      isInitializing.value = false
      isLoading.value = false
    }
  } else {
    isInitializing.value = false
  }
})

onUnmounted(() => {
  window.removeEventListener('divi:tenant-changed', handleCasaSelecionada)
  window.removeEventListener('divi:auth-expired', handleAuthExpired)
  window.removeEventListener('divi:app-error', handleAppError)
  socketService.desconectar()
})
</script>

<template>
  <div class="divi-app-root min-h-screen bg-canvas">
    <Transition name="fade" mode="out-in">
      <div v-if="isInitializing" class="min-h-screen bg-canvas flex flex-col items-center justify-center p-8 space-y-12 animate-in fade-in duration-200">
        <div class="flex flex-col items-center space-y-4">
          <IllustrationMascot variant="ember" :size="80" mood="happy" class="animate-wobble" />
          <h1 class="text-display text-5xl md:text-6xl text-charcoal">
            DIVI<span class="text-ember">.</span>
          </h1>
        </div>
        <div class="w-full max-w-[200px] space-y-4 opacity-40">
          <div class="h-1 bg-stone rounded-full overflow-hidden">
            <div class="h-full bg-ember/40 animate-loading-bar" />
          </div>
          <p class="text-[10px] font-bold text-ash uppercase tracking-[0.25em] text-center">Iniciando aventura</p>
        </div>
      </div>

      <div v-else-if="hasError" class="min-h-screen bg-canvas flex flex-col items-center justify-center p-8 space-y-8">
        <IllustrationMascot variant="coral" :size="80" mood="sad" />
        <div class="text-center space-y-2">
          <h2 class="text-heading text-charcoal">Algo deu errado</h2>
          <p class="text-body text-graphite">{{ errorMessage }}</p>
        </div>
        <button
          class="px-6 py-3 rounded-full bg-midnight text-white font-bold text-sm transition-all duration-200 hover:scale-105 active:scale-95"
          @click="recarregarPagina"
        >
          Tentar novamente
        </button>
      </div>

      <div v-else class="min-h-screen bg-canvas text-graphite font-sans selection:bg-ember/20">
        <ToastNotification />
        <router-view v-slot="{ Component }">
          <Transition name="fade" mode="out-in">
            <component
              :is="Component"
              @auth-success="handleAuthSuccess"
              @forgot-password="router.push('/forgot-password')"
              @back="router.push('/login')"
              @reset-success="router.push('/login')"
              @casa-selecionada="handleCasaSelecionada"
              @logout="handleLogout"
              @recarregar-cartoes="sincronizarDados"
            />
          </Transition>
        </router-view>
      </div>
    </Transition>
  </div>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease-out;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@keyframes loading-bar {
  0% { transform: translateX(-100%); }
  50% { transform: translateX(0); }
  100% { transform: translateX(100%); }
}

.animate-loading-bar {
  animation: loading-bar 2s ease-in-out infinite;
}
</style>
