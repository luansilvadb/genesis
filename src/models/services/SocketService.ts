import { z } from 'zod'
import { logger } from '../../shared/utils/logger'
import {
  GastoResponseSchema,
  CartaoResponseSchema,
  FaturaResponseSchema,
  MembroResponseSchema,
  ContaFixaResponseSchema,
  RolePermissionsObjectSchema,
} from '../../shared/validation/apiSchemas'

interface WSMessage {
  type: string
  payload: unknown
}

/** Tipos de mensagens WebSocket emitidas pelo backend. */
type WSEventType =
  | 'EXPENSE_CREATED'
  | 'EXPENSE_UPDATED'
  | 'EXPENSE_DELETED'
  | 'CARD_CREATED'
  | 'CARD_DELETED'
  | 'INVOICE_UPDATED'
  | 'MEMBER_CREATED'
  | 'MEMBER_UPDATED'
  | 'FIXED_BILL_CREATED'
  | 'FIXED_BILL_UPDATED'
  | 'FIXED_BILL_DELETED'
  | 'PERMISSIONS_UPDATED'

/** Tipos de eventos internos do frontend disparados pelo SocketService. */
type AppEventType =
  | 'gastos_alterados'
  | 'cartoes_alterados'
  | 'faturas_alteradas'
  | 'membros_alterados'
  | 'permissoes_alteradas'
  | 'contas_fixas_alteradas'

// --- Schemas para validação dos payloads WebSocket ---

/** Payload para deleção de um único item. */
const DeletePayloadSchema = z.object({
  id: z.string(),
})

/** Payload para deleção em batch. */
const BatchDeletePayloadSchema = z.object({
  ids: z.array(z.string()),
  action: z.string(),
})

/** Payload para notificação de batch de faturas (criação em lote). */
const FaturaBatchUpdatedPayloadSchema = z.object({
  action: z.literal('batch-updated'),
})

/** Schema união: suporta tanto delete único quanto batch. */
const ExpenseDeletedPayloadSchema = z.union([
  DeletePayloadSchema,
  BatchDeletePayloadSchema,
])

/** Payload para atualização de permissões de uma role. */
const PermissionsUpdatedPayloadSchema = z.object({
  role: z.string(),
  permissions: RolePermissionsObjectSchema,
})

/**
 * Mapeia cada tipo de evento WebSocket ao schema Zod esperado para seu payload.
 * Usado para validar as mensagens recebidas do backend e detectar quebras de contrato.
 */
const WS_PAYLOAD_SCHEMAS: Record<WSEventType, z.ZodType> = {
  EXPENSE_CREATED: GastoResponseSchema,
  EXPENSE_UPDATED: GastoResponseSchema,
  EXPENSE_DELETED: ExpenseDeletedPayloadSchema,
  CARD_CREATED: CartaoResponseSchema,
  CARD_DELETED: DeletePayloadSchema,
  INVOICE_UPDATED: z.union([FaturaResponseSchema, FaturaBatchUpdatedPayloadSchema]),
  MEMBER_CREATED: MembroResponseSchema,
  MEMBER_UPDATED: MembroResponseSchema,
  FIXED_BILL_CREATED: ContaFixaResponseSchema,
  FIXED_BILL_UPDATED: ContaFixaResponseSchema,
  FIXED_BILL_DELETED: DeletePayloadSchema,
  PERMISSIONS_UPDATED: PermissionsUpdatedPayloadSchema,
} as const

const EVENT_MAP: Record<WSEventType, AppEventType[]> = {
  EXPENSE_CREATED: ['gastos_alterados'],
  EXPENSE_UPDATED: ['gastos_alterados'],
  EXPENSE_DELETED: ['gastos_alterados'],
  CARD_CREATED: ['cartoes_alterados'],
  CARD_DELETED: ['cartoes_alterados'],
  INVOICE_UPDATED: ['faturas_alteradas'],
  MEMBER_CREATED: ['membros_alterados', 'permissoes_alteradas'],
  MEMBER_UPDATED: ['membros_alterados', 'permissoes_alteradas'],
  FIXED_BILL_CREATED: ['contas_fixas_alteradas'],
  FIXED_BILL_UPDATED: ['contas_fixas_alteradas'],
  FIXED_BILL_DELETED: ['contas_fixas_alteradas'],
  PERMISSIONS_UPDATED: ['permissoes_alteradas'],
} as const satisfies Record<WSEventType, AppEventType[]>

