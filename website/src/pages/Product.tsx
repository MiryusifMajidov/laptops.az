import { useEffect, useRef, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api, thumbURL, money, gallery, attrVal, hasDiscount, type Product as P } from '../api'
import { useI18n } from '../i18n'
import ProductCard, { PhIcon } from '../ProductCard'

// "önərilən" başlığı — backend i18n-ə toxunmadan əsas dillər üçün
const RECS_LABEL: Record<string, string> = { az: 'Oxşar məhsullar', ru: 'Похожие товары', tr: 'Benzer ürünler', en: 'Similar products' }
import { Price } from '../Price'
import OrderModal from '../OrderModal'

export default function Product() {
  const { id } = useParams()
  const { t, tt, lang } = useI18n()
  const [p, setP] = useState<P | null>(null)
  const [err, setErr] = useState('')
  const [sel, setSel] = useState(0)
  const [order, setOrder] = useState(false)
  const [recs, setRecs] = useState<P[]>([])
  const touchX = useRef<number | null>(null)
  const thumbsRef = useRef<HTMLDivElement | null>(null)

  // aktiv thumbnail-i üfüqi zolaqda görünüşə gətir
  useEffect(() => {
    const el = thumbsRef.current?.children[sel] as HTMLElement | undefined
    el?.scrollIntoView({ behavior: 'smooth', inline: 'center', block: 'nearest' })
  }, [sel])

  useEffect(() => {
    setP(null); setErr('')
    api<P>('/products/' + id + '?lang=' + lang).then((d) => { setP(d); setSel(0) }).catch((e) => setErr((e as Error).message))
  }, [id, lang])

  // önərilən: eyni kateqoriyadan, qiymətcə ən yaxın olanlar
  useEffect(() => {
    if (!p) { setRecs([]); return }
    const qs = new URLSearchParams({ lang })
    if (p.category?.name) qs.set('category', p.category.name)
    api<P[]>('/products?' + qs.toString())
      .then((d) => setRecs(d.filter((x) => x.id !== p.id).sort((a, b) => Math.abs(a.price - p.price) - Math.abs(b.price - p.price)).slice(0, 8)))
      .catch(() => setRecs([]))
  }, [p?.id, lang]) // eslint-disable-line react-hooks/exhaustive-deps

  if (err) return <div className="container center" style={{ padding: '80px 20px' }}>{t('product.notFound')} <Link to="/" style={{ color: 'var(--ink)', fontWeight: 700 }}>{t('common.home')} →</Link></div>
  if (!p) return <div className="container center" style={{ padding: '80px 20px' }}>{t('common.loading')}</div>

  const imgs = [p.card_image, ...gallery(p)].filter(Boolean)
  const brand = attrVal(p, 'Marka') || p.category?.name
  const specs = p.values?.filter((v) => v.value) ?? []

  // əsas şəkli sağa-sola sürüşdürməklə dəyiş
  const swipe = (dir: number) => setSel((s) => Math.max(0, Math.min(imgs.length - 1, s + dir)))
  const onTouchStart = (e: React.TouchEvent) => { touchX.current = e.touches[0].clientX }
  const onTouchEnd = (e: React.TouchEvent) => {
    if (touchX.current == null) return
    const dx = e.changedTouches[0].clientX - touchX.current
    touchX.current = null
    if (Math.abs(dx) > 40) swipe(dx < 0 ? 1 : -1)
  }

  return (
    <div className="container" style={{ paddingTop: 26, paddingBottom: 20 }}>
      <div className="crumb"><Link to="/">{t('common.home')}</Link> <span style={{ color: '#D8D5CD' }}>/</span> <Link to={`/kateqoriya/${p.category?.name}`}>{tt(p.category?.name ?? '')}</Link> <span style={{ color: '#D8D5CD' }}>/</span> <b>{p.name}</b></div>

      <div className="detail">
        <div className="media">
          <div className="main-img" onTouchStart={onTouchStart} onTouchEnd={onTouchEnd}>
            {imgs.length > 0 ? (
              <div className="mislide" style={{ transform: `translateX(-${sel * 100}%)` }}>
                {imgs.map((im, i) => (
                  <div className="mislide-item" key={i}><img src={thumbURL(im, 1000)} alt={p.name} draggable={false} /></div>
                ))}
              </div>
            ) : <span style={{ color: '#C8C5BC' }}><PhIcon /></span>}
            {imgs.length > 1 && (
              <>
                <button className="mi-arrow left" onClick={() => swipe(-1)} disabled={sel === 0} aria-label="Əvvəlki">‹</button>
                <button className="mi-arrow right" onClick={() => swipe(1)} disabled={sel === imgs.length - 1} aria-label="Sonrakı">›</button>
                <div className="mi-dots">{imgs.map((_, i) => <span key={i} className={'mi-dot' + (i === sel ? ' on' : '')} onClick={() => setSel(i)} />)}</div>
              </>
            )}
          </div>
          {imgs.length > 1 && (
            <div className="thumbs" ref={thumbsRef}>
              {imgs.map((im, i) => (
                <div key={i} className={'thumb' + (i === sel ? ' on' : '')} onClick={() => setSel(i)}><img src={thumbURL(im, 150)} alt="" loading="lazy" /></div>
              ))}
            </div>
          )}
        </div>

        <div className="info">
          <div className="eyebrow" style={{ fontSize: 12.5, letterSpacing: '.06em' }}>{tt(brand)}</div>
          <h1>{p.name}</h1>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginTop: 14, flexWrap: 'wrap' }}>
            <Price p={p} />
            {hasDiscount(p) && <span className="badge disc">−{money(p.discount)}</span>}
            <span className="badge good"><span className="dot" />{t('common.available')}</span>
          </div>

          {specs.length > 0 && (
            <div className="specs">
              {Object.entries(specs.reduce((acc, v) => {
                const n = v.attribute?.name ?? ''
                ;(acc[n] ||= []).push(v.value)
                return acc
              }, {} as Record<string, string[]>)).map(([name, vals]) => (
                <div className="specbox" key={name}><div className="l">{tt(name)}</div><div className="v">{vals.map((x) => tt(x)).join(', ')}</div></div>
              ))}
            </div>
          )}

          <div className="pd-actions">
            <div style={{ display: 'flex', gap: 12, marginTop: 24 }}>
              <button className="btn primary" style={{ flex: 1.4 }} onClick={() => setOrder(true)}>{t('common.order')}</button>
              <a className="btn ghost" style={{ flex: 1 }} href="tel:+994708151283">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#14213A" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round"><path d="M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3 19.5 19.5 0 0 1-6-6 19.8 19.8 0 0 1-3-8.6A2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1.9.4 1.8.7 2.7a2 2 0 0 1-.5 2.1L8.1 9.9a16 16 0 0 0 6 6l1.4-1.2a2 2 0 0 1 2.1-.5c.9.3 1.8.6 2.7.7a2 2 0 0 1 1.7 2Z" /></svg>{t('common.call')}
              </a>
            </div>
            <div className="note">
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="#0A7E57" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" style={{ flex: 'none', marginTop: 1 }}><path d="M9 12l2 2 4-4" /><circle cx="12" cy="12" r="9" /></svg>
              <div>{t('product.orderNote')}</div>
            </div>
          </div>
        </div>
      </div>

      {recs.length > 0 && (
        <section className="recs-section" style={{ marginTop: 44 }}>
          <h2 className="sg" style={{ fontSize: 'clamp(19px, 5vw, 26px)', fontWeight: 600, letterSpacing: '-.02em', margin: '0 0 16px' }}>{RECS_LABEL[lang] ?? RECS_LABEL.az}</h2>
          <div className="pgrid cols3">{recs.map((r) => <ProductCard key={r.id} p={r} />)}</div>
        </section>
      )}

      {/* mobil: sabit alt-panel */}
      <div className="buybar">
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 9, fontSize: 11, color: 'var(--good-ink)', fontWeight: 600 }}>
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#0A7E57" strokeWidth={2} strokeLinecap="round"><path d="M9 12l2 2 4-4" /><circle cx="12" cy="12" r="9" /></svg>{t('product.buybarNote')}
        </div>
        <div style={{ display: 'flex', gap: 10 }}>
          <button className="btn primary" style={{ flex: 1 }} onClick={() => setOrder(true)}>{t('common.order')}</button>
          <a className="btn ghost" style={{ width: 52, padding: 0 }} href="tel:+994708151283" aria-label={t('common.call')}>
            <svg width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="#14213A" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round"><path d="M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3 19.5 19.5 0 0 1-6-6 19.8 19.8 0 0 1-3-8.6A2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1.9.4 1.8.7 2.7a2 2 0 0 1-.5 2.1L8.1 9.9a16 16 0 0 0 6 6l1.4-1.2a2 2 0 0 1 2.1-.5c.9.3 1.8.6 2.7.7a2 2 0 0 1 1.7 2Z" /></svg>
          </a>
        </div>
      </div>

      {order && <OrderModal p={p} onClose={() => setOrder(false)} />}
    </div>
  )
}
