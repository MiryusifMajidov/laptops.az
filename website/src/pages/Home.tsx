import { useEffect, useState, type CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import { api, thumbURL, money, attrVal, hasDiscount, finalPrice, fetchSettings, type Product } from '../api'
import { useI18n } from '../i18n'
import { openMap } from '../mapBus'
import ProductCard, { PhIcon } from '../ProductCard'
import { Price, DiscBadge } from '../Price'

export default function Home() {
  const { t, lang } = useI18n()
  const [products, setProducts] = useState<Product[]>([])
  const [settings, setSettings] = useState<Record<string, string>>({})
  useEffect(() => { api<Product[]>('/products?lang=' + lang).then(setProducts).catch(() => {}) }, [lang])
  useEffect(() => { fetchSettings().then(setSettings).catch(() => {}) }, [])

  const featured = products.slice(0, 8)
  // seçilmiş məhsul (Site Settings-dən) — yoxdursa/satılıbsa ən yeni məhsula düşür
  const fid = parseInt(settings.featured_item_id || '', 10)
  const hero = (!isNaN(fid) ? products.find((p) => p.id === fid) : undefined) || products[0]

  // plitə fon şəkli — şəkil varsa tünd örtük + şəkil, yoxdursa sadə kart
  const tileBg = (url?: string): CSSProperties | undefined =>
    url ? { backgroundImage: `linear-gradient(115deg, rgba(20,33,58,.95) 30%, rgba(20,33,58,.5) 100%), url(${thumbURL(url, 800)})`, backgroundSize: 'cover', backgroundPosition: 'center right' } : undefined
  const brand = hero ? (attrVal(hero, 'Marka') || hero.category?.name) : ''

  return (
    <>
      {/* MASAÜSTÜ — marketinq hero */}
      <section className="hero">
        <div className="container grid">
          <div style={{ flex: 1 }}>
            <div className="eyebrow">{t('home.heroEyebrow')}</div>
            <h1 style={{ whiteSpace: 'pre-line' }}>{t('home.heroTitle')}</h1>
            <p>{t('home.heroText')}</p>
            <div className="cta">
              <Link to="/kateqoriya/Notebook" className="btn primary">{t('home.ctaBrowse')}</Link>
              <button type="button" className="btn ghost" onClick={openMap}>{t('home.ctaVisit')}</button>
            </div>
            <div className="live"><span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--good)' }} />{t('home.liveText')}</div>
          </div>
          <div className="hero-media">
            <div className="hero-img">
              {hero?.card_image ? <img src={thumbURL(hero.card_image, 800)} alt={hero.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : <span style={{ color: '#C8C5BC' }}><PhIcon /></span>}
            </div>
            {hero && (
              <Link to={`/mehsul/${hero.id}`} className="hero-badge">
                <div><div style={{ fontSize: 11, color: 'var(--muted)', fontWeight: 700, letterSpacing: '.04em' }}>{t('home.featuredEyebrow')}</div><div style={{ fontSize: 14.5, fontWeight: 800, marginTop: 2 }}>{hero.name}</div></div>
                <div style={{ width: 1, height: 34, background: 'var(--line2)' }} />
                <div><div className="sg" style={{ fontSize: 20, fontWeight: 600 }}>{money(finalPrice(hero))}{hasDiscount(hero) && <span className="old-price" style={{ marginLeft: 6 }}>{money(hero.price)}</span>}</div><div style={{ display: 'flex', alignItems: 'center', gap: 5, marginTop: 2 }}><span style={{ width: 6, height: 6, borderRadius: '50%', background: 'var(--good)' }} /><span style={{ fontSize: 11.5, color: 'var(--good-ink)', fontWeight: 700 }}>{t('common.available')}</span></div></div>
              </Link>
            )}
          </div>
        </div>
      </section>

      <div className="container">
        {/* MOBİL — hero məhsul kartı */}
        {hero && (
          <Link to={`/mehsul/${hero.id}`} className="mhero-card">
            <div className="mhero-img">
              {hero.card_image ? <img src={thumbURL(hero.card_image, 800)} alt={hero.name} /> : <PhIcon />}
              <span className="avail good" style={{ top: 11, left: 11 }}><span className="dot" />{t('common.available')}</span>
              <DiscBadge p={hero} />
            </div>
            <div className="mhero-body">
              <div className="mhero-eyebrow">{brand} · {t('home.featuredEyebrow')}</div>
              <div className="mhero-name">{hero.name}</div>
              <div className="mhero-row"><Price p={hero} /><span className="btn primary sm">{t('common.order')}</span></div>
            </div>
          </Link>
        )}

        {/* MOBİL — kateqoriya çipləri */}
        <div className="catchips">
          <Link to="/mehsullar" className="on">{t('home.chipAll')}</Link>
          <Link to="/kateqoriya/Notebook">{t('nav.notebook')}</Link>
          <Link to={'/kateqoriya/' + encodeURIComponent('Yığım (PC)')}>{t('nav.desktop')}</Link>
          <Link to="/kateqoriya/Telefon">{t('nav.phone')}</Link>
          <Link to="/kateqoriya/Aksesuar">{t('nav.accessory')}</Link>
        </div>

        <section className="section">
          <div className="head">
            <h2>{t('home.featuredTitle')}</h2>
            <Link to="/mehsullar" className="seeall">{t('home.seeAll')} <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round"><path d="M5 12h14M13 6l6 6-6 6" /></svg></Link>
          </div>
          {featured.length === 0 ? <div className="center">{t('home.loadingProducts')}</div> : (
            <div className="pgrid">{featured.map((p) => <ProductCard key={p.id} p={p} />)}</div>
          )}
        </section>

        <section className="section" style={{ paddingTop: 40 }}>
          <div className="tiles">
            <Link to="/kateqoriya/Notebook" className="tile" style={tileBg(settings.tile_notebook_img)}>
              <h3>{t('home.tileNotebookTitle')}</h3>
              <p>{t('home.tileNotebookDesc')}</p>
              <span className="btn ghost sm" style={{ marginTop: 16, width: 'fit-content' }}>{t('home.tileCollection')}</span>
            </Link>
            <Link to={'/kateqoriya/' + encodeURIComponent('Yığım (PC)')} className="tile" style={tileBg(settings.tile_desktop_img)}>
              <h3>{t('home.tileDesktopTitle')}</h3>
              <p>{t('home.tileDesktopDesc')}</p>
              <span className="btn ghost sm" style={{ marginTop: 16, width: 'fit-content' }}>{t('home.tileView')}</span>
            </Link>
          </div>
          <div className="tiles-sm">
            <Link to="/kateqoriya/Telefon" className="tile-sm">
              <div className="ic"><svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="#14213A" strokeWidth={1.7} strokeLinecap="round" strokeLinejoin="round"><rect x="6.5" y="2" width="11" height="20" rx="2.6" /><path d="M10.5 18.5h3" /></svg></div>
              <div style={{ flex: 1 }}><div style={{ fontSize: 18, fontWeight: 800 }}>{t('home.tilePhoneTitle')}</div><div className="muted" style={{ fontSize: 13.5, marginTop: 3 }}>{t('home.tilePhoneDesc')}</div></div>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#14213A" strokeWidth={2.2} strokeLinecap="round"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
            </Link>
            <Link to="/kateqoriya/Aksesuar" className="tile-sm">
              <div className="ic"><svg width="34" height="34" viewBox="0 0 24 24" fill="none" stroke="#14213A" strokeWidth={1.7} strokeLinecap="round" strokeLinejoin="round"><rect x="7.5" y="2.5" width="9" height="19" rx="4.5" /><path d="M12 6.5v4" /></svg></div>
              <div style={{ flex: 1 }}><div style={{ fontSize: 18, fontWeight: 800 }}>{t('home.tileAccessoryTitle')}</div><div className="muted" style={{ fontSize: 13.5, marginTop: 3 }}>{t('home.tileAccessoryDesc')}</div></div>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#14213A" strokeWidth={2.2} strokeLinecap="round"><path d="M5 12h14M13 6l6 6-6 6" /></svg>
            </Link>
          </div>
        </section>
      </div>
    </>
  )
}
