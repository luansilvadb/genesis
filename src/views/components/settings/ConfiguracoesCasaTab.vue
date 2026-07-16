<script setup lang="ts">
import { ref } from 'vue'
import { Shield, ShieldAlert } from 'lucide-vue-next'
import { useMembros } from '../../../viewmodels/useMembros'
import { useToast } from '../../../composables/useToast'
import { mensagemErro } from '../../../shared/utils/mensagemErro'

const props = defineProps<{
  activeTenantId: string | null
}>()

const emit = defineEmits(['focus-change'])

const { tenantPermissions, atualizarPermissions } = useMembros()
const toast = useToast()
const salvando = ref<Record<string, boolean>>({})
const activeRole = ref<'MORADOR' | 'VISUALIZADOR'>('MORADOR')

const togglePermission = async (key: 'ALLOW_LANCAR_GASTO' | 'ALLOW_GERENCIAR_CARTOES' | 'ALLOW_GERENCIAR_CONTAS_FIXAS' | 'ALLOW_REGISTRAR_NETTING' | 'ALLOW_VER_AUDIT_LOGS' | 'ALLOW_ALTERAR_RENDA' | 'ALLOW_ALTERAR_NOME' | 'ALLOW_FECHAR_PERIODO') => {
  const saveKey = `${activeRole.value}-${key}`
  if (salvando.value[saveKey]) return
  salvando.value[saveKey] = true
  
  const currentRolePerms = tenantPermissions.value[activeRole.value] || {
    ALLOW_LANCAR_GASTO: activeRole.value === 'MORADOR',
    ALLOW_GERENCIAR_CARTOES: activeRole.value === 'MORADOR',
    ALLOW_GERENCIAR_CONTAS_FIXAS: activeRole.value === 'MORADOR',
    ALLOW_REGISTRAR_NETTING: activeRole.value === 'MORADOR',
    ALLOW_VER_AUDIT_LOGS: activeRole.value === 'MORADOR',
    ALLOW_ALTERAR_RENDA: activeRole.value === 'MORADOR',
    ALLOW_ALTERAR_NOME: activeRole.value === 'MORADOR',
    ALLOW_FECHAR_PERIODO: activeRole.value === 'MORADOR'
  }
  const currentValue = currentRolePerms[key]

  try {
    await atualizarPermissions(activeRole.value, {
      [key]: !currentValue
    })
    toast.show('Permissão atualizada com sucesso', 'success')
  } catch (error: unknown) {
    toast.show(mensagemErro(error, 'Erro ao atualizar permissão'), 'error')
  } finally {
    salvando.value[saveKey] = false
  }
}
</script>

