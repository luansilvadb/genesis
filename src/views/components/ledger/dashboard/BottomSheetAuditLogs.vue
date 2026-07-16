<script setup lang="ts">
import BottomSheet from '../../ui/BottomSheet.vue'
import IllustrationMascot from '../../ui/IllustrationMascot.vue'
import type { AuditLogDto } from '../../../../models/repositories/http/HttpAuditLogRepository'

defineProps<{
  visible: boolean
  logs: AuditLogDto[]
  loading: boolean
  getMembroNome: (id: string) => string
}>()

const emit = defineEmits(['close'])
</script> 

<template>
  <BottomSheet 
    :model-value="visible" 
    max-height="90dvh" 
    @update:model-value="val => { if (!val) emit('close') }"
  >
    <template #title>
      <h3 class="text-heading text-charcoal font-display">
        Atividades <span class="text-ember">da Casa</span>
      </h3>
      <p class="text-[10px] text-graphite font-bold uppercase tracking-widest mt-1">
        Histórico de auditoria contábil
      </p>
    </template>

    <div class="space-y-4 pt-2">
      <div
        v-if="loading"
        class="flex flex-col items-center justify-center py-12 space-y-4"
      >
        <div class="w-8 h-8 border-4 border-ember border-t-transparent rounded-full animate-spin" />
        <p class="text-xs text-ash font-bold uppercase tracking-wider">
          Carregando...
        </p>
      </div>

      <div
        v-else-if="logs.length === 0"
        class="flex flex-col items-center justify-center py-16 text-center space-y-4"
      >
        <IllustrationMascot
          variant="sky"
          :size="80"
          mood="chill"
          class="opacity-45"
        />
        <p class="text-xs text-ash font-bold italic max-w-[240px]">
          Nenhuma atividade registrada na casa ainda.
        </p>
      </div>

      <div
        v-else
        class="space-y-6"
      >
        <div
          v-for="log in logs"
          :key="log.id"
          class="flex gap-4 items-start pb-4 border-b border-stone/30 last:border-b-0"
        >
          <div
            class="w-8 h-8 rounded-lg flex items-center justify-center text-sm shrink-0 shadow-subtle border"
            :class="{
              'bg-meadow/10 border-meadow/20': log.acao === 'CRIAR_GASTO',
              'bg-sky/10 border-sky/20': log.acao === 'EDITAR_GASTO',
              'bg-coral/10 border-coral/20': log.acao === 'EXCLUIR_GASTO',
              'bg-sunburst/10 border-sunburst/20': log.acao === 'ALTERAR_RENDA',
              'bg-midnight/10 border-midnight/20': log.acao === 'CRIAR_MEMBRO',
              'bg-flamingo/10 border-flamingo/20': log.acao === 'CRIAR_CARTAO',
            }"
          >
            <span v-if="log.acao === 'CRIAR_GASTO'">💸</span>
            <span v-else-if="log.acao === 'EDITAR_GASTO'">✏️</span>
            <span v-else-if="log.acao === 'EXCLUIR_GASTO'">🗑️</span>
            <span v-else-if="log.acao === 'ALTERAR_RENDA'">💰</span>
            <span v-else-if="log.acao === 'CRIAR_MEMBRO'">👤</span>
            <span v-else-if="log.acao === 'CRIAR_CARTAO'">💳</span>
          </div>
          <div class="space-y-1 min-w-0">
            <p class="text-xs text-charcoal font-medium leading-relaxed">
              {{ log.detalhes }}
            </p>
            <div class="flex items-center gap-2 text-[9px] text-ash font-bold uppercase tracking-wide">
              <span>{{ getMembroNome(log.membroId) }}</span>
              <span>•</span>
              <span>{{ log.createdAt ? new Date(log.createdAt).toLocaleString('pt-BR') : '-' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </BottomSheet>
</template>
