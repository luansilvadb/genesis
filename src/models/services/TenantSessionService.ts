import { logger } from '../../shared/utils/logger'
import { z } from 'zod'
import { obterHeadersCsrf, definirCsrfToken } from '../../shared/utils/csrf'
import {
  AuthResponseSchema,
  SessionResponseSchema,
  InvitePreviewResponseSchema,
  LoginRequestSchema,
  RegisterRequestSchema,
  CreateTenantRequestSchema,
  JoinTenantRequestSchema,
} from '../../shared/validation/apiSchemas'

export interface TenantSummary {
  id: string
  name: string
  inviteCode: string
}

export interface InvitePreview {
  id: string
  name: string
  membrosDisponiveis: { id: string; nome: string; avatar: string }[]
}

interface LoginResponse {
  token: string
  user: {
    id: string
    email: string
    nome: string
  }
}

const lerMensagemErro = async (response: Response, fallback: string): Promise<string> => {
  const body = await response.json().catch(() => null) as { message?: string } | null
  return body?.message || fallback
}

/**
 * Valida a resposta de um endpoint de auth contra o schema Zod esperado.
 * Lança erro se o contrato for quebrado, independente do ambiente.
 * Em produção, o log detalhado é registrado para diagnóstico.
 */
function validateResponse<T>(schema: z.ZodType<T>, data: unknown, endpoint: string): T {
  const result = schema.safeParse(data)
  if (!result.success) {
    const issues = result.error.issues.map(i => `${i.path.join('.') || '(root)'}: ${i.message}`)
    const errorMessage = `[Auth Contract] Resposta de ${endpoint} divergiu: ${issues.join('; ')}`
    logger.error(errorMessage)
    if (import.meta.env.DEV || import.meta.env.MODE === 'test') {
      console.error(errorMessage, { raw: data, issues: result.error.issues })
    }
    throw new Error(`Falha na validação da resposta da API. Entre em contato com o suporte.`)
  }
  return data as T
}

export class TenantSessionService {
  private activeTenantId: string | null = null
  private jwtToken: string | null = null
  private currentUserId: string | null = null

  private tenants: TenantSummary[] = []

  constructor() {
    this.activeTenantId = localStorage.getItem('divi_active_tenant_id')
    this.jwtToken = localStorage.getItem('divi_jwt_token')
    this.currentUserId = localStorage.getItem('divi_current_user_id')
  }

  private get baseUrl(): string {
    const url = (import.meta.env.VITE_API_URL as string) || 'http://localhost:3000'
    const normalized = url.replace(/\/+$/, '')
    const base = normalized.endsWith('/api') ? normalized : `${normalized}/api`
    return `${base}/`
  }

