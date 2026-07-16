<script setup lang="ts">
import { TrendingUp } from 'lucide-vue-next'
import Card from '../../ui/Card.vue'
import MembroAvatar from '../../ui/MembroAvatar.vue'
import { formatarBRL } from '../../../../shared/utils/formatarMoeda'

defineProps<{
  membrosVisiveis: { id: string; nome: string }[]
  saldosUnificadosAtivos: Record<string, number>
}>()

const variants: ('ember' | 'meadow' | 'sky' | 'sunburst' | 'flamingo')[] = ['ember', 'sky', 'meadow', 'sunburst', 'flamingo']
</script>

<template>
  <Card class="!p-0 overflow-hidden shadow-subtle bg-white text-graphite">
    <div class="py-5 px-5 sm:py-7 sm:px-6 border-b border-stone bg-parchment flex items-center">
      <div class="flex items-center gap-5">
        <div class="w-11 h-11 rounded-xl bg-meadow text-white flex items-center justify-center shadow-sm">
          <TrendingUp
            class="w-5 h-5"
            aria-hidden="true"
          />
        </div>
        <div>
          <h2 class="font-bold text-lg leading-tight text-charcoal tracking-tight">
            Saldos Unificados
          </h2>
          <p class="text-[11px] text-graphite uppercase tracking-widest mt-0.5 font-semibold">
            Créditos e débitos da casa
          </p>
        </div>
      </div>
    </div>

    <div class="p-3 sm:p-6 space-y-3 sm:space-y-4 relative z-10">
      <div 
        v-for="(m, idx) in membrosVisiveis" 
        :key="m.id" 
        class="group flex justify-between items-center p-3 sm:p-4 rounded-2xl border border-stone bg-canvas hover:border-ember/30 hover:bg-white transition-all duration-300"
      >
        <div class="flex items-center gap-3 sm:gap-4">
          <MembroAvatar
            :nome="m.nome"
            :variant="variants[idx % variants.length]"
            size="md"
          />
          <div>
            <span class="font-bold text-sm sm:text-base block text-charcoal leading-tight">{{ m.nome }}</span>
            <span class="text-[10px] sm:text-[11px] text-graphite block mt-0.5 font-semibold uppercase tracking-wider opacity-60">
              {{ saldosUnificadosAtivos[m.id] > 0.005 ? 'Crédito acumulado' : saldosUnificadosAtivos[m.id] < -0.005 ? 'Débito pendente' : 'Tudo em dia' }}
            </span>
          </div>
        </div>
        <div class="text-right">
          <span :class="['font-display text-xl sm:text-2xl block tracking-tighter', saldosUnificadosAtivos[m.id] > 0.005 ? 'text-meadow' : saldosUnificadosAtivos[m.id] < -0.005 ? 'text-coral' : 'text-graphite']">
            {{ saldosUnificadosAtivos[m.id] > 0.005 ? '+' : '' }}{{ formatarBRL(saldosUnificadosAtivos[m.id]) }}
          </span>
        </div>
      </div>
    </div>
  </Card>
</template>
