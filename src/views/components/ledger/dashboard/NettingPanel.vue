<script setup lang="ts">
import { computed } from 'vue'
import { Sparkles, ArrowRight } from 'lucide-vue-next'
import Card from '../../ui/Card.vue'
import Button from '../../ui/Button.vue'
import MembroAvatar from '../../ui/MembroAvatar.vue'
import type { TransferenciaNetting } from '../../../../models/services/NettingService'
import { formatarBRL } from '../../../../shared/utils/formatarMoeda'

interface Props {
  nettingTransferencias: TransferenciaNetting[]
  faturaSelecionadaFechada: boolean
  getMembroNome: (id: string) => string
  isReadOnly?: boolean
  title?: string
  subtitle?: string
  iconVariant?: 'sunburst' | 'sky' | 'meadow' | 'ember' | 'midnight'
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Pendências',
  subtitle: 'Acertos de contas',
  iconVariant: 'sunburst'
})

defineEmits<{
  (e: 'abrirNetting', t: TransferenciaNetting): void
}>()

// Map variant names to full Tailwind classes so the JIT compiler can detect them.
const ICON_BG_CLASSES: Record<string, string> = {
  sunburst: 'bg-sunburst',
  sky: 'bg-sky',
  meadow: 'bg-meadow',
  ember: 'bg-ember',
  midnight: 'bg-midnight',
}

const iconBgClass = computed(() => ICON_BG_CLASSES[props.iconVariant] || ICON_BG_CLASSES.sunburst)
</script>

<template>
  <Card class="!p-0 overflow-hidden shadow-subtle bg-white text-graphite">
    <!-- Cabeçalho Padronizado -->
    <div class="py-5 px-5 sm:py-7 sm:px-6 border-b border-stone bg-parchment flex justify-between items-center">
      <div class="flex items-center gap-5">
        <div
          :class="[
            'w-11 h-11 rounded-xl text-white flex items-center justify-center shadow-sm',
            iconBgClass
          ]"
        >
          <slot name="icon">
            <Sparkles
              class="w-5 h-5"
              aria-hidden="true"
            />
          </slot>
        </div>
        <div>
          <h2 class="font-bold text-lg leading-tight text-charcoal tracking-tight">
            {{ props.title }}
          </h2>
          <p class="text-[11px] text-graphite uppercase tracking-widest mt-0.5 font-semibold">
            {{ props.subtitle }}
          </p>
        </div>
      </div>
    </div>

    <div class="p-4 sm:p-6 grid gap-4">
      <div 
        v-for="t in nettingTransferencias" 
        :key="`${t.from}-${t.to}-${t.val}`" 
        class="p-5 border border-stone bg-canvas rounded-2xl shadow-subtle hover:border-ember/30 transition-all duration-300"
      >
        <div class="flex flex-col gap-6">
          <div class="flex items-center justify-between gap-4">
            <div class="flex flex-col items-center gap-2 flex-1 min-w-0">
              <MembroAvatar
                :nome="getMembroNome(t.from)"
                size="md"
                variant="sky"
              />
              <span class="text-[10px] font-bold text-charcoal uppercase truncate max-w-full text-center">{{ getMembroNome(t.from) }}</span>
            </div>
            
            <div class="flex flex-col items-center gap-1 shrink-0">
              <ArrowRight class="w-5 h-5 text-ember" />
              <span class="text-[10px] font-bold text-ember uppercase tracking-widest">{{ formatarBRL(t.val) }}</span>
            </div>

            <div class="flex flex-col items-center gap-2 flex-1 min-w-0">
              <MembroAvatar
                :nome="getMembroNome(t.to)"
                size="md"
                variant="meadow"
              />
              <span class="text-[10px] font-bold text-charcoal uppercase truncate max-w-full text-center">{{ getMembroNome(t.to) }}</span>
            </div>
          </div>

          <div class="w-full">
            <Button
              variant="primary"
              class="w-full h-12 font-bold uppercase tracking-widest text-[10px] shadow-sm"
              :disabled="faturaSelecionadaFechada || isReadOnly"
              @click="$emit('abrirNetting', t)"
            >
              Registrar Pagamento
            </Button>
          </div>
        </div>
      </div>
    </div>
  </Card>
</template>
