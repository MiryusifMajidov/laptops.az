import { useMemo, useState } from 'react'
import { api, phoneOwner, type Customer } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import FormModal from '../components/FormModal'

export default function Customers() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: custs } = useFetch<Customer[]>('/customers', [key])
  const [q, setQ] = useState('')
  const [form, setForm] = useState(false)
  const [edit, setEdit] = useState<Customer | null>(null)
  const [nName, setNName] = useState('')
  const [nPhone, setNPhone] = useState('')
  const [nBusy, setNBusy] = useState(false)

  const list = custs ?? []
  const dupe = phoneOwner(list, nPhone)

  function openForm() { setNName(''); setNPhone(''); setForm(true) }
  async function createCustomer() {
    if (!nName.trim() || dupe) return
    setNBusy(true)
    try {
      await api('/customers', { method: 'POST', body: JSON.stringify({ name: nName.trim(), phone: nPhone.trim() }) })
      bump(); toast('Müştəri əlavə edildi'); setForm(false)
    } catch (e) { toast((e as Error).message) } finally { setNBusy(false) }
  }
  const filtered = useMemo(() => list.filter((c) =>
    c.name.toLowerCase().includes(q.toLowerCase()) || (c.phone ?? '').includes(q)
  ), [list, q])

  return (
    <div className="content">
      <div className="toolbar">
        <div className="h" style={{ marginRight: 'auto' }}>Müştəri bazası · {list.length}</div>
        <div className="miniSearch"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#B0AEA8" strokeWidth={2.2} strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4-4" /></svg><input placeholder="Ad və ya telefon…" value={q} onChange={(e) => setQ(e.target.value)} /></div>
        <button className="btn primary sm" onClick={openForm}><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth={2.6} strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>Yeni müştəri</button>
      </div>

      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><circle cx="9" cy="8" r="3" /><path d="M3 19c0-3 2.5-5 6-5" /></svg><div>Excel-də müştərilər dağınıq idi. Burada hər müştərinin profili var; baza satışlar əlavə olundukca böyüyür.</div></div>

      <div className="card" style={{ overflow: 'hidden' }}>
        <table>
          <thead><tr><th>Müştəri</th><th>Telefon</th><th className="tright">Profil</th></tr></thead>
          <tbody>
            {filtered.map((c) => (
              <tr key={c.id}><td className="prod">{c.name}</td><td className="tiny">{c.phone || '—'}</td><td className="tright"><button className="btn ghost sm" onClick={() => setEdit(c)}>Redaktə</button></td></tr>
            ))}
            {filtered.length === 0 && <tr><td colSpan={3} className="center-msg">Müştəri yoxdur</td></tr>}
          </tbody>
        </table>
      </div>

      {form && (
        <>
          <div className="scrim" onClick={() => setForm(false)} />
          <div className="modal" style={{ width: 460 }}>
            <header>
              <div><div className="h">Yeni müştəri</div><div className="tiny">nömrə unikaldır — təkrar olsa altda göstərilir</div></div>
              <button className="x" onClick={() => setForm(false)}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
            </header>
            <div className="body">
              <div className="field full"><label>Ad, soyad *</label><input placeholder="məs. Rauf İsmayılov" value={nName} onChange={(e) => setNName(e.target.value)} autoFocus /></div>
              <div className="field full"><label>Telefon</label>
                <input placeholder="050 xxx xx xx" value={nPhone} onChange={(e) => setNPhone(e.target.value)} style={dupe ? { borderColor: 'var(--bad)' } : undefined} />
                {dupe && <div className="tiny" style={{ color: 'var(--bad)', fontWeight: 700, marginTop: 6 }}>Bu nömrə artıq var: {dupe.name}</div>}
              </div>
            </div>
            <footer>
              <button className="btn ghost" style={{ flex: 1 }} onClick={() => setForm(false)}>İmtina</button>
              <button className="btn primary" style={{ flex: 1.6 }} onClick={createCustomer} disabled={nBusy || !!dupe || !nName.trim()}>{nBusy ? 'Saxlanır…' : 'Əlavə et'}</button>
            </footer>
          </div>
        </>
      )}

      {edit && (
        <FormModal
          title="Müştərini redaktə et" submitLabel="Yadda saxla" onClose={() => setEdit(null)}
          initial={{ name: edit.name, phone: edit.phone }}
          fields={[
            { name: 'name', label: 'Ad, soyad', required: true, full: true },
            { name: 'phone', label: 'Telefon', full: true },
          ]}
          onSubmit={async (v) => { await api(`/customers/${edit.id}`, { method: 'PUT', body: JSON.stringify(v) }); bump(); toast('Müştəri yeniləndi'); setEdit(null) }}
        />
      )}
    </div>
  )
}
