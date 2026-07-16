import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import BottomSheetAuditLogs from './BottomSheetAuditLogs.vue'

const TIPOS_ACAO = ['CRIAR_GASTO', 'EDITAR_GASTO', 'EXCLUIR_GASTO', 'ALTERAR_RENDA', 'CRIAR_MEMBRO', 'CRIAR_CARTAO'] as const

function makeLog(acao: string, index: number) {
  return {
    id: `log-${index}`,
    acao,
    detalhes: `Detalhes ${acao}`,
    membroId: 'm1',
    createdAt: new Date().toISOString(),
  }
}

function createWrapper(props: Record<string, any> = {}) {
  return mount(BottomSheetAuditLogs, {
    props: {
      visible: true,
      logs: [],
      loading: false,
      getMembroNome: () => 'Membro Teste',
      ...props,
    },
    global: {
      stubs: {
        BottomSheet: {
          template: '<div><slot name="title" /><slot /></div>',
        },
        IllustrationMascot: {
          template: '<div />',
        },
      },
    },
    attachTo: document.body,
  })
}

describe('BottomSheetAuditLogs', () => {
  it('deve renderizar todos os tipos de ação sem erro', () => {
    const logs = TIPOS_ACAO.map((acao, i) => makeLog(acao, i))
    const wrapper = createWrapper({ logs })
    TIPOS_ACAO.forEach((acao) => {
      expect(wrapper.text()).toContain(`Detalhes ${acao}`)
    })
  })

  it('cada tipo de ação deve ter uma classe CSS única no badge', () => {
    const logs = TIPOS_ACAO.map((acao, i) => makeLog(acao, i))
    const wrapper = createWrapper({ logs })

    const badges = wrapper.findAll('div.w-8')
    expect(badges.length).toBe(TIPOS_ACAO.length)

    const classes = badges.map(b => b.attributes('class') || '')
    const uniqueClasses = new Set(classes)
    expect(uniqueClasses.size).toBe(TIPOS_ACAO.length)
  })

  it('deve exibir estado vazio quando não há logs', () => {
    const wrapper = createWrapper({ logs: [], loading: false })
    expect(wrapper.text()).toContain('Nenhuma atividade registrada')
  })

  it('deve exibir spinner quando loading for true', () => {
    const wrapper = createWrapper({ logs: [], loading: true })
    expect(wrapper.text()).toContain('Carregando')
  })
})
