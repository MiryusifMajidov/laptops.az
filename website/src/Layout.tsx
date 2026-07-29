import { useState, useEffect } from 'react'
import { Link, NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom'
import AiAssistant from './AiAssistant'
import StoreMap from './StoreMap'
import { openAi } from './aiBus'
import { openMap } from './mapBus'
import { useI18n } from './i18n'
import { imgURL } from './api'

// [mətn açarı, kateqoriya adı (DB)] — link avtomatik qurulur
const NAV: [string, string][] = [
  ['nav.notebook', 'Notebook'],
  ['nav.desktop', 'Yığım (PC)'],
  ['nav.phone', 'Telefon'],
  ['nav.accessory', 'Aksesuar'],
]
const catPath = (cat: string) => '/kateqoriya/' + encodeURIComponent(cat)

function LangSwitch() {
  const { lang, setLang, langs } = useI18n()
  const [open, setOpen] = useState(false)
  if (langs.length < 2) return null
  const cur = langs.find((l) => l.code === lang)
  const flag = (icon: string, code: string) =>
    icon ? <img className="lang-flag" src={imgURL(icon)} alt="" /> : <span className="lang-code">{code.toUpperCase()}</span>
  return (
    <div className="lang-switch">
      <button className="lang-btn" onClick={() => setOpen((o) => !o)} aria-label="Dil / Language">
        {flag(cur?.icon ?? '', cur?.code ?? lang)}
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.4} strokeLinecap="round"><path d="M6 9l6 6 6-6" /></svg>
      </button>
      {open && (
        <>
          <div className="lang-scrim" onClick={() => setOpen(false)} />
          <div className="lang-menu">
            {langs.map((l) => (
              <button key={l.code} className={'lang-item' + (l.code === lang ? ' on' : '')} onClick={() => { setLang(l.code); setOpen(false) }}>
                {flag(l.icon, l.code)}<span>{l.name}</span>
              </button>
            ))}
          </div>
        </>
      )}
    </div>
  )
}

