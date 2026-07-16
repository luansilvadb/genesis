<!-- src/views/components/ledger/dashboard/PersonalBalancePanel.vue -->
<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowUpRight, ArrowDownLeft, UserCheck } from 'lucide-vue-next'
import { Gasto } from '../../../../models/entities/Gasto'
import { Dinheiro } from '../../../../models/entities/Dinheiro'
import { DivisaoDeGasto } from '../../../../models/entities/DivisaoDeGasto'
import { formatarCentavosParaBRL } from '../../../../shared/utils/formatarMoeda'
import { ExtratoPessoalService } from '../../../../models/services/ExtratoPessoalService'
import { useMembros } from '../../../../viewmodels/useMembros'
import { useToast } from '../../../../composables/useToast'
import { useCartoesEFaturas } from '../../../../viewmodels/useCartoesEFaturas'
import { gastoService } from '../../../../shared/container'
import { obterPeriodoSelecionado } from '../../../../shared/utils/periodoStorage'
import Card from '../../ui/Card.vue'
import ActivityFeed from '../ActivityFeed.vue'
import IllustrationMascot from '../../ui/IllustrationMascot.vue'
import NettingPanel from './NettingPanel.vue'
import BottomSheet from '../../ui/BottomSheet.vue'
import BottomSheetAcertoCompensacao from './BottomSheetAcertoCompensacao.vue'
import type { TransferenciaNetting } from '../../../../models/services/NettingService'
import type { LancarGastoInput } from '../../../../models/services/GastoService'
import { logger } from '../../../../shared/utils/logger'

interface Props {
  membros: { id: string; nome: string }[]
  gastos: Gasto[]
}

const props = defineProps<Props>()
const { currentMembro } = useMembros()
const toast = useToast()
const cartoesEFaturas = useCartoesEFaturas()

const userId = computed(() => currentMembro.value?.id || '')

// Filtrar gastos privados envoltos com o membro logado
const gastosPrivados = computed(() => {
  if (!userId.value) return []
  return props.gastos.filter(g => {
    if (!g.isPrivate) return false
    const envolvidoNasDivisoes = g.divisoes.some(d => d.membroId === userId.value)
    return g.compradorId === userId.value || g.cardOwner === userId.value || g.borrowerId === userId.value || envolvidoNasDivisoes
  })
})

const saldosPessoais = computed(() => {
  if (!userId.value) return []
  return ExtratoPessoalService.calcularSaldosPessoais(userId.value, props.gastos)
})

// Totais
const totalAReceberCentavos = computed(() => {
  return saldosPessoais.value
    .filter(s => s.saldo.centavos > 0)
    .reduce((acc, s) => acc + s.saldo.centavos, 0)
})

const totalAPagarCentavos = computed(() => {
  return saldosPessoais.value
    .filter(s => s.saldo.centavos < 0)
    .reduce((acc, s) => acc + Math.abs(s.saldo.centavos), 0)
})

// Fluxo de Liquidação
const liquidandoPendenciaNetting = ref<TransferenciaNetting | null>(null)
const isSubmittingLiquidador = ref(false)

const fecharLiquidar = () => {
  liquidandoPendenciaNetting.value = null
}

const nettingPessoais = computed<TransferenciaNetting[]>(() => {
  const transferencias: TransferenciaNetting[] = []
  if (!userId.value) return transferencias

  saldosPessoais.value.forEach(ext => {
    if (ext.saldo.centavos === 0) return
    const valorReais = Math.abs(ext.saldo.centavos) / 100
    if (ext.saldo.centavos > 0) {
      transferencias.push({ from: ext.id, to: userId.value!, val: valorReais })
    } else {
      transferencias.push({ from: userId.value!, to: ext.id, val: valorReais })
    }
  })
  return transferencias
})

const getMembroNomeNetting = (id: string) => {
  if (id === userId.value) return currentMembro.value?.nome || 'Você'
  if (id.startsWith('externo:')) return id.substring(8)
  return props.membros.find(m => m.id === id)?.nome || '?'
}

const handleAbrirNetting = (t: TransferenciaNetting) => {
  liquidandoPendenciaNetting.value = t
}

