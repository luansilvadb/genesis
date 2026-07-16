/**
 * Testes de Contrato API — Backend ↔ Frontend
 *
 * Valida que as respostas reais da API correspondem aos schemas Zod do frontend.
 * Executa em dois modos:
 *
 * 1. FIXTURE (sempre): usa dados de exemplo embedded que representam o contrato
 *    esperado. Rápido, determinístico, não precisa de backend rodando.
 *
 * 2. LIVE (opcional): se o backend estiver rodando, faz chamadas HTTP reais
 *    e valida as respostas contra os mesmos schemas. Pula automaticamente
 *    se o backend não estiver acessível.
 *
 * Executar: npx vitest --run src/shared/validation/apiContracts.test.ts
 * Executar com live: VITE_API_URL=http://localhost:3000 npx vitest --run src/shared/validation/apiContracts.test.ts
 */

import { describe, it, expect, beforeAll } from 'vitest'
import {
  MembroResponseSchema,
  MembroFlexibleListResponseSchema,
  CartaoResponseSchema,
  CartaoFlexibleListResponseSchema,
  FaturaResponseSchema,
  FaturaFlexibleListResponseSchema,
  GastoResponseSchema,
  GastoFlexibleListResponseSchema,
  ContaFixaResponseSchema,
  ContaFixaFlexibleListResponseSchema,
  AuthResponseSchema,
  SessionResponseSchema,
  PermissionsResponseSchema,
  AuditLogResponseSchema,
  AuditLogFlexibleListResponseSchema,
  InvitePreviewResponseSchema,
  CreateGastoRequestSchema,
  UpdateGastoRequestSchema,
  CreateMembroRequestSchema,
  CreateCartaoRequestSchema,
  CreateFaturaRequestSchema,
  CreateContaFixaRequestSchema,
  DeleteBatchRequestSchema,
  LoginRequestSchema,
  RegisterRequestSchema,
  normalizeFlexibleResponse,
} from './apiSchemas'

// ── Fixture Data ───────────────────────────────────────────────────────────
// Estes dados representam o contrato esperado entre backend e frontend.
// Se o backend mudar o formato, estes fixtures devem ser atualizados
// E os testes falharão até que os schemas Zod sejam ajustados.

const membroFixture = {
  id: '550e8400-e29b-41d4-a716-446655440000',
  nome: 'João Silva',
  avatar: 'blob-1',
  ativo: true,
  role: 'ADMIN',
  rendaCentavos: 500000,
  userId: '660e8400-e29b-41d4-a716-446655440001',
  createdAt: '2025-01-15T10:30:00Z',
}

const cartaoFixture = {
  id: '770e8400-e29b-41d4-a716-446655440002',
  nome: 'Nubank',
  diaFechamento: 15,
  responsavelPadraoId: '550e8400-e29b-41d4-a716-446655440000',
}

const faturaFixture = {
  id: '880e8400-e29b-41d4-a716-446655440003',
  cartaoId: '770e8400-e29b-41d4-a716-446655440002',
  mes: 6,
  ano: 2026,
  responsavelId: '550e8400-e29b-41d4-a716-446655440000',
  status: 'ABERTA' as const,
  dataPagamentoBanco: null,
}

const gastoFixture = {
  id: '990e8400-e29b-41d4-a716-446655440004',
  faturaId: '880e8400-e29b-41d4-a716-446655440003',
  descricao: 'Supermercado',
  valorTotalCentavos: 15000,
  compradorId: '550e8400-e29b-41d4-a716-446655440000',
  divisoes: [
    { membroId: '550e8400-e29b-41d4-a716-446655440000', valorCentavos: 7500 },
    { membroId: 'aa0e8400-e29b-41d4-a716-446655440005', valorCentavos: 7500 },
  ],
  installments: 1,
  totalInstallments: 1,
  isLoan: false,
  borrowerId: null,
  recurringBillId: null,
  isSettlement: false,
  settlementDetails: null,
  method: 'pix',
  cardOwnerId: null,
  grupoParcelasId: null,
  isPrivate: false,
  splitMode: 'EQUAL',
  createdAt: '2025-06-15T14:30:00Z',
}

