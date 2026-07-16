<script setup lang="ts">
import { computed } from 'vue'
import BottomSheet from '../../ui/BottomSheet.vue'
import Button from '../../ui/Button.vue'
import { Home, Check, Copy, LogOut } from 'lucide-vue-next'
import type { CasasModalView } from '../../../../viewmodels/useCasasMultitenant'

const props = defineProps<{
  visible: boolean
  casasMultitenant: CasasModalView
}>()

const emit = defineEmits(['close'])

const nomeNovaCasa = computed({
  get: () => props.casasMultitenant.form.nomeNovaCasa,
  // eslint-disable-next-line vue/no-mutating-props
  set: (val: string) => { props.casasMultitenant.form.nomeNovaCasa = val }
})

const codigoConvite = computed({
  get: () => props.casasMultitenant.form.codigoConvite,
  // eslint-disable-next-line vue/no-mutating-props
  set: (val: string) => { props.casasMultitenant.form.codigoConvite = val }
})
</script>

<template>
  <BottomSheet 
    :model-value="visible" 
    subtitle="Selecione uma casa ativa ou gerencie seus grupos financeiros. Compartilhe o código de convite para trazer novos membros." 
    max-height="90dvh"
    @update:model-value="val => { if (!val) emit('close') }"
  >
    <template #title>
      <h3 class="text-3xl font-display text-charcoal leading-tight">
        Minhas <span class="text-ember">Casas</span>
      </h3>
    </template>

    <div class="space-y-6 pt-2">
      <div class="space-y-3">
        <h4 class="text-[10px] font-bold uppercase tracking-widest text-graphite ml-1 block">
          Alternar de Casa
        </h4>
        <div class="grid gap-2">
          <div 
            v-for="casa in casasMultitenant.casas" 
            :key="casa.id"
            class="p-3.5 rounded-2xl border cursor-pointer transition-all flex items-center justify-between"
            :class="casasMultitenant.activeTenantId === casa.id ? 'border-ember bg-ember/5 text-ember font-semibold' : 'border-stone bg-canvas text-charcoal hover:bg-stone/30'"
            @click="casasMultitenant.selecionarCasa(casa.id); emit('close')"
          >
            <div class="flex items-center gap-3">
              <Home
                class="w-4 h-4 shrink-0"
                :class="casasMultitenant.activeTenantId === casa.id ? 'text-ember' : 'text-ash'"
              />
              <span class="text-sm font-medium">{{ casa.name }}</span>
            </div>
            <div
              class="flex items-center gap-2"
              @click.stop
            >
              <code class="text-[10px] bg-stone/50 px-2 py-1 rounded text-ash font-mono select-all">
                {{ casa.inviteCode }}
              </code>
              <button 
                class="p-1 hover:bg-stone rounded transition-colors border-none bg-transparent cursor-pointer" 
                @click="casasMultitenant.copyInviteCode(casa.inviteCode)"
              >
                <Check
                  v-if="casasMultitenant.copiedCode === casa.inviteCode"
                  class="w-3.5 h-3.5 text-meadow"
                />
                <Copy
                  v-else
                  class="w-3.5 h-3.5 text-ash"
                />
              </button>
            </div>
          </div>
        </div>
      </div>

      <div class="h-px bg-stone/60 my-2" />

      <div class="space-y-3">
        <h4 class="text-[10px] font-bold uppercase tracking-widest text-graphite ml-1 block">
          Criar Nova Casa
        </h4>
        <div class="flex gap-2">
          <input 
            v-model="nomeNovaCasa"
            placeholder="Ex: República Central"
            class="flex-1 bg-canvas border border-stone rounded-xl px-4 py-3.5 text-sm font-bold text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all"
            @keyup.enter="casasMultitenant.criarNovaCasa"
          >
          <Button
            :loading="casasMultitenant.isCreating"
            class="font-bold uppercase tracking-widest text-[10px] h-12"
            @click="casasMultitenant.criarNovaCasa"
          >
            Criar
          </Button>
        </div>
      </div>

      <div class="space-y-3 pt-2">
        <h4 class="text-[10px] font-bold uppercase tracking-widest text-graphite ml-1 block">
          Entrar com Código
        </h4>
        <div class="flex gap-2">
          <input 
            v-model="codigoConvite"
            placeholder="Ex: CASA-7F2A1"
            class="flex-1 bg-canvas border border-stone rounded-xl px-4 py-3.5 text-sm font-bold text-charcoal placeholder:text-ash focus:outline-none focus:border-ember transition-all"
            @keyup.enter="casasMultitenant.entrarPorCodigo"
          >
          <Button
            :loading="casasMultitenant.isEntering"
            class="font-bold uppercase tracking-widest text-[10px] h-12"
            @click="casasMultitenant.entrarPorCodigo"
          >
            Entrar
          </Button>
        </div>
      </div>

      <div
        v-if="casasMultitenant.form.errorCasa"
        class="text-xs text-coral font-medium pt-2"
      >
        {{ casasMultitenant.form.errorCasa }}
      </div>
    </div>
    
    <template #footer>
      <div class="flex justify-between items-center w-full gap-4">
        <Button 
          variant="ghost"
          class="flex items-center gap-2 text-coral !p-0 h-auto" 
          @click="casasMultitenant.handleLogoutClick"
        >
          <LogOut class="w-4 h-4" />
          <span class="text-[10px] uppercase tracking-widest font-bold">Sair da Conta</span>
        </Button>
        <Button
          variant="secondary"
          class="font-bold uppercase tracking-widest text-[10px] h-12 px-8"
          @click="emit('close')"
        >
          Fechar
        </Button>
      </div>
    </template>
  </BottomSheet>
</template>
