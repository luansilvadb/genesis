<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import BottomSheet from '../../ui/BottomSheet.vue'
import Button from '../../ui/Button.vue'
import { ChevronDown, Lock, CheckCircle2, AlertTriangle } from 'lucide-vue-next'
import type { Fatura } from '../../../../models/entities/Fatura'
import type { ContaFixa } from '../../../../models/entities/ContaFixa'
import type { DashboardViewModel } from '../../../../viewmodels/useDashboardViewModel'
import { formatarBRL } from '../../../../shared/utils/formatarMoeda'

type HistoricoViewModel = Pick<
  DashboardViewModel,
  | 'periodoSelecionado'
  | 'setPeriodoSelecionado'
  | 'mesesAbertosOpcoes'
  | 'mesesTrancadosOpcoes'
  | 'formatarMesAno'
  | 'faturaSelecionadaFechada'
  | 'faturaAtivaVisualizada'
  | 'totalPeriodoSelecionado'
  | 'totalLancamentosPeriodoSelecionado'
  | 'nettingTransferencias'
  | 'contasFixas'
  | 'gastosFaturaSelecionada'
  | 'formatarDinheiro'
  | 'confirmarNovoPeriodo'
  | 'reabrirPeriodoSelecionado'
>

const props = defineProps<{
  visible: boolean
  vm: HistoricoViewModel
  faturasFechadas: Fatura[]
  isFecharPeriodoBloqueado?: boolean
}>()

const emit = defineEmits(['close'])

const {
  periodoSelecionado,
  setPeriodoSelecionado,
  mesesAbertosOpcoes,
  mesesTrancadosOpcoes,
  formatarMesAno,
  faturaSelecionadaFechada,
  faturaAtivaVisualizada,
  totalPeriodoSelecionado,
  totalLancamentosPeriodoSelecionado,
  nettingTransferencias,
  contasFixas,
  gastosFaturaSelecionada,
  formatarDinheiro,
  confirmarNovoPeriodo,
  reabrirPeriodoSelecionado
} = props.vm

// ── Estado interno ──
const modoConfirmacao = ref(false)
const isSubmitting = ref(false)
const itemSelecionadoRef = ref<HTMLElement | null>(null)
const dropdownContainerRef = ref<HTMLElement | null>(null)
const isDropdownAbertosOpen = ref(false)

const temContaFixaPendente = computed(() => contasFixas.value.some(
  (conta: ContaFixa) => !gastosFaturaSelecionada.value.some(gasto => gasto.recurringBillId === conta.id)
))

// Resetar modo confirmação quando o bottom sheet fechar
watch(() => props.visible, (v) => {
  if (!v) {
    modoConfirmacao.value = false
  }
})

watch(isDropdownAbertosOpen, async (aberto) => {
  if (aberto) {
    await nextTick()
    if (itemSelecionadoRef.value) {
      itemSelecionadoRef.value.scrollIntoView({ block: 'nearest' })
    } else {
      // Período selecionado não está na lista de abertos (ex: foi arquivado).
      // A lista já está ordenada por proximidade — o topo é o mês mais relevante.
      dropdownContainerRef.value?.scrollTo({ top: 0 })
    }
  }
})

// ── Ações ──
const handleEncerrar = () => {
  modoConfirmacao.value = true
}

const handleConfirmarEncerramento = async () => {
  isSubmitting.value = true
  try {
    await confirmarNovoPeriodo()
    emit('close')
  } finally {
    isSubmitting.value = false
  }
}

const handleReabrir = async () => {
  isSubmitting.value = true
  try {
    await reabrirPeriodoSelecionado()
  } finally {
    isSubmitting.value = false
  }
}

const handleCancelarConfirmacao = () => {
  modoConfirmacao.value = false
}

const handleSelecionarPeriodo = (mes: number, ano: number) => {
  setPeriodoSelecionado(mes, ano)
  isDropdownAbertosOpen.value = false
}
</script>

