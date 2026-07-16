<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Gasto } from '../../../models/entities/Gasto'
import { Dinheiro } from '../../../models/entities/Dinheiro'
import { DivisaoDeGasto } from '../../../models/entities/DivisaoDeGasto'
import { formatarBRL, aplicarMascaraBRLText } from '../../../shared/utils/formatarMoeda'
import Button from '../ui/Button.vue'
import BottomSheet from '../ui/BottomSheet.vue'
import MembroAvatar from '../ui/MembroAvatar.vue'
import { Check, CreditCard, Wallet, Info, Minus, Plus } from 'lucide-vue-next'

interface Props {
  visible: boolean
  gasto: Gasto | null
  membros: { id: string; nome: string }[]
  cartoes: { id: string; nome: string; responsavelPadraoId: string }[]
  faturas?: { id: string; cartaoId: string }[]
  loading?: boolean
}

const props = defineProps<Props>()
const emit = defineEmits(['cancel', 'save'])

const descInput = ref('')
const valorInput = ref(0)
const valorFormatado = ref('')
const quemPaga = ref('')
const activeMethod = ref<'pix' | 'card'>('pix')
const activeCardOwner = ref<string | null>(null)
const selectedSplit = ref<string[]>([])
const installmentsInput = ref(1)

watch(() => props.gasto, (newG) => {
  if (newG) {
    descInput.value = newG.descricao || ''
    const valorEmReais = newG.valorTotal?.centavos ? newG.valorTotal.centavos / 100 : 0
    valorInput.value = valorEmReais
    valorFormatado.value = valorEmReais > 0 ? formatarBRL(valorEmReais, false) : ''
    quemPaga.value = newG.compradorId || ''
    activeMethod.value = newG.method === 'card' ? 'card' : 'pix'
    
    let cardIdResolved: string | null = null
    if (newG.method === 'card') {
      if (newG.faturaId && props.faturas) {
        const fat = props.faturas.find(f => f.id === newG.faturaId)
        if (fat) {
          cardIdResolved = fat.cartaoId
        }
      }
      if (!cardIdResolved && newG.cardOwner) {
        const card = props.cartoes.find(c => c.id === newG.cardOwner || c.responsavelPadraoId === newG.cardOwner)
        if (card) {
          cardIdResolved = card.id
        }
      }
    }
    activeCardOwner.value = cardIdResolved

    selectedSplit.value = newG.divisoes ? newG.divisoes.map(d => d.membroId) : []
    installmentsInput.value = newG.installments || 1
  }
}, { immediate: true })

const handleValorInput = (e: Event) => {
  const target = e.target as HTMLInputElement
  const mascarado = aplicarMascaraBRLText(target.value)
  valorFormatado.value = mascarado
  if (mascarado === '') {
    valorInput.value = 0
  } else {
    const cleanValue = mascarado.replace(/\./g, '').replace(',', '.')
    valorInput.value = parseFloat(cleanValue)
  }
}

const selectMethod = (method: 'pix' | 'card', cardOwner: string | null) => {
  activeMethod.value = method
  activeCardOwner.value = cardOwner
}

const toggleSplit = (memberId: string) => {
  const idx = selectedSplit.value.indexOf(memberId)
  if (idx > -1) {
    if (selectedSplit.value.length === 1) return // mínimo 1 participante
    selectedSplit.value = selectedSplit.value.filter((_, i) => i !== idx)
  } else {
    selectedSplit.value = [...selectedSplit.value, memberId]
  }
}

const ajustarParcelas = (delta: number) => {
  installmentsInput.value = Math.max(1, installmentsInput.value + delta)
}

const infoParcelamento = computed(() => {
  if (installmentsInput.value <= 1) return 'À vista'
  const val = Number(valorInput.value) || 0
  const parcela = val / installmentsInput.value
  return `${installmentsInput.value}x de ${formatarBRL(parcela)}`
})

const calculatedSharesDesc = computed(() => {
  const n = selectedSplit.value.length
  if (n === 0 || valorInput.value <= 0) return 'Digite um valor e selecione participantes'
  
  if (installmentsInput.value > 1) {
    const shareTotal = valorInput.value / n
    const shareParcela = shareTotal / installmentsInput.value
    const formattedTotal = formatarBRL(shareTotal)
    const formattedParcela = formatarBRL(shareParcela)
    
    if (n === props.membros.length) {
      return `Dividido igualmente com todos. Cada um paga ${formattedParcela}/mês (${formattedTotal} no total em ${installmentsInput.value}x)`
    } else if (n === 1) {
      const name = props.membros.find(m => m.id === selectedSplit.value[0])?.nome || ''
      return `Só de ${name}. Assume 100% pagando ${formattedParcela}/mês (${formattedTotal} no total em ${installmentsInput.value}x)`
    } else {
      const names = selectedSplit.value.map(id => props.membros.find(m => m.id === id)?.nome).join(' e ')
      return `Entre ${names}. Cada um paga ${formattedParcela}/mês (${formattedTotal} no total em ${installmentsInput.value}x)`
    }
  }

  const share = valorInput.value / n
  const formatted = formatarBRL(share)
  
  if (n === props.membros.length) {
    return `Dividido igualmente com todos. Cada um paga ${formatted}`
  } else if (n === 1) {
    const name = props.membros.find(m => m.id === selectedSplit.value[0])?.nome || ''
    return `Só de ${name}. Assume 100% no valor de ${formatted}`
  } else {
    const names = selectedSplit.value.map(id => props.membros.find(m => m.id === id)?.nome).join(' e ')
    return `Entre ${names}. Cada um paga ${formatted}`
  }
})

