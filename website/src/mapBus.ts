// Mağaza xəritəsi modalını istənilən yerdən açmaq üçün sadə bus (aiBus kimi)
let handler: (() => void) | null = null

export function openMap() { handler?.() }

export function onOpenMap(h: () => void) {
  handler = h
  return () => { if (handler === h) handler = null }
}
