import { describe, it, expect, beforeEach, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import ConfiguracoesMembros from './ConfiguracoesMembros.vue'
import { useMembros } from '../../viewmodels/useMembros'

vi.mock('../../viewmodels/useMembros', () => ({
  useMembros: vi.fn()
}))

vi.mock('../../viewmodels/useCasasMultitenant', () => ({
  useCasasMultitenant: () => ({
    activeTenantId: ref('tenant-123')
  })
}))

vi.mock('../../viewmodels/useCartoesEFaturas', () => ({
  useCartoesEFaturas: () => ({
    resetar: vi.fn(),
    cartoes: ref([]),
    adicionarCartao: vi.fn(),
    excluirCartao: vi.fn()
  })
}))

describe('ConfiguracoesMembros', () => {
  const mockMembros = [
    { id: '1', nome: 'Luan', ativo: true, role: 'ADMIN', userId: 'user-luan' },
    { id: '2', nome: 'Maria', ativo: true, role: 'MORADOR', userId: 'user-maria' },
    { id: '3', nome: 'Joao', ativo: false, role: 'MORADOR', userId: 'user-joao' }
  ]

  const mockAdicionarMembro = vi.fn()
  const mockDesativarMembro = vi.fn()
  const mockAtivarMembro = vi.fn()
  const mockAtualizarRoleMembro = vi.fn()
  const mockAtualizarNomeMembro = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    
    ;(useMembros as any).mockReturnValue({
      membros: ref(mockMembros),
      ativos: ref(mockMembros.filter(m => m.ativo)),
      currentMembro: ref(mockMembros[0]), // Luan logado (ADMIN)
      adicionarMembro: mockAdicionarMembro,
      desativarMembro: mockDesativarMembro,
      ativarMembro: mockAtivarMembro,
      atualizarRoleMembro: mockAtualizarRoleMembro,
      atualizarNomeMembro: mockAtualizarNomeMembro,
      carregar: vi.fn()
    })
  })

  it('deve renderizar a aba Meu Perfil por padrao com dados do usuario logado', () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true
        } 
      }
    })
    expect(wrapper.text()).toContain('Perfil dos Usuários')
    expect(wrapper.text()).toContain('Luan') // Nome do currentMembro
    expect(wrapper.text()).toContain('Sair da Conta')
    expect(wrapper.findComponent({ name: 'ConfiguracoesCartoes' }).exists()).toBe(true)
  })

  it('deve renderizar a lista de moradores ao alternar para a aba Acessos', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true
        } 
      }
    })
    
    const botoes = wrapper.findAll('button')
    const botaoAcesso = botoes.find(b => b.text().includes('Acessos'))
    expect(botaoAcesso).toBeDefined()
    await botaoAcesso?.trigger('click')

    expect(wrapper.text()).toContain('Quem mora aqui')
    expect(wrapper.text()).toContain('Luan')
    expect(wrapper.text()).toContain('Maria')
    expect(wrapper.text()).toContain('Joao')
  })

  it('deve chamar adicionarMembro ao enviar o formulario de novo membro', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true,
          CargoFormBottomSheet: true, 
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true
        } 
      }
    })
    
    const botoes = wrapper.findAll('button')
    const botaoAcesso = botoes.find(b => b.text().includes('Acessos'))
    await botaoAcesso?.trigger('click')

    const botaoNovoMorador = wrapper.findAll('button').find(b => b.text().includes('Novo Morador'))
    await botaoNovoMorador?.trigger('click')

    const form = wrapper.findComponent({ name: 'MembroFormBottomSheet' })
    expect(form.exists()).toBe(true)
    await form.vm.$emit('salvar', { nome: 'Nova', email: 'nova@email.com', password: '123' })
    expect(mockAdicionarMembro).toHaveBeenCalledWith('Nova', 'nova@email.com', '123', undefined)
  })

  it('deve emitir logout ao clicar no botao Sair da Conta', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true
        } 
      }
    })

    const botaoLogout = wrapper.findAll('button').find(b => b.text().includes('Sair da Conta'))
    expect(botaoLogout).toBeDefined()
    await botaoLogout?.trigger('click')

    expect(wrapper.emitted('logout')).toBeTruthy()
  })

  it('deve habilitar o modo de edicao inline do nome ao clicar no botao de editar', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true,
          Edit2: true,
          Check: true,
          X: true
        } 
      }
    })

    expect(wrapper.find('h3.truncate').text()).toBe('Luan')

    const botaoEditar = wrapper.find('button[aria-label="Editar nome"]')
    expect(botaoEditar.exists()).toBe(true)
    await botaoEditar.trigger('click')

    const inputNome = wrapper.find('input[type="text"]')
    expect(inputNome.exists()).toBe(true)
    expect((inputNome.element as HTMLInputElement).value).toBe('Luan')
  })

  it('deve chamar atualizarNomeMembro com sucesso ao preencher nome valido e confirmar', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true,
          Edit2: true,
          Check: true,
          X: true
        } 
      }
    })

    await wrapper.find('button[aria-label="Editar nome"]').trigger('click')

    const inputNome = wrapper.find('input[type="text"]')
    await inputNome.setValue('Luan Editado')

    const botaoSalvar = wrapper.find('button[aria-label="Salvar nome"]')
    await botaoSalvar.trigger('click')

    expect(mockAtualizarNomeMembro).toHaveBeenCalledWith('1', 'Luan Editado')
  })

  it('deve cancelar a edicao do nome ao clicar no botao de cancelar', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true,
          Edit2: true,
          Check: true,
          X: true
        } 
      }
    })

    await wrapper.find('button[aria-label="Editar nome"]').trigger('click')

    expect(wrapper.find('input[type="text"]').exists()).toBe(true)

    await wrapper.find('button[aria-label="Cancelar edição"]').trigger('click')

    expect(wrapper.find('input[type="text"]').exists()).toBe(false)
    expect(wrapper.find('h3.truncate').text()).toBe('Luan')
  })

  it('deve ocultar cabecalho, abas, card de perfil e rodape sob o estado isCartaoFocus (Modo Foco)', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true
        } 
      }
    })

    // Estado inicial: Modo Foco inativo
    expect(wrapper.find('h2.text-heading-sm').exists()).toBe(true)
    expect(wrapper.text()).toContain('Perfil dos Usuários')
    expect(wrapper.text()).toContain('Meu Perfil')
    expect(wrapper.text()).toContain('Sair da Conta')
    expect(wrapper.text()).toContain('Voltar ao Dashboard')

    // Ativar o Modo Foco
    const configuracoesCartoes = wrapper.findComponent({ name: 'ConfiguracoesCartoes' })
    await configuracoesCartoes.vm.$emit('focus-change', true)
    await flushPromises()

    // Elementos devem sumir
    expect(wrapper.find('h2.text-heading-sm').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Perfil dos Usuários')
    expect(wrapper.text()).not.toContain('Meu Perfil')
    expect(wrapper.text()).not.toContain('Sair da Conta')
    expect(wrapper.text()).not.toContain('Voltar ao Dashboard')

    // Desativar o Modo Foco
    await configuracoesCartoes.vm.$emit('focus-change', false)
    await flushPromises()

    // Elementos devem reaparecer
    expect(wrapper.find('h2.text-heading-sm').exists()).toBe(true)
    expect(wrapper.text()).toContain('Perfil dos Usuários')
    expect(wrapper.text()).toContain('Meu Perfil')
    expect(wrapper.text()).toContain('Sair da Conta')
    expect(wrapper.text()).toContain('Voltar ao Dashboard')
  })

  it('deve ocultar cabecalho, abas, card de perfil e rodape ao abrir a edicao de morador (Modo Foco)', async () => {
    const wrapper = mount(ConfiguracoesMembros, {
      global: { 
        stubs: { 
          BottomSheet: true, 
          MembroFormBottomSheet: true, 
          CargoFormBottomSheet: true,
          ConfiguracoesCartoes: true,
          MembroAvatar: true,
          ChevronRight: true,
          User: true,
          Shield: true,
          Plus: true
        } 
      }
    })

    // Ir para a aba controle de acesso
    const botoes = wrapper.findAll('button')
    const botaoAcesso = botoes.find(b => b.text().includes('Acessos'))
    await botaoAcesso?.trigger('click')
    await flushPromises()

    // Estado inicial: Modo Foco inativo
    expect(wrapper.find('h2.text-heading-sm').exists()).toBe(true)
    expect(wrapper.text()).toContain('Perfil dos Usuários')
    expect(wrapper.text()).toContain('Voltar ao Dashboard')

    // Clicar em um membro para editar (abrir edicaoMembro/mostrarBottomSheet)
    const items = wrapper.findAll('.cursor-pointer')
    const itemMembro = items.find(i => i.text().includes('Maria'))
    await itemMembro?.trigger('click')
    await flushPromises()

    // Elementos devem sumir devido ao Modo Foco na edição
    expect(wrapper.find('h2.text-heading-sm').exists()).toBe(false)
    expect(wrapper.text()).not.toContain('Perfil dos Usuários')
    expect(wrapper.text()).not.toContain('Voltar ao Dashboard')

    // Clicar no botão Cancelar da edição para fechar
    const botaoCancelar = wrapper.findAll('button').find(b => b.text().includes('Cancelar'))
    await botaoCancelar?.trigger('click')
    await flushPromises()

    // Elementos devem reaparecer
    expect(wrapper.find('h2.text-heading-sm').exists()).toBe(true)
    expect(wrapper.text()).toContain('Perfil dos Usuários')
    expect(wrapper.text()).toContain('Voltar ao Dashboard')
  })
})