export default function Layout() {
  const nav = useNavigate()
  const loc = useLocation()
  const { t } = useI18n()
  const [q, setQ] = useState(() => new URLSearchParams(window.location.search).get('q') ?? '')
  const [menu, setMenu] = useState(false)
  const go = (to: string) => { setMenu(false); nav(to) }

  // real-time axtarış: yazdıqca (220ms debounce) nəticələr səhifəsinə keç/süz
  useEffect(() => {
    const term = q.trim()
    const id = setTimeout(() => {
      const onList = loc.pathname.startsWith('/mehsullar') || loc.pathname.startsWith('/kateqoriya')
      if (term) nav('/mehsullar?q=' + encodeURIComponent(term), { replace: onList })
      else if (onList && loc.search.includes('q=')) nav('/mehsullar', { replace: true })
    }, 220)
    return () => clearTimeout(id)
  }, [q]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <>
      <div className="promo">{t('promo.banner')}</div>

      <header className="site-header">
        <div className="container inner">
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <button className="hamburger" onClick={() => setMenu(true)} aria-label="Menyu">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M3 6h18M3 12h18M3 18h18" /></svg>
            </button>
            <Link to="/"><img className="logo" src="/logo-trim.png" alt="Laptops.az" /></Link>
          </div>
          <nav className="nav">
            {NAV.map(([label, cat]) => (
              <NavLink key={cat} to={catPath(cat)} className={({ isActive }) => (isActive ? 'active' : '')}>{t(label)}</NavLink>
            ))}
          </nav>
          <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
            <button className="ai-header-btn" onClick={openAi}>
              <svg width="15" height="15" viewBox="0 0 24 24" fill="#4C86FF"><path d="M12 2.5l2.2 5.8 5.8 2.2-5.8 2.2L12 18.5l-2.2-5.8L4 10.5l5.8-2.2z" /></svg>
              {t('ai.button')}
            </button>
            <div className="hsearch">
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#98A0AC" strokeWidth={2.2} strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4-4" /></svg>
              <input placeholder={t('search.placeholder')} value={q} onChange={(e) => setQ(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter' && q.trim()) nav('/mehsullar?q=' + encodeURIComponent(q.trim())) }} />
            </div>
            <LangSwitch />
          </div>
        </div>
      </header>

      {/* MOBİL axtarış zolağı */}
      <div className="mobile-search">
        <div className="msearch-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#98A0AC" strokeWidth={2.2} strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4-4" /></svg>
          <input placeholder={t('search.placeholderLong')} value={q} onChange={(e) => setQ(e.target.value)}
            onKeyDown={(e) => { if (e.key === 'Enter' && q.trim()) nav('/mehsullar?q=' + encodeURIComponent(q.trim())) }} />
          {q && <button onClick={() => q.trim() && nav('/mehsullar?q=' + encodeURIComponent(q.trim()))} aria-label="Axtar" style={{ color: 'var(--ink)' }}><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.4} strokeLinecap="round"><path d="M5 12h14M13 6l6 6-6 6" /></svg></button>}
        </div>
      </div>

      <Outlet />

      <footer className="footer" id="magaza">
        <div className="container">
          <div className="cols">
            <div style={{ maxWidth: 280 }}>
              <img src="/logo-white.png" alt="Laptops.az" />
              <div className="txt" style={{ color: '#8595B5', marginTop: 14 }}>{t('footer.tagline')}</div>
            </div>
            <div>
              <div className="lbl">{t('footer.storeLabel')}</div>
              <div className="txt" style={{ whiteSpace: 'pre-line' }}>{t('footer.address')}</div>
              <button className="footer-maplink" onClick={openMap}>
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" /><circle cx="12" cy="10" r="3" /></svg>
                {t('map.view')}
              </button>
            </div>
            <div>
              <div className="lbl">{t('footer.contactLabel')}</div>
              <div className="txt sg" style={{ whiteSpace: 'pre-line' }}>{t('footer.phones')}</div>
            </div>
          </div>
          <div className="bar">{t('footer.copyright')}</div>
        </div>
      </footer>

      {menu && (
        <div className="mobile-menu">
          <div className="mm-head">
            <img src="/logo-white.png" alt="Laptops.az" style={{ height: 22 }} />
            <button onClick={() => setMenu(false)} aria-label="Bağla"><svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth={2.2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
          </div>
          <div className="mm-links">
            {NAV.map(([label, cat]) => (
              <a key={cat} onClick={() => go(catPath(cat))}>{t(label)}<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#4C86FF" strokeWidth={2.2} strokeLinecap="round"><path d="M9 6l6 6-6 6" /></svg></a>
            ))}
            <a onClick={() => { setMenu(false); openAi() }}>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 10 }}>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="#4C86FF"><path d="M12 2.5l2.2 5.8 5.8 2.2-5.8 2.2L12 18.5l-2.2-5.8L4 10.5l5.8-2.2z" /></svg>{t('ai.button')}
              </span>
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#4C86FF" strokeWidth={2.2} strokeLinecap="round"><path d="M9 6l6 6-6 6" /></svg>
            </a>
          </div>
          <div className="mm-store">
            <div style={{ fontSize: 11.5, fontWeight: 700, letterSpacing: '.1em', color: '#5C6B8C', marginBottom: 10 }}>{t('footer.storeLabel')}</div>
            <div style={{ fontSize: 13.5, color: '#CDD6E8', lineHeight: 1.7, whiteSpace: 'pre-line' }}>{t('menu.storeInfo')}</div>
            <a href="tel:+994708151283" className="sg" style={{ fontSize: 14, color: '#fff', fontWeight: 600, marginTop: 10, display: 'inline-block' }}>+994 70 815 12 83</a>
          </div>
        </div>
      )}

      <AiAssistant />
      <StoreMap />
    </>
  )
}
