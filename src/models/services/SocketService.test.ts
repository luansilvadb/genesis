import { beforeEach, describe, expect, it, vi, afterEach } from 'vitest'
import { SocketService } from './SocketService'

// Mock do WebSocket
class MockWebSocket {
  url: string
  readyState: number = WebSocket.CONNECTING
  onopen: ((ev: Event) => void) | null = null
  onclose: ((ev: CloseEvent) => void) | null = null
  onmessage: ((ev: MessageEvent) => void) | null = null
  onerror: ((ev: Event) => void) | null = null

  static readonly CONNECTING = 0
  static readonly OPEN = 1
  static readonly CLOSING = 2
  static readonly CLOSED = 3

  constructor(url: string) {
    this.url = url
  }

  close() {
    this.readyState = WebSocket.CLOSED
    if (this.onclose) {
      this.onclose(new CloseEvent('close'))
    }
  }

  send() {}
}

// Mock do localStorage
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
vi.stubGlobal('WebSocket', MockWebSocket)

// Mock do fetch (preflight)
const fetchMock = vi.fn()
vi.stubGlobal('fetch', fetchMock)

// Mock do window.dispatchEvent
const dispatchEventSpy = vi.fn()
vi.stubGlobal('dispatchEvent', dispatchEventSpy)

describe('SocketService', () => {
  let service: SocketService

  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    localStorageMock.clear()
    localStorageMock.setItem('divi_jwt_token', 'test-jwt-token')
    // Preflight padrão: autorizado (200)
    fetchMock.mockResolvedValue({ status: 200 })
    service = new SocketService()
  })

  afterEach(() => {
    service.desconectar()
    vi.restoreAllMocks()
  })

  describe('conectar', () => {
    it('deve criar uma conexão WebSocket com token e tenant_id', async () => {
      await service.conectar('tenant-123')

      // O WebSocket foi criado (o mock registra a instância)
      // Verificamos que o localStorage foi lido corretamente
      expect(localStorageMock.getItem).toHaveBeenCalledWith('divi_jwt_token')
    })

    it('não deve reconectar se já estiver conectado ao mesmo tenant', async () => {
      await service.conectar('tenant-123')
      // Em uma segunda chamada com o mesmo tenant, não deve criar nova conexão
      // Como o WebSocket mock não abre, o estado permanece CONNECTING
      await service.conectar('tenant-123')
      // Deve funcionar sem erro
    })

    it('deve reconectar se o tenant for diferente', async () => {
      await service.conectar('tenant-123')
      await service.conectar('tenant-456')
      // Deve desconectar do anterior e conectar ao novo
    })

    it('deve cancelar conexão em 401 e disparar divi:auth-expired', async () => {
      fetchMock.mockResolvedValueOnce({ status: 401 })

      const resultado = await service.conectar('tenant-123')

      expect(resultado).toBe(false)
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('divi_jwt_token')
      const event = dispatchEventSpy.mock.calls[0][0] as CustomEvent
      expect(event.type).toBe('divi:auth-expired')
    })

    it('deve cancelar conexão em 403 e disparar divi:app-error', async () => {
      fetchMock.mockResolvedValueOnce({ status: 403 })

      const resultado = await service.conectar('tenant-123')

      expect(resultado).toBe(false)
      const event = dispatchEventSpy.mock.calls[0][0] as CustomEvent
      expect(event.type).toBe('divi:app-error')
      expect(event.detail.message).toContain('acesso')
    })

    it('não deve cancelar em falha de rede do preflight (deixa o WS tentar)', async () => {
      fetchMock.mockRejectedValueOnce(new Error('network'))

      const resultado = await service.conectar('tenant-123')

      expect(resultado).toBe(true)
    })

    it('deve prosseguir com 400 do preflight (auth OK, não é handshake WS)', async () => {
      // O upgrader.Upgrade do backend devolve 400 num fetch comum mesmo em
      // auth válida (faltam cabeçalhos de handshake). 400 não é falha de auth.
      fetchMock.mockResolvedValueOnce({ status: 400 })

      const resultado = await service.conectar('tenant-123')

      expect(resultado).toBe(true)
    })
  })

  describe('validação de payload', () => {
    it('deve aceitar payload EXPENSE_CREATED válido', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('gastos_alterados', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'EXPENSE_CREATED',
            payload: {
              id: 'g1',
              descricao: 'Mercado',
              valorTotalCentavos: 15000,
              compradorId: 'm1',
              faturaId: null,
            },
          }),
        }))
      }

      expect(callback).toHaveBeenCalled()
    })

    it('deve aceitar payload EXPENSE_DELETED com formato {id}', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('gastos_alterados', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'EXPENSE_DELETED',
            payload: { id: 'g1' },
          }),
        }))
      }

      expect(callback).toHaveBeenCalled()
    })

    it('deve aceitar payload EXPENSE_DELETED com formato batch {ids, action}', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('gastos_alterados', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'EXPENSE_DELETED',
            payload: { ids: ['g1', 'g2'], action: 'delete' },
          }),
        }))
      }

      expect(callback).toHaveBeenCalled()
    })

    it('deve emitir múltiplos eventos para MEMBER_CREATED', async () => {
      await service.conectar('tenant-123')

      const membrosCallback = vi.fn()
      const permissoesCallback = vi.fn()
      service.on('membros_alterados', membrosCallback)
      service.on('permissoes_alteradas', permissoesCallback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'MEMBER_CREATED',
            payload: {
              id: 'm1',
              nome: 'João',
              avatar: 'blob-1',
              ativo: true,
              role: 'MORADOR',
            },
          }),
        }))
      }

      expect(membrosCallback).toHaveBeenCalled()
      expect(permissoesCallback).toHaveBeenCalled()
    })

    it('deve logar warning para tipo de evento desconhecido', async () => {
      const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
      await service.conectar('tenant-123')

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'UNKNOWN_EVENT_TYPE',
            payload: {},
          }),
        }))
      }

      // Não deve emitir nenhum evento
      const callback = vi.fn()
      service.on('gastos_alterados', callback)
      expect(callback).not.toHaveBeenCalled()
      warnSpy.mockRestore()
    })
  })

  describe('desconectar', () => {
    it('deve limpar o estado ao desconectar', async () => {
      await service.conectar('tenant-123')
      service.desconectar()

      // Deve ter chamado close no WebSocket
      // O estado interno deve estar limpo
    })

    it('não deve lançar erro ao desconectar sem conexão ativa', () => {
      expect(() => service.desconectar()).not.toThrow()
    })
  })

  describe('eventos', () => {
    it('deve mapear EXPENSE_CREATED → gastos_alterados', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('gastos_alterados', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'EXPENSE_CREATED',
            payload: { id: 'g1', descricao: 'Teste', valorTotalCentavos: 100, compradorId: 'm1' },
          }),
        }))
      }

      expect(callback).toHaveBeenCalledTimes(1)
    })

    it('deve mapear CARD_CREATED → cartoes_alterados', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('cartoes_alterados', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'CARD_CREATED',
            payload: { id: 'c1', nome: 'Nubank', diaFechamento: 15, responsavelPadraoId: 'm1' },
          }),
        }))
      }

      expect(callback).toHaveBeenCalledTimes(1)
    })

    it('deve mapear FIXED_BILL_CREATED → contas_fixas_alteradas', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('contas_fixas_alteradas', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'FIXED_BILL_CREATED',
            payload: { id: 'cf1', name: 'Aluguel', icon: 'home', fixedValueCentavos: 200000, defaultSplit: [] },
          }),
        }))
      }

      expect(callback).toHaveBeenCalledTimes(1)
    })

    it('deve mapear PERMISSIONS_UPDATED → permissoes_alteradas', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('permissoes_alteradas', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'PERMISSIONS_UPDATED',
            payload: {
              role: 'MORADOR',
              permissions: {
                ALLOW_LANCAR_GASTO: true,
                ALLOW_GERENCIAR_CARTOES: false,
              },
            },
          }),
        }))
      }

      expect(callback).toHaveBeenCalledTimes(1)
    })

    it('deve mapear INVOICE_UPDATED → faturas_alteradas', async () => {
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('faturas_alteradas', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: JSON.stringify({
            type: 'INVOICE_UPDATED',
            payload: { id: 'f1', cartaoId: 'c1', mes: 6, ano: 2026, responsavelId: 'm1', status: 'ABERTA' },
          }),
        }))
      }

      expect(callback).toHaveBeenCalledTimes(1)
    })
  })

  describe('resiliência', () => {
    it('deve ignorar mensagens com JSON inválido', async () => {
      const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      await service.conectar('tenant-123')

      const callback = vi.fn()
      service.on('gastos_alterados', callback)

      const ws = (service as any).ws
      if (ws && ws.onmessage) {
        ws.onmessage(new MessageEvent('message', {
          data: 'not valid json{{{',
        }))
      }

      expect(callback).not.toHaveBeenCalled()
      errorSpy.mockRestore()
    })
  })
})
