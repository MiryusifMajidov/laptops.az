import { useState } from 'react'
import { api, type Consignment, type Item } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money, dateAz } from '../lib/format'
import FormModal from '../components/FormModal'

const STAT: Record<string, [string, string]> = {
  out: ['Bizdə deyil', 'bad'],
  sold_unpaid: ['Satılıb · borc', 'warn'],
  sold_paid: ['Satılıb · ödənilib', 'good'],
  returned: ['Qaytarıldı', 'neut'],
}

export default function Realizasiya() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: cons } = useFetch<Consignment[]>('/consignments', [key])
  const { data: stockItems } = useFetch<Item[]>('/items?status=in_stock', [key])
  const [form, setForm] = useState(false)
  const [statusFor, setStatusFor] = useState<Consignment | null>(null)
  const list = cons ?? []
  const outValue = list.filter((c) => c.status === 'out').reduce((a, c) => a + c.cost, 0)
  const debt = list.reduce((a, c) => a + c.debt, 0)
  const sold = list.filter((c) => c.status.startsWith('sold')).length
  const stores = new Set(list.map((c) => c.store_name)).size

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M2.5 6h11v9H2.5Z" /><path d="M13.5 9h4l3 3v3h-7Z" /><circle cx="7" cy="18" r="1.6" /><circle cx="17" cy="18" r="1.6" /></svg><div><b>Başqa mağazalara ofdan (topdan) verilən cihazlar.</b> Digər mağazanın müştəri sifarişi olanda, özlərində olmayan cihazı bizdən götürürlər. Satılanda bizə ödəyirlər.</div></div>

      <div className="grid4">
        <div className="card kpi"><div className="eyebrow">Çöldə olan mal (dəyər)</div><div className="v">{money(outValue)}</div></div>
        <div className="card kpi"><div className="eyebrow" style={{ color: 'var(--bad)' }}>Ödənilməmiş qalıq</div><div className="v" style={{ color: 'var(--bad)' }}>{money(debt)}</div></div>
        <div className="card kpi"><div className="eyebrow">Satılan</div><div className="v">{sold}</div></div>
        <div className="card kpi"><div className="eyebrow">Aktiv mağaza</div><div className="v">{stores}</div></div>
      </div>

      <div className="toolbar"><div className="h" style={{ marginRight: 'auto' }}>Verilən cihazlar</div><button className="btn primary sm" onClick={() => setForm(true)}>+ Mal ver</button></div>

      <div className="card" style={{ overflow: 'hidden' }}>
        <table>
          <thead><tr><th>Mağaza</th><th>Məhsul</th><th>Seriya</th><th>Veriliş</th><th className="tright">Alış</th><th className="tright">Verilmə qiyməti</th><th>Status</th><th className="tright">Borc</th><th></th></tr></thead>
          <tbody>
            {list.map((c) => {
              const st = STAT[c.status] ?? [c.status, 'neut']
              return (
                <tr key={c.id}>
                  <td className="prod">{c.store_name}</td>
                  <td>{c.item_name}</td>
                  <td className="ser">{c.serial || '—'}</td>
                  <td className="tiny">{dateAz(c.given_at)}</td>
                  <td className="tright data cost">{money(c.cost)}</td>
                  <td className="tright data" style={{ fontWeight: 600 }}>{money(c.given_price)}</td>
                  <td><span className={'pill ' + st[1]}>{st[0]}</span></td>
                  <td className="tright data" style={{ color: c.debt ? 'var(--bad)' : 'var(--muted)', fontWeight: c.debt ? 700 : 400 }}>{c.debt ? money(c.debt) : '—'}</td>
                  <td className="tright"><button className="btn ghost sm" onClick={() => setStatusFor(c)}>Status</button></td>
                </tr>
              )
            })}
            {list.length === 0 && <tr><td colSpan={9} className="center-msg">Çöldə mal yoxdur</td></tr>}
          </tbody>
        </table>
      </div>

      {form && (
        <FormModal
          title="Başqa mağazaya mal ver" subtitle="cihaz stokdan çıxıb realizasiyaya keçir" submitLabel="Ver" onClose={() => setForm(false)}
          fields={[
            { name: 'store_name', label: 'Mağaza', required: true, full: true, placeholder: 'məs. Salman (mağaza)' },
            { name: 'item_id', label: 'Stokdakı cihaz', type: 'searchselect', required: true, full: true, placeholder: 'ad, seriya və ya marka yaz…', options: (stockItems ?? []).map((i) => ({ value: i.id, label: `${i.name}${i.serial ? ' · ' + i.serial : ''}` })) },
            { name: 'given_price', label: 'Verilmə qiyməti (₼)', type: 'number', full: true },
          ]}
          onSubmit={async (v) => { await api('/consignments', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Mal realizasiyaya verildi — stokdan çıxdı'); setForm(false) }}
        />
      )}

      {statusFor && (
        <FormModal
          title="Status dəyiş" subtitle={statusFor.item_name} submitLabel="Yadda saxla" onClose={() => setStatusFor(null)}
          initial={{ status: statusFor.status }}
          fields={[{ name: 'status', label: 'Status', type: 'select', required: true, full: true, options: [
            { value: 'out', label: 'Bizdə deyil (çöldə)' },
            { value: 'sold_unpaid', label: 'Satılıb · borc qalıb' },
            { value: 'sold_paid', label: 'Satılıb · ödənilib' },
            { value: 'returned', label: 'Qaytarıldı' },
          ] }]}
          onSubmit={async (v) => { await api(`/consignments/${statusFor.id}`, { method: 'PUT', body: JSON.stringify(v) }); bump(); toast('Status yeniləndi'); setStatusFor(null) }}
        />
      )}
    </div>
  )
}
