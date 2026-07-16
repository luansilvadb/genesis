<script setup lang="ts">
import type { ContaFixa } from '../../../models/entities/ContaFixa'
import type { Gasto } from '../../../models/entities/Gasto'
import { Repeat, Plus } from 'lucide-vue-next'
import Card from '../ui/Card.vue'
import ContasFixasCard from './ContasFixasCard.vue'
import IllustrationMascot from '../ui/IllustrationMascot.vue'

const props = defineProps<{
  contasFixas: ContaFixa[]
  gastos: Gasto[]
  membros: { id: string; nome: string }[]
  isMonthClosed: boolean
  isReadOnly?: boolean
}>()

defineEmits<{
  (e: 'lancar', bill: ContaFixa): void
  (e: 'configurar', bill: ContaFixa): void
  (e: 'novo'): void
  (e: 'estornar', bill: ContaFixa): void
}>()

const obterGasto = (conta: ContaFixa) => props.gastos.find(g => g.recurringBillId === conta.id)

const obterNomeMembro = (id: string) => props.membros.find(m => m.id === id)?.nome || 'Membro'
</script>

<template>
  <Card class="!p-0 overflow-hidden shadow-subtle bg-white text-graphite">
    <div class="py-5 px-5 sm:py-7 sm:px-6 border-b border-stone bg-parchment flex items-center">
      <div class="flex items-center gap-5">
        <div class="w-11 h-11 rounded-xl bg-sky text-white flex items-center justify-center shadow-sm">
          <Repeat
            class="w-5 h-5"
            aria-hidden="true"
          />
        </div>
        <div>
          <h2 class="font-bold text-lg leading-tight text-charcoal tracking-tight">
            Contas Fixas
          </h2>
          <p class="text-[11px] text-graphite uppercase tracking-widest mt-0.5 font-semibold">
            Recorrentes do mês
          </p>
        </div>
      </div>
    </div>

    <div class="p-4 sm:p-6 grid gap-3">
      <div
        v-if="contasFixas.length === 0"
        class="text-center py-10 px-4 border border-dashed border-stone rounded-2xl space-y-5 bg-canvas/30 animate-in fade-in duration-700"
      >
        <div class="flex justify-center">
          <IllustrationMascot
            variant="meadow"
            :size="80"
            mood="happy"
            class="animate-float"
          />
        </div>
        <div class="space-y-2">
          <p class="text-sm font-semibold text-charcoal uppercase tracking-widest">
            Nenhuma conta agendada
          </p>
          <p class="text-xs text-graphite max-w-[240px] mx-auto leading-relaxed font-medium">
            Cadastre aluguel, luz ou internet para fazer lançamentos recorrentes ultra rápidos.
          </p>
        </div>
      </div>

      <template v-else>
        <ContasFixasCard 
          v-for="bill in contasFixas" 
          :key="bill.id" 
          :bill="bill"
          :gasto="obterGasto(bill)"
          :obter-nome-membro="obterNomeMembro"
          :is-month-closed="props.isMonthClosed"
          :is-read-only="props.isReadOnly"
          @lancar="$emit('lancar', bill)"
          @estornar="$emit('estornar', bill)"
          @configurar="$emit('configurar', bill)"
        />
      </template>

      <div
        v-if="!props.isReadOnly && !props.isMonthClosed"
        class="flex flex-col items-center gap-2 mt-3"
      >
        <button
          class="relative overflow-hidden group w-full flex justify-center items-center gap-3 p-4 rounded-xl border border-dashed border-stone bg-canvas/50 text-ash font-semibold text-[10px] uppercase tracking-[0.2em] transition-all duration-300 select-none cursor-pointer hover:border-ember hover:bg-ember/5 hover:text-ember active:scale-[0.97]"
          data-testid="nova-conta-fixa"
          @click="$emit('novo')"
        >
          <Plus class="w-4 h-4 transition-transform group-hover:scale-110" />
          <span>Adicionar Nova Conta</span>
        </button>
      </div>
    </div>
  </Card>
</template>
