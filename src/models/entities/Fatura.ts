export type FaturaStatus = 'ABERTA' | 'FECHADA'

export interface FaturaPeriodo {
  mes: number // 1 a 12
  ano: number
}

interface FaturaProps {
  id: string
  cartaoId: string
  periodo: FaturaPeriodo
  responsavelId: string
  status: FaturaStatus
  dataPagamentoBanco?: Date
}


export class Fatura {
  public readonly id: string
  public readonly cartaoId: string
  public readonly periodo: FaturaPeriodo
  public readonly responsavelId: string
  public readonly status: FaturaStatus
  public readonly dataPagamentoBanco?: Date

  constructor(props: FaturaProps) {
    this.id = props.id
    this.cartaoId = props.cartaoId
    this.periodo = props.periodo
    this.responsavelId = props.responsavelId
    this.status = props.status
    this.dataPagamentoBanco = props.dataPagamentoBanco
  }

  fechar(opts: { responsavelId?: string; dataPagamentoBanco: Date }): Fatura {
    return new Fatura({
      id: this.id,
      cartaoId: this.cartaoId,
      periodo: this.periodo,
      responsavelId: opts?.responsavelId || this.responsavelId,
      status: 'FECHADA',
      dataPagamentoBanco: opts.dataPagamentoBanco
    })
  }


  reabrir(): Fatura {
    return new Fatura({
      id: this.id,
      cartaoId: this.cartaoId,
      periodo: this.periodo,
      responsavelId: this.responsavelId,
      status: 'ABERTA',
      dataPagamentoBanco: undefined
    })
  }
}
