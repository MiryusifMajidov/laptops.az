import { useEffect, useMemo, useRef, useState } from 'react'
import { useParams, useSearchParams, Link } from 'react-router-dom'
import { api, attrVal, money, CAT_LABEL, type Product } from '../api'
import { useI18n } from '../i18n'
import ProductCard from '../ProductCard'

const ORDER = ['Marka', 'Prosessor', 'RAM', 'SSD', 'Ekran kartı', 'Ekran', 'Yaddaş', 'Vəziyyət', 'Rəng', 'Növ', 'Ölçü', 'Tezlik']
const SORTS: [string, string][] = [['pop', 'cat.sortPopular'], ['cheap', 'cat.sortCheap'], ['exp', 'cat.sortExp']]

export default function Category() {
  const { name } = useParams()
  const { t, tt, lang } = useI18n()
  const [sp] = useSearchParams()
  const q = sp.get('q') ?? ''
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(true)
  const [sel, setSel] = useState<Record<string, string[]>>({})
  const [maxPrice, setMaxPrice] = useState(0)
  const [sort, setSort] = useState('pop')
  const [sheet, setSheet] = useState(false)

  const PAGE_SIZE = 24
  const [shown, setShown] = useState(PAGE_SIZE)
  const sentinelRef = useRef<HTMLDivElement | null>(null)

  // Kateqoriya üzrə bir dəfə çəkilir; axtarış (q) client-də real-time süzülür — server sorğusu yox.
  useEffect(() => {
    setLoading(true); setSel({})
    const qs = new URLSearchParams()
    if (name) qs.set('category', name)
    qs.set('lang', lang)
    api<Product[]>('/products?' + qs.toString())
      .then((d) => { setProducts(d); setMaxPrice(Math.max(1, ...d.map((p) => p.price))) })
      .catch(() => setProducts([]))
      .finally(() => setLoading(false))
  }, [name, lang])

  const attrGroups = useMemo(() => {
    const m = new Map<string, Set<string>>()
    for (const p of products) for (const v of p.values ?? []) {
      if (!v.attribute?.name || !v.value) continue
      if (!m.has(v.attribute.name)) m.set(v.attribute.name, new Set())
      m.get(v.attribute.name)!.add(v.value)
    }
    const names = [...m.keys()].sort((a, b) => (ORDER.indexOf(a) < 0 ? 99 : ORDER.indexOf(a)) - (ORDER.indexOf(b) < 0 ? 99 : ORDER.indexOf(b)))
    return names.map((n) => ({ name: n, values: [...m.get(n)!].sort() }))
  }, [products])

  const priceMax = useMemo(() => Math.max(1, ...products.map((p) => p.price)), [products])

  const filtered = useMemo(() => {
    const term = q.trim().toLowerCase()
    let r = products.filter((p) => {
      if (term) {
        const hay = (p.name + ' ' + (p.values ?? []).map((v) => v.value).join(' ')).toLowerCase()
        if (!hay.includes(term)) return false
      }
      for (const [an, vals] of Object.entries(sel)) {
        if (vals.length === 0) continue
        if (!vals.includes(attrVal(p, an))) return false
      }
      return maxPrice === 0 || p.price <= maxPrice
    })
    if (sort === 'cheap') r = [...r].sort((a, b) => a.price - b.price)
    else if (sort === 'exp') r = [...r].sort((a, b) => b.price - a.price)
    return r
  }, [products, sel, maxPrice, sort, q])

  // infinite scroll — sona yaxınlaşanda avtomatik daha çox göstər (server sorğusu yox)
  const paged = filtered.slice(0, shown)
  const hasMore = shown < filtered.length
  useEffect(() => { setShown(PAGE_SIZE) }, [q, name, sort, sel, maxPrice]) // süzgəc dəyişəndə əvvələ
  useEffect(() => { window.scrollTo({ top: 0 }) }, [q, name])              // axtarış/kateqoriya dəyişəndə yuxarı
  useEffect(() => {
    if (!hasMore) return
    const el = sentinelRef.current
    if (!el) return
    const obs = new IntersectionObserver(
      (e) => { if (e[0].isIntersecting) setShown((s) => s + PAGE_SIZE) },
      { rootMargin: '800px' }, // ekrana çatmadan əvvəl yüklə → hiss olunmadan
    )
    obs.observe(el)
    return () => obs.disconnect()
  }, [hasMore, filtered.length, shown])

  const toggle = (an: string, val: string) => setSel((s) => {
    const cur = s[an] ?? []
    return { ...s, [an]: cur.includes(val) ? cur.filter((x) => x !== val) : [...cur, val] }
  })
  const clearAll = () => { setSel({}); setMaxPrice(priceMax) }
  const activeCount = Object.values(sel).reduce((a, v) => a + v.length, 0) + (maxPrice < priceMax ? 1 : 0)
  const catDisp = (n: string) => { const x = tt(n); return x !== n ? x : (CAT_LABEL[n] ?? n) }
  const title = q ? `${t('cat.searchTitle')}: ${q}` : (name ? catDisp(name) : t('cat.allProducts'))
  const crumb = q ? t('cat.searchTitle') : title

  const filterBody = (
    <>
      <div className="fgroup">
        <h4>{t('cat.price')}</h4>
        <input type="range" min={0} max={priceMax} value={maxPrice} onChange={(e) => setMaxPrice(Number(e.target.value))} style={{ width: '100%', accentColor: '#14213A' }} />
        <div className="sg" style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: 'var(--muted)', fontWeight: 600, marginTop: 4 }}><span>₼0</span><span>{money(maxPrice)}</span></div>
      </div>
      {attrGroups.map((g) => (
        <div className="fgroup" key={g.name}>
          <h4>{tt(g.name)}</h4>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 7 }}>
            {g.values.map((val) => (
              <span key={val} className={'chip' + ((sel[g.name] ?? []).includes(val) ? ' on' : '')} onClick={() => toggle(g.name, val)}>{tt(val)}</span>
            ))}
          </div>
        </div>
      ))}
      {attrGroups.length === 0 && !loading && <div className="muted" style={{ fontSize: 13 }}>{t('cat.noFilter')}</div>}
    </>
  )

  return (
    <div className="container" style={{ paddingTop: 26, paddingBottom: 30 }}>
      <div className="crumb"><Link to="/">{t('common.home')}</Link> <span style={{ color: '#D8D5CD' }}>/</span> <b>{crumb}</b></div>
      <div className="cat-head">
        <div style={{ minWidth: 0 }}>
          <h1 className="sg" style={{ fontSize: 'clamp(21px, 6vw, 30px)', fontWeight: 600, letterSpacing: '-.02em', margin: 0 }}>{title}</h1>
          <div className="muted" style={{ fontSize: 13.5, fontWeight: 600, marginTop: 4 }}>{filtered.length} {t('cat.productsSuffix')}</div>
        </div>
        <div style={{ display: 'flex', gap: 10, flex: 'none' }}>
          <button className="filter-trigger" onClick={() => setSheet(true)}>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round"><path d="M4 6h16M7 12h10M10 18h4" /></svg>{t('cat.filter')}{activeCount > 0 ? ` (${activeCount})` : ''}
          </button>
          <select className="sort-select" value={sort} onChange={(e) => setSort(e.target.value)}>
            {SORTS.map(([v, l]) => <option key={v} value={v}>{t(l)}</option>)}
          </select>
        </div>
      </div>

      <div className="catwrap">
        <aside className="filters">
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
            <div style={{ fontSize: 14, fontWeight: 800 }}>{t('cat.filters')}</div>
            {activeCount > 0 && <button onClick={clearAll} style={{ fontSize: 12, fontWeight: 700, color: 'var(--warn)' }}>{t('cat.clear')}</button>}
          </div>
          {filterBody}
        </aside>

        <div style={{ flex: 1, minWidth: 0 }}>
          {loading ? <div className="center">{t('common.loading')}</div>
            : filtered.length === 0 ? <div className="center">{t('cat.noProducts')}</div>
              : <>
                  <div className="pgrid cols3">{paged.map((p) => <ProductCard key={p.id} p={p} />)}</div>
                  {hasMore && <div ref={sentinelRef} className="load-more">{t('common.loading')}</div>}
                </>}
        </div>
      </div>

      {sheet && (
        <div className="sheet-scrim" onClick={() => setSheet(false)}>
          <div className="sheet" onClick={(e) => e.stopPropagation()}>
            <div className="handle" />
            <div className="shead">
              <span className="sg" style={{ fontSize: 19, fontWeight: 600, letterSpacing: '-.02em' }}>{t('cat.filter')}</span>
              <button onClick={clearAll} style={{ fontSize: 13, fontWeight: 700, color: 'var(--ink2)' }}>{t('cat.reset')}</button>
            </div>
            <div className="sbody">
              <div className="fgroup">
                <h4>{t('cat.sort')}</h4>
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 7 }}>
                  {SORTS.map(([v, l]) => (
                    <span key={v} className={'chip' + (sort === v ? ' on' : '')} onClick={() => setSort(v)}>{t(l)}</span>
                  ))}
                </div>
              </div>
              {filterBody}
            </div>
            <div className="sfoot">
              <button className="btn primary" style={{ width: '100%' }} onClick={() => setSheet(false)}>{filtered.length} {t('cat.showResults')}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