<template>
  <div class="space-y-6 animate-in fade-in duration-300">
    <div class="bg-white border border-stone/30 rounded-2xl shadow-subtle overflow-hidden">
      <!-- Header -->
      <div class="px-6 pt-6 pb-4 border-b border-stone/20">
        <h3 class="text-heading-sm text-charcoal flex items-center gap-2">
          <Shield class="w-5 h-5 text-ember" />
          Permissões da Casa
        </h3>
        <p class="text-xs text-ash font-medium mt-1 uppercase tracking-wider">
          Defina o que cada papel de membro pode gerenciar na casa
        </p>
      </div>

      <!-- Role Selector -->
      <div class="px-6 pt-6">
        <div class="flex gap-2 p-1 bg-parchment/60 border border-stone/30 rounded-xl">
          <button 
            @click="activeRole = 'MORADOR'"
            type="button"
            class="flex-1 py-2 px-3 text-xs font-bold rounded-lg border-none cursor-pointer transition-all duration-200"
            :class="activeRole === 'MORADOR' ? 'bg-charcoal text-white shadow-subtle' : 'bg-transparent text-ash hover:text-charcoal hover:bg-stone/10'"
          >
            Morador
          </button>
          <button 
            @click="activeRole = 'VISUALIZADOR'"
            type="button"
            class="flex-1 py-2 px-3 text-xs font-bold rounded-lg border-none cursor-pointer transition-all duration-200"
            :class="activeRole === 'VISUALIZADOR' ? 'bg-charcoal text-white shadow-subtle' : 'bg-transparent text-ash hover:text-charcoal hover:bg-stone/10'"
          >
            Visualizador
          </button>
        </div>
      </div>

      <!-- Info Alert -->
      <div class="p-4 mx-6 mt-6 bg-parchment border border-stone rounded-2xl flex items-start gap-3">
        <ShieldAlert class="w-5 h-5 text-ember shrink-0 mt-0.5" />
        <div class="space-y-1">
          <h4 class="text-xs font-bold text-charcoal">Configuração por papel de acesso</h4>
          <p class="text-xs text-ash leading-relaxed">
            As restrições abaixo aplicam-se aos membros que possuem a Role selecionada. Administradores possuem soberania absoluta e mantêm acesso irrestrito a todas as funcionalidades do sistema, independentemente destas configurações.
          </p>
        </div>
      </div>

      <!-- Toggles List -->
      <div class="p-6 divide-y divide-stone/20 space-y-4">
        <!-- Lançar Gastos -->
        <div class="flex items-center justify-between gap-4 pt-1 first:pt-0">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Lançar Gastos e Despesas</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite lançar novos gastos comuns, parcelados e compras em cartões de crédito.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_LANCAR_GASTO')"
            :disabled="salvando[`${activeRole}-ALLOW_LANCAR_GASTO`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_LANCAR_GASTO ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_LANCAR_GASTO`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_LANCAR_GASTO ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <!-- Gerenciar Cartões -->
        <div class="flex items-center justify-between gap-4 pt-4">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Gerenciar Cartões de Crédito</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite cadastrar novos cartões de crédito e excluir os cartões existentes da moradia.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_GERENCIAR_CARTOES')"
            :disabled="salvando[`${activeRole}-ALLOW_GERENCIAR_CARTOES`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_GERENCIAR_CARTOES ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_GERENCIAR_CARTOES`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_GERENCIAR_CARTOES ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <!-- Gerenciar Contas Fixas -->
        <div class="flex items-center justify-between gap-4 pt-4">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Gerenciar Contas Fixas</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite criar novos modelos de contas fixas e excluir/estornar lançamentos automáticos.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_GERENCIAR_CONTAS_FIXAS')"
            :disabled="salvando[`${activeRole}-ALLOW_GERENCIAR_CONTAS_FIXAS`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_GERENCIAR_CONTAS_FIXAS ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_GERENCIAR_CONTAS_FIXAS`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_GERENCIAR_CONTAS_FIXAS ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <!-- Registrar Netting -->
        <div class="flex items-center justify-between gap-4 pt-4">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Registrar Acertos (Netting)</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite dar baixa e registrar pagamentos de netting no painel de fechamento do mês.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_REGISTRAR_NETTING')"
            :disabled="salvando[`${activeRole}-ALLOW_REGISTRAR_NETTING`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_REGISTRAR_NETTING ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_REGISTRAR_NETTING`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_REGISTRAR_NETTING ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <!-- Ver Histórico / Audit Logs -->
        <div class="flex items-center justify-between gap-4 pt-4">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Visualizar Histórico de Auditoria</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite acessar a gaveta de logs de auditoria e ver quem executou cada ação na moradia.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_VER_AUDIT_LOGS')"
            :disabled="salvando[`${activeRole}-ALLOW_VER_AUDIT_LOGS`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_VER_AUDIT_LOGS ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_VER_AUDIT_LOGS`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_VER_AUDIT_LOGS ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <!-- Alterar Renda -->
        <div class="flex items-center justify-between gap-4 pt-4">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Alterar Renda do Perfil</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite que os membros editem o valor de sua renda no perfil de morador.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_ALTERAR_RENDA')"
            :disabled="salvando[`${activeRole}-ALLOW_ALTERAR_RENDA`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_ALTERAR_RENDA ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_ALTERAR_RENDA`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_ALTERAR_RENDA ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <!-- Alterar Nome -->
        <div class="flex items-center justify-between gap-4 pt-4">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Alterar Nome de Exibição</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite que os membros alterem o seu nome de exibição no perfil de morador.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_ALTERAR_NOME')"
            :disabled="salvando[`${activeRole}-ALLOW_ALTERAR_NOME`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_ALTERAR_NOME ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_ALTERAR_NOME`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_ALTERAR_NOME ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>

        <!-- Encerrar e Reabrir Período -->
        <div class="flex items-center justify-between gap-4 pt-4">
          <div class="space-y-1">
            <h4 class="text-xs font-bold text-charcoal">Encerrar e Reabrir Período</h4>
            <p class="text-xs text-ash leading-relaxed">
              Permite que os membros encerrem o mês atual ou reabram períodos já fechados.
            </p>
          </div>
          <button
            @click="togglePermission('ALLOW_FECHAR_PERIODO')"
            :disabled="salvando[`${activeRole}-ALLOW_FECHAR_PERIODO`]"
            class="w-11 h-6 rounded-full p-0.5 border-none cursor-pointer transition-colors focus:outline-none shrink-0"
            :class="[
              (tenantPermissions[activeRole]?.ALLOW_FECHAR_PERIODO ?? (activeRole === 'MORADOR')) ? 'bg-meadow' : 'bg-stone',
              salvando[`${activeRole}-ALLOW_FECHAR_PERIODO`] ? 'opacity-50 cursor-not-allowed' : ''
            ]"
          >
            <div
              class="bg-white w-5 h-5 rounded-full shadow-subtle transform transition-transform"
              :class="(tenantPermissions[activeRole]?.ALLOW_FECHAR_PERIODO ?? (activeRole === 'MORADOR')) ? 'translate-x-5' : 'translate-x-0'"
            />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
