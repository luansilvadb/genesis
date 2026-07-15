import { HttpBaseRepository } from './HttpBaseRepository'
import type { ContaFixa, ContaFixaSplitItem } from '../../entities/ContaFixa'
import type { IContaFixaRepository } from '../IContaFixaRepository'
import {
  ContaFixaFlexibleListResponseSchema,
  ContaFixaResponseSchema,
  CreateContaFixaRequestSchema,
  normalizeFlexibleResponse,
} from '../../../shared/validation/apiSchemas'

interface ContaFixaSplitItemDto {
  membroId: string
  valorCentavos: number
}

interface ContaFixaDto {
  id: string
  name: string
  icon: string
  fixedValueCentavos: number | null
  defaultSplit: ContaFixaSplitItemDto[]
}

export class HttpContaFixaRepository extends HttpBaseRepository implements IContaFixaRepository {
  async listarTodas(): Promise<ContaFixa[]> {
    const raw = await this.validatedRequest(ContaFixaFlexibleListResponseSchema, '/contas-fixas')
    const list = normalizeFlexibleResponse<ContaFixaDto>(raw)
    return list.map(item => ({
      id: item.id,
      name: item.name,
      icon: item.icon,
      fixedValueCentavos: item.fixedValueCentavos,
      defaultSplit: (item.defaultSplit || []).map(s => ({
        membroId: s.membroId,
        valorCentavos: s.valorCentavos,
      }))
    }))
  }

  async salvar(conta: ContaFixa): Promise<void> {
    const body = {
      name: conta.name,
      icon: conta.icon,
      fixedValueCentavos: conta.fixedValueCentavos,
      defaultSplit: conta.defaultSplit.map(s => ({ membroId: s.membroId, valorCentavos: s.valorCentavos }))
    }
    CreateContaFixaRequestSchema.parse(body)
    await this.validatedRequest(ContaFixaResponseSchema, '/contas-fixas', {
      method: 'POST',
      body: JSON.stringify(body)
    })
  }

  async atualizar(id: string, conta: ContaFixa): Promise<void> {
    const body = {
      name: conta.name,
      icon: conta.icon,
      fixedValueCentavos: conta.fixedValueCentavos,
      defaultSplit: conta.defaultSplit.map(d => ({
        membroId: d.membroId,
        valorCentavos: d.valorCentavos
      }))
    }
    await this.validatedRequest(ContaFixaResponseSchema, `/contas-fixas/${id}`, {
      method: 'PUT',
      body: JSON.stringify(body)
    })
  }

  async excluir(id: string): Promise<void> {
    await this.request(`/contas-fixas/${id}`, {
      method: 'DELETE'
    })
  }
}
