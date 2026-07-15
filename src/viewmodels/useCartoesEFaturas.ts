import { ref, computed } from 'vue'
import { Cartao } from '../models/entities/Cartao'
import { Fatura } from '../models/entities/Fatura'
import { Gasto } from '../models/entities/Gasto'
import { cartaoRepository, gastoRepository, faturaService, gastoService } from '../shared/container'

const cartoes = ref<Cartao[]>([])
const faturas = ref<Fatura[]>([])
const gastos = ref<Gasto[]>([])
export const useCartoesEFaturas = () => {
  const sync = async () => {
    const [c, g] = await Promise.all([cartaoRepository.listarTodos(), gastoRepository.listarTodos()])
    cartoes.value = c
    gastos.value = g
    const hoje = new Date()
    faturas.value = await faturaService.assegurarFaturasAbertas(c, hoje.getMonth() + 1, hoje.getFullYear())
  }
  const mutar = async (acao: () => Promise<unknown>) => {
    await acao()
    await sync()
  }

  return {
    cartoes, faturas, gastos,
    faturasAbertas: computed(() => faturas.value.filter(f => f.status === 'ABERTA')),
    faturasFechadas: computed(() => faturas.value.filter(f => f.status !== 'ABERTA')),
    
    inicializar: sync,
    adicionarCartao: (nome: string, diaFechamento: number, responsavelPadraoId: string) => mutar(() => cartaoRepository.salvar(new Cartao({ id: crypto.randomUUID(), nome, diaFechamento, responsavelPadraoId }))),
    excluirCartao: (id: string) => mutar(() => cartaoRepository.excluir(id)),
    fecharFatura: (faturaId: string, responsavelId?: string) => mutar(() => faturaService.fecharFatura(faturaId, responsavelId, new Date())),
    reabrirFatura: (faturaId: string) => mutar(() => faturaService.reabrirFatura(faturaId)),
    atualizarGasto: (gastoId: string, dados: Parameters<typeof gastoService.atualizarGastoCompleto>[1]) => mutar(() => gastoService.atualizarGastoCompleto(gastoId, dados)),
    
    resetar: () => {
      cartoes.value = []
      faturas.value = []
      gastos.value = []
    }
  }
}
