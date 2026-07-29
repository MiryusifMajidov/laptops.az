// AI paneli istənilən yerdən açmaq üçün kiçik "bus" (context-siz, route-lar arası işləyir)
type Listener = () => void
let listener: Listener | null = null

export function onOpenAi(fn: Listener): () => void {
  listener = fn
  return () => { if (listener === fn) listener = null }
}

export function openAi(): void {
  listener?.()
}
