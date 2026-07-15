import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useContasFixas } from './useContasFixas'
import { Gasto } from '../models/entities/Gasto'
import { Dinheiro } from '../models/entities/Dinheiro'
import { DivisaoDeGasto } from '../models/entities/DivisaoDeGasto'
import type { ContaFixa } from '../models/entities/ContaFixa'

vi.mock('../shared/container', () => ({
  contaFixaRepository: {
    listarTodas: vi.fn(),
    salvar: vi.fn(),
    atualizar: vi.fn(),
    excluir: vi.fn()
  },
  gastoService: {
    lancarGastoContaFixa: vi.fn(),
    lancarGastoOuEmprestimo: vi.fn(),
    excluirGasto: vi.fn(),
    registrarAcertoNetting: vi.fn(),
    removerAssociacaoContaFixa: vi.fn()
  },
  tenantSessionService: {
    isAuthenticated: vi.fn(() => false)
  }
}))

import { contaFixaRepository, gastoService } from '../shared/container'

describe('useContasFixas', () => {
  let mockContas: ContaFixa[] = []
  const mockGastoService = gastoService as any

  beforeEach(() => {
    localStorage.clear()
    mockContas = []
    vi.clearAllMocks()

    vi.mocked(contaFixaRepository.listarTodas).mockImplementation(async () => mockContas)
    vi.mocked(contaFixaRepository.salvar).mockImplementation(async (conta: any) => {
      const idx = mockContas.findIndex(c => c.id === conta.id)
      if (idx > -1) mockContas[idx] = conta
      else mockContas.push(conta)
    })
    vi.mocked(contaFixaRepository.atualizar).mockImplementation(async (_id: string, conta: any) => {
      const idx = mockContas.findIndex(c => c.id === conta.id)
      if (idx > -1) mockContas[idx] = conta
      else mockContas.push(conta)
    })
    vi.mocked(contaFixaRepository.excluir).mockImplementation(async (id: string) => {
      mockContas = mockContas.filter(c => c.id !== id)
    })

    const { resetar } = useContasFixas()
    resetar()
  })

  const esperarTick = () => new Promise(resolve => setTimeout(resolve, 0))

  it('deve manter a lista de contas vazia ao inicializar se vazio', async () => {
    const { contasFixas, carregarTemplates } = useContasFixas()
    await carregarTemplates()
    expect(contasFixas.value.length).toBe(0)
  })

  it('deve cadastrar, atualizar e remover um template customizado', async () => {
    const { contasFixas, salvarContaFixa, excluirContaFixa, carregarTemplates } = useContasFixas()
    await carregarTemplates()
    
    await salvarContaFixa({
      id: 'new_bill',
      name: 'Academia',
      icon: '💪',
      fixedValueCentavos: 10000,
      defaultSplit: [{ membroId: 'luciana', valorCentavos: 0 }]
    })

    expect(contasFixas.value.some(c => c.id === 'new_bill')).toBe(true)

    await salvarContaFixa({
      id: 'new_bill',
      name: 'Academia VIP',
      icon: '💪',
      fixedValueCentavos: 15000,
      defaultSplit: [{ membroId: 'luciana', valorCentavos: 0 }]
    })
    expect(contasFixas.value.find(c => c.id === 'new_bill')?.name).toBe('Academia VIP')

    await excluirContaFixa('new_bill')
    expect(contasFixas.value.some(c => c.id === 'new_bill')).toBe(false)
  })

  it('deve calcular dinamicamente o status de pagamento baseado em gastos reais', async () => {
    const { verificarStatusPaga } = useContasFixas()
    await esperarTick()
    
    const contaAluguel = {
      id: 'aluguel',
      name: 'Aluguel',
      icon: '🔑',
      fixedValueCentavos: 150000,
      defaultSplit: [{ membroId: 'luciana', valorCentavos: 0 }, { membroId: 'luan', valorCentavos: 0 }, { membroId: 'joao', valorCentavos: 0 }]
    }

    expect(verificarStatusPaga(contaAluguel, [])).toBeNull()

    const gastoAluguel = new Gasto({
      id: 'g1',
      faturaId: 'f1',
      descricao: 'Talão: Aluguel',
      valorTotal: Dinheiro.deReais(1500),
      compradorId: 'luciana',
      divisoes: [new DivisaoDeGasto('luciana', Dinheiro.deReais(1500))],
      recurringBillId: 'aluguel'
    })

    const status = verificarStatusPaga(contaAluguel, [gastoAluguel])
    expect(status).not.toBeNull()
    expect(status?.valorCentavos).toBe(150000)
    expect(status?.pagoPor).toBe('luciana')
  })

  it('deve lancar gasto de conta fixa delegando ao GastoService', async () => {
    const { lancarGastoContaFixa } = useContasFixas()
    await esperarTick()

    const contaAluguel = {
      id: 'aluguel',
      name: 'Aluguel',
      icon: '🔑',
      fixedValueCentavos: 150000,
      defaultSplit: [{ membroId: 'luciana', valorCentavos: 0 }, { membroId: 'luan', valorCentavos: 0 }]
    }

    await lancarGastoContaFixa('f1', contaAluguel, 150000, 'luciana', ['luciana', 'luan'])

    expect(mockGastoService.lancarGastoContaFixa).toHaveBeenCalledWith({
      faturaId: 'f1',
      conta: contaAluguel,
      valorCentavos: 150000,
      compradorId: 'luciana',
      participantes: ['luciana', 'luan']
    })
  })

  it('deve chamar removerAssociacaoContaFixa do GastoService ao excluir um template de conta fixa', async () => {
    const { excluirContaFixa } = useContasFixas()
    await esperarTick()

    await excluirContaFixa('aluguel')
    expect(mockGastoService.removerAssociacaoContaFixa).toHaveBeenCalledWith('aluguel')
  })
})
