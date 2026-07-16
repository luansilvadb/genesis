import { ref, computed, watch } from 'vue'
import { Fatura } from '../models/entities/Fatura'
import { Cartao } from '../models/entities/Cartao'
import { formatarMesAno, NOMES_MESES } from '../shared/utils/meses'
import { obterPeriodoSelecionado, salvarPeriodoSelecionado } from '../shared/utils/periodoStorage'

// ── Pure helpers ──

function criarFaturaVirtual(p: { mes: number; ano: number }, cartaoId: string, responsavelId: string): Fatura {
  return new Fatura({
    id: `${cartaoId}-${p.mes}-${p.ano}`,
    cartaoId,
    periodo: { mes: p.mes, ano: p.ano },
    responsavelId,
    status: 'ABERTA'
  })
}

function findNearestInvoice(faturas: Fatura[], mesAtual: number, anoAtual: number): Fatura | undefined {
  if (faturas.length === 0) return undefined
  return faturas.reduce((melhor, f) => {
    const diffMelhor = Math.abs((melhor.periodo.ano - anoAtual) * 12 + (melhor.periodo.mes - mesAtual))
    const diffAtual = Math.abs((f.periodo.ano - anoAtual) * 12 + (f.periodo.mes - mesAtual))
    return diffAtual < diffMelhor ? f : melhor
  })
}

function buscarFaturasNoPeriodo(
  p: { mes: number; ano: number },
  abertas: Fatura[],
  fechadas: Fatura[]
): Fatura[] {
  const abertasNoPeriodo = abertas.filter(f => f.periodo.mes === p.mes && f.periodo.ano === p.ano)
  const fechadasNoPeriodo = fechadas.filter(f => f.periodo.mes === p.mes && f.periodo.ano === p.ano)
  return [...abertasNoPeriodo, ...fechadasNoPeriodo]
}

function resolverFaturaPix(
  p: { mes: number; ano: number },
  faturasExistentes: Fatura[],
  membros: { id: string }[],
  forceOwner?: string
): Fatura[] {
  const pixExistente = faturasExistentes.find(f => f.cartaoId === 'PIX_DEFAULT_ID')
  if (pixExistente) return [pixExistente]
  const fallbackOwner = forceOwner || (membros.length > 0 ? membros[0].id : 'virtual-member')
  return [criarFaturaVirtual(p, 'PIX_DEFAULT_ID', fallbackOwner)]
}

function resolverFaturasDeCartoes(
  p: { mes: number; ano: number },
  faturasExistentes: Fatura[],
  todosCartoes: Cartao[],
  membros: { id: string }[]
): Fatura[] {
  return todosCartoes.map(cartao => {
    const existente = faturasExistentes.find(f => f.cartaoId === cartao.id)
    if (existente) return existente
    const defaultOwner = cartao.responsavelPadraoId || (membros.length > 0 ? membros[0].id : 'virtual-member')
    return criarFaturaVirtual(p, cartao.id, defaultOwner)
  })
}

function resolverFaturasOrfas(
  faturasExistentes: Fatura[],
  listaAtual: Fatura[],
  todosCartoes: Cartao[]
): Fatura[] {
  const orfas: Fatura[] = []
  for (const fatura of faturasExistentes) {
    if (fatura.cartaoId === 'PIX_DEFAULT_ID') continue
    const jaIncluida = listaAtual.some(f => f.id === fatura.id)
    const cartaoAindaExiste = todosCartoes.some(cartao => cartao.id === fatura.cartaoId)
    if (!jaIncluida && !cartaoAindaExiste) {
      orfas.push(fatura)
    }
  }
  return orfas
}

function sortFaturasPixLast(a: Fatura, b: Fatura): number {
  if (a.cartaoId === 'PIX_DEFAULT_ID') return 1
  if (b.cartaoId === 'PIX_DEFAULT_ID') return -1
  return a.cartaoId.localeCompare(b.cartaoId)
}

