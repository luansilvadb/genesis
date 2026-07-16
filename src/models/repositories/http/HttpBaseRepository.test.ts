import { beforeEach, describe, expect, it, vi } from 'vitest'
import { HttpBaseRepository } from './HttpBaseRepository'
import { z } from 'zod'

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}
  return {
    getItem: vi.fn((key: string) => store[key] ?? null),
    setItem: vi.fn((key: string, value: string) => { store[key] = value }),
    removeItem: vi.fn((key: string) => { delete store[key] }),
    clear: vi.fn(() => { store = {} }),
  }
})()

Object.defineProperty(globalThis, 'localStorage', { value: localStorageMock })

class TestRepository extends HttpBaseRepository {
  async getTest(url: string): Promise<unknown> {
    return this.request<unknown>(url)
  }

  async postTest(url: string, body?: unknown): Promise<unknown> {
    return this.request<unknown>(url, { method: 'POST', body: JSON.stringify(body) })
  }

  async putTest(url: string, body?: unknown): Promise<unknown> {
    return this.request<unknown>(url, { method: 'PUT', body: JSON.stringify(body) })
  }

  async deleteTest(url: string): Promise<unknown> {
    return this.request<unknown>(url, { method: 'DELETE' })
  }

  async getValidated<T>(schema: z.ZodType<T>, url: string): Promise<T> {
    return this.validatedRequest<T>(schema, url)
  }
}

/** Helper: builds a mock Response with an optional X-CSRF-Token header. */
function createMockResponse(
  overrides: Partial<Response> & { csrfHeader?: string | null } = {},
): Response {
  const headers = new Headers()
  if (overrides.csrfHeader !== null && overrides.csrfHeader !== undefined) {
    headers.set('X-CSRF-Token', overrides.csrfHeader)
  } else if (overrides.csrfHeader === undefined) {
    // Default: include a token so GET requests prime the CSRF cache
    headers.set('X-CSRF-Token', 'default-test-token')
  }

  return {
    ok: true,
    status: 200,
    json: async () => ({}),
    headers,
    ...overrides,
  } as Response
}

// Helper to extract the X-CSRF-Token header that was sent with a fetch call
function csrfHeaderFromCall(mock: ReturnType<typeof vi.fn>, callIndex: number): string | null {
  const call = mock.mock.calls[callIndex]
  if (!call?.[1]?.headers) return null
  return (call[1].headers as Headers).get('X-CSRF-Token') ?? null
}

