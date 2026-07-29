import { useMemo, useState } from 'react'
import { api, type Branch, type Category, type Item, STATUS_AZ, STATUS_TAG } from '../api'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'
import { money, ageDays, dateFullAz } from '../lib/format'
import NewProductModal from '../components/NewProductModal'

export default function Stock({ search = '' }: { search?: string }) {
  const { key, bump } = useRefresh()
  const { data: items, loading } = useFetch<Item[]>('/items', [key])
  const { data: cats } = useFetch<Category[]>('/categories', [])
  const { data: branches } = useFetch<Branch[]>('/branches', [])
  const [cat, setCat] = useState('all')
  const [branch, setBranch] = useState('all')
  const [status, setStatus] = useState('in_stock')
  const [modal, setModal] = useState<{ item?: Item } | null>(null)

  // realizasiyaya verilən cihazlar stokda görünmür (yalnız Realizasiya səhifəsində)
  const list = (items ?? []).filter((i) => i.status !== 'consignment')
  const filtered = useMemo(() => list.filter((i) =>
    (cat === 'all' || i.category?.name === cat) &&
    (branch === 'all' || String(i.branch_id) === branch) &&
    (status === 'all' || i.status === status) &&
    (search === '' || i.name.toLowerCase().includes(search.toLowerCase()) || i.serial.toLowerCase().includes(search.toLowerCase()))
  ), [list, cat, branch, status, search])

  // Excel-dəki kimi gəldiyi (created_at) günə görə qruplaşdır — yeni gün üstdə
  const groups = useMemo(() => {
    const sorted = [...filtered].sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
    const m = new Map<string, Item[]>()
    for (const i of sorted) {
      const day = (i.created_at || '').slice(0, 10)
      if (!m.has(day)) m.set(day, [])
      m.get(day)!.push(i)
    }
    return [...m.entries()]
  }, [filtered])

  // KPI kartları — kateqoriya/filial/axtarışa görə yenilənir (status filtrindən asılı deyil,
  // çünki kartlar onsuz da statusa görə bölünür: Stokda / Rezerv / Ölü)
  const kpiBase = useMemo(() => list.filter((i) =>
    (cat === 'all' || i.category?.name === cat) &&
    (branch === 'all' || String(i.branch_id) === branch) &&
    (search === '' || i.name.toLowerCase().includes(search.toLowerCase()) || i.serial.toLowerCase().includes(search.toLowerCase()))
  ), [list, cat, branch, search])
  const inStock = kpiBase.filter((i) => i.status === 'in_stock')
  const stockValue = inStock.reduce((a, i) => a + i.cost * (i.quantity || 1), 0)
  const reserved = kpiBase.filter((i) => i.status === 'reserved').length
  const dead = inStock.filter((i) => ageDays(i.created_at) > 90).length

  async function toggleSite(i: Item, e: React.MouseEvent) {
    e.stopPropagation()
    try { await api(`/items/${i.id}`, { method: 'PUT', body: JSON.stringify({ show_on_site: !i.show_on_site }) }); bump() } catch { /* ignore */ }
  }

  return (
    <div className="content">
      <div className="grid4">
        <div className="card kpi"><div className="eyebrow">Stokda</div><div className="v">{inStock.length} <span className="tiny">cihaz</span></div></div>
        <div className="card kpi"><div className="eyebrow">Anbar dəyəri (alış)</div><div className="v">{money(stockValue)}</div></div>
        <div className="card kpi"><div className="eyebrow">Rezervdə</div><div className="v">{reserved}</div></div>
        <div className="card kpi"><div className="eyebrow" style={{ color: 'var(--warn)' }}>Ölü stok · 90+ gün</div><div className="v" style={{ color: 'var(--warn)' }}>{dead}</div></div>
      </div>

      <div className="toolbar">
        <div className="chips">
          <span className={'chip' + (cat === 'all' ? ' on' : '')} onClick={() => setCat('all')}>Hamısı</span>
          {cats?.map((c) => <span key={c.id} className={'chip' + (cat === c.name ? ' on' : '')} onClick={() => setCat(c.name)}>{c.name}</span>)}
        </div>
        <select className="select" value={branch} onChange={(e) => setBranch(e.target.value)} style={{ cursor: 'pointer' }}>
          <option value="all">Filial: Hamısı</option>
          {branches?.map((b) => <option key={b.id} value={String(b.id)}>{b.name}</option>)}
        </select>
        <select className="select" value={status} onChange={(e) => setStatus(e.target.value)} style={{ cursor: 'pointer' }}>
          <option value="all">Status: Hamısı</option>
          <option value="in_stock">Stokda</option>
          <option value="reserved">Rezerv</option>
          <option value="sold">Satıldı</option>
          <option value="returned">Qaytarıldı</option>
        </select>
        {search && <span className="tiny">axtarış: «{search}»</span>}
        <button className="btn primary sm" style={{ marginLeft: 'auto' }} onClick={() => setModal({})}><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth={2.6} strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>Yeni məhsul</button>
      </div>

      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><circle cx="12" cy="12" r="9" /><path d="M12 8h.01M11 12h1v4h1" /></svg><div><b>Excel-dəki kimi — gəldiyi günə görə bölünmüş.</b> Hər gün ayrıca cədvəldir. Sətrə klik → redaktə et. Satılan cihaz da burada görünür (status «Satıldı»).</div></div>

      {loading && <div className="card"><div className="center-msg">Yüklənir…</div></div>}
      {!loading && groups.length === 0 && <div className="card"><div className="center-msg">Nəticə yoxdur</div></div>}

      {groups.map(([day, rows]) => {
        const dayValue = rows.reduce((a, i) => a + i.cost * (i.quantity || 1), 0)
        return (
          <div className="card" key={day} style={{ overflow: 'hidden' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 18px', background: 'var(--surface)', borderBottom: '1px solid var(--line)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#5A6474" strokeWidth={2} strokeLinecap="round"><rect x="3" y="4.5" width="18" height="17" rx="2" /><path d="M3 9h18M8 2.5v4M16 2.5v4" /></svg>
                <span style={{ fontWeight: 800, fontSize: 14 }}>{dateFullAz(day)}</span>
                <span className="pill neut">{rows.length} cihaz</span>
              </div>
              <div style={{ display: 'flex', gap: 18 }}>
                <span className="tiny">Dəyər (alış): <b className="data" style={{ color: 'var(--ink)' }}>{money(dayValue)}</b></span>
              </div>
            </div>
            <table>
              <thead><tr><th style={{ width: 44, textAlign: 'center' }}>№</th><th>Məhsul</th><th>Kateqoriya</th><th>Filial</th><th className="tright">Alış</th><th className="tright">Optavoy</th><th className="tright">Satış</th><th>Status</th><th>Saytda</th></tr></thead>
              <tbody>
                {rows.map((i, idx) => (
                  <tr key={i.id} style={{ cursor: 'pointer' }} onClick={() => setModal({ item: i })}>
                    <td className="data" style={{ textAlign: 'center', color: 'var(--muted)', fontWeight: 700 }}>{idx + 1}</td>
                    <td><div className="prod">{i.name}{i.quantity > 1 && <span className="pill" style={{ marginLeft: 6, background: 'var(--card)', border: '1px solid var(--line)', fontWeight: 700 }}>{i.quantity} ədəd</span>}</div><div className="ser">{i.serial || '—'}</div><div className="attrchips">{i.values?.map((v) => <span key={v.id} className="attr">{v.value}</span>)}</div></td>
                    <td>{i.category?.name}</td>
                    <td>{i.branch?.name}</td>
                    <td className="tright cost data">{money(i.cost)}</td>
                    <td className="tright data" style={{ color: 'var(--muted)' }}>{i.wholesale_price ? money(i.wholesale_price) : '—'}</td>
                    <td className="tright data" style={{ fontWeight: 600 }}>
                      {i.price
                        ? (i.discount > 0 && i.discount < i.price
                          ? <>{money(i.price - i.discount)}<div className="tiny" style={{ fontWeight: 500 }}><span style={{ textDecoration: 'line-through', color: 'var(--muted)' }}>{money(i.price)}</span> <span style={{ color: 'var(--bad)' }}>−{money(i.discount)}</span></div></>
                          : money(i.price))
                        : '—'}
                    </td>
                    <td><span className={'pill ' + STATUS_TAG[i.status]}>{STATUS_AZ[i.status]}{i.status === 'reserved' ? ' · 24s' : ''}</span></td>
                    <td><div className={'tg' + (i.show_on_site ? ' on' : '')} onClick={(e) => toggleSite(i, e)} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      })}

      {modal && <NewProductModal edit={modal.item} onClose={() => setModal(null)} />}
    </div>
  )
}
