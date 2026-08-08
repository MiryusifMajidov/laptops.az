import { useMemo, useState } from 'react'
import { api, type Category, type Sale, type Item, type Branch, CHANNEL_AZ, isAdmin, currentBranchName } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money, dateFullAz } from '../lib/format'
import FormModal from '../components/FormModal'

const chTag = (ch: string) => ({ installment: 'warn', credit: 'bad' } as Record<string, string>)[ch] || 'neut'
const CH_OPTS = [{ value: 'cash', label: 'Nağd' }, { value: 'card', label: 'Kart' }, { value: 'installment', label: 'Taksit' }, { value: 'credit', label: 'Kredit' }]
const PAGE = 100

export default function Sales() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const [page, setPage] = useState(0)
  const [cat, setCat] = useState('all')
  const [ch, setCh] = useState('all')
  const [q, setQ] = useState('')
  const [edit, setEdit] = useState<Sale | null>(null)
  const [confirming, setConfirming] = useState<Item | null>(null) // təsdiq gözləyən cihaz (status=satıldı)
  const [selDays, setSelDays] = useState<Set<string>>(new Set())
  const admin = isAdmin()
  const [branch, setBranch] = useState('all') // yalnız admin dəyişir; satıcı öz filialına kilidlidir (backend məcbur edir)

  const { data: cats } = useFetch<Category[]>('/categories', [])
  const { data: branches } = useFetch<Branch[]>(admin ? '/branches' : null, [])

  const qs = new URLSearchParams({ limit: String(PAGE), offset: String(page * PAGE) })
  if (cat !== 'all') qs.set('category', cat)
  if (ch !== 'all') qs.set('channel', ch)
  if (q.trim()) qs.set('q', q.trim())
  if (admin && branch !== 'all') qs.set('branch_id', branch)
  const { data: sales } = useFetch<Sale[]>('/sales?' + qs.toString(), [key, page, cat, ch, q, branch])
  const list = sales ?? []
  // təsdiq gözləyənlər — statusu «satıldı», amma hələ satışı olmayan cihazlar (məhsullar cədvəlindən)
  const { data: pendingData } = useFetch<Item[]>('/sales/pending', [key])
  const pending = pendingData ?? []

  // KPI kartları — filtrə uyğun BÜTÜN satışların cəmi (səhifədən asılı deyil)
  const sumQs = new URLSearchParams()
  if (cat !== 'all') sumQs.set('category', cat)
  if (ch !== 'all') sumQs.set('channel', ch)
  if (q.trim()) sumQs.set('q', q.trim())
  if (admin && branch !== 'all') sumQs.set('branch_id', branch)
  const { data: sum } = useFetch<{ count: number; turnover: number; profit: number }>('/sales/summary?' + sumQs.toString(), [key, cat, ch, q, branch])

  // günlərə görə qruplaşdır (backend sold_at desc qaytarır)
  const groups = useMemo(() => {
    const m = new Map<string, Sale[]>()
    for (const s of list) {
      const day = (s.sold_at || '').slice(0, 10)
      if (!m.has(day)) m.set(day, [])
      m.get(day)!.push(s)
    }
    return [...m.entries()]
  }, [list])

  // seçilmiş günlərin cəmi (kənardakı panel üçün)
  const selTotals = useMemo(() => {
    let t = 0, p = 0, n = 0
    for (const [day, rows] of groups) {
      if (selDays.has(day)) {
        t += rows.reduce((a, s) => a + s.sale_price, 0)
        p += rows.reduce((a, s) => a + s.profit, 0)
        n += rows.length
      }
    }
    return { t, p, n, days: selDays.size }
  }, [groups, selDays])
  const toggleDay = (day: string) => setSelDays((s) => { const n = new Set(s); if (n.has(day)) n.delete(day); else n.add(day); return n })

  const margin = sum && sum.turnover > 0 ? ((sum.profit / sum.turnover) * 100).toFixed(1) : '0.0'
  const filtered = cat !== 'all' || ch !== 'all' || q.trim() !== ''
  const reset = (fn: () => void) => { fn(); setPage(0) }

  return (
    <div className="content">
      <div className="grid4">
        <div className="card kpi"><div className="eyebrow">{filtered ? 'Satış (filtr)' : 'Ümumi · satış'}</div><div className="v">{sum?.count ?? 0}</div></div>
        <div className="card kpi"><div className="eyebrow">Dövriyyə</div><div className="v">{money(sum?.turnover ?? 0)}</div></div>
        <div className="card kpi"><div className="eyebrow">Mənfəət</div><div className="v" style={{ color: 'var(--good-ink)' }}>{money(sum?.profit ?? 0)}</div></div>
        <div className="card kpi"><div className="eyebrow">Orta marja</div><div className="v">{margin}%</div></div>
      </div>

      <div className="toolbar">
        <div className="chips">
          <span className={'chip' + (cat === 'all' ? ' on' : '')} onClick={() => reset(() => setCat('all'))}>Hamısı</span>
          {cats?.map((c) => <span key={c.id} className={'chip' + (cat === c.name ? ' on' : '')} onClick={() => reset(() => setCat(c.name))}>{c.name}</span>)}
        </div>
        <select className="select" value={ch} onChange={(e) => reset(() => setCh(e.target.value))} style={{ cursor: 'pointer' }}>
          <option value="all">Kanal: Hamısı</option>
          {CH_OPTS.map((c) => <option key={c.value} value={c.value}>{c.label}</option>)}
        </select>
        {admin ? (
          <select className="select" value={branch} onChange={(e) => reset(() => setBranch(e.target.value))} style={{ cursor: 'pointer' }} title="Filial üzrə bax">
            <option value="all">Filial: Hamısı</option>
            {branches?.map((b) => <option key={b.id} value={String(b.id)}>{b.name}</option>)}
          </select>
        ) : (
          <span className="pill neut" style={{ alignSelf: 'center' }} title="Yalnız öz filialınızın satışları">Filial: {currentBranchName() || '—'}</span>
        )}
        <div className="miniSearch"><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#B0AEA8" strokeWidth={2.2} strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4-4" /></svg><input placeholder="Ad və ya seriya…" value={q} onChange={(e) => reset(() => setQ(e.target.value))} /></div>
      </div>

      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><circle cx="12" cy="12" r="9" /><path d="M12 8h.01M11 12h1v4h1" /></svg><div><b>Excel-dəki «продажа» kimi — günlərə bölünmüş.</b> Hər gün ayrıca cədvəldir. Sətrə klik → redaktə. Filtr: kateqoriya, kanal, ad/seriya axtarışı.</div></div>

      {pending.length > 0 && (
        <div className="card" style={{ overflow: 'hidden', border: '1px solid var(--warn)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '14px 18px', background: 'rgba(194,65,12,.08)', borderBottom: '1px solid var(--line)' }}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--warn)" strokeWidth={2} strokeLinecap="round"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
            <span style={{ fontWeight: 800, fontSize: 14, color: 'var(--warn)' }}>Təsdiq gözləyən</span>
            <span className="pill warn">{pending.length}</span>
            <span className="tiny" style={{ marginLeft: 'auto' }}>Status «satıldı» edilib — pul gəldikdə qiyməti yazıb təsdiqləyin</span>
          </div>
          <table>
            <thead><tr><th>Məhsul</th><th>Seriya</th><th>Filial</th><th className="tright">Alış</th><th className="tright">Əməliyyat</th></tr></thead>
            <tbody>
              {pending.map((it) => (
                <tr key={it.id}>
                  <td className="prod" style={{ fontSize: 13 }}>{it.name}</td>
                  <td className="ser">{it.serial || '—'}</td>
                  <td>{it.branch?.name ?? '—'}</td>
                  <td className="tright cost data">{money(it.cost)}</td>
                  <td className="tright"><button className="btn primary sm" onClick={() => setConfirming(it)}>Təsdiqlə →</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {groups.length === 0 && pending.length === 0 && <div className="card"><div className="center-msg">Nəticə yoxdur</div></div>}

      {groups.map(([day, rows]) => {
        const turnover = rows.reduce((a, s) => a + s.sale_price, 0)
        const profit = rows.reduce((a, s) => a + s.profit, 0)
        return (
          <div className="card" key={day} style={{ overflow: 'hidden' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 18px', background: 'var(--surface)', borderBottom: '1px solid var(--line)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <input type="checkbox" checked={selDays.has(day)} onChange={() => toggleDay(day)} title="Bu günü seç (toplama əlavə et)" style={{ width: 16, height: 16, cursor: 'pointer', accentColor: 'var(--ink)' }} />
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#5A6474" strokeWidth={2} strokeLinecap="round"><rect x="3" y="4.5" width="18" height="17" rx="2" /><path d="M3 9h18M8 2.5v4M16 2.5v4" /></svg>
                <span style={{ fontWeight: 800, fontSize: 14 }}>{dateFullAz(day)}</span>
                <span className="pill neut">{rows.length} satış</span>
              </div>
              <div style={{ display: 'flex', gap: 18 }}>
                <span className="tiny">Dövriyyə: <b className="data" style={{ color: 'var(--ink)' }}>{money(turnover)}</b></span>
                <span className="tiny">Mənfəət: <b className="data profit">+{money(profit)}</b></span>
              </div>
            </div>
            <table>
              <thead><tr><th>Məhsul</th><th>Seriya</th><th>Müştəri</th><th className="tright">Alış</th><th className="tright">Satış</th><th className="tright">Mənfəət</th><th>Kanal</th><th>Zəmanət</th></tr></thead>
              <tbody>
                {rows.map((s) => (
                  <tr key={s.id} style={{ cursor: 'pointer' }} onClick={() => setEdit(s)}>
                    <td className="prod" style={{ fontSize: 13 }}>{s.item?.name}</td>
                    <td className="ser">{s.item?.serial || '—'}</td>
                    <td>{s.customer?.name ?? '—'}</td>
                    <td className="tright cost data">{money(s.item?.cost ?? 0)}</td>
                    <td className="tright data" style={{ fontWeight: 600 }}>{money(s.sale_price)}</td>
                    <td className="tright profit data">+{money(s.profit)}</td>
                    <td><span className={'pill ' + chTag(s.channel)}>{CHANNEL_AZ[s.channel] ?? s.channel}</span></td>
                    <td className="tiny">{s.warranty_months ? `${s.warranty_months} ay` : '—'}</td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr style={{ background: 'var(--surface)', borderTop: '2px solid var(--line)' }}>
                  <td colSpan={4} style={{ fontWeight: 800, fontSize: 12.5 }}>Günün cəmi · {rows.length} satış</td>
                  <td className="tright data" style={{ fontWeight: 800 }}>{money(turnover)}</td>
                  <td className="tright data profit" style={{ fontWeight: 800 }}>+{money(profit)}</td>
                  <td colSpan={2}></td>
                </tr>
              </tfoot>
            </table>
          </div>
        )
      })}

      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 14 }}>
        <button className="btn ghost sm" disabled={page === 0} onClick={() => setPage((p) => Math.max(0, p - 1))}>← Əvvəlki</button>
        <span className="tiny">Səhifə {page + 1}</span>
        <button className="btn ghost sm" disabled={list.length < PAGE} onClick={() => setPage((p) => p + 1)}>Növbəti →</button>
      </div>

      {selTotals.days > 0 && (
        <div style={{ position: 'fixed', right: 24, bottom: 24, zIndex: 50, background: 'var(--ink)', color: '#fff', borderRadius: 14, padding: '14px 18px', boxShadow: '0 12px 34px -8px rgba(20,33,58,.55)', minWidth: 250 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
            <span style={{ fontWeight: 800, fontSize: 13 }}>Seçilmiş {selTotals.days} gün</span>
            <span style={{ cursor: 'pointer', fontSize: 12, opacity: .85, textDecoration: 'underline' }} onClick={() => setSelDays(new Set())}>Təmizlə</span>
          </div>
          <div style={{ display: 'flex', gap: 22 }}>
            <div><div style={{ fontSize: 11, opacity: .7 }}>Dövriyyə</div><div className="data" style={{ fontSize: 19, fontWeight: 700, marginTop: 2 }}>{money(selTotals.t)}</div></div>
            <div><div style={{ fontSize: 11, opacity: .7 }}>Mənfəət</div><div className="data" style={{ fontSize: 19, fontWeight: 700, marginTop: 2 }}>+{money(selTotals.p)}</div></div>
          </div>
          <div style={{ fontSize: 11, opacity: .7, marginTop: 8 }}>{selTotals.n} satış</div>
        </div>
      )}

      {edit && (
        <FormModal
          title="Satışı redaktə et" subtitle={edit.item?.name} submitLabel="Yadda saxla" onClose={() => setEdit(null)}
          initial={{ sale_price: edit.sale_price, channel: edit.channel, warranty_months: edit.warranty_months, customer_id: edit.customer?.id ?? '', sold_at: (edit.sold_at ?? '').slice(0, 10) }}
          fields={[
            { name: 'sale_price', label: 'Satış qiyməti (₼)', type: 'number', required: true },
            { name: 'sold_at', label: 'Satılma tarixi', type: 'date' },
            { name: 'channel', label: 'Kanal', type: 'select', options: CH_OPTS },
            { name: 'warranty_months', label: 'Zəmanət (ay)', type: 'number' },
            { name: 'customer_id', label: 'Müştəri', type: 'customer', full: true, placeholder: 'müştəri axtar və ya yeni yarat…' },
          ]}
          onSubmit={async (v) => { const body: Record<string, unknown> = { ...v }; if (!body.customer_id) delete body.customer_id; await api(`/sales/${edit.id}`, { method: 'PUT', body: JSON.stringify(body) }); bump(); toast('Satış yeniləndi'); setEdit(null) }}
        />
      )}

      {confirming && (
        <FormModal
          title="Satışı təsdiqlə" subtitle={`${confirming.name} · alış ${money(confirming.cost)}`} submitLabel="Təsdiqlə" onClose={() => setConfirming(null)}
          fields={[
            { name: 'sale_price', label: `Satış qiyməti (₼) — sayt qiyməti: ${money(confirming.price ?? 0)}`, type: 'number', required: true, full: true, placeholder: 'əl ilə yazın' },
          ]}
          onSubmit={async (v) => { await api('/sales/confirm', { method: 'POST', body: JSON.stringify({ item_id: confirming.id, sale_price: Number(v.sale_price) || 0 }) }); bump(); toast('Satış təsdiqləndi'); setConfirming(null) }}
        />
      )}
    </div>
  )
}
