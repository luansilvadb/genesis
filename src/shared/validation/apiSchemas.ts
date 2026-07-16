import { z } from 'zod'

// --- Split Item ---
const SplitItemSchema = z.object({
  membroId: z.string(),
  valorCentavos: z.number().int(),
})

// --- Membro ---
export const MembroResponseSchema = z.object({
  id: z.string(),
  nome: z.string(),
  avatar: z.string(),
  ativo: z.boolean().optional(),
  role: z.string().optional(),
  rendaCentavos: z.number().int().nullish(),
  userId: z.string().nullish(),
  createdAt: z.string().optional(),
})

// --- Cartao ---
export const CartaoResponseSchema = z.object({
  id: z.string(),
  nome: z.string(),
  diaFechamento: z.number().int().min(1).max(31),
  responsavelPadraoId: z.string(),
})

// --- Fatura ---
export const FaturaResponseSchema = z.object({
  id: z.string(),
  cartaoId: z.string(),
  mes: z.number().int().min(1).max(12),
  ano: z.number().int(),
  responsavelId: z.string(),
  status: z.enum(['ABERTA', 'FECHADA']),
  dataPagamentoBanco: z.string().nullish(),
})

// --- Gasto ---
export const GastoResponseSchema = z.object({
  id: z.string(),
  faturaId: z.string().nullish(),
  descricao: z.string(),
  valorTotalCentavos: z.number().int(),
  compradorId: z.string(),
  divisoes: z.array(SplitItemSchema).optional(),
  installments: z.number().int().optional(),
  totalInstallments: z.number().int().optional(),
  isLoan: z.boolean().optional(),
  borrowerId: z.string().nullish(),
  recurringBillId: z.string().nullish(),
  isSettlement: z.boolean().optional(),
  settlementDetails: z.unknown().nullish(),
  method: z.string().optional(),
  cardOwnerId: z.string().nullish(),
  grupoParcelasId: z.string().nullish(),
  isPrivate: z.boolean().optional(),
  splitMode: z.string().optional(),
  createdAt: z.string().optional(),
})

// --- Conta Fixa ---
const ContaFixaSplitItemSchema = z.object({
  membroId: z.string(),
  valorCentavos: z.number().int(),
})

export const ContaFixaResponseSchema = z.object({
  id: z.string(),
  name: z.string(),
  icon: z.string(),
  fixedValueCentavos: z.number().int().nullish(),
  defaultSplit: z.array(ContaFixaSplitItemSchema),
  createdAt: z.string().optional(),
})

// --- Auth ---
export const AuthResponseSchema = z.object({
  token: z.string(),
  user: z.object({
    id: z.string(),
    email: z.string().email(),
    nome: z.string(),
  }),
})

export const SessionResponseSchema = z.object({
  tenants: z.array(z.object({
    id: z.string(),
    name: z.string(),
    inviteCode: z.string(),
  })),
})

// --- Paginated ---
const PaginatedResponseSchema = <T extends z.ZodTypeAny>(itemSchema: T) =>
  z.object({
    data: z.array(itemSchema),
    total: z.number().int(),
    page: z.number().int(),
    page_size: z.number().int(),
    totalPages: z.number().int(),
  })

/**
 * Schema flexível que aceita tanto uma resposta array direta quanto
 * uma resposta paginada. Usado para endpoints GET que podem retornar
 * ambos os formatos dependendo do query param ?paginated=true.
 *
 * Em modo DEV, divergências são logadas no console como warning.
 * Em produção, o dado bruto é retornado mesmo se não validar.
 */
export const FlexibleListResponseSchema = <T extends z.ZodTypeAny>(itemSchema: T) =>
  z.union([
    z.array(itemSchema),
    PaginatedResponseSchema(itemSchema),
  ])

/**
 * Normaliza uma resposta flexível (array | paginado) para array,
 * extraindo .data se for paginado.
 */
export function normalizeFlexibleResponse<T>(
  raw: z.infer<ReturnType<typeof FlexibleListResponseSchema>>,
): T[] {
  if (Array.isArray(raw)) {
    return raw as T[]
  }
  return (raw as { data: T[] }).data
}

// Schemas flexíveis pré-construídos para cada entidade:
export const MembroFlexibleListResponseSchema = FlexibleListResponseSchema(MembroResponseSchema)
export const CartaoFlexibleListResponseSchema = FlexibleListResponseSchema(CartaoResponseSchema)
export const FaturaFlexibleListResponseSchema = FlexibleListResponseSchema(FaturaResponseSchema)
export const GastoFlexibleListResponseSchema = FlexibleListResponseSchema(GastoResponseSchema)
export const ContaFixaFlexibleListResponseSchema = FlexibleListResponseSchema(ContaFixaResponseSchema)

// --- Audit Log ---
export const AuditLogResponseSchema = z.object({
  id: z.string(),
  tenantId: z.string(),
  membroId: z.string(),
  acao: z.string(),
  detalhes: z.string(),
  createdAt: z.string(),
})

export const AuditLogFlexibleListResponseSchema = FlexibleListResponseSchema(AuditLogResponseSchema)

// --- Invite Preview ---
const InvitePreviewMembroSchema = z.object({
  id: z.string(),
  nome: z.string(),
  avatar: z.string(),
})

