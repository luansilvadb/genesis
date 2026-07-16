import { logger } from '../../../shared/utils/logger'
import type { z } from 'zod'
import { lerCsrfToken, definirCsrfToken, limparCsrfToken } from '../../../shared/utils/csrf'

// ---------------------------------------------------------------------------
// CSRF token state — module-scoped (session lifetime, not persisted)
// ---------------------------------------------------------------------------

let csrfRefreshPromise: Promise<string> | null = null

function isMutatingMethod(method: string): boolean {
  const m = method.toUpperCase()
  return m === 'POST' || m === 'PUT' || m === 'DELETE' || m === 'PATCH'
}

/**
 * Base class for HTTP-based repository implementations.
 *
 * KNOWN LIMITATION: buscarPorId and buscarPorFatura methods in concrete
 * repositories fetch the entire collection and filter locally (O(n)).
 * For datasets with 10k+ records, consider adding dedicated GET /resource/:id
 * and GET /resource?fatura_id=X endpoints to the backend API.
 */
export class HttpBaseRepository {
  protected get baseUrl(): string {
    const url = (import.meta.env.VITE_API_URL as string) || 'http://localhost:3000'
    const normalized = url.replace(/\/+$/, '')
    const base = normalized.endsWith('/api') ? normalized : `${normalized}/api`
    return `${base}/`
  }

  protected get token(): string | null {
    return localStorage.getItem('divi_jwt_token')
  }

  protected get tenantId(): string | null {
    return localStorage.getItem('divi_active_tenant_id')
  }

  // -----------------------------------------------------------------------
  // CSRF helpers
  // -----------------------------------------------------------------------

  /**
   * Returns the currently cached CSRF token, or an empty string if none
   * has been obtained yet. The token is primed by normal GET responses
   * (captured from the X-CSRF-Token header) during the app lifecycle.
   *
   * Falls back to reading the csrf_token cookie (httpOnly=false) when the
   * in-memory cache is empty — this covers the case where the first
   * mutating request happens before any GET response populates the cache.
   */
  private getCsrfToken(): string {
    return lerCsrfToken()
  }

  /**
   * Fetches a fresh CSRF token from the backend via a lightweight GET.
   * Concurrent callers reuse a single in-flight promise (dedup).
   * Called automatically by the 403 CSRF retry handler.
   */
  private async refreshCsrfToken(): Promise<string> {
    if (csrfRefreshPromise) return csrfRefreshPromise

    csrfRefreshPromise = (async (): Promise<string> => {
      try {
        const headers = new Headers()
        headers.set('Content-Type', 'application/json')
        if (this.token) headers.set('Authorization', `Bearer ${this.token}`)
        if (this.tenantId) headers.set('X-Tenant-ID', this.tenantId)

        const response = await fetch(`${this.baseUrl}auth/me`, {
          headers,
          credentials: 'include',
        })

        const newToken = response.headers?.get('X-CSRF-Token') ?? null
        if (newToken) {
          definirCsrfToken(newToken)
          return newToken
        }

        logger.warn('CSRF token não disponível na resposta do servidor')
        return ''
      } catch {
        logger.warn('Falha ao obter CSRF token do servidor')
        return ''
      } finally {
        csrfRefreshPromise = null
      }
    })()

    return csrfRefreshPromise
  }

  // -----------------------------------------------------------------------
  // HTTP request
  // -----------------------------------------------------------------------

  protected async request<T>(url: string, options: RequestInit = {}): Promise<T> {
    return this.sendRequest<T>(url, options, false)
  }

  // ── Request pipeline (split for lower cyclomatic complexity) ──────────