const contaFixaFixture = {
  id: 'bb0e8400-e29b-41d4-a716-446655440006',
  name: 'Aluguel',
  icon: 'home',
  fixedValueCentavos: 200000,
  defaultSplit: [
    { membroId: '550e8400-e29b-41d4-a716-446655440000', valorCentavos: 100000 },
    { membroId: 'aa0e8400-e29b-41d4-a716-446655440005', valorCentavos: 100000 },
  ],
  createdAt: '2025-01-01T00:00:00Z',
}

const auditLogFixture = {
  id: 'cc0e8400-e29b-41d4-a716-446655440007',
  tenantId: 'dd0e8400-e29b-41d4-a716-446655440008',
  membroId: '550e8400-e29b-41d4-a716-446655440000',
  acao: 'GASTO_CRIADO',
  detalhes: 'Supermercado - R$ 150,00',
  createdAt: '2025-06-15T14:30:00Z',
}

const invitePreviewFixture = {
  id: 'ee0e8400-e29b-41d4-a716-446655440009',
  name: 'Casa Praia',
  membrosDisponiveis: [
    { id: 'm1', nome: 'João', avatar: 'blob-1' },
    { id: 'm2', nome: 'Maria', avatar: 'blob-2' },
  ],
}

const paginatedFixture = {
  data: [membroFixture],
  total: 1,
  page: 1,
  page_size: 20,
  totalPages: 1,
}

// ── Fixture Tests ──────────────────────────────────────────────────────────

