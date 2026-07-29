import { useState } from 'react'
import CustomerPicker from './CustomerPicker'

export type Field = {
  name: string
  label: string
  type?: 'text' | 'number' | 'date' | 'select' | 'searchselect' | 'customer' | 'multiselect'
  options?: { value: string | number; label: string }[]
  placeholder?: string
  required?: boolean
  full?: boolean
}

// axtarışlı seçim (çoxlu məhsul arasından yazaraq tapmaq üçün)
function SearchSelect({ options, value, onChange, placeholder }: {
  options: { value: string | number; label: string }[]
  value: unknown
  onChange: (v: string | number) => void
  placeholder?: string
}) {
  const [open, setOpen] = useState(false)
  const [q, setQ] = useState('')
  const selected = options.find((o) => String(o.value) === String(value))
  const query = q.trim().toLowerCase()
  const filtered = query ? options.filter((o) => o.label.toLowerCase().includes(query)) : options
  return (
    <div className="ss">
      <input
        placeholder={placeholder ?? 'yazaraq axtar…'}
        value={open ? q : (selected?.label ?? '')}
        onFocus={() => { setOpen(true); setQ('') }}
        onChange={(e) => { setQ(e.target.value); setOpen(true) }}
        onBlur={() => setTimeout(() => setOpen(false), 150)}
      />
      {open && (
        <div className="ss-list">
          {filtered.slice(0, 60).map((o) => (
            <div key={o.value} className="ss-opt" onMouseDown={() => { onChange(o.value); setOpen(false); setQ('') }}>{o.label}</div>
          ))}
          {filtered.length === 0 && <div className="ss-opt muted">tapılmadı</div>}
          {filtered.length > 60 && <div className="ss-opt muted">…daha dəqiq yaz ({filtered.length} nəticə)</div>}
        </div>
      )}
    </div>
  )
}

export default function FormModal({ title, subtitle, fields, submitLabel, onClose, onSubmit, initial }: {
  title: string
  subtitle?: string
  fields: Field[]
  submitLabel: string
  onClose: () => void
  onSubmit: (v: Record<string, unknown>) => Promise<void> | void
  initial?: Record<string, unknown>
}) {
  const [v, setV] = useState<Record<string, unknown>>(initial ?? {})
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const set = (n: string, val: unknown) => setV((s) => ({ ...s, [n]: val }))

  async function submit() {
    for (const f of fields) {
      if (f.required && !v[f.name]) { setErr(`«${f.label}» vacibdir`); return }
    }
    setErr('')
    const out: Record<string, unknown> = {}
    for (const f of fields) {
      let val = v[f.name]
      if (f.type === 'number') val = val === '' || val == null ? 0 : Number(val)
      if ((f.type === 'select' || f.type === 'searchselect') && typeof val === 'string' && val !== '' && !isNaN(Number(val))) val = Number(val)
      out[f.name] = val ?? (f.type === 'multiselect' ? [] : '')
    }
    setBusy(true)
    try { await onSubmit(out) } catch (e) { setErr((e as Error).message) } finally { setBusy(false) }
  }

  return (
    <>
      <div className="scrim" onClick={onClose} />
      <div className="modal" style={{ width: 520 }}>
        <header>
          <div><div className="h">{title}</div>{subtitle && <div className="tiny">{subtitle}</div>}</div>
          <button className="x" onClick={onClose}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
        </header>
        <div className="body">
          {fields.map((f) => (
            <div className={'field' + (f.full || f.type === 'multiselect' ? ' full' : '')} key={f.name}>
              <label>{f.label}{f.required && ' *'}</label>
              {f.type === 'select' ? (
                <select value={(v[f.name] as string) ?? ''} onChange={(e) => set(f.name, e.target.value)}>
                  <option value="">— seç —</option>
                  {f.options?.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
                </select>
              ) : f.type === 'searchselect' ? (
                <SearchSelect options={f.options ?? []} value={v[f.name]} onChange={(val) => set(f.name, val)} placeholder={f.placeholder} />
              ) : f.type === 'customer' ? (
                <CustomerPicker value={v[f.name] as number} onChange={(id) => set(f.name, id)} placeholder={f.placeholder} />
              ) : f.type === 'multiselect' ? (
                <div className="optchips">
                  {f.options?.map((o) => {
                    const arr = (v[f.name] as (string | number)[]) ?? []
                    const on = arr.includes(o.value)
                    return (
                      <span key={o.value} className="optchip" style={on ? { background: 'var(--ink)', color: '#fff', borderStyle: 'solid', cursor: 'pointer' } : { cursor: 'pointer' }}
                        onClick={() => set(f.name, on ? arr.filter((x) => x !== o.value) : [...arr, o.value])}>{o.label}</span>
                    )
                  })}
                </div>
              ) : (
                <input type={f.type === 'number' ? 'number' : f.type === 'date' ? 'date' : 'text'} placeholder={f.placeholder}
                  value={(v[f.name] as string) ?? ''} onChange={(e) => set(f.name, e.target.value)} />
              )}
            </div>
          ))}
          {err && <div className="full login-err">{err}</div>}
        </div>
        <footer>
          <button className="btn ghost" style={{ flex: 1 }} onClick={onClose}>İmtina</button>
          <button className="btn primary" style={{ flex: 1.6 }} onClick={submit} disabled={busy}>{busy ? 'Saxlanır…' : submitLabel}</button>
        </footer>
      </div>
    </>
  )
}
