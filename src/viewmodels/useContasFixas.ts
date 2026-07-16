import { ref } from 'vue'
import type { ContaFixa } from '../models/entities/ContaFixa'
import type { Gasto } from '../models/entities/Gasto'

import { logger } from '../shared/utils/logger'
import { contaFixaRepository, gastoService, tenantSessionService } from '../shared/container'

const contasFixas = ref<ContaFixa[]>([])
let promiseInicializacao: Promise<void> | null = null

export function useContasFixas() {

  const carregarTemplates = async () => {
    if (promiseInicializacao) {
      return promiseInicializacao
    }

    if (tenantSessionService.isAuthenticated() && !tenantSessionService.getActiveTenantId()) {
      contasFixas.value = []
      return
    }

    const carregar = async () => {
      try {
        const saved = await contaFixaRepository.listarTodas()
        contasFixas.value = saved || []
      } catch (e) {
        logger.error('Erro ao carregar contas fixas:', e)
      }
    }

    promiseInicializacao = carregar()
    try {
      await promiseInicializacao
    } finally {
      promiseInicializacao = null
    }
  }

  const salvarContaFixa = async (template: ContaFixa) => {
    const idx = contasFixas.value.findIndex(c => c.id === template.id)
    if (idx > -1) {
      contasFixas.value[idx] = template
      await contaFixaRepository.atualizar(template.id, template)
    } else {
      contasFixas.value.push(template)
      await contaFixaRepository.salvar(template)
    }
  }

  const excluirContaFixa = async (id: string) => {
    contasFixas.value = contasFixas.value.filter(c => c.id !== id)
    await contaFixaRepository.excluir(id)
    await gastoService.removerAssociacaoContaFixa(id)
  }

  const verificarStatusPaga = (conta: ContaFixa, gastos: Gasto[]) => {
    const gasto = gastos.find(g => g.recurringBillId === conta.id)
    if (!gasto) return null
    return {
      valorCentavos: gasto.valorTotal.centavos,
      pagoPor: gasto.compradorId
    }
  }

  const lancarGastoContaFixa = async (
    faturaId: string,
    conta: ContaFixa,
    valorCentavos: number,
    compradorId: string,
    participantes: string[]
  ) => {
    await gastoService.lancarGastoContaFixa({
      faturaId,
      conta,
      valorCentavos,
      compradorId,
      participantes
    })
  }

  const resetar = () => {
    contasFixas.value = []
  }

  return {
    contasFixas,
    salvarContaFixa,
    excluirContaFixa,
    verificarStatusPaga,
    lancarGastoContaFixa,
    carregarTemplates,
    resetar
  }
}
