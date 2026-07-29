// Sayt üçün public API (auth yoxdur) — eyni origin-dən (nisbi).
// Dev: Vite proxy → :8080; Prod: Go binary eyni portdan verir.
const HOST = ''
const BASE = HOST + '/api/public'

export async function api<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, { headers: { 'Content-Type': 'application/json' }, ...opts })
  if (!res.ok) {
    const m = await res.json().catch(() => ({ error: res.statusText }))
    const err = new Error(m.error || 'Xəta baş verdi') as Error & { status?: number }
    err.status = res.status // backend cavab verdi (limit, 502 və s.) — status var; şəbəkə xətasında yoxdur
    throw err
  }
  return res.json() as Promise<T>
}

export const imgURL = (p: string) => (p ? (p.startsWith('http') ? p : HOST + p) : '')

// sayt ziyarəti — sessiya başına bir dəfə loglanır (admin panel statistikası üçün)
export function trackVisit() {
  try {
    if (sessionStorage.getItem('visited')) return
    sessionStorage.setItem('visited', '1')
    fetch(HOST + '/api/public/visit', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path: location.pathname }), keepalive: true,
    }).catch(() => {})
  } catch { /* ignore */ }
}

// kart/siyahı üçün kiçildilmiş, sıxılmış şəkil (backend /img proksisi — həm /uploads, həm xarici linklər).
// Tam ölçü yalnız məhsul detailində (imgURL) qalır.
export const thumbURL = (p: string, w = 400) => (p ? `${HOST}/img?u=${encodeURIComponent(p)}&w=${w}` : '')

// AI köməkçi — backend Gemini proksisinə mesaj göndər (açar server tərəfdədir)
export interface AiChatMsg { role: 'user' | 'assistant'; text: string }
export interface AiChatReply { reply: string; products: Record<string, Product> }
// söhbəti admin-də qruplaşdırmaq üçün sessiya-əsaslı sabit id (chat da sessionStorage-dədir)
function aiConvId(): string {
  try {
    let id = sessionStorage.getItem('ai_cid')
    if (!id) { id = 'web-' + Date.now().toString(36) + Math.random().toString(36).slice(2, 8); sessionStorage.setItem('ai_cid', id) }
    return id
  } catch { return 'web-' + Math.random().toString(36).slice(2, 10) }
}
export const aiChat = (messages: AiChatMsg[], lang: string) =>
  api<AiChatReply>('/ai-chat?lang=' + encodeURIComponent(lang), { method: 'POST', body: JSON.stringify({ messages, conversation_id: aiConvId(), source: 'web' }) })

export interface Attr { id: number; name: string }
export interface ItemValue { attribute_id: number; attribute: Attr; value: string }
export interface Category { id: number; name: string }
export interface Product {
  id: number; name: string; serial: string; price: number; discount: number
  category: Category; card_image: string; gallery: string
  values: ItemValue[]; created_at: string
}
export interface CatCount { name: string; count: number }

export const money = (n: number) => '₼' + Math.round(n).toLocaleString('en-US')

// endirim (₼ ilə) — varsa real qiymət price - discount
export const hasDiscount = (p: { discount?: number; price: number }) => (p.discount ?? 0) > 0 && (p.discount ?? 0) < p.price
export const finalPrice = (p: { discount?: number; price: number }) => (hasDiscount(p) ? p.price - (p.discount || 0) : p.price)

// dəyər tapıcı (xüsusiyyət adına görə)
export const attrVal = (p: Product, name: string) =>
  p.values?.find((v) => v.attribute?.name === name)?.value ?? ''

// qısa spesifikasiya sətri (kart üçün)
export function specLine(p: Product): string {
  const parts = ['Prosessor', 'RAM', 'SSD', 'Ekran kartı', 'Yaddaş']
    .map((n) => attrVal(p, n)).filter(Boolean)
  return parts.slice(0, 3).join(' · ')
}

export function gallery(p: Product): string[] {
  try { const g = p.gallery ? JSON.parse(p.gallery) : []; return Array.isArray(g) ? g : [] } catch { return [] }
}

// sayt kateqoriya adları (DB) → menyu adı
export const CAT_LABEL: Record<string, string> = {
  Notebook: 'Notebooklar', İşlənmiş: 'İşlənmiş', Telefon: 'Telefon', Aksesuar: 'Aksesuar',
}

// ---- çoxdilli ----
export interface Language {
  id: number; code: string; name: string; icon: string
  enabled: boolean; is_default: boolean; sort: number
}
export const fetchLanguages = () => api<Language[]>('/languages')
export const fetchI18n = (lang: string) =>
  api<Record<string, string>>('/i18n?lang=' + encodeURIComponent(lang))
// kataloq terminləri (kateqoriya/xüsusiyyət/dəyər) — {azText: tərcümə}
export const fetchTerms = (lang: string) =>
  api<Record<string, string>>('/terms?lang=' + encodeURIComponent(lang))
// sayt tənzimləmələri (xəritə koordinatı, seçilmiş məhsul)
export const fetchSettings = () => api<Record<string, string>>('/settings')