  private async authenticate(endpoint: string, body: Record<string, unknown>, errorLabel: string, catchLabel: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}${endpoint}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body)
      })

      if (!response.ok) {
        logger.error(`Erro de ${errorLabel}:`, await lerMensagemErro(response, response.statusText))
        return false
      }

      const data = validateResponse(AuthResponseSchema, await response.json(), endpoint)
      this.persistSession(data)
      await this.carregarSessaoUsuario()
      return true
    } catch (err) {
      logger.error(`Falha de conexão ${catchLabel}:`, err)
      return false
    }
  }

  async login(email: string, passwordSecret: string): Promise<boolean> {
    try {
      LoginRequestSchema.parse({ email, password: passwordSecret })
    } catch (validationErr) {
      logger.error('Erro de validação no login:', validationErr)
      return false
    }
    return this.authenticate('auth/login', { email, password: passwordSecret }, 'login', 'ao fazer login')
  }

  async loginComGoogle(credential: string, inviteCode?: string, membroId?: string): Promise<boolean> {
    return this.authenticate('auth/google', { credential, inviteCode, membroId }, 'login com Google', 'ao fazer login com Google')
  }

  async register(email: string, nome: string, passwordSecret: string, inviteCode?: string, membroId?: string): Promise<boolean> {
    try {
      RegisterRequestSchema.parse({ email, nome, password: passwordSecret, inviteCode, membroId })
    } catch (validationErr) {
      logger.error('Erro de validação no registro:', validationErr)
      return false
    }
    return this.authenticate('auth/register', { email, nome, password: passwordSecret, inviteCode, membroId }, 'cadastro', 'ao registrar')
  }

  private persistSession(data: LoginResponse): void {
    this.jwtToken = data.token
    this.currentUserId = data.user.id

    localStorage.setItem('divi_jwt_token', data.token)
    localStorage.setItem('divi_current_user_id', data.user.id)
    localStorage.setItem('divi_user_email', data.user.email)
    localStorage.setItem('divi_username', data.user.nome)
  }

  async getInvitePreview(code: string): Promise<InvitePreview> {
    const response = await fetch(`${this.baseUrl}tenants/invite/${code}`)
    if (!response.ok) {
      throw new Error(await lerMensagemErro(response, 'Convite inválido'))
    }
    return validateResponse(InvitePreviewResponseSchema, await response.json(), 'tenants/invite/:code')
  }

  async forgotPassword(email: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}auth/forgot-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email })
      })
      if (!response.ok) {
        logger.error('Erro forgotPassword:', await lerMensagemErro(response, response.statusText))
        return false
      }
      return true
    } catch (err) {
      logger.error('Falha de conexão em forgotPassword:', err)
      return false
    }
  }

  async resetPassword(token: string, newPassword: string): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}auth/reset-password`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token, password: newPassword })
      })
      if (!response.ok) {
        logger.error('Erro resetPassword:', await lerMensagemErro(response, response.statusText))
        return false
      }
      return true
    } catch (err) {
      logger.error('Falha de conexão em resetPassword:', err)
      return false
    }
  }

  async logout(): Promise<void> {
    this.jwtToken = null
    this.currentUserId = null
    this.activeTenantId = null
    this.tenants = []
    localStorage.removeItem('divi_jwt_token')
    localStorage.removeItem('divi_current_user_id')
    localStorage.removeItem('divi_active_tenant_id')
    localStorage.removeItem('divi_username')
  }

  isAuthenticated(): boolean {
    if (!this.jwtToken) return false
    // Basic expiration check: decode payload without verifying signature
    // (the server will reject expired tokens; this is a client-side early-out).
    try {
      const payload = JSON.parse(atob(this.jwtToken.split('.')[1]))
      if (payload.exp && Date.now() >= payload.exp * 1000) {
        this.logout()
        return false
      }
    } catch {
      // Token malformed — treat as not authenticated
      return false
    }
    return true
  }

  /** Carrega a sessão do usuário (tenants) ao inicializar o app. Deve ser chamado antes de qualquer fetch de dados. */
  async inicializarSessao(): Promise<void> {
    if (this.jwtToken) {
      await this.carregarSessaoUsuario()
    }
  }

  getActiveTenantId(): string | null {
    return this.activeTenantId
  }

  setActiveTenant(tenantId: string): void {
    this.activeTenantId = tenantId || null
    if (tenantId) {
      localStorage.setItem('divi_active_tenant_id', tenantId)
    } else {
      localStorage.removeItem('divi_active_tenant_id')
    }
  }

  getCurrentUserId(): string | null {
    return this.currentUserId
  }

  getTenants(): TenantSummary[] {
    return [...this.tenants]
  }

  /** Cria uma nova casa e seleciona ela automaticamente */
  async criarCasa(nome: string): Promise<TenantSummary> {
    CreateTenantRequestSchema.parse({ name: nome })

    const response = await fetch(`${this.baseUrl}tenants`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.jwtToken}`,
        ...obterHeadersCsrf(),
      },
      body: JSON.stringify({ name: nome }),
      credentials: 'include',
    })

    if (!response.ok) {
      throw new Error(await lerMensagemErro(response, 'Erro ao criar a casa'))
    }

    const raw = await response.json()
    const tenant = validateResponse(
      z.object({ id: z.string(), name: z.string(), inviteCode: z.string() }),
      raw,
      'tenants',
    )
    this.setActiveTenant(tenant.id)
    this.tenants = [...this.tenants, tenant]
    return tenant
  }

  /** Entra em uma casa existente pelo código de convite */
  async entrarCasa(inviteCode: string): Promise<TenantSummary> {
    JoinTenantRequestSchema.parse({ inviteCode })

    const response = await fetch(`${this.baseUrl}tenants/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${this.jwtToken}`,
        ...obterHeadersCsrf(),
      },
      body: JSON.stringify({ inviteCode }),
      credentials: 'include',
    })

    if (!response.ok) {
      throw new Error(await lerMensagemErro(response, 'Código de convite inválido ou casa não encontrada.'))
    }

    const raw = await response.json()
    const tenant = validateResponse(
      z.object({ id: z.string(), name: z.string(), inviteCode: z.string() }),
      raw,
      'tenants/join',
    )
    this.setActiveTenant(tenant.id)
    if (!this.tenants.find(t => t.id === tenant.id)) {
      this.tenants = [...this.tenants, tenant]
    }
    return tenant
  }

  private async carregarSessaoUsuario(): Promise<void> {
    if (!this.jwtToken) return
    try {
      const response = await fetch(`${this.baseUrl}auth/me`, {
        headers: {
          'Authorization': `Bearer ${this.jwtToken}`
        },
        credentials: 'include',
      })

      // Captura o token CSRF do header para requisições POST subsequentes
      const csrfHeader = response.headers?.get?.('X-CSRF-Token') ?? null
      if (csrfHeader) definirCsrfToken(csrfHeader)

      if (response.status === 401) {
        await this.logout()
        return
      }
      if (!response.ok) return

      const data = validateResponse(SessionResponseSchema, await response.json(), 'auth/me')
      this.tenants = data.tenants || []
      if (this.tenants.length === 0) {
        this.setActiveTenant('')
        return
      }
      if (!this.activeTenantId || !this.tenants.some(t => t.id === this.activeTenantId)) {
        this.setActiveTenant(this.tenants[0].id)
      }
    } catch (err) {
      // Não relançamos a exceção para não quebrar o fluxo de login.
      // O token JWT já foi salvo; o usuário pode recarregar os tenants depois.
      logger.error('Falha ao carregar sessão do usuário:', err)
    }
  }
}
