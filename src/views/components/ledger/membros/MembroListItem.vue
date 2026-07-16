<script setup lang="ts">
import { ChevronRight } from 'lucide-vue-next'
import MembroAvatar from '../../ui/MembroAvatar.vue'
import type { Membro } from '../../../../models/entities/Membro'

interface Props {
  membro: Membro
  variant: 'ember' | 'meadow' | 'sky' | 'sunburst' | 'flamingo'
  clickable?: boolean
}

defineProps<Props>()
const emit = defineEmits(['click'])
</script>

<template>
  <div 
    class="p-4 flex justify-between items-center bg-canvas/50 border border-stone/50 rounded-2xl transition-all duration-500 group"
    :class="[
      { 'opacity-60 grayscale-[0.5]': !membro.ativo },
      clickable !== false ? 'cursor-pointer active:scale-[0.98] hover:border-ember/40' : 'cursor-default'
    ]"
    @click="clickable !== false && emit('click')"
  >
    <div class="flex items-center gap-4 min-w-0">
      <MembroAvatar
        :nome="membro.nome"
        :variant="variant"
        size="sm"
      />
      <div class="min-w-0">
        <span class="text-sm font-bold text-charcoal leading-none block truncate">{{ membro.nome }}</span>
        <p
          class="text-caption text-[9px] mt-1"
          :class="membro.ativo ? 'text-meadow' : 'text-ash'"
        >
          {{ membro.ativo ? 'Ativo na Casa' : 'Acesso Suspenso' }}
        </p>
      </div>
    </div>
    
    <div class="flex items-center gap-3 shrink-0">
      <!-- ADMIN: pill ember (existente, sem alteração) -->
      <span
        v-if="membro.role === 'ADMIN'"
        class="px-2.5 py-1 rounded-pill text-[9px] font-bold bg-ember/10 text-ember uppercase tracking-widest border border-ember/20"
      >Admin</span>

      <!-- MORADOR sem Cargo: pill neutra (novo) -->
      <span
        v-else-if="membro.role === 'MORADOR'"
        class="px-2.5 py-1 rounded-pill text-[9px] font-bold bg-stone text-ash uppercase tracking-widest border border-stone"
      >Morador</span>

      <!-- VISUALIZADOR sem Cargo: pill neutra (novo) -->
      <span
        v-else-if="membro.role === 'VISUALIZADOR'"
        class="px-2.5 py-1 rounded-pill text-[9px] font-bold bg-sky/10 text-sky uppercase tracking-widest border border-sky/20"
      >Visualizador</span>

      <ChevronRight 
        v-if="clickable !== false" 
        class="w-4 h-4 text-ash group-hover:text-charcoal group-hover:translate-x-1 transition-all" 
      />
    </div>
  </div>
</template>
