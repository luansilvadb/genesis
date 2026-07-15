import { HttpBaseRepository } from './HttpBaseRepository'
import { Gasto, type PaymentMethod, type SplitMode } from '../../entities/Gasto'
import { DivisaoDeGasto } from '../../entities/DivisaoDeGasto'
import { Dinheiro } from '../../entities/Dinheiro'
import type { IGastoRepository } from '../IGastoRepository'
import {
  GastoFlexibleListResponseSchema,
  GastoResponseSchema,
  CreateGastoRequestSchema,
  UpdateGastoRequestSchema,
  DeleteBatchRequestSchema,
  normalizeFlexibleResponse,
} from '../../../shared/validation/apiSchemas'

interface GastoDto {
  id: string
  faturaId: string
  descricao: string
  valorTotalCentavos: number
  compradorId: string
  divisoes?: { membroId: string; valorCentavos: number }[]
  installments?: number
  totalInstallments?: number
  isLoan?: boolean
  borrowerId?: string | null
  recurringBillId?: string | null
  isSettlement?: boolean
  settlementDetails?: Gasto['settlementDetails']
  method?: PaymentMethod
  cardOwnerId?: string | null
  grupoParcelasId?: string | null
  isPrivate?: boolean
  splitMode?: 'EQUAL' | 'INCOME' | 'CUSTOM'
  createdAt?: string
}

const toDomainSplitMode = (splitMode?: GastoDto['splitMode']): SplitMode => {
  if (splitMode === 'EQUAL') return 'equal'
  if (splitMode === 'INCOME') return 'income'
  return 'custom'
}

const toApiSplitMode = (splitMode: SplitMode): NonNullable<GastoDto['splitMode']> => {
  if (splitMode === 'equal') return 'EQUAL'
  if (splitMode === 'income') return 'INCOME'
  return 'CUSTOM'
}

export class HttpGastoRepository extends HttpBaseRepository implements IGastoRepository {
  private mapToEntity(item: GastoDto): Gasto {
    const divisoes = (item.divisoes || []).map(d => new DivisaoDeGasto(
      d.membroId,
      Dinheiro.deCentavos(d.valorCentavos)
    ))

    let settlementDetails = item.settlementDetails
    if (typeof settlementDetails === 'string' && settlementDetails) {
      try {
        settlementDetails = JSON.parse(settlementDetails)
      } catch {
        settlementDetails = null
      }
    }

    return new Gasto({
      id: item.id,
      faturaId: item.faturaId,
      descricao: item.descricao,
      valorTotal: Dinheiro.deCentavos(item.valorTotalCentavos),
      compradorId: item.compradorId,
      divisoes,
      installments: item.installments,
      totalInstallments: item.totalInstallments,
      isLoan: item.isLoan,
      borrowerId: item.borrowerId,
      recurringBillId: item.recurringBillId,
      isSettlement: item.isSettlement,
      settlementDetails,
      method: item.method,
      cardOwner: item.cardOwnerId,
      grupoParcelasId: item.grupoParcelasId,
      isPrivate: item.isPrivate,
      splitMode: toDomainSplitMode(item.splitMode),
      createdAt: item.createdAt
    })
  }

  async buscarPorFatura(faturaId: string): Promise<Gasto[]> {
    const list = await this.listarTodos()
    return list.filter(g => g.faturaId === faturaId)
  }

  async buscarPorId(id: string): Promise<Gasto | null> {
    const list = await this.listarTodos()
    return list.find(g => g.id === id) || null
  }

  private mapToDto(gasto: Gasto) {
    return {
      id: gasto.id,
      faturaId: gasto.faturaId,
      descricao: gasto.descricao,
      valorTotalCentavos: gasto.valorTotal.centavos,
      compradorId: gasto.compradorId,
      installments: gasto.installments,
      totalInstallments: gasto.totalInstallments,
      isLoan: gasto.isLoan,
      borrowerId: gasto.borrowerId,
      recurringBillId: gasto.recurringBillId,
      isSettlement: gasto.isSettlement,
      settlementDetails: gasto.settlementDetails ?? null,
      method: gasto.method,
      cardOwnerId: gasto.cardOwner,
      grupoParcelasId: gasto.grupoParcelasId,
      isPrivate: gasto.isPrivate,
      splitMode: toApiSplitMode(gasto.splitMode),
      createdAt: gasto.createdAt ? gasto.createdAt.toISOString() : undefined,
      divisoes: gasto.divisoes.map(d => ({
        membroId: d.membroId,
        valorCentavos: d.valor.centavos
      }))
    }
  }

  async salvar(gasto: Gasto): Promise<void> {
    const dto = this.mapToDto(gasto)
    CreateGastoRequestSchema.parse(dto)
    await this.validatedRequest(GastoResponseSchema, '/gastos', {
      method: 'POST',
      body: JSON.stringify(dto)
    })
  }

  async atualizar(id: string, gasto: Gasto): Promise<void> {
    const dto = this.mapToDto(gasto)
    UpdateGastoRequestSchema.parse(dto)
    await this.validatedRequest(GastoResponseSchema, `/gastos/${id}`, {
      method: 'PUT',
      body: JSON.stringify(dto)
    })
  }

  async salvarMuitos(gastos: Gasto[]): Promise<void> {
    const dtos = gastos.map(g => {
      const dto = this.mapToDto(g)
      CreateGastoRequestSchema.parse(dto)
      return dto
    })
    await this.request('/gastos/batch', {
      method: 'POST',
      body: JSON.stringify(dtos)
    })
  }

  async excluir(id: string): Promise<void> {
    await this.request(`/gastos/${id}`, {
      method: 'DELETE'
    })
  }

  async excluirMuitos(ids: string[]): Promise<void> {
    DeleteBatchRequestSchema.parse({ ids })
    await this.request('/gastos/delete-batch', {
      method: 'POST',
      body: JSON.stringify({ ids })
    })
  }

  async listarTodos(): Promise<Gasto[]> {
    const raw = await this.validatedRequest(GastoFlexibleListResponseSchema, '/gastos')
    const list = normalizeFlexibleResponse<GastoDto>(raw)
    return list.map(item => this.mapToEntity(item))
  }

}