describe('HttpBaseRepository', () => {
  let repo: TestRepository

  beforeEach(() => {
    vi.restoreAllMocks()
    localStorageMock.clear()
    // Re-create repo to get a fresh instance (note: module-scoped CSRF
    // state persists across tests, which is acceptable since it mirrors
    // real session behavior — each test primes the state it needs.)
    repo = new TestRepository()
  })

  // -----------------------------------------------------------------------
  // Existing behaviour (unchanged contract)
  // -----------------------------------------------------------------------

  describe('request', () => {
    it('deve incluir header Authorization quando token existe', async () => {
      localStorageMock.setItem('divi_jwt_token', 'test-token')
      const fetchMock = vi.fn().mockResolvedValue(createMockResponse())
      vi.stubGlobal('fetch', fetchMock)

      await repo.getTest('/test')

      const headers = fetchMock.mock.calls[0][1].headers
      expect(headers.get('Authorization')).toBe('Bearer test-token')
    })

    it('deve incluir header X-Tenant-ID quando tenantId existe', async () => {
      localStorageMock.setItem('divi_active_tenant_id', 'tenant-123')
      const fetchMock = vi.fn().mockResolvedValue(createMockResponse())
      vi.stubGlobal('fetch', fetchMock)

      await repo.getTest('/test')

      const headers = fetchMock.mock.calls[0][1].headers
      expect(headers.get('X-Tenant-ID')).toBe('tenant-123')
    })

    it('deve retornar null para resposta 204', async () => {
      const fetchMock = vi.fn().mockResolvedValue({
        ok: true,
        status: 204,
        headers: new Headers(),
      })
      vi.stubGlobal('fetch', fetchMock)

      const result = await repo.getTest('/test')
      expect(result).toBeNull()
    })

    it('deve disparar divi:auth-expired em resposta 401', async () => {
      localStorageMock.setItem('divi_jwt_token', 'expired-token')
      localStorageMock.setItem('divi_active_tenant_id', 't1')
      const eventSpy = vi.fn()
      window.addEventListener('divi:auth-expired', eventSpy)

      const fetchMock = vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        headers: new Headers(),
        json: async () => ({ message: 'token expirado' }),
      })
      vi.stubGlobal('fetch', fetchMock)

      await expect(repo.getTest('/test')).rejects.toThrow('Sessão expirada')
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('divi_jwt_token')
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('divi_active_tenant_id')
      expect(eventSpy).toHaveBeenCalled()

      window.removeEventListener('divi:auth-expired', eventSpy)
    })

    it('deve extrair mensagem de erro do corpo da resposta', async () => {
      const fetchMock = vi.fn().mockResolvedValue({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        headers: new Headers(),
        json: async () => ({ message: 'Campo obrigatório ausente' }),
      })
      vi.stubGlobal('fetch', fetchMock)

      await expect(repo.getTest('/test')).rejects.toThrow('Campo obrigatório ausente')
    })

    it('deve tratar erro de conexão com mensagem amigável', async () => {
      const fetchMock = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'))
      vi.stubGlobal('fetch', fetchMock)

      await expect(repo.getTest('/test')).rejects.toThrow('Não foi possível se conectar ao servidor do DIVI')
    })
  })

  // -----------------------------------------------------------------------
  // CSRF — token inclusion
  // -----------------------------------------------------------------------

  describe('CSRF token inclusion', () => {
    it('deve incluir header X-CSRF-Token em requisições POST', async () => {
      // Prime the CSRF token via a GET first
      let fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'prime-token' }))
      vi.stubGlobal('fetch', fetchMock)
      await repo.getTest('/cartoes')
      vi.restoreAllMocks()

      // Now POST — the CSRF token should be attached
      fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'prime-token' }))
      vi.stubGlobal('fetch', fetchMock)

      await repo.postTest('/gastos', { valor: 100 })

      const sentToken = csrfHeaderFromCall(fetchMock, 0)
      expect(sentToken).toBe('prime-token')
    })

    it('deve incluir header X-CSRF-Token em requisições PUT', async () => {
      let fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'put-token' }))
      vi.stubGlobal('fetch', fetchMock)
      await repo.getTest('/cartoes')
      vi.restoreAllMocks()

      fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'put-token' }))
      vi.stubGlobal('fetch', fetchMock)

      await repo.putTest('/gastos/1', { valor: 200 })

      expect(csrfHeaderFromCall(fetchMock, 0)).toBe('put-token')
    })

    it('deve incluir header X-CSRF-Token em requisições DELETE', async () => {
      let fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'del-token' }))
      vi.stubGlobal('fetch', fetchMock)
      await repo.getTest('/cartoes')
      vi.restoreAllMocks()

      fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'del-token' }))
      vi.stubGlobal('fetch', fetchMock)

      await repo.deleteTest('/cartoes/1')

      expect(csrfHeaderFromCall(fetchMock, 0)).toBe('del-token')
    })

    it('NÃO deve bloquear requisições GET (CSRF não é enviado mas também não falha)', async () => {
      const fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'get-token' }))
      vi.stubGlobal('fetch', fetchMock)

      const result = await repo.getTest('/cartoes')

      expect(result).toBeDefined()
      // GET requests don't send CSRF header (it's not required)
      expect(csrfHeaderFromCall(fetchMock, 0)).toBeNull()
    })
  })

  // -----------------------------------------------------------------------
  // credentials: 'include'
  // -----------------------------------------------------------------------

  describe('credentials include', () => {
    it('deve incluir credentials: include em requisições GET', async () => {
      const fetchMock = vi.fn().mockResolvedValue(createMockResponse())
      vi.stubGlobal('fetch', fetchMock)

      await repo.getTest('/test')

      expect(fetchMock.mock.calls[0][1].credentials).toBe('include')
    })

    it('deve incluir credentials: include em requisições POST', async () => {
      // Prime CSRF token
      let fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'cred-token' }))
      vi.stubGlobal('fetch', fetchMock)
      await repo.getTest('/cartoes')
      vi.restoreAllMocks()

      fetchMock = vi.fn().mockResolvedValue(createMockResponse({ csrfHeader: 'cred-token' }))
      vi.stubGlobal('fetch', fetchMock)

      await repo.postTest('/gastos', { valor: 50 })

      expect(fetchMock.mock.calls[0][1].credentials).toBe('include')
    })
  })

  // -----------------------------------------------------------------------
  // CSRF — 403 retry logic
  // -----------------------------------------------------------------------

  describe('CSRF 403 retry', () => {
    it('deve renovar token CSRF e retentar em caso de 403 com mensagem CSRF', async () => {
      // Step 1: POST returns 403 "CSRF token ausente"
      // Step 2: getCsrfToken() calls GET /auth/me → returns 200 with new header
      // Step 3: POST retry succeeds
      const fetchMock = vi.fn()
        // First call: POST → 403 CSRF
        .mockResolvedValueOnce({
          ok: false,
          status: 403,
          statusText: 'Forbidden',
          headers: new Headers(),
          json: async () => ({ message: 'CSRF token ausente' }),
        })
        // Second call: GET /auth/me (CSRF refresh) → success with header
        .mockResolvedValueOnce(
          createMockResponse({ csrfHeader: 'fresh-token', json: async () => ({ id: 'u1', nome: 'User' }) }),
        )
        // Third call: POST retry → success
        .mockResolvedValueOnce(
          createMockResponse({ csrfHeader: 'fresh-token', json: async () => ({ id: 'g1', valor: 100 }) }),
        )

      vi.stubGlobal('fetch', fetchMock)

      const result = await repo.postTest('/gastos', { valor: 100 })

      expect(result).toEqual({ id: 'g1', valor: 100 })
      expect(fetchMock).toHaveBeenCalledTimes(3)

      // The retry request should have the fresh token
      expect(csrfHeaderFromCall(fetchMock, 2)).toBe('fresh-token')
    })

    it('deve propagar erro se 403 não for relacionado a CSRF', async () => {
      const fetchMock = vi.fn().mockResolvedValue({
        ok: false,
        status: 403,
        statusText: 'Forbidden',
        headers: new Headers(),
        json: async () => ({ message: 'Acesso negado — permissão insuficiente' }),
      })
      vi.stubGlobal('fetch', fetchMock)

      await expect(repo.postTest('/admin-only', {})).rejects.toThrow(
        'Acesso negado — permissão insuficiente',
      )
      // Should NOT have made any extra calls (no retry)
      expect(fetchMock).toHaveBeenCalledTimes(1)
    })

    it('não deve retentar mais de uma vez em caso de 403 CSRF consecutivos', async () => {
      const fetchMock = vi.fn()
        // First call: POST → 403 CSRF
        .mockResolvedValueOnce({
          ok: false,
          status: 403,
          statusText: 'Forbidden',
          headers: new Headers(),
          json: async () => ({ message: 'CSRF token inválido' }),
        })
        // Second call: GET /auth/me (CSRF refresh)
        .mockResolvedValueOnce(
          createMockResponse({ csrfHeader: 'another-token' }),
        )
        // Third call: POST retry → still 403 CSRF
        .mockResolvedValueOnce({
          ok: false,
          status: 403,
          statusText: 'Forbidden',
          headers: new Headers(),
          json: async () => ({ message: 'CSRF token inválido' }),
        })

      vi.stubGlobal('fetch', fetchMock)

      await expect(repo.postTest('/gastos', { valor: 100 })).rejects.toThrow('CSRF token inválido')

      // Exactly 3 calls: original POST, refresh GET, retry POST — no 4th call
      expect(fetchMock).toHaveBeenCalledTimes(3)
    })
  })

  // -----------------------------------------------------------------------
  // validatedRequest (unchanged)
  // -----------------------------------------------------------------------

  describe('validatedRequest', () => {
    const TestSchema = z.object({
      id: z.string(),
      nome: z.string(),
    })

    it('deve retornar dados validados quando schema passa', async () => {
      const fetchMock = vi.fn().mockResolvedValue(createMockResponse({
        json: async () => ({ id: '1', nome: 'Teste' }),
      }))
      vi.stubGlobal('fetch', fetchMock)

      const result = await repo.getValidated(TestSchema, '/test')
      expect(result).toEqual({ id: '1', nome: 'Teste' })
    })

    it('deve lançar erro em ambiente de teste quando schema falha', async () => {
      const fetchMock = vi.fn().mockResolvedValue(createMockResponse({
        json: async () => ({ id: 123, nome: 'Teste' }), // id não é string
      }))
      vi.stubGlobal('fetch', fetchMock)

      await expect(repo.getValidated(TestSchema, '/test')).rejects.toThrow('[API Contract]')
    })
  })
})
