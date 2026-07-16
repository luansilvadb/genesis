import { Gasto } from '../entities/Gasto'
import { valorParcelaAtual } from '../entities/ParcelaCalculator'

export interface TransferenciaNetting {
  from: string
  to: string
  val: number
}

// ── Helpers privados para cada tipo de gasto ──

function processarGastoEmprestimo(g: Gasto, saldos: Record<string, number>): void {
  const vp = valorParcelaAtual(g.valorTotal, g.installments, g.totalInstallments)
  if (vp.centavos === 0) return
  const c = vp.centavos
  if (g.compradorId) saldos[g.compradorId] = (saldos[g.compradorId] || 0) + c
  if (g.borrowerId) saldos[g.borrowerId] = (saldos[g.borrowerId] || 0) - c
}

function processarGastoAcerto(g: Gasto, saldos: Record<string, number>): void {
  if (!g.settlementDetails) return
  const vp = valorParcelaAtual(g.valorTotal, g.installments, g.totalInstallments)
  if (vp.centavos === 0) return
  saldos[g.settlementDetails.fromMemberId] = (saldos[g.settlementDetails.fromMemberId] || 0) + vp.centavos
  saldos[g.settlementDetails.toMemberId] = (saldos[g.settlementDetails.toMemberId] || 0) - vp.centavos
}

function processarGastoComum(g: Gasto, saldos: Record<string, number>): void {
  const pagadorId = (g.method === 'card' && g.cardOwner) ? g.cardOwner : g.compradorId
  let totalDebitos = 0
  g.divisoes.forEach(div => {
    const valorDebito = valorParcelaAtual(div.valor, g.installments, g.totalInstallments)
    if (valorDebito.centavos === 0) return
    const c = valorDebito.centavos
    saldos[div.membroId] = (saldos[div.membroId] || 0) - c
    totalDebitos += c
  })
  if (pagadorId) saldos[pagadorId] = (saldos[pagadorId] || 0) + totalDebitos
}

// ── API pública ──

/**
 * Calcula saldos unificados de crédito/débito entre membros a partir dos gastos.
 * Função pura sem dependência de framework.
 */
export function calcularSaldosUnificados(
  membros: { id: string }[],
  gastos: Gasto[]
): Record<string, number> {
  const saldos: Record<string, number> = {}
  membros.forEach(m => { saldos[m.id] = 0 })

  for (const g of gastos) {
    if (g.isPrivate) continue
    if (g.isLoan) processarGastoEmprestimo(g, saldos)
    else if (g.isSettlement) processarGastoAcerto(g, saldos)
    else processarGastoComum(g, saldos)
  }

  return saldos
}

/**
 * Algoritmo guloso de netting: gera o conjunto mínimo de transferências
 * para zerar os saldos entre membros trabalhando em centavos inteiros absolutos.
 */
export function calcularTransacoesNetting(saldosCentavos: Record<string, number>): TransferenciaNetting[] {
  const creditors: { id: string; val: number }[] = []
  const debtors: { id: string; val: number }[] = []

  for (const mId in saldosCentavos) {
    const val = saldosCentavos[mId]
    if (val > 0) {
      creditors.push({ id: mId, val })
    } else if (val < 0) {
      debtors.push({ id: mId, val: -val })
    }
  }

  creditors.sort((a, b) => b.val - a.val)
  debtors.sort((a, b) => b.val - a.val)

  const transferencias: TransferenciaNetting[] = []
  let cIdx = 0
  let dIdx = 0

  while (cIdx < creditors.length && dIdx < debtors.length) {
    const creditor = creditors[cIdx]
    const debtor = debtors[dIdx]
    const amount = Math.min(creditor.val, debtor.val)

    if (amount > 0) {
      transferencias.push({
        from: debtor.id,
        to: creditor.id,
        val: amount / 100 // Exposto em Reais para a visualização externa
      })
    }

    creditor.val -= amount
    debtor.val -= amount

    if (creditor.val === 0) cIdx++
    if (debtor.val === 0) dIdx++
  }

  return transferencias
}
