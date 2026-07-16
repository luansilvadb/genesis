interface CartaoProps {
  id: string
  nome: string
  diaFechamento: number
  responsavelPadraoId: string
}

export class Cartao {
  public readonly id: string
  public readonly nome: string
  public readonly diaFechamento: number
  public readonly responsavelPadraoId: string

  constructor(props: CartaoProps) {
    this.id = props.id
    this.nome = props.nome
    this.diaFechamento = props.diaFechamento
    this.responsavelPadraoId = props.responsavelPadraoId
  }
}
