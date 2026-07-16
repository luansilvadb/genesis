<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { auditLogRepository } from '../../../shared/container'
import { useMembros } from '../../../viewmodels/useMembros'
import IllustrationMascot from '../ui/IllustrationMascot.vue'
import type { AuditLogDto } from '../../../models/repositories/http/HttpAuditLogRepository'

const props = defineProps<{
  activeTenantId: string | null
}>()

defineEmits<{
  (e: 'focusChange', active: boolean): void
}>()

const logs = ref<AuditLogDto[]>([])
const loading = ref(false)
const erro = ref('')

const { membros, carregar: carregarMembros } = useMembros()

const getMembroNome = (id: string) => {
  return membros.value.find(m => m.id === id)?.nome || 'Desconhecido'
}

const carregarLogs = async () => {
  loading.value = true
  erro.value = ''
  try {
    logs.value = await auditLogRepository.listarTodos()
  } catch {
    erro.value = 'Erro ao carregar histórico de atividades.'
    logs.value = []
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await carregarMembros()
  await carregarLogs()
})

watch(() => props.activeTenantId, async () => {
  await carregarMembros()
  await carregarLogs()
})
</script>

<template>
  <div class="space-y-6">
    <div
      v-if="loading"
      class="flex flex-col items-center justify-center py-16 space-y-5"
    >
      <div class="relative w-10 h-10">
        <div class="absolute inset-0 rounded-full border-2 border-stone/30" />
        <div class="absolute inset-0 rounded-full border-2 border-transparent border-t-ember/50 animate-spin" />
      </div>
      <p class="text-[10px] text-ash/50 font-medium uppercase tracking-[0.2em]">
        Buscando atividades...
      </p>
    </div>

    <div
      v-else-if="erro"
      class="flex flex-col items-center justify-center py-16 text-center space-y-4"
    >
      <IllustrationMascot
        variant="coral"
        :size="80"
        mood="confused"
        class="opacity-45"
      />
      <p class="text-xs text-coral font-bold">
        {{ erro }}
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
          <div class="flex items-center gap-2 text-[10px] sm:text-xs text-ash font-bold uppercase tracking-wide">
            <span>{{ getMembroNome(log.membroId) }}</span>
            <span>•</span>
            <span>{{ log.createdAt ? new Date(log.createdAt).toLocaleString('pt-BR') : '-' }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