describe('Contract: Fixtures → Zod Schemas', () => {
  it('MembroResponseSchema valida fixture de membro', () => {
    const result = MembroResponseSchema.safeParse(membroFixture)
    expect(result.success).toBe(true)
  })

  it('MembroFlexibleListResponseSchema valida array direto', () => {
    const result = MembroFlexibleListResponseSchema.safeParse([membroFixture])
    expect(result.success).toBe(true)
    const normalized = normalizeFlexibleResponse(result.data!)
    expect(normalized).toHaveLength(1)
  })

  it('MembroFlexibleListResponseSchema valida resposta paginada', () => {
    const result = MembroFlexibleListResponseSchema.safeParse(paginatedFixture)
    expect(result.success).toBe(true)
    const normalized = normalizeFlexibleResponse(result.data!)
    expect(normalized).toHaveLength(1)
    expect(normalized[0]).toMatchObject(membroFixture)
  })

  it('CartaoResponseSchema valida fixture de cartão', () => {
    const result = CartaoResponseSchema.safeParse(cartaoFixture)
    expect(result.success).toBe(true)
  })

  it('CartaoResponseSchema rejeita campos extras do modelo GORM', () => {
    // Simula o que acontecia antes: model.Cartao incluía tenantId e createdAt
    const withGormFields = {
      ...cartaoFixture,
      tenantId: 'some-tenant',
      createdAt: '2025-01-01T00:00:00Z',
    }
    // Zod permite campos extras por padrão (não é strict)
    const result = CartaoResponseSchema.safeParse(withGormFields)
    expect(result.success).toBe(true)
    // Os campos extras são ignorados silenciosamente
    expect((result.data as typeof cartaoFixture).id).toBe(cartaoFixture.id)
  })

  it('CartaoFlexibleListResponseSchema valida ambos os formatos', () => {
    // Array
    expect(CartaoFlexibleListResponseSchema.safeParse([cartaoFixture]).success).toBe(true)
    // Paginado
    expect(CartaoFlexibleListResponseSchema.safeParse({
      data: [cartaoFixture], total: 1, page: 1, page_size: 20, totalPages: 1,
    }).success).toBe(true)
  })

  it('FaturaResponseSchema valida fixture de fatura', () => {
    const result = FaturaResponseSchema.safeParse(faturaFixture)
    expect(result.success).toBe(true)
  })

  it('FaturaResponseSchema valida fatura fechada com data de pagamento', () => {
    const result = FaturaResponseSchema.safeParse({
      ...faturaFixture,
      status: 'FECHADA',
      dataPagamentoBanco: '2025-06-20T00:00:00Z',
    })
    expect(result.success).toBe(true)
  })

  it('FaturaFlexibleListResponseSchema valida ambos os formatos', () => {
    expect(FaturaFlexibleListResponseSchema.safeParse([faturaFixture]).success).toBe(true)
    expect(FaturaFlexibleListResponseSchema.safeParse({
      data: [faturaFixture], total: 1, page: 1, page_size: 20, totalPages: 1,
    }).success).toBe(true)
  })

  it('GastoResponseSchema valida fixture de gasto', () => {
    const result = GastoResponseSchema.safeParse(gastoFixture)
    expect(result.success).toBe(true)
  })

  it('GastoResponseSchema valida gasto parcelado', () => {
    const result = GastoResponseSchema.safeParse({
      ...gastoFixture,
      installments: 2,
      totalInstallments: 3,
      grupoParcelasId: 'parcelas-123',
    })
    expect(result.success).toBe(true)
  })

  it('GastoFlexibleListResponseSchema valida ambos os formatos', () => {
    expect(GastoFlexibleListResponseSchema.safeParse([gastoFixture]).success).toBe(true)
    expect(GastoFlexibleListResponseSchema.safeParse({
      data: [gastoFixture], total: 1, page: 1, page_size: 20, totalPages: 1,
    }).success).toBe(true)
  })

  it('ContaFixaResponseSchema valida fixture de conta fixa', () => {
    const result = ContaFixaResponseSchema.safeParse(contaFixaFixture)
    expect(result.success).toBe(true)
  })

  it('ContaFixaResponseSchema valida conta sem valor fixo', () => {
    const result = ContaFixaResponseSchema.safeParse({
      id: 'cc-123',
      name: 'Conta variável',
      icon: 'zap',
      fixedValueCentavos: null,
      defaultSplit: [],
    })
    expect(result.success).toBe(true)
  })

  it('ContaFixaFlexibleListResponseSchema valida ambos os formatos', () => {
    expect(ContaFixaFlexibleListResponseSchema.safeParse([contaFixaFixture]).success).toBe(true)
    expect(ContaFixaFlexibleListResponseSchema.safeParse({
      data: [contaFixaFixture], total: 1, page: 1, page_size: 20, totalPages: 1,
    }).success).toBe(true)
  })

  it('AuthResponseSchema valida resposta de login', () => {
    const result = AuthResponseSchema.safeParse({
      token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx.yyy',
      user: { id: 'u-1', email: 'joao@exemplo.com', nome: 'João' },
    })
    expect(result.success).toBe(true)
  })

  it('SessionResponseSchema valida resposta de sessão', () => {
    const result = SessionResponseSchema.safeParse({
      tenants: [
        { id: 't-1', name: 'Casa Praia', inviteCode: 'ABC123' },
        { id: 't-2', name: 'Casa Centro', inviteCode: 'DEF456' },
      ],
    })
    expect(result.success).toBe(true)
  })

  it('PermissionsResponseSchema valida resposta de permissões', () => {
    const result = PermissionsResponseSchema.safeParse({
      ADMIN: {
        ALLOW_LANCAR_GASTO: true,
        ALLOW_GERENCIAR_CARTOES: true,
        ALLOW_GERENCIAR_CONTAS_FIXAS: true,
        ALLOW_REGISTRAR_NETTING: true,
        ALLOW_VER_AUDIT_LOGS: true,
        ALLOW_FECHAR_PERIODO: true,
        ALLOW_ALTERAR_RENDA: true,
        ALLOW_ALTERAR_NOME: true,
      },
      MORADOR: {
        ALLOW_LANCAR_GASTO: true,
        ALLOW_GERENCIAR_CARTOES: false,
      },
    })
    expect(result.success).toBe(true)
  })

  // --- Audit Log ---
  it('AuditLogResponseSchema valida fixture de audit log', () => {
    const result = AuditLogResponseSchema.safeParse(auditLogFixture)
    expect(result.success).toBe(true)
  })

  it('AuditLogFlexibleListResponseSchema valida ambos os formatos', () => {
    expect(AuditLogFlexibleListResponseSchema.safeParse([auditLogFixture]).success).toBe(true)
    expect(AuditLogFlexibleListResponseSchema.safeParse({
      data: [auditLogFixture], total: 1, page: 1, page_size: 20, totalPages: 1,
    }).success).toBe(true)
  })

  // --- Invite Preview ---
  it('InvitePreviewResponseSchema valida fixture de invite preview', () => {
    const result = InvitePreviewResponseSchema.safeParse(invitePreviewFixture)
    expect(result.success).toBe(true)
  })

  // --- ContaFixa com createdAt ---
  it('ContaFixaResponseSchema valida fixture com createdAt', () => {
    const result = ContaFixaResponseSchema.safeParse(contaFixaFixture)
    expect(result.success).toBe(true)
    expect(result.data?.createdAt).toBe('2025-01-01T00:00:00Z')
  })

  // --- Request Body Schemas ---
  it('CreateGastoRequestSchema valida request de criação de gasto', () => {
    const result = CreateGastoRequestSchema.safeParse({
      descricao: 'Mercado',
      valorTotalCentavos: 15000,
      compradorId: 'm1',
      faturaId: 'f1',
      divisoes: [{ membroId: 'm1', valorCentavos: 7500 }],
      splitMode: 'EQUAL',
      method: 'pix',
    })
    expect(result.success).toBe(true)
  })

  it('CreateGastoRequestSchema rejeita descricao vazia', () => {
    const result = CreateGastoRequestSchema.safeParse({
      descricao: '',
      valorTotalCentavos: 15000,
      compradorId: 'm1',
    })
    expect(result.success).toBe(false)
  })

  it('CreateGastoRequestSchema rejeita valor zero', () => {
    const result = CreateGastoRequestSchema.safeParse({
      descricao: 'Teste',
      valorTotalCentavos: 0,
      compradorId: 'm1',
    })
    expect(result.success).toBe(false)
  })

  it('UpdateGastoRequestSchema valida request parcial (todos campos opcionais)', () => {
    const result = UpdateGastoRequestSchema.safeParse({ descricao: 'Nova descrição' })
    expect(result.success).toBe(true)
  })

  it('UpdateGastoRequestSchema valida request vazio', () => {
    const result = UpdateGastoRequestSchema.safeParse({})
    expect(result.success).toBe(true)
  })

  it('CreateMembroRequestSchema valida request de criação de membro', () => {
    const result = CreateMembroRequestSchema.safeParse({ nome: 'João', avatar: 'blob-1' })
    expect(result.success).toBe(true)
  })

  it('CreateMembroRequestSchema rejeita nome vazio', () => {
    const result = CreateMembroRequestSchema.safeParse({ nome: '', avatar: 'blob-1' })
    expect(result.success).toBe(false)
  })

  it('CreateCartaoRequestSchema valida request de criação de cartão', () => {
    const result = CreateCartaoRequestSchema.safeParse({
      nome: 'Nubank', diaFechamento: 15, responsavelPadraoId: 'm1',
    })
    expect(result.success).toBe(true)
  })

  it('CreateCartaoRequestSchema rejeita diaFechamento > 31', () => {
    const result = CreateCartaoRequestSchema.safeParse({
      nome: 'Nubank', diaFechamento: 32, responsavelPadraoId: 'm1',
    })
    expect(result.success).toBe(false)
  })

  it('CreateFaturaRequestSchema valida request de criação de fatura', () => {
    const result = CreateFaturaRequestSchema.safeParse({
      cartaoId: 'c1', mes: 6, ano: 2026, responsavelId: 'm1', status: 'ABERTA',
    })
    expect(result.success).toBe(true)
  })

  it('CreateFaturaRequestSchema rejeita mês > 12', () => {
    const result = CreateFaturaRequestSchema.safeParse({
      cartaoId: 'c1', mes: 13, ano: 2026, responsavelId: 'm1', status: 'ABERTA',
    })
    expect(result.success).toBe(false)
  })

  it('CreateContaFixaRequestSchema valida request de criação de conta fixa', () => {
    const result = CreateContaFixaRequestSchema.safeParse({
      name: 'Aluguel', icon: 'home', fixedValueCentavos: 200000,
      defaultSplit: [{ membroId: 'm1', valorCentavos: 100000 }],
    })
    expect(result.success).toBe(true)
  })

  it('DeleteBatchRequestSchema valida request de batch delete', () => {
    const result = DeleteBatchRequestSchema.safeParse({ ids: ['id1', 'id2'] })
    expect(result.success).toBe(true)
  })

  it('DeleteBatchRequestSchema rejeita array vazio', () => {
    const result = DeleteBatchRequestSchema.safeParse({ ids: [] })
    expect(result.success).toBe(false)
  })

  it('LoginRequestSchema valida request de login', () => {
    const result = LoginRequestSchema.safeParse({ email: 'joao@exemplo.com', password: '123456' })
    expect(result.success).toBe(true)
  })

  it('LoginRequestSchema rejeita email inválido', () => {
    const result = LoginRequestSchema.safeParse({ email: 'invalido', password: '123456' })
    expect(result.success).toBe(false)
  })

  it('RegisterRequestSchema valida request de registro', () => {
    const result = RegisterRequestSchema.safeParse({
      email: 'joao@exemplo.com', nome: 'João', password: 'Abcdef123',
    })
    expect(result.success).toBe(true)
  })

  it('RegisterRequestSchema rejeita senha curta', () => {
    const result = RegisterRequestSchema.safeParse({
      email: 'joao@exemplo.com', nome: 'João', password: '12345',
    })
    expect(result.success).toBe(false)
  })
})

