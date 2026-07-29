import { useState } from 'react'
import { api, type SupplyBatch } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money, dateAz } from '../lib/format'
import FormModal from '../components/FormModal'

export default function Supplies() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data } = useFetch<SupplyBatch[]>('/supplies', [key])
  const [form, setForm] = useState(false)
  const list = data ?? []

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M12 3l8 4.5v9L12 21l-8-4.5v-9z" /></svg><div>Gələn mal partiyaları (Dubay, İstanbul, yerli). Alış qiymətləri <b>manatla</b>. Partiya təsdiqlənəndə içindəki cihazlar Stok/Anbara düşür.</div></div>

      <div className="toolbar"><div className="h" style={{ marginRight: 'auto' }}>Mal partiyaları</div><button className="btn primary sm" onClick={() => setForm(true)}>+ Yeni partiya</button></div>

      <div className="card" style={{ overflow: 'hidden' }}>
        <table>
          <thead><tr><th>Tarix</th><th>Mənbə</th><th>Təchizatçı</th><th className="tright">Cihaz sayı</th><th className="tright">Ümumi alış</th><th>Status</th><th></th></tr></thead>
          <tbody>
            {list.map((s) => (
              <tr key={s.id}>
                <td className="tiny">{dateAz(s.date)}</td>
                <td><span className="pill neut">{s.source}</span></td>
                <td>{s.supplier}</td>
                <td className="tright data">{s.item_count}</td>
                <td className="tright data cost">{money(s.total_cost)}</td>
                <td><span className={'pill ' + (s.status === 'pending' ? 'warn' : 'good')}>{s.status === 'pending' ? 'Gözləyir' : 'Stoka düşdü'}</span></td>
                <td className="tright">{s.status === 'pending' && <button className="btn ghost sm" onClick={async () => { try { await api(`/supplies/${s.id}`, { method: 'PUT', body: JSON.stringify({ status: 'received' }) }); bump(); toast('Partiya təsdiqləndi') } catch (e) { toast((e as Error).message) } }}>Təsdiqlə</button>}</td>
              </tr>
            ))}
            {list.length === 0 && <tr><td colSpan={7} className="center-msg">Partiya yoxdur</td></tr>}
          </tbody>
        </table>
      </div>

      {form && (
        <FormModal
          title="Yeni partiya" submitLabel="Əlavə et" onClose={() => setForm(false)}
          fields={[
            { name: 'source', label: 'Mənbə', type: 'select', required: true, options: ['Dubay', 'İstanbul', 'Yerli'].map((x) => ({ value: x, label: x })) },
            { name: 'supplier', label: 'Təchizatçı', placeholder: 'məs. NoteTech' },
            { name: 'item_count', label: 'Cihaz sayı', type: 'number' },
            { name: 'total_cost', label: 'Ümumi alış (₼)', type: 'number' },
            { name: 'status', label: 'Status', type: 'select', options: [{ value: 'pending', label: 'Gözləyir' }, { value: 'received', label: 'Stoka düşdü' }] },
          ]}
          onSubmit={async (v) => { await api('/supplies', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Partiya əlavə edildi'); setForm(false) }}
        />
      )}
    </div>
  )
}
