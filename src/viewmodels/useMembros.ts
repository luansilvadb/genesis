import { ref, computed } from 'vue'
import { Membro, type MembroRole } from '../models/entities/Membro'
import type { RolePermissions } from '../models/repositories/IMembroRepository'

import { logger } from '../shared/utils/logger'
import { membroRepository, membroService, tenantSessionService } from '../shared/container'
const membros = ref<Membro[]>([])
const tenantPermissions = ref<Record<string, RolePermissions>>({
  MORADOR: {
    ALLOW_LANCAR_GASTO: true,
    ALLOW_GERENCIAR_CARTOES: true,
    ALLOW_GERENCIAR_CONTAS_FIXAS: true,
    ALLOW_REGISTRAR_NETTING: true,
    ALLOW_VER_AUDIT_LOGS: true,
    ALLOW_FECHAR_PERIODO: true,
    ALLOW_ALTERAR_RENDA: true,
    ALLOW_ALTERAR_NOME: true
  },
  VISUALIZADOR: {
    ALLOW_LANCAR_GASTO: false,
    ALLOW_GERENCIAR_CARTOES: false,
    ALLOW_GERENCIAR_CONTAS_FIXAS: false,
    ALLOW_REGISTRAR_NETTING: false,
    ALLOW_VER_AUDIT_LOGS: false,
    ALLOW_FECHAR_PERIODO: false,
    ALLOW_ALTERAR_RENDA: false,
    ALLOW_ALTERAR_NOME: false
  }
})
const inicializado = ref(false)
let promiseInicializacao: Promise<void> | null = null

export function useMembros() {

  const ativos = computed(() => membros.value.filter(m => m.ativo))

  const currentMembro = computed(() => {
    const currentUserId = tenantSessionService.getCurrentUserId()
    return membros.value.find(m => m.userId === currentUserId)
  })

  const carregar = async () => {
    if (tenantSessionService.isAuthenticated() && !tenantSessionService.getActiveTenantId()) {
      membros.value = []
      inicializado.value = true
      return
    }
    const lista = await membroRepository.listarTodos()
    membros.value = lista

    if (tenantSessionService.getActiveTenantId()) {
      try {
        const perms = await membroRepository.obterPermissions()
        tenantPermissions.value = perms
      } catch (err) {
        logger.error('Erro ao carregar permissões do tenant:', err)
      }
    }

    inicializado.value = true
  }

  const inicializar = async () => {
    if (inicializado.value) return
    if (promiseInicializacao) return promiseInicializacao
    promiseInicializacao = carregar().catch((err) => {
      promiseInicializacao = null
      throw err
    })
    return promiseInicializacao
  }

  const adicionarMembro = async (nome: string, email?: string, password?: string, rendaCentavos?: number) => {
    await membroService.adicionarMembro(nome, email, password, rendaCentavos)
    await carregar()
  }

  const desativarMembro = async (id: string) => {
    await membroService.desativarMembro(id)
    await carregar()
  }

  const ativarMembro = async (id: string) => {
    await membroService.ativarMembro(id)
    await carregar()
  }

  const atualizarRoleMembro = async (id: string, role: MembroRole) => {
    await membroService.atualizarRoleMembro(id, role)
    await carregar()
  }

  const atualizarNomeMembro = async (id: string, nome: string) => {
    await membroService.atualizarNomeMembro(id, nome)
    await carregar()
  }

  const atualizarRendaMembro = async (id: string, rendaCentavos?: number) => {
    await membroService.atualizarRendaMembro(id, rendaCentavos)
    await carregar()
  }

  const atualizarPermissions = async (role: string, perms: Partial<RolePermissions>) => {
    try {
      const updated = await membroRepository.atualizarPermissions(role, perms)
      tenantPermissions.value = updated
    } catch (err) {
      logger.error('Erro ao atualizar permissões do tenant:', err)
      throw err
    }
  }

  return {
    membros,
    ativos,
    currentMembro,
    tenantPermissions,
    adicionarMembro,
    desativarMembro,
    ativarMembro,
    atualizarRoleMembro,
    atualizarNomeMembro,
    atualizarRendaMembro,
    atualizarPermissions,
    inicializar,
    carregar
  }
}
