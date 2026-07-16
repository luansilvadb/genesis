const PERIODO_STORAGE_KEY = 'divi_periodo_selecionado'

export interface Periodo {
  mes: number
  ano: number
}

export function obterPeriodoSelecionado(fallbackPeriodo?: Periodo): Periodo {
  const salvo = localStorage.getItem(PERIODO_STORAGE_KEY)
  if (salvo) {
    try {
      const parsed = JSON.parse(salvo)
      if (parsed.mes && parsed.ano) {
        const hoje = new Date()
        const diffMeses = Math.abs((Number(parsed.ano) - hoje.getFullYear()) * 12 + (Number(parsed.mes) - (hoje.getMonth() + 1)))
        // Ignora período salvo se estiver mais de 6 meses distante do atual (evita períodos espúrios)
        if (diffMeses <= 6) {
          return { mes: Number(parsed.mes), ano: Number(parsed.ano) }
        }
      }
    } catch { /* JSON inválido no localStorage — ignora e usa fallback */ }
  }
  if (fallbackPeriodo) return fallbackPeriodo
  return { mes: new Date().getMonth() + 1, ano: new Date().getFullYear() }
}

export function salvarPeriodoSelecionado(periodo: Periodo): void {
  localStorage.setItem(PERIODO_STORAGE_KEY, JSON.stringify(periodo))
}
