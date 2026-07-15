import type { ContaFixa } from '../entities/ContaFixa'

export interface IContaFixaRepository {
  listarTodas(): Promise<ContaFixa[]>
  salvar(conta: ContaFixa): Promise<void>
  atualizar(id: string, conta: ContaFixa): Promise<void>
  excluir(id: string): Promise<void>
}