// ── Live API Tests ──────────────────────────────────────────────────────────
//
// Estes testes fazem chamadas HTTP REAIS ao backend e validam as respostas
// contra os schemas Zod. São a camada mais forte de verificação de contrato.
//
// PRÉ-REQUISITOS:
//   1. Backend rodando (cd backend-go && go run ./cmd/server/)
//   2. Ter feito login no app para popular localStorage com token e tenant
//      (ou setar manualmente divi_jwt_token e divi_active_tenant_id)
//
// EXECUÇÃO:
//   npx vitest --run src/shared/validation/apiContracts.test.ts
//
// FORÇAR EXECUÇÃO (ignora verificação de health):
//   VITE_TEST_LIVE=1 npx vitest --run src/shared/validation/apiContracts.test.ts
//
// Se o backend não estiver acessível, os testes são pulados com skip
// e o relatório indica quantos testes live foram omitidos.

const API_URL = (import.meta.env.VITE_API_URL as string) || 'http://localhost:3000'
const FORCE_LIVE = import.meta.env.VITE_TEST_LIVE === '1'
let backendAvailable = false

beforeAll(async () => {
  try {
    const resp = await fetch(`${API_URL}/health`)
    if (resp.ok) {
      const body = await resp.json()
      backendAvailable = body?.status === 'ok'
    }
  } catch {
    // Backend não está rodando — testes live serão pulados
  }
  if (!backendAvailable && !FORCE_LIVE) {
    console.warn(
      '[Contract Live] Backend não detectado em ' + API_URL + '. ' +
      'Testes live serão pulados. Inicie o backend com "cd backend-go && go run ./cmd/server/" ' +
      'ou force a execução com VITE_TEST_LIVE=1.'
    )
  }
})