export function useDashboardPeriodos(
  getFaturasAbertas: () => Fatura[],
  getFaturasFechadas: () => Fatura[],
  getCartoes: () => Cartao[],
  getMembros: () => { id: string }[],
  emit: (event: 'periodoStatusChanged', isLocked: boolean) => void
) {
  const obterPeriodoInicial = () => {
    const hoje = new Date()
    const mesAtual = hoje.getMonth() + 1
    const anoAtual = hoje.getFullYear()

    const faturasAbertas = getFaturasAbertas() || []
    const faturasFechadas = getFaturasFechadas() || []
    // Última fatura fechada = mais recente (assumindo ordem cronológica de criação)
    const faturaFechadaRecente = faturasFechadas.length > 0
      ? faturasFechadas[faturasFechadas.length - 1]
      : undefined
    const faturaProxima = findNearestInvoice(faturasAbertas, mesAtual, anoAtual)
      ?? faturaFechadaRecente

    const fallback = faturaProxima?.periodo
      ? { mes: faturaProxima.periodo.mes, ano: faturaProxima.periodo.ano }
      : undefined
    return obterPeriodoSelecionado(fallback)
  }

  const periodoSelecionado = ref<{ mes: number; ano: number }>(obterPeriodoInicial())

  watch(periodoSelecionado, (novos) => {
    salvarPeriodoSelecionado(novos)
  }, { deep: true, immediate: true })

  const faturasPeriodoSelecionado = computed(() => {
    const p = periodoSelecionado.value
    const abertas = getFaturasAbertas()
    const fechadas = getFaturasFechadas()
    const faturasExistentes = buscarFaturasNoPeriodo(p, abertas, fechadas)
    const todosCartoes = getCartoes()
    const membros = getMembros()

    if (todosCartoes.length === 0) {
      const pix = resolverFaturaPix(p, faturasExistentes, membros)
      return [...faturasExistentes.filter(f => f.cartaoId !== 'PIX_DEFAULT_ID'), ...pix]
    }

    const faturasCartoes = resolverFaturasDeCartoes(p, faturasExistentes, todosCartoes, membros)
    const pix = resolverFaturaPix(p, faturasExistentes, membros, 'PIX_SYSTEM_OWNER')
    const orfas = resolverFaturasOrfas(faturasExistentes, faturasCartoes, todosCartoes)

    return [...faturasCartoes, ...pix, ...orfas].sort(sortFaturasPixLast)
  })

  const faturaPixPeriodoSelecionado = computed(() => {
    const p = periodoSelecionado.value
    const pix = faturasPeriodoSelecionado.value.find(f => f.cartaoId === 'PIX_DEFAULT_ID')
    if (pix) return pix

    return criarFaturaVirtual(p, 'PIX_DEFAULT_ID', 'PIX_SYSTEM_OWNER')
  })

  const isPixAbertoNoPeriodo = (p: { mes: number; ano: number }) => {
    return getFaturasAbertas().some(f => f.cartaoId === 'PIX_DEFAULT_ID' && f.periodo.mes === p.mes && f.periodo.ano === p.ano)
  }

  const isCartaoFechadoNoPeriodo = (cartaoId: string, p: { mes: number; ano: number }) => {
    return getFaturasFechadas().some(f => f.cartaoId === cartaoId && f.periodo.mes === p.mes && f.periodo.ano === p.ano)
  }

  const verificarPeriodoTrancado = (p: { mes: number; ano: number }): boolean => {
    if (isPixAbertoNoPeriodo(p)) return false

    const cartoes = getCartoes()
    if (cartoes.length > 0) {
      return cartoes.every(cartao => isCartaoFechadoNoPeriodo(cartao.id, p))
    }

    return getFaturasFechadas().some(f => f.periodo.mes === p.mes && f.periodo.ano === p.ano)
  }

  const faturaSelecionadaFechada = computed(() => verificarPeriodoTrancado(periodoSelecionado.value))

  watch(faturaSelecionadaFechada, (isLocked) => {
    emit('periodoStatusChanged', isLocked)
  }, { immediate: true })

  const faturaAtivaVisualizada = computed(() => {
    return faturasPeriodoSelecionado.value[0]
  })

  const listaMesesSeletor = computed(() => {
    const hoje = new Date()
    const mesAtual = hoje.getMonth() + 1
    const anoAtual = hoje.getFullYear()
    const list = []
    // Ordenação: futuro primeiro (+1..+12), depois mês atual (0), depois passado (-1..-12).
    // Quando o período selecionado foi arquivado e o scroll cai no topo,
    // o primeiro item visível sempre é o próximo mês à frente — nunca um passado.
    const offsets: number[] = []
    for (let i = 1; i <= 12; i++) offsets.push(i)   // futuro
    offsets.push(0)                                   // atual
    for (let i = -1; i >= -12; i--) offsets.push(i)  // passado
    for (const offset of offsets) {
      const d = new Date(anoAtual, mesAtual - 1 + offset, 1)
      const mesIdx = d.getMonth() + 1
      const anoIdx = d.getFullYear()
      const estaFechada = getFaturasFechadas().some(f => f.periodo.mes === mesIdx && f.periodo.ano === anoIdx)
      list.push({
        mes: mesIdx,
        ano: anoIdx,
        nome: formatarMesAno(mesIdx, anoIdx),
        status: (estaFechada ? 'FECHADA' : 'ABERTA') as 'FECHADA' | 'ABERTA'
      })
    }
    return list
  })

  const mesesAbertosOpcoes = computed(() => listaMesesSeletor.value.filter(item => item.status === 'ABERTA'))
  const mesesTrancadosOpcoes = computed(() =>
    listaMesesSeletor.value
      .filter(item => item.status === 'FECHADA')
      .sort((a, b) => a.ano !== b.ano ? a.ano - b.ano : a.mes - b.mes)
  )

  const currentMonthName = computed(() => {
    const fat = faturaAtivaVisualizada.value
    if (!fat) return 'Mês'
    return NOMES_MESES[fat.periodo.mes - 1]
  })

  const currentYear = computed(() => {
    const fat = faturaAtivaVisualizada.value
    if (!fat) return 'Atual'
    return fat.periodo.ano.toString()
  })

  const setPeriodoSelecionado = (mes: number, ano: number) => {
    periodoSelecionado.value = { mes, ano }
  }

  return {
    periodoSelecionado,
    setPeriodoSelecionado,
    faturaSelecionadaFechada,
    faturaAtivaVisualizada,
    faturasPeriodoSelecionado,
    faturaPixPeriodoSelecionado,
    listaMesesSeletor,
    mesesAbertosOpcoes,
    mesesTrancadosOpcoes,
    currentMonthName,
    currentYear
  }
}