export const InvitePreviewResponseSchema = z.object({
  id: z.string(),
  name: z.string(),
  membrosDisponiveis: z.array(InvitePreviewMembroSchema),
})

// --- Permissions ---

/** Schema para um único objeto de permissões de role. */
export const RolePermissionsObjectSchema = z.object({
  ALLOW_LANCAR_GASTO: z.boolean().optional(),
  ALLOW_GERENCIAR_CARTOES: z.boolean().optional(),
  ALLOW_GERENCIAR_CONTAS_FIXAS: z.boolean().optional(),
  ALLOW_REGISTRAR_NETTING: z.boolean().optional(),
  ALLOW_VER_AUDIT_LOGS: z.boolean().optional(),
  ALLOW_FECHAR_PERIODO: z.boolean().optional(),
  ALLOW_ALTERAR_RENDA: z.boolean().optional(),
  ALLOW_ALTERAR_NOME: z.boolean().optional(),
})

/** Schema para a resposta completa de permissões (mapa role → permissões). */
export const PermissionsResponseSchema = z.record(
  z.string(),
  RolePermissionsObjectSchema,
)

// ── Request Body Schemas ────────────────────────────────────────────────────
// Validam o que o frontend envia para o backend, garantindo que o contrato
// de request também seja respeitado.

export const CreateGastoRequestSchema = z.object({
  id: z.string().optional(), // enviado pelo frontend, ignorado pelo backend
  descricao: z.string().min(1),
  valorTotalCentavos: z.number().int().min(1),
  compradorId: z.string().min(1),
  faturaId: z.string().nullish(),
  installments: z.number().int().positive().nullish(),
  totalInstallments: z.number().int().positive().nullish(),
  isLoan: z.boolean().nullish(),
  borrowerId: z.string().nullish(),
  method: z.enum(['pix', 'card', 'cash']).nullish(),
  cardOwnerId: z.string().nullish(),
  isPrivate: z.boolean().nullish(),
  isSettlement: z.boolean().nullish(),
  settlementDetails: z.unknown().nullish(),
  grupoParcelasId: z.string().nullish(),
  recurringBillId: z.string().nullish(),
  splitMode: z.enum(['EQUAL', 'INCOME', 'CUSTOM']).nullish(),
  divisoes: z.array(SplitItemSchema).nullish(),
  createdAt: z.string().nullish(),
})

export const UpdateGastoRequestSchema = z.object({
  id: z.string().optional(), // enviado pelo frontend, ignorado pelo backend
  descricao: z.string().nullish(),
  valorTotalCentavos: z.number().int().min(1).nullish(),
  compradorId: z.string().nullish(),
  faturaId: z.string().nullish(),
  installments: z.number().int().positive().nullish(),
  totalInstallments: z.number().int().positive().nullish(),
  isLoan: z.boolean().nullish(),
  borrowerId: z.string().nullish(),
  method: z.enum(['pix', 'card', 'cash']).nullish(),
  cardOwnerId: z.string().nullish(),
  isPrivate: z.boolean().nullish(),
  isSettlement: z.boolean().nullish(),
  settlementDetails: z.unknown().nullish(),
  grupoParcelasId: z.string().nullish(),
  recurringBillId: z.string().nullish(),
  splitMode: z.enum(['EQUAL', 'INCOME', 'CUSTOM']).nullish(),
  divisoes: z.array(SplitItemSchema).nullish(),
})

export const CreateMembroRequestSchema = z.object({
  nome: z.string().min(1),
  avatar: z.string().min(1),
})

export const CreateMembroWithAccountRequestSchema = z.object({
  nome: z.string().min(1),
  avatar: z.string().min(1),
  email: z.string().email().optional(),
  password: z.string().min(8).max(128).optional(),
})

export const UpdateMembroRequestSchema = z.object({
  nome: z.string().nullish(),
  avatar: z.string().nullish(),
  ativo: z.boolean().nullish(),
  role: z.string().nullish(),
  rendaCentavos: z.number().int().nullish(),
})

export const CreateCartaoRequestSchema = z.object({
  nome: z.string().min(1),
  diaFechamento: z.number().int().min(1).max(31),
  responsavelPadraoId: z.string().min(1),
})

export const CreateFaturaRequestSchema = z.object({
  cartaoId: z.string().min(1),
  mes: z.number().int().min(1).max(12),
  ano: z.number().int(),
  responsavelId: z.string().min(1),
  status: z.enum(['ABERTA', 'FECHADA']),
  dataPagamentoBanco: z.string().nullish(),
})

export const CreateContaFixaRequestSchema = z.object({
  name: z.string().min(1),
  icon: z.string().min(1),
  fixedValueCentavos: z.number().int().nullish(),
  defaultSplit: z.array(ContaFixaSplitItemSchema).nullish(),
})

export const DeleteBatchRequestSchema = z.object({
  ids: z.array(z.string().min(1)).min(1),
})

export const LoginRequestSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
})

export const RegisterRequestSchema = z.object({
  email: z.string().email(),
  nome: z.string().min(1),
  password: z.string().min(8).max(128),
  inviteCode: z.string().nullish(),
  membroId: z.string().nullish(),
})

export const CreateTenantRequestSchema = z.object({
  name: z.string().min(1),
})

export const JoinTenantRequestSchema = z.object({
  inviteCode: z.string().min(1),
})