<template>
  <BottomSheet 
    :model-value="visible" 
    @update:model-value="val => { if (!val) emit('close') }" 
    :subtitle="modoConfirmacao
      ? (faturaAtivaVisualizada ? `Revise os números antes de arquivar o mês de ${formatarMesAno(faturaAtivaVisualizada.periodo.mes, faturaAtivaVisualizada.periodo.ano)}. O saldo será consolidado e os acertos serão gerados automaticamente.` : '')
      : 'Gerencie o período atual e navegue pelo histórico de meses da casa.'"
    :content-class="`px-6 pb-8 flex-grow ${isDropdownAbertosOpen ? 'overflow-visible' : 'overflow-y-auto'}`"
  >
    <template #title>
      <h3 class="text-3xl font-display text-charcoal leading-tight">
        <template v-if="modoConfirmacao">
          Fechamento de <span class="text-ember">Período</span>
        </template>
        <template v-else>
          Navegar nos <span class="text-ember">Períodos</span>
        </template>
      </h3>
    </template>

    <!-- ═══ MODO CONFIRMAÇÃO ═══ -->
    <div v-if="modoConfirmacao" class="space-y-6 pt-2">
      <div v-if="faturaAtivaVisualizada" class="grid grid-cols-2 gap-3">
        <div class="bg-parchment p-4 rounded-2xl border border-stone shadow-subtle">
          <p class="text-[10px] font-bold uppercase text-graphite tracking-widest mb-1">Total do Mês</p>
          <p class="text-2xl font-display text-charcoal break-words">{{ formatarBRL(formatarDinheiro(totalPeriodoSelecionado)) }}</p>
          <p class="text-[10px] text-graphite font-bold mt-1 uppercase opacity-60">{{ totalLancamentosPeriodoSelecionado }} lançamentos</p>
        </div>

        <div class="bg-parchment p-4 rounded-2xl border border-stone shadow-subtle">
          <p class="text-[10px] font-bold uppercase text-graphite tracking-widest mb-1">Impacto (Pix)</p>
          <p class="text-2xl font-display text-ember">{{ nettingTransferencias.length }} Acertos</p>
          <p class="text-[10px] text-graphite font-bold mt-1 uppercase opacity-60">serão cobrados</p>
        </div>
      </div>

      <div v-if="faturaAtivaVisualizada && temContaFixaPendente" class="p-4 rounded-2xl bg-ember/5 border border-ember/20 flex gap-3 items-start animate-in fade-in slide-in-from-top-1">
        <div class="w-8 h-8 rounded-full bg-ember/10 flex items-center justify-center shrink-0 mt-0.5">
          <AlertTriangle class="w-4 h-4 text-ember" />
        </div>
        <div>
          <p class="text-xs font-bold text-ember uppercase tracking-tight">Contas fixas pendentes!</p>
          <p class="text-[11px] text-graphite font-semibold mt-0.5 leading-tight">Ainda existem contas fixas deste mês que não foram lançadas. Deseja fechar mesmo assim?</p>
        </div>
      </div>

    </div>

    <!-- ═══ MODO NAVEGAÇÃO ═══ -->
    <div v-else class="space-y-6 pt-2">
      <!-- ── Navegação de Períodos Abertos ── -->
      <div class="space-y-3">
        <h4 class="text-[10px] font-bold uppercase tracking-widest text-graphite ml-1">Gerenciar Período Aberto</h4>
        <div class="relative" @focusout="isDropdownAbertosOpen = false">
          <div 
            @click="isDropdownAbertosOpen = !isDropdownAbertosOpen"
            @keydown.enter.prevent="isDropdownAbertosOpen = !isDropdownAbertosOpen"
            @keydown.space.prevent="isDropdownAbertosOpen = !isDropdownAbertosOpen"
            aria-label="Gerenciar período aberto"
            role="button"
            tabindex="0"
            :aria-expanded="isDropdownAbertosOpen ? 'true' : 'false'"
            class="w-full px-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-sm text-charcoal cursor-pointer flex justify-between items-center transition-all hover:bg-stone/30"
          >
            <span class="flex items-center gap-2.5">
              <Lock v-if="faturaSelecionadaFechada" class="w-3.5 h-3.5 text-graphite opacity-50 shrink-0" />
              <span v-else class="w-2.5 h-2.5 rounded-full bg-meadow animate-pulse shrink-0" />
              <span class="block truncate">
                {{ periodoSelecionado
                  ? formatarMesAno(periodoSelecionado.mes, periodoSelecionado.ano)
                  : 'Selecionar mês...'
                }}
              </span>
            </span>
            <ChevronDown class="w-4 h-4 text-graphite pointer-events-none transition-transform duration-300" :class="isDropdownAbertosOpen ? 'rotate-180' : ''" />
          </div>
          
          <transition name="dropdown-slide">
            <div v-if="isDropdownAbertosOpen" ref="dropdownContainerRef" class="absolute left-0 w-full mt-1.5 max-h-48 overflow-y-auto bg-card border border-stone rounded-xl shadow-xl z-50 py-2 custom-scrollbar">
              <div 
                v-for="op in mesesAbertosOpcoes" 
                :key="op.nome" 
                :ref="(el) => { if (el && periodoSelecionado.mes === op.mes && periodoSelecionado.ano === op.ano) itemSelecionadoRef = el as HTMLElement }"
                @mousedown.prevent="handleSelecionarPeriodo(op.mes, op.ano)" 
                role="button"
                tabindex="0"
                class="px-4 py-3.5 text-sm font-bold hover:bg-stone cursor-pointer transition-colors flex items-center gap-3"
                :class="periodoSelecionado.mes === op.mes && periodoSelecionado.ano === op.ano ? 'text-ember bg-ember/5' : 'text-charcoal'"
              >
                <span class="w-2 h-2 rounded-full bg-meadow animate-pulse shrink-0" />
                {{ op.nome }}
              </div>
            </div>
          </transition>
        </div>
      </div>

      <!-- ── Histórico de Fechados ── -->
      <div class="space-y-3">
        <h4 class="text-[10px] font-bold uppercase tracking-widest text-graphite ml-1">Histórico de Fechados</h4>
        <div class="grid gap-2">
          <div 
            v-for="item in mesesTrancadosOpcoes" 
            :key="item.nome"
            @click="handleSelecionarPeriodo(item.mes, item.ano)"
            :aria-label="'Selecionar período arquivado ' + item.nome"
            role="button"
            tabindex="0"
            class="p-3.5 rounded-2xl border cursor-pointer transition-all flex items-center justify-between"
            :class="periodoSelecionado.mes === item.mes && periodoSelecionado.ano === item.ano ? 'border-ember bg-ember/5 text-ember font-bold' : 'border-stone bg-canvas text-charcoal hover:bg-stone/30'"
          >
            <div class="flex items-center gap-3">
              <span class="w-2.5 h-2.5 rounded-full bg-graphite opacity-30" />
              <span class="text-sm font-bold">{{ item.nome }}</span>
            </div>
            <div class="flex items-center gap-2">
              <span class="text-[10px] uppercase font-bold text-graphite opacity-60">Arquivado</span>
              <Lock class="w-3.5 h-3.5 text-graphite opacity-40 shrink-0" />
            </div>
          </div>
          
          <div v-if="mesesTrancadosOpcoes.length === 0" class="text-center py-4 border border-dashed border-stone rounded-2xl">
            <p class="text-[11px] text-graphite font-semibold opacity-60 italic">Nenhum mês arquivado ainda.</p>
          </div>
        </div>
      </div>

      <div class="h-px bg-stone/60 my-2" />

      <!-- ── Status do Período Atual ── -->
      <div class="bg-parchment p-5 rounded-2xl border border-stone shadow-subtle">
        <div class="flex items-start gap-4">
          <div v-if="faturaSelecionadaFechada" class="w-10 h-10 rounded-xl bg-meadow/10 flex items-center justify-center shrink-0 border border-meadow/20">
            <CheckCircle2 class="w-5 h-5 text-meadow" stroke-width="3" />
          </div>
          <div v-else class="w-10 h-10 rounded-xl bg-ember/10 flex items-center justify-center shrink-0 border border-ember/20">
            <CheckCircle2 class="w-5 h-5 text-ember opacity-30" stroke-width="3" />
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="font-bold text-base leading-tight text-charcoal tracking-tight">
              {{ faturaAtivaVisualizada ? formatarMesAno(faturaAtivaVisualizada.periodo.mes, faturaAtivaVisualizada.periodo.ano) : formatarMesAno(periodoSelecionado.mes, periodoSelecionado.ano) }}
            </h3>
            <p class="text-[12px] text-graphite mt-1 font-medium leading-relaxed">
              {{ faturaSelecionadaFechada ? 'Este mês está oficialmente arquivado. O histórico está protegido para auditoria.' : 'Finalize o mês atual para gerar os acertos finais e preparar o próximo ciclo da casa.' }}
            </p>
          </div>
        </div>

      </div>
    </div>
    
    <!-- Footer unificado -->
    <template #footer>
      <div v-if="modoConfirmacao" class="grid grid-cols-2 gap-3">
        <Button variant="secondary" class="font-bold uppercase tracking-widest text-[10px] h-12" @click="handleCancelarConfirmacao" :disabled="isSubmitting">Cancelar</Button>
        <Button variant="primary" class="font-bold uppercase tracking-widest text-[10px] h-12" @click="handleConfirmarEncerramento" :loading="isSubmitting">Arquivar Mês</Button>
      </div>
      <div v-else class="grid grid-cols-2 gap-3">
        <Button variant="secondary" class="font-bold uppercase tracking-widest text-[10px] h-12" @click="emit('close')">Fechar</Button>
        <Button
          v-if="faturaSelecionadaFechada"
          variant="secondary"
          class="bg-white border-stone text-charcoal font-bold uppercase tracking-widest text-[10px] h-12 shadow-sm"
          :disabled="isFecharPeriodoBloqueado || isSubmitting"
          :loading="isSubmitting"
          @click="handleReabrir"
        >Reabrir Período</Button>
        <Button v-else variant="primary" class="!bg-midnight hover:!bg-charcoal text-white font-bold uppercase tracking-widest text-[10px] h-12 shadow-md" :disabled="isFecharPeriodoBloqueado" @click="handleEncerrar">Encerrar Mês</Button>
      </div>
    </template>
  </BottomSheet>
</template>