const handleConfirm = () => {
  if (!descInput.value.trim()) return
  if (valorInput.value <= 0) return

  const totalCents = Math.round(valorInput.value * 100)
  const floorShare = Math.floor(totalCents / selectedSplit.value.length)
  const divisoes: DivisaoDeGasto[] = []
  
  selectedSplit.value.forEach(mId => {
    divisoes.push(new DivisaoDeGasto(mId, Dinheiro.deCentavos(floorShare)))
  })

  let remainder = totalCents - (floorShare * selectedSplit.value.length)
  let idx = 0
  const order = [...selectedSplit.value]
  
  if (quemPaga.value && order.includes(quemPaga.value)) {
    const index = order.indexOf(quemPaga.value)
    order.splice(index, 1)
    order.unshift(quemPaga.value)
  }

  while (remainder > 0) {
    const targetMemberId = order[idx % order.length]
    const currentDivIdx = divisoes.findIndex(d => d.membroId === targetMemberId)
    if (currentDivIdx > -1) {
      const originalDiv = divisoes[currentDivIdx]
      divisoes[currentDivIdx] = new DivisaoDeGasto(targetMemberId, Dinheiro.deCentavos(originalDiv.valor.centavos + 1))
    }
    remainder--
    idx++
  }

  emit('save', {
    descricao: descInput.value.trim(),
    valorTotal: Dinheiro.deCentavos(totalCents),
    compradorId: quemPaga.value,
    method: activeMethod.value,
    cardOwner: activeCardOwner.value,
    divisoes,
    installments: installmentsInput.value
  })
}
</script>

