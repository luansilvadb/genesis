import { Membro } from '../entities/Membro'

export interface RolePermissions {
  ALLOW_LANCAR_GASTO?: boolean
  ALLOW_GERENCIAR_CARTOES?: boolean
  ALLOW_GERENCIAR_CONTAS_FIXAS?: boolean
  ALLOW_REGISTRAR_NETTING?: boolean
  ALLOW_VER_AUDIT_LOGS?: boolean
  ALLOW_FECHAR_PERIODO?: boolean
  ALLOW_ALTERAR_RENDA?: boolean
  ALLOW_ALTERAR_NOME?: boolean
}

export interface MembroPatch {
  nome?: string
  avatar?: string
  ativo?: boolean
  role?: string
  rendaCentavos?: number | null
}

export interface IMembroRepository {
  salvar(membro: Membro, credentials?: { email?: string; password?: string }): Promise<void>
  atualizar(id: string, patch: MembroPatch): Promise<void>
  listarTodos(): Promise<Membro[]>
  buscarPorId(id: string): Promise<Membro | null>
  obterPermissions(): Promise<Record<string, RolePermissions>>
  atualizarPermissions(role: string, permissions: Partial<RolePermissions>): Promise<Record<string, RolePermissions>>
}
