import { useMemo, useState } from 'react'
import { api, type Item } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money } from '../lib/format'
import FormModal from './FormModal'

const CHANNELS: [string, string][] = [
  ['cash', 'Nağd'], ['card', 'Kart'], ['installment', 'Taksit'], ['credit', 'Kredit'],
]

export default function NewSaleDrawer({ onClose, prefill, orderId, orderInfo }: {
  onClose: () => void
  prefill?: Item | null
  orderId?: number
  orderInfo?: { name: string; phone: string }
}) {
  const toast = useToast()
  const { bump } = useRefresh()
  const { data: items } = useFetch<Item[]>('/items?status=in_stock', [])
  const [q, setQ] = useState('')
  const [picked, setPicked] = useState<Item | null>(prefill ?? null)
  const [price, setPrice] = useState(prefill?.price ? String(prefill.price) : '')
  const [qty, setQty] = useState('1')
  const [channel, setChannel] = useState('cash')
  const [warranty, setWarranty] = useState('')
  const [creditOpen, setCreditOpen] = useState(false)
  const [quickOpen, setQuickOpen] = useState(false) // stokda yoxdursa — tez məhsul yarat
  const [saving, setSaving] = useState(false)

  const results = useMemo(() => {
    const list = items ?? []
    const term = q.toLowerCase()
    return list.filter((i) => i.name.toLowerCase().includes(term) || i.serial.toLowerCase().includes(term)).slice(0, 6)
  }, [items, q])

  const stock = picked ? Math.max(1, picked.quantity || 1) : 1 // stokda mövcud ədəd (köhnə data → 1)
  const n = Math.max(1, Number(qty) || 1)
  const total = picked && price ? Number(price) * n : 0 // satışın cəmi
  const profit = picked && price ? (Number(price) - picked.cost) * n : 0
  const margin = picked && Number(price) > 0 ? Math.round(((Number(price) - picked.cost) / Number(price)) * 100) : null

  async function doSale(extra: Record<string, unknown>) {
    setSaving(true)
    try {
      await api('/sales', { method: 'POST', body: JSON.stringify({ item_id: picked!.id, sale_price: Number(price), quantity: n, channel, warranty_months: Number(warranty) || 0, order_id: orderId, ...extra }) })
      bump()
      toast(channel === 'credit' ? 'Kreditli satış qeydə alındı · borc yaradıldı' : `Satış qeydə alındı · cihaz «Satıldı» oldu (${CHANNELS.find((c) => c[0] === channel)?.[1]})`)
      setCreditOpen(false); onClose()
    } catch (e) {
      toast((e as Error).message)
    } finally {
      setSaving(false)
    }
  }

  function confirm() {
    if (!picked || !price) { toast('Cihaz və satış qiyməti seç'); return }
    if (n > stock) { toast(`Stokda yalnız ${stock} ədəd var`); return }
    if (channel === 'credit') { setCreditOpen(true); return } // əlavə pəncərə: kredit məlumatları
    doSale({})
  }

  return (
    <>
      <div className="scrim" onClick={onClose} />
      <aside className="drawer">
        <header>
          <div><div className="h">Yeni satış</div><div className="tiny">{orderInfo ? `Onlayn sifariş · ${orderInfo.name}${orderInfo.phone ? ' · ' + orderInfo.phone : ''}` : 'stokdan cihaz seç → qiymət → kanal'}</div></div>
          <button className="x" onClick={onClose}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
        </header>

        <div className="body">
          {!picked && (
            <div className="field">
              <label>1 · Cihazı tap (ad, seriya və ya barkod)</label>
              <div className="miniSearch" style={{ width: '100%' }}>
                <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#B0AEA8" strokeWidth={2.2} strokeLinecap="round"><rect x="3" y="5" width="18" height="14" rx="2" /><path d="M7 9v6M10 9v6M13 9v6M16 9v6" /></svg>
                <input autoFocus placeholder="Skan et və ya yaz…" value={q} onChange={(e) => setQ(e.target.value)} />
              </div>
              <div className="res" style={{ marginTop: 8 }}>
                {results.length === 0 && <div className="it" style={{ cursor: 'default', color: 'var(--muted)' }}>Stokda tapılmadı</div>}
                {results.map((i) => (
                  <div key={i.id} className="it" onClick={() => { setPicked(i); setQty('1'); if (i.price) setPrice(String(i.price)) }}>
                    <div><div style={{ fontWeight: 700, fontSize: 13 }}>{i.name}</div><div className="ser">{i.serial} · {i.branch?.name}</div></div>
                    <span className="data cost">{money(i.cost)}</span>
                  </div>
                ))}
              </div>
              <button className="btn ghost sm" style={{ marginTop: 10, width: '100%' }} onClick={() => setQuickOpen(true)}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.4} strokeLinecap="round" style={{ marginRight: 5 }}><path d="M12 5v14M5 12h14" /></svg>
                Stokda yoxdur → yeni məhsul yarat
              </button>
            </div>
          )}

          {picked && (
            <>
              <div className="picked">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                  <div><div className="p">{picked.name}</div><div className="ser" style={{ color: 'rgba(255,255,255,.6)' }}>{picked.serial}</div></div>
                  <button className="btn sm" style={{ background: 'rgba(255,255,255,.15)', color: '#fff' }} onClick={() => { setPicked(null); setPrice('') }}>Dəyiş</button>
                </div>
                <div style={{ marginTop: 10, fontSize: 12, color: 'rgba(255,255,255,.7)' }}>Alış qiyməti (avtomatik): <b className="data" style={{ color: '#fff' }}>{money(picked.cost)}</b></div>
              </div>

              <div className="field">
                <label>2 · {stock > 1 ? 'Bir ədədin qiyməti (₼)' : 'Satış qiyməti (₼)'}</label>
                <input className="data" type="number" placeholder="0" value={price} onChange={(e) => setPrice(e.target.value)} autoFocus />
              </div>

              {stock > 1 && (
                <div className="field">
                  <label>Say (ədəd) <span className="tiny" style={{ fontWeight: 600 }}>stokda {stock} ədəd</span></label>
                  <input className="data" type="number" min={1} max={stock} placeholder="1" value={qty} onChange={(e) => setQty(e.target.value)} />
                  {n > 1 && Number(price) > 0 && <div className="tiny" style={{ marginTop: 4 }}>Cəmi: <b className="data">{money(total)}</b> ({n} × {money(Number(price))})</div>}
                </div>
              )}

              <div className="bigprofit">
                <div><div className="eyebrow" style={{ color: 'var(--good-ink)' }}>Mənfəət{n > 1 ? ` (${n} ədəd)` : ''}</div><div className="tiny" style={{ color: '#5FA588' }}>{margin === null ? 'marja —' : `marja ${margin}%`}</div></div>
                <div className="n">{profit >= 0 ? '+' : '−'}{money(Math.abs(profit))}</div>
              </div>

              <div className="field">
                <label>3 · Ödəniş kanalı</label>
                <div className="chanrow">
                  {CHANNELS.map(([k, l]) => (
                    <div key={k} className={'chan' + (channel === k ? ' on' : '')} onClick={() => setChannel(k)}>{l}</div>
                  ))}
                </div>
                {channel === 'credit' && <div className="tiny" style={{ marginTop: 8, color: 'var(--warn)' }}>Kredit — təsdiqləyəndə müştəri və ilkin ödəniş soruşulacaq, borc avtomatik yaranacaq.</div>}
              </div>

              <div className="field">
                <label>4 · Zəmanət (ay) <span className="tiny" style={{ fontWeight: 600 }}>(istəyə görə)</span></label>
                <input className="data" type="number" placeholder="məs. 12" value={warranty} onChange={(e) => setWarranty(e.target.value)} />
              </div>
            </>
          )}
        </div>

        <footer>
          <button className="btn ghost" style={{ flex: 1 }} onClick={onClose}>İmtina</button>
          <button className="btn primary" style={{ flex: 1.6 }} onClick={confirm} disabled={saving}>{saving ? 'Yadda saxlanır…' : (channel === 'credit' ? 'Davam et →' : 'Satışı təsdiqlə')}</button>
        </footer>
      </aside>

      {creditOpen && picked && (
        <FormModal
          title="Kredit (nisyə) məlumatları" subtitle={`${picked.name} · ${money(Number(price) || 0)}`} submitLabel="Satışı və borcu yarat" onClose={() => setCreditOpen(false)}
          fields={[
            { name: 'customer_id', label: 'Müştəri (borclu)', type: 'customer', required: true, full: true, placeholder: 'müştəri axtar və ya yeni yarat…' },
            { name: 'down_payment', label: 'İlkin ödəniş (₼)', type: 'number' },
            { name: 'next_due', label: 'Növbəti ödəniş tarixi', type: 'date' },
          ]}
          onSubmit={async (v) => { await doSale(v) }}
        />
      )}

      {quickOpen && (
        <FormModal
          title="Yeni məhsul (tez)" subtitle="stokda yoxdursa — yalnız ad, alış və satış" submitLabel="Yarat və seç" onClose={() => setQuickOpen(false)}
          fields={[
            { name: 'name', label: 'Məhsulun adı', required: true, full: true, placeholder: 'məs. Lenovo IdeaPad Slim 3' },
            { name: 'cost', label: 'Alış qiyməti (₼)', type: 'number', required: true, placeholder: '0' },
            { name: 'price', label: 'Satış qiyməti (₼)', type: 'number', required: true, placeholder: '0' },
          ]}
          onSubmit={async (v) => {
            const created = await api<Item>('/items', { method: 'POST', body: JSON.stringify({ name: v.name, cost: Number(v.cost) || 0, price: Number(v.price) || 0, quantity: 1 }) })
            bump()
            setPicked(created)
            setQty('1')
            setPrice(created.price ? String(created.price) : String(v.price || ''))
            setQuickOpen(false)
            toast('Yeni məhsul yaradıldı və seçildi')
          }}
        />
      )}
    </>
  )
}