<template>
  <BottomSheet
    :model-value="props.visible"
    @update:model-value="val => { if (!val) emit('cancel') }"
  >
    <template #title>
      <h3 class="text-3xl font-display text-charcoal leading-tight">
        Corrigir <span class="text-ember">Lançamento</span>
      </h3>
    </template>

    <div class="space-y-6 pt-2">
      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Descrição</label>
        <input 
          v-model="descInput" 
          type="text" 
          class="w-full px-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-sm text-charcoal focus:border-ember transition-all" 
          placeholder="Ex: Supermercado"
        >
      </div>

      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Valor Total</label>
        <div class="relative">
          <span class="absolute left-4 top-1/2 -translate-y-1/2 text-graphite text-sm font-bold">R$</span>
          <input 
            :value="valorFormatado"
            type="text"
            inputmode="numeric"
            class="w-full pl-10 pr-4 py-3.5 rounded-xl border border-stone bg-canvas outline-none font-bold text-sm text-charcoal focus:border-ember transition-all"
            placeholder="0,00"
            @input="handleValorInput"
          >
        </div>
      </div>

      <div
        v-if="!props.gasto?.isLoan"
        class="space-y-2"
      >
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Método / Cartão</label>
        <div class="grid grid-cols-3 gap-2">
          <button 
            class="flex flex-col items-center gap-2 py-3 rounded-xl transition-all duration-200 cursor-pointer"
            :class="activeMethod === 'pix' ? 'border-2 border-charcoal bg-white text-charcoal font-bold shadow-sm' : 'border-2 border-transparent bg-stone hover:bg-ash/20 text-charcoal'"
            @click="selectMethod('pix', null)"
          >
            <Wallet
              class="w-4 h-4"
              aria-hidden="true"
            />
            <span class="text-[9px] font-bold uppercase tracking-wider">Pix</span>
          </button>
          <button 
            v-for="c in props.cartoes"
            :key="c.id"
            class="flex flex-col items-center gap-2 py-3 rounded-xl transition-all duration-200 cursor-pointer"
            :class="activeMethod === 'card' && activeCardOwner === c.id ? 'border-2 border-charcoal bg-white text-charcoal font-bold shadow-sm' : 'border-2 border-transparent bg-stone hover:bg-ash/20 text-charcoal'"
            @click="selectMethod('card', c.id)"
          >
            <CreditCard
              class="w-4 h-4"
              aria-hidden="true"
            />
            <span class="text-[9px] font-bold uppercase tracking-wider">{{ c.nome }}</span>
          </button>
        </div>
      </div>

      <div
        v-if="activeMethod === 'card' || props.gasto?.isLoan"
        class="space-y-2"
      >
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Parcelamento</label>
        <div class="flex items-center justify-between gap-3 p-3 rounded-xl border border-stone bg-canvas shadow-subtle">
          <button
            type="button"
            class="w-10 h-10 rounded-full bg-stone hover:bg-ash/20 text-charcoal flex items-center justify-center transition-colors border-none cursor-pointer"
            @click="ajustarParcelas(-1)"
          >
            <Minus
              class="w-4 h-4"
              aria-hidden="true"
            />
          </button>
          <div class="text-center">
            <span class="text-lg font-bold text-charcoal tracking-tight">{{ installmentsInput }}x</span>
            <p class="text-[10px] font-semibold text-graphite">
              {{ infoParcelamento }}
            </p>
          </div>
          <button
            type="button"
            class="w-10 h-10 rounded-full bg-stone hover:bg-ash/20 text-charcoal flex items-center justify-center transition-colors border-none cursor-pointer"
            @click="ajustarParcelas(1)"
          >
            <Plus
              class="w-4 h-4"
              aria-hidden="true"
            />
          </button>
        </div>
      </div>

      <div class="space-y-2">
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">
          {{ props.gasto?.isLoan ? 'Quem emprestou?' : activeMethod === 'pix' ? 'Quem fez o Pix?' : `Quem usou o cartão?` }}
        </label>
        <div class="grid grid-cols-3 gap-2">
          <button 
            v-for="m in props.membros"
            :key="m.id"
            class="group py-3 rounded-xl font-bold text-[11px] uppercase tracking-wider transition-all duration-300 cursor-pointer flex flex-col items-center gap-2"
            :class="quemPaga === m.id ? 'border-2 border-charcoal bg-white shadow-sm scale-[1.02] text-charcoal' : 'border-2 border-transparent bg-stone hover:bg-stone/80 text-charcoal'"
            @click="quemPaga = m.id"
          >
            <MembroAvatar
              :nome="m.nome"
              size="sm"
              :variant="quemPaga === m.id ? 'ember' : 'sky'"
            />
            <span class="truncate max-w-full px-1">{{ m.nome }}</span>
          </button>
        </div>
      </div>

      <div
        v-if="!props.gasto?.isLoan"
        class="space-y-2"
      >
        <label class="block text-[10px] font-bold uppercase text-graphite tracking-widest ml-1">Dividir com</label>
        <div class="grid grid-cols-3 gap-2">
          <button 
            v-for="m in props.membros"
            :key="m.id"
            class="group relative py-3 rounded-xl font-bold text-[11px] uppercase tracking-wider transition-all duration-300 flex flex-col items-center gap-2 border-none cursor-pointer"
            :class="selectedSplit.includes(m.id) ? 'bg-white shadow-subtle scale-[1.02] text-charcoal' : 'bg-stone/50 text-graphite opacity-60 hover:opacity-100'"
            @click="toggleSplit(m.id)"
          >
            <MembroAvatar
              :nome="m.nome"
              size="sm"
              :variant="selectedSplit.includes(m.id) ? 'meadow' : 'sky'"
            />
            <span class="truncate max-w-full px-1">{{ m.nome }}</span>
            <div
              v-if="selectedSplit.includes(m.id)"
              class="absolute top-1.5 right-1.5 animate-in zoom-in-50 duration-300"
            >
              <Check class="w-3.5 h-3.5 text-meadow" />
            </div>
          </button>
        </div>
      </div>

      <div
        v-if="!props.gasto?.isLoan"
        class="flex gap-4 p-4 rounded-xl bg-meadow/5 border border-meadow/20 text-meadow"
      >
        <Info
          class="w-5 h-5 shrink-0 mt-0.5"
          aria-hidden="true"
        />
        <div class="space-y-1">
          <p class="text-[10px] font-bold uppercase tracking-widest">
            Resumo do Rateio
          </p>
          <p class="text-xs font-semibold leading-relaxed">
            {{ calculatedSharesDesc }}
          </p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="grid grid-cols-2 gap-3">
        <Button
          variant="secondary"
          class="font-bold uppercase tracking-widest text-[10px] h-12"
          @click="emit('cancel')"
        >
          Voltar
        </Button>
        <Button
          variant="primary"
          class="font-bold uppercase tracking-widest text-[10px] h-12"
          :disabled="!descInput.trim() || valorInput <= 0 || props.loading"
          :loading="props.loading"
          @click="handleConfirm"
        >
          Salvar Alterações
        </Button>
      </div>
    </template>
  </BottomSheet>
</template>

<style scoped>
</style>