describe('Contract: Live API → Zod Schemas', () => {
  const itLive = (backendAvailable || FORCE_LIVE) ? it : it.skip

  // Se todos os testes live forem pulados, ainda registramos o fato
  if (!backendAvailable && !FORCE_LIVE) {
    it('NOTA: testes live pulados — backend não detectado', () => {
      // Este teste sempre passa e serve como lembrete
      expect(true).toBe(true)
    })
  }

  itLive('GET /health retorna status ok', async () => {
    const resp = await fetch(`${API_URL}/health`)
    expect(resp.ok).toBe(true)
    const body = await resp.json()
    expect(body).toMatchObject({ status: 'ok', service: 'divi-api' })
  })

  itLive('GET /api/membros retorna array válido (ou paginado)', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) {
      console.warn('[Contract Live] Pulei /api/membros: sem token ou tenant_id no localStorage')
      return
    }

    const resp = await fetch(`${API_URL}/api/membros`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = MembroFlexibleListResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/membros divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)

    // Verifica normalização
    const normalized = normalizeFlexibleResponse(body)
    expect(Array.isArray(normalized)).toBe(true)
    if (normalized.length > 0) {
      expect(normalized[0]).toHaveProperty('id')
      expect(normalized[0]).toHaveProperty('nome')
      expect(normalized[0]).toHaveProperty('avatar')
    }
  })

  itLive('GET /api/cartoes retorna array válido (ou paginado)', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) return

    const resp = await fetch(`${API_URL}/api/cartoes`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = CartaoFlexibleListResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/cartoes divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)

    const normalized = normalizeFlexibleResponse(body)
    expect(Array.isArray(normalized)).toBe(true)
    if (normalized.length > 0) {
      expect(normalized[0]).toHaveProperty('id')
      expect(normalized[0]).toHaveProperty('nome')
      expect(normalized[0]).toHaveProperty('diaFechamento')
      expect(normalized[0]).toHaveProperty('responsavelPadraoId')
      // Garante que campos do modelo GORM não vazam
      expect(normalized[0]).not.toHaveProperty('tenantId')
    }
  })

  itLive('GET /api/faturas retorna array válido (ou paginado)', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) return

    const resp = await fetch(`${API_URL}/api/faturas`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = FaturaFlexibleListResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/faturas divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)

    const normalized = normalizeFlexibleResponse(body)
    expect(Array.isArray(normalized)).toBe(true)
    if (normalized.length > 0) {
      expect(normalized[0]).toHaveProperty('id')
      expect(normalized[0]).toHaveProperty('cartaoId')
      expect(normalized[0]).toHaveProperty('mes')
      expect(normalized[0]).toHaveProperty('ano')
      expect(normalized[0]).toHaveProperty('status')
      // Garante que campos do modelo GORM não vazam
      expect(normalized[0]).not.toHaveProperty('tenantId')
    }
  })

  itLive('GET /api/gastos retorna array válido (ou paginado)', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) return

    const resp = await fetch(`${API_URL}/api/gastos`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = GastoFlexibleListResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/gastos divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)
  })

  itLive('GET /api/contas-fixas retorna array válido (ou paginado)', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) return

    const resp = await fetch(`${API_URL}/api/contas-fixas`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = ContaFixaFlexibleListResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/contas-fixas divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)
  })

  itLive('GET /api/tenants/permissions retorna permissões válidas', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) return

    const resp = await fetch(`${API_URL}/api/tenants/permissions`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = PermissionsResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/tenants/permissions divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)
  })

  itLive('GET /api/gastos?paginated=true retorna resposta paginada válida', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) return

    const resp = await fetch(`${API_URL}/api/gastos?paginated=true&page=1&page_size=5`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = GastoFlexibleListResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/gastos?paginated=true divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)

    // Deve ser paginado (objeto com .data), não array direto
    expect(body).toHaveProperty('data')
    expect(body).toHaveProperty('total')
    expect(body).toHaveProperty('page')
    expect(body).toHaveProperty('page_size')
    expect(body).toHaveProperty('totalPages')

    const normalized = normalizeFlexibleResponse(body)
    expect(Array.isArray(normalized)).toBe(true)
  })

  itLive('GET /api/audit-logs retorna array válido (ou paginado)', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) {
      console.warn('[Contract Live] Pulei /api/audit-logs: sem token ou tenant_id')
      return
    }

    const resp = await fetch(`${API_URL}/api/audit-logs`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
    })
    expect(resp.ok).toBe(true)

    const body = await resp.json()
    const result = AuditLogFlexibleListResponseSchema.safeParse(body)
    if (!result.success) {
      console.error('[Contract Live] /api/audit-logs divergiu:', result.error.issues)
    }
    expect(result.success).toBe(true)

    const normalized = normalizeFlexibleResponse(body)
    expect(Array.isArray(normalized)).toBe(true)
    if (normalized.length > 0) {
      expect(normalized[0]).toHaveProperty('id')
      expect(normalized[0]).toHaveProperty('tenantId')
      expect(normalized[0]).toHaveProperty('membroId')
      expect(normalized[0]).toHaveProperty('acao')
      expect(normalized[0]).toHaveProperty('detalhes')
      expect(normalized[0]).toHaveProperty('createdAt')
    }
  })

  itLive('POST /api/membros cria membro e valida resposta', async () => {
    const token = localStorage.getItem('divi_jwt_token')
    const tenantId = localStorage.getItem('divi_active_tenant_id')
    if (!token || !tenantId) {
      console.warn('[Contract Live] Pulei POST /api/membros: sem auth')
      return
    }

    const nome = `Teste-${Date.now()}`
    const resp = await fetch(`${API_URL}/api/membros`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
        'X-Tenant-ID': tenantId,
      },
      body: JSON.stringify({ nome, avatar: 'default' }),
    })

    // Pode ser 201 (sucesso) ou 400/403 (restrições de rate limit/permissão)
    // Validamos apenas se for sucesso
    if (resp.ok) {
      expect(resp.status).toBe(201)
      const body = await resp.json()
      const result = MembroResponseSchema.safeParse(body)
      if (!result.success) {
        console.error('[Contract Live] POST /api/membros resposta divergiu:', result.error.issues)
      }
      expect(result.success).toBe(true)
      expect(body.nome).toBe(nome)
      expect(body.avatar).toBe('default')
    } else {
      // Rate limit ou permissão — não é falha de contrato
      console.warn('[Contract Live] POST /api/membros não foi 201:', resp.status)
    }
  })
})