const confirmarLiquidar = async (dados: { from: string; to: string; valor: number; method: 'pix' | 'cash'; descricao: string }) => {
  if (!liquidandoPendenciaNetting.value || !userId.value) return
  isSubmittingLiquidador.value = true
  try {
    const valorReais = dados.valor
    const valorAbsCentavos = Math.round(valorReais * 100)

    const compradorId = dados.from
    const destinatarioId = dados.to

    const data = {
      flow: 'settlement',
      paymentMethod: dados.method,
      compradorId,
      valor: valorReais,
      descricao: dados.descricao,
      divisoes: [
        new DivisaoDeGasto(destinatarioId, Dinheiro.deCentavos(valorAbsCentavos))
      ],
      installments: 1,
      cardOwnerId: null,
      borrowerId: null,
      periodo: obterPeriodoSelecionado(),
      isPrivate: true,
      splitMode: 'custom',
      settlementDetails: {
        fromMemberId: compradorId,
        toMemberId: destinatarioId,
        method: dados.method
      }
    }

    await gastoService.lancarGastoOuEmprestimo(data as LancarGastoInput)
    await cartoesEFaturas.inicializar()
    toast.show(`Acerto registrado com sucesso!`, 'success')
    fecharLiquidar()
  } catch (err) {
    logger.error(err instanceof Error ? err.message : String(err))
    toast.show('Erro ao registrar acerto.', 'error')
  } finally {
    isSubmittingLiquidador.value = false
  }
}

// Estorno e Ajuste de Gastos Privados
const gastoParaEstornar = ref<Gasto | null>(null)

const solicitarEstorno = (g: Gasto) => {
  gastoParaEstornar.value = g
}

const confirmarEstorno = async () => {
  const g = gastoParaEstornar.value
  if (!g) return
  gastoParaEstornar.value = null
  try {
    await gastoService.excluirGasto(g.id)
    await cartoesEFaturas.inicializar()
    toast.show('Estornado com sucesso!', 'success')
  } catch (err) {
    logger.error(err instanceof Error ? err.message : String(err))
    toast.show('Erro ao estornar gasto.', 'error')
  }
}
</script>

