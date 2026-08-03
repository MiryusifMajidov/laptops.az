// API klienti — eyni origin-dən (nisbi /api), token ilə.
// Dev-də Vite proxy /api və /uploads-u :8080 backend-ə yönləndirir (vite.config.ts).
// Prod-da Go binary həm API-ni həm frontend-i eyni portdan verir → nisbi ünvan hər yerdə işləyir.
const HOST = ''
const BASE = HOST + '/api'

function authHeaders(): Record<string, string> {
  const t = localStorage.getItem('token')
  return t ? { Authorization: 'Bearer ' + t } : {}
}

export async function api<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    ...opts,
    headers: { 'Content-Type': 'application/json', ...authHeaders(), ...(opts?.headers || {}) },
  })
  if (res.status === 401 && path !== '/login') {
    localStorage.clear()
    location.reload()
    throw new Error('Sessiya bitdi — yenidən daxil olun')
  }
  if (!res.ok) {
    const msg = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(msg.error || 'Server xətası')
  }
  return res.json() as Promise<T>
}

// şəkil yüklə → URL qaytarır
export async function uploadFile(file: File): Promise<string> {
  const fd = new FormData()
  fd.append('file', file)
  const res = await fetch(BASE + '/upload', { method: 'POST', headers: authHeaders(), body: fd })
  if (!res.ok) throw new Error('şəkil yüklənmədi')
  const d = await res.json()
  return d.url as string
}

// fayl endir (Excel) — token ilə fetch, blob
export async function downloadFile(path: string, filename: string) {
  const res = await fetch(BASE + path, { headers: authHeaders() })
  if (!res.ok) throw new Error('endirilmədi')
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

export const imgURL = (p: string) => (p ? (p.startsWith('http') ? p : HOST + p) : '')

// ---- tiplər ----
export interface Branch { id: number; name: string; is_main: boolean; address: string }
export interface AttributeOption { id: number; attribute_id: number; value: string }
export interface Attribute { id: number; name: string; multiselect?: boolean; show_on_site?: boolean; options: AttributeOption[] }
export interface Category { id: number; name: string; attributes: Attribute[] }
export interface ItemValue { id: number; attribute_id: number; attribute: Attribute; value: string }
export interface ItemTranslation { lang: string; name: string }
export interface Item {
  id: number; name: string; serial: string
  category_id: number; category: Category
  branch_id: number; branch: Branch
  cost: number; price: number; wholesale_price: number; discount: number; quantity: number; status: string; show_on_site: boolean
  card_image: string; gallery: string
  values: ItemValue[]; translations?: ItemTranslation[]; created_at: string; sold_at?: string
}
export interface Customer { id: number; name: string; phone: string }
export interface Sale {
  id: number; item: Item; sale_price: number; quantity: number; profit: number
  channel: string; customer?: Customer; warranty_months: number; sold_at: string
}
export interface CreditPayment { id: number; credit_plan_id: number; amount: number; created_at: string }
export interface Credit {
  id: number; customer: Customer; item_name: string
  total: number; paid: number; next_due: string; status: string
  payments?: CreditPayment[]
  closed: boolean; sale_id?: number
}
export interface Consignment {
  id: number; store_name: string; item_id?: number; item_name: string; serial: string
  given_at: string; cost: number; given_price: number; status: string; debt: number
}
export interface Expense { id: number; date: string; category: string; note: string; amount: number }
export interface Dashboard {
  stock_count: number; stock_value: number; reserved_count: number
  open_credit: number; overdue_count: number; sales_count: number
  turnover: number; gross_profit: number; expenses: number; net_profit: number
  by_category: { name: string; turnover: number }[]
}
export interface OnlineOrder {
  id: number; customer_name: string; phone: string; item: Item
  status: string; created_at: string; expires_at: string
}
export interface DailyClose { id: number; date: string; cash: number; card: number; installment: number; credit: number; total: number }
export interface KassaData {
  today: { cash: number; card: number; installment: number; credit: number; total: number }
  closes: DailyClose[]
}
export interface SupplyBatch { id: number; date: string; source: string; supplier: string; item_count: number; total_cost: number; status: string }
export interface Reports {
  top_models: { name: string; count: number }[]
  branch_perf: { name: string; profit: number }[]
  dead_stock: { name: string; cost: number; age_days: number }[]
}
export interface AuditLog { id: number; username: string; role: string; action: string; path: string; detail: string; created_at: string }
export interface EmailTemplate { id: number; name: string; emails: string }
// çoxdilli
export interface Language { id: number; code: string; name: string; icon: string; enabled: boolean; is_default: boolean; sort: number }
export interface UiStringRow { key: string; section: string; values: Record<string, string> }
export interface UiStringsData { languages: Language[]; strings: UiStringRow[] }
export interface TermRow { id: number; kind: string; az: string; values: Record<string, string> }
export interface TermsData { languages: Language[]; categories: TermRow[]; attributes: TermRow[]; options: TermRow[] }
// tərəfdaşlıq
export interface PartnerApplication { id: number; name: string; store_name: string; phone: string; status: string; created_at: string }
export interface AppUser { id: number; username: string; role: string; name: string }

export const STATUS_AZ: Record<string, string> = {
  in_stock: 'Stokda', reserved: 'Rezerv', sold: 'Satıldı', returned: 'Qaytarıldı', consignment: 'Realizasiyada',
}
export const STATUS_TAG: Record<string, string> = {
  in_stock: 'good', reserved: 'warn', sold: 'neut', returned: 'bad', consignment: 'warn',
}
export const CHANNEL_AZ: Record<string, string> = {
  cash: 'Nağd', card: 'Kart', installment: 'Taksit', credit: 'Kredit',
}
// telefonu müqayisə üçün normallaşdır (yalnız rəqəmlər)
export const normPhone = (p?: string) => (p || '').replace(/\D/g, '')

// verilmiş nömrəli müştəri (varsa) — unikallıq yoxlaması üçün
export const phoneOwner = (customers: Customer[], phone: string, excludeId?: number): Customer | undefined => {
  const np = normPhone(phone)
  if (!np) return undefined
  return customers.find((c) => c.id !== excludeId && normPhone(c.phone) === np)
}

// cari istifadəçi (localStorage-dən)
export const currentRole = () => localStorage.getItem('role') || 'user'
export const currentName = () => localStorage.getItem('name') || 'İstifadəçi'
export const currentBranchId = () => Number(localStorage.getItem('branch_id') || 0)
export const currentBranchName = () => localStorage.getItem('branch_name') || ''
export const isAdmin = () => currentRole() === 'admin'