  /** Builds Headers for an outgoing request, attaching auth, tenant, and CSRF. */
  private buildHeaders(method: string, baseHeaders?: HeadersInit): Headers {
    const headers = new Headers(baseHeaders)
    headers.set('Content-Type', 'application/json')

    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`)
    }

    if (this.tenantId) {
      headers.set('X-Tenant-ID', this.tenantId)
    }

    if (isMutatingMethod(method)) {
      const csrf = this.getCsrfToken()
      if (csrf) {
        headers.set('X-CSRF-Token', csrf)
      }
    }

    return headers
  }

  /**
   * Dispatches the actual fetch with a 15s timeout, translating network
   * errors into user-facing messages.
   */
  private async fetchWithTimeout(
    url: string,
    options: RequestInit,
  ): Promise<Response> {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 15000)

    try {
      return await fetch(url, { ...options, signal: controller.signal, credentials: 'include' })
    } catch (err: unknown) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        throw new Error('A requisição excedeu o tempo limite. Verifique sua conexão e tente novamente.', { cause: err })
      }
      logger.error(`Falha de conexão para ${url}:`, err)
      throw new Error('Não foi possível se conectar ao servidor do DIVI. Certifique-se de que a API está ativa e que há conexão com a internet.', { cause: err })
    } finally {
      clearTimeout(timeoutId)
    }
  }

  /**
   * Handles HTTP response status codes. Returns the parsed JSON body on
   * success, or throws with a descriptive error message on failure.
   *
   * Special cases:
   * - 403 with CSRF message → triggers token renewal + single retry
   * - 401 → clears session and dispatches divi:auth-expired
   * - 204 → returns null
   */
  private async handleResponse<T>(
    response: Response,
    url: string,
    options: RequestInit,
    csrfRetry: boolean,
  ): Promise<T> {
    // Capture CSRF token from successful GET responses for future mutating requests
    const method = (options.method || 'GET').toUpperCase()
    if (method === 'GET' && response.ok) {
      const newToken = response.headers?.get('X-CSRF-Token') ?? null
      if (newToken) {
        definirCsrfToken(newToken)
      }
    }

    if (response.status === 403) {
      return this.handle403<T>(response, url, options, csrfRetry)
    }

    if (!response.ok) {
      return this.handleErrorResponse<T>(response)
    }

    if (response.status === 204) {
      return null as T
    }

    return response.json()
  }

  /** Handles 403 responses with CSRF token renewal and single retry. */
  private async handle403<T>(
    response: Response,
    url: string,
    options: RequestInit,
    csrfRetry: boolean,
  ): Promise<T> {
    const errBody = await response.json().catch(() => ({}))
    if (/csrf token/i.test(errBody.message || '')) {
      limparCsrfToken()
      if (!csrfRetry) {
        logger.warn('CSRF token expirado ou ausente, renovando...')
        await this.refreshCsrfToken()
        return this.sendRequest<T>(url, options, true)
      }
    }
    throw new Error(errBody.message || `Erro HTTP: ${response.status} ${response.statusText}`)
  }

  /** Handles non-403 error responses. 401 triggers auth-expired cleanup. */
  private async handleErrorResponse<T>(response: Response): Promise<T> {
    if (response.status === 401) {
      localStorage.removeItem('divi_jwt_token')
      localStorage.removeItem('divi_active_tenant_id')
      window.dispatchEvent(new CustomEvent('divi:auth-expired'))
      throw new Error('Sessão expirada. Faça login novamente.')
    }
    const errBody = await response.json().catch(() => ({}))
    throw new Error(errBody.message || `Erro HTTP: ${response.status} ${response.statusText}`)
  }

  /** Core request implementation: build headers → fetch → handle response. */
  private async sendRequest<T>(
    url: string,
    options: RequestInit,
    csrfRetry: boolean,
  ): Promise<T> {
    const method = (options.method || 'GET').toUpperCase()
    const headers = this.buildHeaders(method, options.headers)

    const cleanUrl = url.startsWith('/') ? url.slice(1) : url

    const response = await this.fetchWithTimeout(`${this.baseUrl}${cleanUrl}`, {
      ...options,
      headers,
    })

    return this.handleResponse<T>(response, url, options, csrfRetry)
  }

  /**
   * Like request(), but validates the JSON response against a Zod schema.
   *
   * Em desenvolvimento e testes, lança erro se a resposta não corresponder
   * ao schema esperado — quebras de contrato são detectadas imediatamente.
   *
   * Em produção, loga o erro mas retorna os dados brutos para evitar
   * interrupção do fluxo do usuário.
   */
  protected async validatedRequest<T>(
    schema: z.ZodType<T>,
    url: string,
    options: RequestInit = {},
  ): Promise<T> {
    const raw = await this.request<unknown>(url, options)
    const result = schema.safeParse(raw)
    if (!result.success) {
      const endpoint = url.split('?')[0]
      const issues = result.error.issues.map(i => `${i.path.join('.') || '(root)'}: ${i.message}`)
      const errorMessage = `[API Contract] Resposta de ${endpoint} divergiu do schema esperado: ${issues.join('; ')}`

      logger.error(errorMessage)

      if (import.meta.env.DEV || import.meta.env.MODE === 'test') {
        // Em dev/teste, lança erro para que quebras de contrato sejam detectadas imediatamente
        console.error(errorMessage, { raw, issues: result.error.issues })
        throw new Error(errorMessage)
      }
      // Em produção, retorna os dados brutos para não interromper o fluxo,
      // mas registra um aviso visível no console para facilitar debugging.
      console.warn(
        `[API Contract Warning] ${endpoint} response did not match expected schema. ` +
        `This may cause runtime errors. Issues: ${issues.join('; ')}`,
        { raw }
      )
    }
    return raw as T
  }
}