export class SocketService {
  private ws: WebSocket | null = null
  private currentTenantId: string | null = null
  private listeners = new Map<string, Set<(payload?: unknown) => void>>()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private shouldReconnect = false
  private reconnectAttempts = 0
  private readonly MAX_RECONNECT_DELAY = 30000
  private readonly MAX_RECONNECT_ATTEMPTS = 8

  private get baseUrl(): string {
    const url = (import.meta.env.VITE_API_URL as string) || 'http://localhost:3000'
    return url.replace(/\/+$/, '')
  }

  private get wsUrl(): string {
    const httpUrl = this.baseUrl
    const protocol = httpUrl.startsWith('https') ? 'wss' : 'ws'
    const host = httpUrl.replace(/^https?:\/\//, '')
    // O token JWT NÃO é mais enviado via query parameter (exposto em logs de proxy,
    // Referer headers, etc.). Em vez disso, é enviado como subprotocolo WebSocket
    // (Sec-WebSocket-Protocol), que não é logado por proxies e não aparece na URL.
    return `${protocol}://${host}/ws?tenant_id=${encodeURIComponent(this.currentTenantId || '')}`
  }

  /** Endpoint HTTP espelho do /ws — o backend valida token + membro ANTES do upgrade,
   *  retornando 401/403 com JSON legível em vez do upgrade silencioso falho do navegador. */
  private get preflightUrl(): string {
    const httpUrl = this.baseUrl
    return `${httpUrl}/ws?tenant_id=${encodeURIComponent(this.currentTenantId || '')}`
  }

  /** Retorna o token JWT para autenticação no WebSocket. */
  private get wsToken(): string {
    const token = localStorage.getItem('divi_jwt_token')
    if (!token) {
      throw new Error('Token não disponível para conexão WebSocket')
    }
    return token
  }

  /**
   * Verifica via HTTP (antes de abrir o WS) se a conexão seria autorizada.
   * O handler /ws do backend valida token e membership ANTES de chamar Upgrade,
   * portanto falhas de auth/acesso chegam como 401/403 com JSON, não como o
   * "WebSocket connection failed" opaco do navegador.
   *
   * Retorna 'ok' | 'auth' | 'forbidden' | 'network'.
   */
  private async preflight(tenantId: string): Promise<'ok' | 'auth' | 'forbidden' | 'network'> {
    let token: string
    try {
      token = this.wsToken
    } catch {
      return 'auth'
    }
    const controller = new AbortController()
    const timeout = setTimeout(() => controller.abort(), 8000)
    try {
      const resp = await fetch(this.preflightUrl, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`,
          'X-Tenant-ID': tenantId,
        },
        signal: controller.signal,
      })
      // O handler /ws valida token + membership ANTES de tentar o upgrade.
      // Assim, 401 (token inválido/expirado) e 403 (não-membro) chegam como JSON.
      // Em auth OK, o handler chama upgrader.Upgrade, que sem cabeçalhos de
      // handshake WebSocket num fetch comum devolve 400 — sinal de que auth
      // passou. Apenas 401/403 são falhas permanentes; todo o resto (incl. 400)
      // libera a abertura do WebSocket.
      if (resp.status === 401) return 'auth'
      if (resp.status === 403) return 'forbidden'
      return 'ok'
    } catch {
      return 'network'
    } finally {
      clearTimeout(timeout)
    }
  }

  /**
   * Conecta ao WebSocket do tenant. Faz um preflight HTTP para detectar falhas
   * permanentes (token inválido/expirado → 401, ou usuário não-membro → 403)
   * ANTES de abrir o socket, evitando o loop de reconexão e o erro opaco
   * "WebSocket connection failed" do navegador.
   *
   * Em 401 dispara o evento global `divi:auth-expired` (forçando re-login).
   * Em 403 loga erro claro de acesso e não tenta reconectar.
   * Em falha de rede, abre o WS e usa backoff limitado.
   */
  async conectar(tenantId: string): Promise<boolean> {
    if (this.ws && this.currentTenantId === tenantId && this.ws.readyState === WebSocket.OPEN) {
      return true
    }

    this.desconectar()
    this.currentTenantId = tenantId
    this.shouldReconnect = true

    const check = await this.preflight(tenantId)
    if (check === 'auth') {
      logger.error('[SocketService] Token ausente, inválido ou expirado (401). Disparando re-login.')
      localStorage.removeItem('divi_jwt_token')
      window.dispatchEvent(new CustomEvent('divi:auth-expired'))
      this.shouldReconnect = false
      this.currentTenantId = null
      return false
    }
    if (check === 'forbidden') {
      logger.error('[SocketService] Acesso negado ao núcleo (403): o usuário não é membro ativo deste tenant. Conexão WS cancelada.')
      window.dispatchEvent(new CustomEvent('divi:app-error', {
        detail: { message: 'Você não tem acesso a esta casa. Selecione outra casa ou faça login novamente.' },
      }))
      this.shouldReconnect = false
      this.currentTenantId = null
      return false
    }

    // 'ok' ou 'network' — tenta abrir o WebSocket (rede instável ainda pode conectar)
    this.conectarWebSocket()
    return true
  }

  private conectarWebSocket(): void {
    if (!this.currentTenantId) return

    let token: string
    try {
      token = this.wsToken
    } catch (err) {
      logger.error('[SocketService] Falha ao obter token:', err)
      return
    }

    try {
      this.ws = new WebSocket(this.wsUrl, [`divi.${token}`])

      this.ws.onopen = () => {
        logger.info('[SocketService] Conectado ao WebSocket')
        this.reconnectAttempts = 0
      }

      this.ws.onmessage = (event: MessageEvent) => {
        try {
          const msg: WSMessage = JSON.parse(event.data)

          // Validar o payload contra o schema esperado para este tipo de evento
          this.validatePayload(msg)

          const events = EVENT_MAP[msg.type as WSEventType]
          if (events) {
            for (const eventName of events) {
              this.emit(eventName, msg.payload)
            }
          }
        } catch {
          logger.warn('[SocketService] Mensagem inválida recebida')
        }
      }

      this.ws.onclose = () => {
        logger.info('[SocketService] Conexão fechada')
        this.ws = null
        if (this.shouldReconnect) {
          this.scheduleReconnect()
        }
      }

      this.ws.onerror = (error) => {
        logger.error('[SocketService] Erro na conexão:', error)
      }
    } catch (err) {
      logger.error('[SocketService] Falha ao conectar:', err)
    }
  }

  /** Agenda reconexão com backoff exponencial até o limite de tentativas. */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.MAX_RECONNECT_ATTEMPTS) {
      logger.error(
        `[SocketService] Limite de ${this.MAX_RECONNECT_ATTEMPTS} tentativas de reconexão atingido. Parando.`,
      )
      this.shouldReconnect = false
      return
    }
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), this.MAX_RECONNECT_DELAY)
    this.reconnectAttempts++
    this.reconnectTimer = setTimeout(() => this.conectarWebSocket(), delay)
  }

  /**
   * Valida o payload de uma mensagem WebSocket contra o schema Zod esperado.
   * Em desenvolvimento, loga warnings detalhados quando o contrato é quebrado.
   * Em produção, loga via logger.error para monitoramento.
   *
   * Nunca lança exceção — a validação é puramente observacional para não
   * interromper o fluxo de dados em produção.
   */
  private validatePayload(msg: WSMessage): void {
    const schema = WS_PAYLOAD_SCHEMAS[msg.type as WSEventType]
    if (!schema) {
      // Tipo de evento desconhecido — pode ser novo evento do backend
      logger.warn(`[SocketService] Tipo de evento WS desconhecido: "${msg.type}"`)
      return
    }

    const result = schema.safeParse(msg.payload)
    if (!result.success) {
      const issues = result.error.issues.map(i =>
        `${i.path.join('.') || '(root)'}: ${i.message}`,
      )
      logger.error(
        `[WS Contract] Payload do evento "${msg.type}" não passou na validação:`,
        issues,
      )
      if (import.meta.env.DEV) {
        console.warn(
          `[WS Contract] "${msg.type}" divergiu do schema esperado.`,
          { payload: msg.payload, issues: result.error.issues },
        )
      }
    }
  }

  desconectar(): void {
    this.shouldReconnect = false
    this.reconnectAttempts = 0
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.onclose = null
      this.ws.close()
      this.ws = null
    }
    this.currentTenantId = null
    // Clear all accumulated event listeners to prevent duplicate callbacks
    // on reconnect or tenant switch.
    this.listeners.clear()
  }

  on(event: string, callback: (payload?: unknown) => void): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
  }

  off(event: string, callback: (payload?: unknown) => void): void {
    const callbacks = this.listeners.get(event)
    if (callbacks) {
      callbacks.delete(callback)
      if (callbacks.size === 0) {
        this.listeners.delete(event)
      }
    }
  }

  private emit(event: string, payload?: unknown): void {
    const callbacks = this.listeners.get(event)
    if (callbacks) {
      for (const cb of callbacks) {
        cb(payload)
      }
    }
  }
}
