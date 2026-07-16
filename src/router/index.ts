import { createRouter, createWebHashHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { tenantSessionService } from '../shared/container'

const PUBLIC_PATHS = ['/login', '/forgot-password', '/reset-password']
const isPublic = (path: string) => PUBLIC_PATHS.some(p => path.startsWith(p))

const LoginScreen = () => import('../views/screens/LoginScreen.vue')
const ForgotPasswordScreen = () => import('../views/screens/ForgotPasswordScreen.vue')
const ResetPasswordScreen = () => import('../views/screens/ResetPasswordScreen.vue')
const TenantSelectorScreen = () => import('../views/screens/TenantSelectorScreen.vue')
const DashboardView = () => import('../router/DashboardView.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: () => {
      if (!tenantSessionService.isAuthenticated()) return '/login'
      if (!tenantSessionService.getActiveTenantId()) return '/select-tenant'
      return '/dashboard'
    }
  },
  {
    path: '/login',
    name: 'login',
    component: LoginScreen,
  },
  {
    path: '/forgot-password',
    name: 'forgotPassword',
    component: ForgotPasswordScreen,
  },
  {
    path: '/reset-password',
    name: 'resetPassword',
    component: ResetPasswordScreen,
    props: (route) => ({ token: route.query.token || '' }),
  },
  {
    path: '/select-tenant',
    name: 'selectTenant',
    component: TenantSelectorScreen,
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: DashboardView,
  },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// Navigation guards: enforce auth and tenant selection.
router.beforeEach(async (to, _from, next) => {
  const isAuthed = tenantSessionService.isAuthenticated()
  const hasTenant = !!tenantSessionService.getActiveTenantId()

  if (!isAuthed && !isPublic(to.path)) {
    return next('/login')
  }

  if (isAuthed && !hasTenant && to.path !== '/select-tenant' && !isPublic(to.path)) {
    return next('/select-tenant')
  }

  if (isAuthed && hasTenant && (to.path.startsWith('/login') || to.path.startsWith('/select-tenant'))) {
    return next('/dashboard')
  }

  next()
})

export default router
