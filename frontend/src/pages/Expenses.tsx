import { useMemo, useState } from 'react'
import { api, type Dashboard, type Expense } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money, dateFullAz } from '../lib/format'
import FormModal from '../components/FormModal'

const CATS = ['Kirayə', 'Maaş', 'Kommunal', 'Nəqliyyat', 'Reklam', 'Digər']

export default function Expenses() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: exp } = useFetch<Expense[]>('/expenses', [key])
  const { data: d } = useFetch<Dashboard>('/dashboard', [key])
  const [form, setForm] = useState(false)
  const [edit, setEdit] = useState<Expense | null>(null)
  const list = exp ?? []

  // günlərə görə qruplaşdır (backend date desc qaytarır)
  const groups = useMemo(() => {
    const m = new Map<string, Expense[]>()
    for (const e of list) {
      const day = (e.date || '').slice(0, 10)
      if (!m.has(day)) m.set(day, [])
      m.get(day)!.push(e)
    }
    return [...m.entries()]
  }, [list])

  async function remove(e: Expense, ev: React.MouseEvent) {
    ev.stopPropagation()
    if (!window.confirm(`«${e.category}${e.note ? ' · ' + e.note : ''}» (${money(e.amount)}) silinsin?`)) return
    try { await api(`/expenses/${e.id}`, { method: 'DELETE' }); bump(); toast('Xərc silindi') }
    catch (err) { toast((err as Error).message) }
  }

  const fields = [
    { name: 'category', label: 'Kateqoriya', type: 'select' as const, required: true, options: CATS.map((c) => ({ value: c, label: c })) },
    { name: 'amount', label: 'Məbləğ (₼)', type: 'number' as const, required: true },
    { name: 'date', label: 'Tarix', type: 'date' as const },
    { name: 'note', label: 'Qeyd', placeholder: 'məs. Mağaza icarəsi', full: true },
  ]

  return (
    <div className="content">
      <div className="row">
        <div className="card kpi" style={{ flex: 1 }}><div className="eyebrow">Brüt mənfəət</div><div className="v" style={{ color: 'var(--good-ink)' }}>{money(d?.gross_profit ?? 0)}</div></div>
        <div className="card kpi" style={{ flex: 1 }}><div className="eyebrow">Xərclər</div><div className="v" style={{ color: 'var(--bad)' }}>− {money(d?.expenses ?? 0)}</div></div>
        <div className="card kpi" style={{ flex: 1, background: 'var(--ink)', borderColor: 'var(--ink)' }}><div className="eyebrow" style={{ color: 'rgba(255,255,255,.6)' }}>Xalis mənfəət</div><div className="v" style={{ color: '#fff' }}>{money(d?.net_profit ?? 0)}</div></div>
      </div>

      <div className="banner" style={{ justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
          <div><b>Günlərə bölünmüş.</b> Hər gün ayrıca cədvəldir. Sətrə klik → redaktə et; sağdakı ilə sil.</div>
        </div>
        <button className="btn ghost sm" style={{ flex: 'none' }} onClick={() => setForm(true)}><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.6} strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>Xərc əlavə et</button>
      </div>

      {groups.length === 0 && <div className="card"><div className="center-msg">Xərc yoxdur</div></div>}

      {groups.map(([day, rows]) => {
        const dayTotal = rows.reduce((a, e) => a + e.amount, 0)
        return (
          <div className="card" key={day} style={{ overflow: 'hidden' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 18px', background: 'var(--surface)', borderBottom: '1px solid var(--line)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#5A6474" strokeWidth={2} strokeLinecap="round"><rect x="3" y="4.5" width="18" height="17" rx="2" /><path d="M3 9h18M8 2.5v4M16 2.5v4" /></svg>
                <span style={{ fontWeight: 800, fontSize: 14 }}>{dateFullAz(day)}</span>
                <span className="pill neut">{rows.length} xərc</span>
              </div>
              <span className="tiny">Cəmi: <b className="data" style={{ color: 'var(--bad)' }}>− {money(dayTotal)}</b></span>
            </div>
            <table>
              <thead><tr><th style={{ width: 44, textAlign: 'center' }}>№</th><th>Kateqoriya</th><th>Qeyd</th><th className="tright">Məbləğ</th><th style={{ width: 48 }}></th></tr></thead>
              <tbody>
                {rows.map((e, idx) => (
                  <tr key={e.id} style={{ cursor: 'pointer' }} onClick={() => setEdit(e)}>
                    <td className="data" style={{ textAlign: 'center', color: 'var(--muted)', fontWeight: 700 }}>{idx + 1}</td>
                    <td><span className="pill neut">{e.category}</span></td>
                    <td>{e.note || '—'}</td>
                    <td className="tright data" style={{ color: 'var(--bad)', fontWeight: 600 }}>− {money(e.amount)}</td>
                    <td style={{ textAlign: 'center' }}>
                      <button onClick={(ev) => remove(e, ev)} title="Sil" style={{ color: 'var(--bad)', display: 'inline-flex', padding: 4 }}>
                        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14" /></svg>
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      })}

      {form && (
        <FormModal
          title="Yeni xərc" submitLabel="Əlavə et" onClose={() => setForm(false)}
          fields={fields}
          onSubmit={async (v) => { await api('/expenses', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Xərc əlavə edildi'); setForm(false) }}
        />
      )}

      {edit && (
        <FormModal
          title="Xərci redaktə et" submitLabel="Yadda saxla" onClose={() => setEdit(null)}
          fields={fields}
          initial={{ category: edit.category, amount: edit.amount, date: (edit.date ?? '').slice(0, 10), note: edit.note }}
          onSubmit={async (v) => { await api(`/expenses/${edit.id}`, { method: 'PUT', body: JSON.stringify(v) }); bump(); toast('Xərc yeniləndi'); setEdit(null) }}
        />
      )}
    </div>
  )
}