<template>
  <div class="space-y-8 animate-in fade-in duration-300">
    <!-- Header e Boas-Vindas Pessoais -->
    <div class="p-6 rounded-3xl bg-parchment/70 border border-stone/50 flex flex-col md:flex-row justify-between items-center gap-6">
      <div class="flex items-center gap-5">
        <IllustrationMascot
          variant="sky"
          :size="80"
          mood="happy"
          class="shrink-0"
        />
        <div>
          <h2 class="font-display text-2xl text-charcoal">
            Controle <span class="text-ember">Pessoal</span>
          </h2>
          <p class="text-[11px] text-graphite uppercase tracking-widest font-semibold mt-1">
            Minhas pendências isoladas e acertos diretos
          </p>
        </div>
      </div>
    </div>

    <!-- Mini Cards de Totais -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <Card class="p-5 flex items-center justify-between bg-white shadow-subtle border-none">
        <div class="space-y-1">
          <span class="text-[10px] font-bold uppercase tracking-widest text-ash">A Receber</span>
          <p :class="['font-display text-2xl', totalAReceberCentavos > 0 ? 'text-meadow' : 'text-graphite']">
            {{ totalAReceberCentavos > 0 ? '+' : '' }}{{ formatarCentavosParaBRL(totalAReceberCentavos) }}
          </p>
        </div>
        <div :class="['w-10 h-10 rounded-xl flex items-center justify-center border', totalAReceberCentavos > 0 ? 'bg-meadow/5 text-meadow border-meadow/10' : 'bg-stone text-graphite border-stone']">
          <ArrowUpRight class="w-5 h-5 stroke-[2.5px]" />
        </div>
      </Card>

      <Card class="p-5 flex items-center justify-between bg-white shadow-subtle border-none">
        <div class="space-y-1">
          <span class="text-[10px] font-bold uppercase tracking-widest text-ash">A Pagar</span>
          <p :class="['font-display text-2xl', totalAPagarCentavos > 0 ? 'text-coral' : 'text-graphite']">
            {{ totalAPagarCentavos > 0 ? '-' : '' }}{{ formatarCentavosParaBRL(totalAPagarCentavos) }}
          </p>
        </div>
        <div :class="['w-10 h-10 rounded-xl flex items-center justify-center border', totalAPagarCentavos > 0 ? 'bg-coral/5 text-coral border-coral/10' : 'bg-stone text-graphite border-stone']">
          <ArrowDownLeft class="w-5 h-5 stroke-[2.5px]" />
        </div>
      </Card>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-12 gap-8">
      <!-- Coluna Esquerda: Saldos com Externos (5/12 cols) -->
      <div class="lg:col-span-5 space-y-6">
        <NettingPanel
          v-if="saldosPessoais.length > 0"
          :netting-transferencias="nettingPessoais"
          :fatura-selecionada-fechada="false"
          :get-membro-nome="getMembroNomeNetting"
          :is-read-only="false"
          title="Pendências"
          subtitle="Acertos de contas"
          icon-variant="sky"
          @abrir-netting="handleAbrirNetting"
        >
          <template #icon>
            <UserCheck
              class="w-5 h-5"
              aria-hidden="true"
            />
          </template>
        </NettingPanel>

        <Card
          v-else
          class="!p-0 overflow-hidden shadow-subtle bg-white border-none flex flex-col min-h-[300px]"
        >
          <div class="py-5 px-5 sm:py-6 sm:px-6 border-b border-stone bg-parchment flex justify-between items-center shrink-0">
            <div class="flex items-center gap-5">
              <div class="w-11 h-11 rounded-xl bg-sky text-white flex items-center justify-center shadow-sm">
                <UserCheck
                  class="w-5 h-5"
                  aria-hidden="true"
                />
              </div>
              <div>
                <h3 class="font-bold text-lg leading-tight text-charcoal tracking-tight">
                  Pendências
                </h3>
                <p class="text-[11px] text-graphite uppercase tracking-widest mt-0.5 font-semibold">
                  Acertos de contas
                </p>
              </div>
            </div>
          </div>

          <div class="flex-1 flex flex-col items-center justify-center p-8 text-center space-y-4">
            <IllustrationMascot
              variant="meadow"
              :size="70"
              mood="sleeping"
            />
            <p class="text-xs text-ash font-medium leading-relaxed max-w-[200px]">
              Nenhum saldo pendente na sua área pessoal neste período.
            </p>
          </div>
        </Card>
      </div>

      <!-- Coluna Direita: Feed de Lançamentos Privados (7/12 cols) -->
      <div class="lg:col-span-7">
        <ActivityFeed
          :gastos="gastosPrivados"
          :membros="props.membros"
          :is-month-closed="false"
          :is-read-only="false"
          @excluir="solicitarEstorno"
          @ajustar="() => toast.show('Use o estorno e relance para modificar gastos pessoais.', 'info')"
        />
      </div>
    </div>

    <!-- Modal Confirmar Liquidação Pessoal via Acerto Inteligente -->
    <BottomSheetAcertoCompensacao
      :visible="!!liquidandoPendenciaNetting"
      :from-id="liquidandoPendenciaNetting?.from"
      :to-id="liquidandoPendenciaNetting?.to"
      :from-name="liquidandoPendenciaNetting ? getMembroNomeNetting(liquidandoPendenciaNetting.from) : ''"
      :to-name="liquidandoPendenciaNetting ? getMembroNomeNetting(liquidandoPendenciaNetting.to) : ''"
      :suggested-value="liquidandoPendenciaNetting?.val || 0"
      :loading="isSubmittingLiquidador"
      subtitle="Liquidando um saldo pessoal. Será registrado como uma transação privada."
      @confirm="confirmarLiquidar"
      @cancel="fecharLiquidar"
    />

    <BottomSheet
      :model-value="!!gastoParaEstornar"
      title="Confirmar Estorno"
      subtitle="Deseja realmente estornar este lançamento privado?"
      content-class="px-6 pb-6"
      @update:model-value="val => { if (!val) gastoParaEstornar = null }"
    >
      <div class="flex gap-3 pt-4">
        <button
          class="flex-1 px-4 py-3 rounded-xl border border-stone bg-stone/30 font-bold text-sm text-charcoal"
          @click="gastoParaEstornar = null"
        >
          Cancelar
        </button>
        <button
          class="flex-1 px-4 py-3 rounded-xl bg-coral text-white font-bold text-sm"
          @click="confirmarEstorno"
        >
          Sim, estornar
        </button>
      </div>
    </BottomSheet>
  </div>
</template>
