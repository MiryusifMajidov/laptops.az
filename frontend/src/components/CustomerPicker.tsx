import { useEffect, useState } from 'react'
import { api, phoneOwner, type Customer } from '../api'

// Axtarışlı müştəri seçimi + yerindəcə "yeni müştəri yarat"
// (arxaya qayıtmadan). FormModal-da type: 'customer' kimi işlənir.
export default function CustomerPicker({ value, onChange, placeholder }: {
  value?: number | string
  onChange: (id: number) => void
  placeholder?: string
}) {
  const [customers, setCustomers] = useState<Customer[]>([])
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const [creating, setCreating] = useState(false)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [busy, setBusy] = useState(false)

  const load = () => api<Customer[]>('/customers').then(setCustomers).catch(() => {})
  useEffect(() => { load() }, [])

  const selected = customers.find((c) => String(c.id) === String(value))
  const query = q.trim().toLowerCase()
  const filtered = query
    ? customers.filter((c) => `${c.name} ${c.phone}`.toLowerCase().includes(query))
    : customers

  const dupe = phoneOwner(customers, phone)

  async function create() {
    if (!name.trim() || dupe) return
    setBusy(true)
    try {
      const c = await api<Customer>('/customers', { method: 'POST', body: JSON.stringify({ name: name.trim(), phone: phone.trim() }) })
      await load()
      onChange(c.id)
      setCreating(false); setName(''); setPhone(''); setQ(''); setOpen(false)
    } catch { /* toast yoxdur burda — sükutla ötür */ } finally { setBusy(false) }
  }

  if (creating) {
    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, background: 'var(--surface)', border: '1px solid var(--line)', borderRadius: 10, padding: 10 }}>
        <div className="tiny" style={{ fontWeight: 700 }}>Yeni müştəri</div>
        <input placeholder="Ad, soyad" value={name} onChange={(e) => setName(e.target.value)} autoFocus />
        <input placeholder="Telefon" value={phone} onChange={(e) => setPhone(e.target.value)}
          style={dupe ? { borderColor: 'var(--bad)' } : undefined} />
        {dupe && <div className="tiny" style={{ color: 'var(--bad)', fontWeight: 700 }}>Bu nömrə artıq var: {dupe.name}</div>}
        <div style={{ display: 'flex', gap: 8 }}>
          <button type="button" className="btn ghost sm" style={{ flex: 1 }} onClick={() => setCreating(false)}>Ləğv</button>
          <button type="button" className="btn primary sm" style={{ flex: 1.4 }} onClick={create} disabled={busy || !!dupe}>{busy ? '…' : 'Yarat və seç'}</button>
        </div>
      </div>
    )
  }

  return (
    <div className="ss">
      <input
        placeholder={placeholder ?? 'müştəri axtar…'}
        value={open ? q : (selected ? `${selected.name} — ${selected.phone}` : '')}
        onFocus={() => { setOpen(true); setQ('') }}
        onChange={(e) => { setQ(e.target.value); setOpen(true) }}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
      />
      {open && (
        <div className="ss-list">
          <div className="ss-opt" style={{ fontWeight: 700, color: 'var(--good-ink)' }}
            onMouseDown={() => { setCreating(true); setName(q) }}>
            + Yeni müştəri yarat{q ? ` («${q}»)` : ''}
          </div>
          {filtered.slice(0, 60).map((c) => (
            <div key={c.id} className="ss-opt" onMouseDown={() => { onChange(c.id); setOpen(false); setQ('') }}>{c.name} — {c.phone}</div>
          ))}
          {filtered.length === 0 && <div className="ss-opt muted">müştəri tapılmadı — yuxarıdan yeni yarat</div>}
        </div>
      )}
    </div>
  )
}
