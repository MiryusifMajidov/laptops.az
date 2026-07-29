import { useEffect, useState } from 'react'
import { api, type Branch, type Category } from '../api'
import { useFetch } from '../lib/hooks'
import { money } from '../lib/format'

interface CalcResult {
  turnover: number; profit: number; cost: number; count: number
  expenses: number; net: number; credit_pending: number; credit_pending_count: number
}

function iso(d: Date): string {
  const p = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

export default function Calculator() {
  const { data: branches } = useFetch<Branch[]>('/branches', [])
  const { data: cats } = useFetch<Category[]>('/categories', [])

  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [branchSel, setBranchSel] = useState<Set<number>>(new Set())
  const [catSel, setCatSel] = useState<Set<string>>(new Set())
  const [res, setRes] = useState<CalcResult | null>(null)
  const [loading, setLoading] = useState(false)

  const toggleBranch = (id: number) => setBranchSel((s) => { const n = new Set(s); if (n.has(id)) n.delete(id); else n.add(id); return n })
  const toggleCat = (name: string) => setCatSel((s) => { const n = new Set(s); if (n.has(name)) n.delete(name); else n.add(name); return n })

  function preset(kind: 'today' | 'week' | 'month' | 'year' | 'all') {
    const now = new Date()
    if (kind === 'all') { setFrom(''); setTo('') ; return }
    if (kind === 'today') { const t = iso(now); setFrom(t); setTo(t); return }
    if (kind === 'week') { const d = new Date(now); d.setDate(d.getDate() - 6); setFrom(iso(d)); setTo(iso(now)); return }
    if (kind === 'month') { setFrom(iso(new Date(now.getFullYear(), now.getMonth(), 1))); setTo(iso(now)); return }
    if (kind === 'year') { setFrom(iso(new Date(now.getFullYear(), 0, 1))); setTo(iso(now)); return }
  }

  useEffect(() => {
    let alive = true
    setLoading(true)
    const qs = new URLSearchParams()
    if (from) qs.set('from', from)
    if (to) qs.set('to', to)
    if (branchSel.size) qs.set('branches', [...branchSel].join(','))
    if (catSel.size) qs.set('categories', [...catSel].join(','))
    api<CalcResult>('/calc?' + qs.toString())
      .then((r) => { if (alive) setRes(r) })
      .catch(() => { if (alive) setRes(null) })
      .finally(() => { if (alive) setLoading(false) })
    return () => { alive = false }
  }, [from, to, branchSel, catSel])

  const chip = (label: string, active: boolean, onClick: () => void) => (
    <span className={'chip' + (active ? ' on' : '')} onClick={onClick} style={{ cursor: 'pointer' }}>{label}</span>
  )

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><rect x="4" y="2.5" width="16" height="19" rx="2" /><path d="M8 6.5h8M8 15h4M16 15v3" /></svg><div><b>Dövriyyə / net qazanc kalkulyatoru.</b> Tarix aralığı, filial(lar) və kateqoriya(lar) seç — sistem seçdiyinə uyğun dövriyyəni, mənfəəti, xərcləri və <b>əlimizdə olmalı xalis pulu</b> hesablayır. Bağlanmamış kredit (əlimizə gəlməyən pul) ayrıca göstərilir.</div></div>

      {/* 1 · Tarix aralığı */}
      <div className="card pad">
        <div className="h" style={{ fontSize: 14, marginBottom: 10 }}>1 · Tarix aralığı</div>
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'flex-end' }}>
          <div className="field" style={{ margin: 0 }}><label>Başlanğıc</label><input type="date" value={from} onChange={(e) => setFrom(e.target.value)} /></div>
          <div className="field" style={{ margin: 0 }}><label>Son</label><input type="date" value={to} onChange={(e) => setTo(e.target.value)} /></div>
          <div className="chips" style={{ marginBottom: 2 }}>
            {chip('Bu gün', false, () => preset('today'))}
            {chip('Bu həftə', false, () => preset('week'))}
            {chip('Bu ay', false, () => preset('month'))}
            {chip('Bu il', false, () => preset('year'))}
            {chip('Hamısı', from === '' && to === '', () => preset('all'))}
          </div>
        </div>
      </div>

      {/* 2 · Filiallar */}
      <div className="card pad">
        <div className="h" style={{ fontSize: 14, marginBottom: 10 }}>2 · Filial(lar) <span className="tiny" style={{ fontWeight: 600 }}>seçilməsə — hamısı</span></div>
        <div className="chips">
          {chip('Hamısı', branchSel.size === 0, () => setBranchSel(new Set()))}
          {branches?.map((b) => chip(b.name, branchSel.has(b.id), () => toggleBranch(b.id)))}
        </div>
      </div>

      {/* 3 · Kateqoriyalar */}
      <div className="card pad">
        <div className="h" style={{ fontSize: 14, marginBottom: 10 }}>3 · Kateqoriya(lar) <span className="tiny" style={{ fontWeight: 600 }}>seçilməsə — hamısı</span></div>
        <div className="chips">
          {chip('Hamısı', catSel.size === 0, () => setCatSel(new Set()))}
          {cats?.map((c) => chip(c.name, catSel.has(c.name), () => toggleCat(c.name)))}
        </div>
      </div>

      {/* Nəticə */}
      <div className="grid4">
        <div className="card kpi"><div className="eyebrow">Dövriyyə</div><div className="v">{money(res?.turnover ?? 0)}</div><div className="tiny">{res?.count ?? 0} satış</div></div>
        <div className="card kpi"><div className="eyebrow">Ümumi mənfəət</div><div className="v" style={{ color: 'var(--good-ink)' }}>{money(res?.profit ?? 0)}</div><div className="tiny">alış: {money(res?.cost ?? 0)}</div></div>
        <div className="card kpi"><div className="eyebrow" style={{ color: 'var(--bad)' }}>Əlavə xərclər</div><div className="v" style={{ color: 'var(--bad)' }}>{money(res?.expenses ?? 0)}</div><div className="tiny">bu aralıqda</div></div>
        <div className="card kpi" style={{ background: 'var(--ink)', color: '#fff' }}><div className="eyebrow" style={{ color: 'rgba(255,255,255,.7)' }}>Net — əlimizdə olmalı</div><div className="v" style={{ color: '#fff' }}>{money(res?.net ?? 0)}</div><div className="tiny" style={{ color: 'rgba(255,255,255,.6)' }}>mənfəət − xərclər</div></div>
      </div>

      <div className="card pad" style={{ display: 'flex', alignItems: 'center', gap: 12, borderLeft: '3px solid var(--warn)' }}>
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--warn)" strokeWidth={2} strokeLinecap="round"><path d="M12 9v4M12 17h.01M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0Z" /></svg>
        <div style={{ flex: 1 }}>
          <div style={{ fontWeight: 700, fontSize: 13.5 }}>Bağlanmamış kredit (əlimizə hələ gəlməyib): <span className="data" style={{ color: 'var(--warn)' }}>{money(res?.credit_pending ?? 0)}</span></div>
          <div className="tiny">{res?.credit_pending_count ?? 0} kredit satışı — tam ödənəndə və ya «Biz bağladıq» seçiləndə yuxarıdakı hesaba əlavə olunur.</div>
        </div>
        {loading && <span className="tiny">hesablanır…</span>}
      </div>
    </div>
  )
}
