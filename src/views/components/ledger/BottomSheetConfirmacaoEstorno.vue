<template>
  <BottomSheet 
    :model-value="visible" 
    :title="itemType === 'Conta Fixa' ? 'Excluir Modelo' : `Estornar ${itemType}`" 
    :subtitle="itemType === 'Conta Fixa' ? 'Este modelo de conta fixa será removido da sua lista de recorrências.' : 'O lançamento será removido e os saldos dos moradores serão recalculados.'"
    @update:model-value="(val: boolean) => { if (!val) $emit('cancel') }"
  >
    <template #header>
      <h3 class="text-3xl font-display text-charcoal">
        {{ itemType === 'Conta Fixa' ? 'Excluir' : 'Estornar' }} <span class="text-ember">{{ itemType === 'Conta Fixa' ? 'Modelo' : itemType }}</span>?
      </h3>
    </template>

    <div class="flex flex-col items-center text-center space-y-6 pt-4">
      <!-- Ilustração Sad Blob (Estilo Family) -->
      <div class="relative w-20 h-20 mb-1">
        <svg
          viewBox="0 0 100 100"
          class="w-full h-full drop-shadow-sm"
        >
          <!-- Corpo do Blob (Coral Red para indicar perigo/tristeza) -->
          <path 
            d="M20,50 Q20,20 50,20 Q80,20 80,50 Q80,80 50,85 Q20,80 20,50 Z" 
            fill="var(--color-coral)" 
            class="animate-pulse"
            style="animation-duration: 4s;"
          />
          <!-- Olhinhos tristes (Stick limbs style) -->
          <circle
            cx="40"
            cy="45"
            r="3"
            fill="white"
          />
          <circle
            cx="60"
            cy="45"
            r="3"
            fill="white"
          />
          <!-- Boca curva de preocupação -->
          <path
            d="M42,62 Q50,55 58,62"
            stroke="white"
            stroke-width="2.5"
            stroke-linecap="round"
            fill="none"
          />
          <!-- Perninhas de palito -->
          <line
            x1="40"
            y1="83"
            x2="35"
            y2="95"
            stroke="var(--color-charcoal)"
            stroke-width="3"
            stroke-linecap="round"
          />
          <line
            x1="60"
            y1="83"
            x2="65"
            y2="95"
            stroke="var(--color-charcoal)"
            stroke-width="3"
            stroke-linecap="round"
          />
        </svg>
      </div>

      <!-- Detalhes do Item (Visual Recessed Panel) -->
      <div
        v-if="itemName || (itemValue !== undefined && itemValue !== null)"
        class="w-full p-4 rounded-2xl bg-parchment shadow-subtle border-none"
      >
        <div class="flex justify-between items-center gap-4">
          <div class="text-left flex-1 min-w-0">
            <p class="text-[10px] font-bold uppercase text-graphite tracking-widest mb-1">
              {{ itemType === 'Conta Fixa' ? 'Nome da Conta' : 'Item selecionado' }}
            </p>
            <p class="text-sm font-bold text-charcoal truncate">
              {{ itemName || 'Sem descrição' }}
            </p>
          </div>
          <div
            v-if="itemValue"
            class="text-right shrink-0"
          >
            <p class="text-[10px] font-bold uppercase text-graphite tracking-widest mb-1">
              Valor
            </p>
            <p class="text-base font-bold text-ember tracking-tight">
              {{ formatarBRL(itemValue) }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex flex-col gap-2.5 w-full">
        <Button 
          variant="primary" 
          class="w-full h-12 text-xs font-bold uppercase tracking-widest"
          @click="$emit('confirm')"
        >
          {{ itemType === 'Conta Fixa' ? 'Sim, excluir modelo' : 'Sim, confirmar estorno' }}
        </Button>
        <Button 
          variant="secondary"
          class="w-full h-12 text-xs font-bold uppercase tracking-widest"
          @click="$emit('cancel')"
        >
          {{ itemType === 'Conta Fixa' ? 'Não, manter modelo' : 'Não, manter lançamento' }}
        </Button>
      </div>
    </template>
  </BottomSheet>
</template>

<script setup lang="ts">
import BottomSheet from '../ui/BottomSheet.vue'
import Button from '../ui/Button.vue'
import { formatarBRL } from '../../../shared/utils/formatarMoeda'

defineProps<{
  visible: boolean
  itemType: string // ex: "Lançamento", "Conta Fixa"
  itemName?: string
  itemValue?: number
}>()

defineEmits(['confirm', 'cancel'])
</script>
