<script setup lang="ts">
import { Edit3, Trash2 } from 'lucide-vue-next'
import { Gasto } from '../../../models/entities/Gasto'
import { computed } from 'vue'
import Card from '../ui/Card.vue'
import Button from '../ui/Button.vue'
import MembroAvatar from '../ui/MembroAvatar.vue'
import IllustrationMascot from '../ui/IllustrationMascot.vue'
import { formatarCentavosParaBRL } from '../../../shared/utils/formatarMoeda'

const props = defineProps<{
  gastos: Gasto[]
  membros: { id: string; nome: string }[]
  isMonthClosed: boolean
  isReadOnly?: boolean
}>()

const emit = defineEmits(['excluir', 'ajustar'])

const getMembroNome = (id: string) => {
  if (id && id.startsWith('externo:')) {
    return id.substring(8)
  }
  return props.membros.find(m => m.id === id)?.nome || '?'
}

const sortedGastos = computed(() => {
  return [...props.gastos].sort((a, b) => b.id.localeCompare(a.id))
})

function badgeLabel(g: Gasto): string {
  if (g.id.startsWith('forecast-bill-')) return 'Previsão Fixa'
  if (g.id.startsWith('audit-settlement-')) return 'Rateio'
  if (g.isLoan) return 'Empréstimo'
  if (g.isSettlement) return 'Acerto'
  if (g.method === 'card') return 'Cartão'
  return 'Pix'
}

function badgeClass(g: Gasto): Record<string, boolean> {
  if (g.id.startsWith('forecast-')) return { 'bg-stone text-graphite': true }
  if (g.id.startsWith('audit-settlement-')) return { 'bg-midnight text-white': true }
  return { 'bg-ember text-white': true }
}
</script>

<template>
  <Card class="!p-0 overflow-hidden shadow-subtle bg-white text-graphite min-h-[400px] flex flex-col">
    <div class="py-5 px-5 sm:py-7 sm:px-6 border-b border-stone bg-parchment flex justify-between items-center shrink-0">
      <div class="flex items-center gap-5">
        <div class="w-11 h-11 rounded-xl bg-midnight text-white flex items-center justify-center shadow-sm">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="24"
            height="24"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="w-5 h-5"
          ><path d="M12 20v-6M9 20V10M15 20V4M3 20h18" /></svg>
        </div>
        <div>
          <h2 class="font-bold text-lg leading-tight text-charcoal tracking-tight">
            Atividades
          </h2>
          <p class="text-[11px] text-graphite uppercase tracking-widest mt-0.5 font-semibold">
            Últimos registros da casa
          </p>
        </div>
      </div>
    </div>

    <div
      v-if="sortedGastos.length === 0"
      class="flex-1 flex flex-col items-center justify-center p-12 text-center space-y-6"
    >
      <div class="animate-float">
        <IllustrationMascot
          variant="sunburst"
          :size="100"
          mood="happy"
        />
      </div>
      <div class="space-y-1">
        <p class="text-sm font-semibold text-charcoal uppercase tracking-widest">
          Tudo limpo por aqui
        </p>
        <p class="text-xs text-graphite max-w-[200px] mx-auto leading-relaxed font-medium">
          Nenhum gasto registrado. Clique no botão de + para começar!
        </p>
      </div>
    </div>

    <div
      v-else
      class="p-4 sm:p-6 space-y-4 flex-1 overflow-y-auto custom-scrollbar"
    >
      <TransitionGroup
        name="list"
        tag="div"
        class="space-y-4"
      >
        <div 
          v-for="g in sortedGastos" 
          :key="g.id"
          class="group flex flex-col p-4 rounded-xl border border-stone bg-canvas hover:border-ember/30 transition-all duration-200 space-y-4"
        >
          <div class="flex justify-between items-start gap-4">
            <div class="space-y-1">
              <span class="font-semibold text-sm text-charcoal flex items-center gap-1.5">
                <span>{{ g.descricao }}</span>
                <span
                  v-if="g.totalInstallments > 1"
                  class="text-xs text-graphite font-semibold"
                >
                  ({{ g.totalInstallments - g.installments + 1 }}/{{ g.totalInstallments }})
                </span>
              </span>
              <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
                <span 
                  class="text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded-md"
                  :class="badgeClass(g)"
                >
                  {{ badgeLabel(g) }}
                </span>
                <span
                  v-if="g.id.startsWith('forecast-bill-')"
                  class="text-[10px] text-graphite font-semibold italic"
                >
                  (Aguardando lançamento)
                </span>
                <div
                  v-else
                  class="flex items-center gap-1.5"
                >
                  <MembroAvatar
                    :nome="getMembroNome(g.compradorId)"
                    size="sm"
                    variant="sky"
                    class="!w-5 !h-5 !text-[8px] animate-in zoom-in-50 duration-300"
                  />
                  <span class="text-[10px] text-graphite font-medium">
                    Pago por <strong class="text-charcoal font-bold">{{ getMembroNome(g.compradorId) }}</strong>
                  </span>
                </div>
              </div>
            </div>
            <div class="text-right flex flex-col items-end">
              <span class="font-display text-lg text-charcoal">
                {{ formatarCentavosParaBRL(g.totalInstallments > 1 ? Math.round(g.valorTotal.centavos / g.totalInstallments) : g.valorTotal.centavos) }}
              </span>
              <span
                v-if="g.totalInstallments > 1"
                class="text-[10px] text-graphite font-semibold block"
              >
                Total: {{ formatarCentavosParaBRL(g.valorTotal.centavos) }}
              </span>
            </div>
          </div>

          <!-- Ações do Feed -->
          <div
            v-if="!props.isReadOnly"
            class="flex flex-col items-end gap-2 pt-3 border-t border-stone transition-opacity"
          >
            <div class="flex justify-end gap-2 w-full">
              <Button 
                v-if="!g.isSettlement"
                variant="secondary"
                size="sm"
                class="h-9 px-4 text-xs border border-stone font-semibold"
                :disabled="props.isMonthClosed"
                :aria-label="'Ajustar ' + g.descricao"
                @click="emit('ajustar', g)"
              >
                <Edit3
                  class="w-3.5 h-3.5 mr-1.5"
                  aria-hidden="true"
                />
                Ajustar
              </Button>
              <Button 
                variant="secondary" 
                size="sm"
                class="h-9 px-4 text-xs text-coral hover:bg-coral/5 border border-transparent font-semibold"
                :disabled="props.isMonthClosed"
                :aria-label="'Estornar ' + g.descricao"
                @click="emit('excluir', g)"
              >
                <Trash2
                  class="w-3.5 h-3.5 mr-1.5"
                  aria-hidden="true"
                />
                Estornar
              </Button>
            </div>
          </div>
        </div>
      </TransitionGroup>
    </div>
  </Card>
</template>

<style scoped>
.list-enter-active,
.list-leave-active {
  transition: all 0.5s var(--ease-spring);
}
.list-enter-from {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}
.list-leave-to {
  opacity: 0;
  transform: scale(0.9) translateY(-10px);
}
.list-move {
  transition: transform 0.5s var(--ease-spring);
}

</style>
