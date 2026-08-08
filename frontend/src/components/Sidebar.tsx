import type { ReactNode } from 'react'
import { type Dashboard, type OnlineOrder, api, currentName, currentRole } from '../api'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'

export type View =
  | 'dashboard' | 'satislar' | 'stok' | 'silinmis' | 'onlayn'
  | 'kredit' | 'kassa' | 'xercler'
  | 'musteriler' | 'filiallar' | 'realizasiya' | 'techizat' | 'kateqoriya'
  | 'partnyor' | 'hesabatlar' | 'kalkulyator' | 'mail' | 'poct' | 'sayt' | 'aisohbet' | 'ziyaretchi' | 'audit' | 'ayarlar'

type NavItem = { view: View; label: string; icon: ReactNode; adminOnly?: boolean }
type NavGroup = { label: string; items: NavItem[] }

const s = { fill: 'none', stroke: 'currentColor', strokeWidth: 1.9, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const }

export const NAV: NavGroup[] = [
  { label: 'ƏMƏLİYYAT', items: [
    { view: 'dashboard', label: 'Ana Panel', icon: <svg viewBox="0 0 24 24" {...s}><rect x="3" y="3" width="7" height="7" rx="1.6" /><rect x="14" y="3" width="7" height="7" rx="1.6" /><rect x="3" y="14" width="7" height="7" rx="1.6" /><rect x="14" y="14" width="7" height="7" rx="1.6" /></svg> },
    { view: 'satislar', label: 'Satışlar', icon: <svg viewBox="0 0 24 24" {...s}><path d="M5 3h14v18l-3-2-2 2-2-2-2 2-3-2z" /><path d="M9 8h6M9 12h6" /></svg> },
    { view: 'stok', label: 'Stok / Anbar', icon: <svg viewBox="0 0 24 24" {...s}><path d="M21 8 L12 3 L3 8 v8 l9 5 9-5 Z" /><path d="M3 8 l9 5 9-5" /><path d="M12 13 v8" /></svg> },
    { view: 'silinmis', label: 'Silinmiş məhsullar', icon: <svg viewBox="0 0 24 24" {...s}><path d="M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14M10 11v6M14 11v6" /></svg> },
    { view: 'onlayn', label: 'Onlayn Sifarişlər', icon: <svg viewBox="0 0 24 24" {...s}><circle cx="9" cy="21" r="1.5" /><circle cx="18" cy="21" r="1.5" /><path d="M2 3h3l2.5 13h11l2-9H6" /></svg> },
  ]},
  { label: 'MALİYYƏ', items: [
    { view: 'kredit', label: 'Kredit / Borclar', icon: <svg viewBox="0 0 24 24" {...s}><rect x="2.5" y="5" width="19" height="14" rx="2.5" /><path d="M2.5 9.5h19" /></svg> },
    { view: 'kassa', label: 'Kassa', icon: <svg viewBox="0 0 24 24" {...s}><rect x="3" y="6" width="18" height="13" rx="2" /><path d="M3 10h18M7 15h4" /></svg> },
    { view: 'xercler', label: 'Xərclər', icon: <svg viewBox="0 0 24 24" {...s}><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" /></svg> },
  ]},
  { label: 'BAZA', items: [
    { view: 'musteriler', label: 'Müştərilər', icon: <svg viewBox="0 0 24 24" {...s}><circle cx="9" cy="8" r="3.2" /><path d="M3.5 19c0-3.3 2.5-5.2 5.5-5.2s5.5 1.9 5.5 5.2" /><path d="M16 5.4a3 3 0 0 1 0 5.4" /><path d="M17.6 13.6c2 .5 3.4 2 3.4 4.1" /></svg> },
    { view: 'filiallar', label: 'Filiallar', icon: <svg viewBox="0 0 24 24" {...s}><path d="M3 21V9l9-6 9 6v12" /><path d="M9 21v-6h6v6" /></svg> },
    { view: 'realizasiya', label: 'Realizasiya', icon: <svg viewBox="0 0 24 24" {...s}><path d="M2.5 6h11v9H2.5Z" /><path d="M13.5 9h4l3 3v3h-7Z" /><circle cx="7" cy="18" r="1.6" /><circle cx="17" cy="18" r="1.6" /></svg> },
    { view: 'techizat', label: 'Təchizat', icon: <svg viewBox="0 0 24 24" {...s}><path d="M12 3l8 4.5v9L12 21l-8-4.5v-9z" /><path d="M12 12l8-4.5M12 12v9M12 12L4 7.5" /></svg> },
    { view: 'kateqoriya', label: 'Kateqoriyalar', icon: <svg viewBox="0 0 24 24" {...s}><path d="M4 6h16M4 12h16M4 18h10" /><circle cx="19" cy="18" r="2" /></svg> },
    { view: 'partnyor', label: 'Tərəfdaşlıq', icon: <svg viewBox="0 0 24 24" {...s}><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" /></svg> },
  ]},
  { label: 'HESABAT & SİSTEM', items: [
    { view: 'hesabatlar', label: 'Hesabatlar', icon: <svg viewBox="0 0 24 24" {...s}><path d="M4 20V10M10 20V4M16 20v-7M20.5 20H3.5" /></svg> },
    { view: 'kalkulyator', label: 'Kalkulyator', icon: <svg viewBox="0 0 24 24" {...s}><rect x="4" y="2.5" width="16" height="19" rx="2" /><path d="M8 6.5h8M8 11h.01M12 11h.01M16 11h.01M8 15h.01M12 15h.01M16 15v3M8 18h4" /></svg> },
    { view: 'mail', label: 'Mail & Excel', icon: <svg viewBox="0 0 24 24" {...s}><rect x="3" y="5" width="18" height="14" rx="2" /><path d="M3 7l9 6 9-6" /></svg> },
    { view: 'poct', label: 'Poçt (info@)', adminOnly: true, icon: <svg viewBox="0 0 24 24" {...s}><path d="M4 4h16v16H4z" /><path d="M4 8l8 5 8-5" /><path d="M4 4l8 5 8-5" /></svg> },
    { view: 'sayt', label: 'Sayt Tənzimləmələri', icon: <svg viewBox="0 0 24 24" {...s}><circle cx="12" cy="12" r="9" /><path d="M3 12h18M12 3a15 15 0 0 1 0 18M12 3a15 15 0 0 0 0 18" /></svg> },
    { view: 'aisohbet', label: 'AI söhbətləri', icon: <svg viewBox="0 0 24 24" {...s}><path d="M21 11.5a8.38 8.38 0 0 1-8.5 8.5 8.5 8.5 0 0 1-3.8-.9L3 21l1.9-5.7a8.5 8.5 0 0 1-.9-3.8A8.38 8.38 0 0 1 12.5 3 8.38 8.38 0 0 1 21 11.5z" /><path d="M8.5 11.5h.01M12 11.5h.01M15.5 11.5h.01" /></svg> },
    { view: 'ziyaretchi', label: 'Ziyarətçilər', icon: <svg viewBox="0 0 24 24" {...s}><circle cx="12" cy="12" r="9" /><path d="M3 12h18M12 3a15 15 0 0 1 0 18M12 3a15 15 0 0 0 0 18" /><path d="M9 8l3-2 3 2" /></svg> },
    { view: 'audit', label: 'Audit jurnalı', adminOnly: true, icon: <svg viewBox="0 0 24 24" {...s}><path d="M9 11l3 3 8-8" /><path d="M20 12v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h9" /></svg> },
    { view: 'ayarlar', label: 'Tənzimləmələr', icon: <svg viewBox="0 0 24 24" {...s}><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.6 1.6 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.6 1.6 0 0 0-2.7 1.1V21a2 2 0 1 1-4 0v-.1A1.6 1.6 0 0 0 6.6 19l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1A1.6 1.6 0 0 0 4 12.6H4a2 2 0 1 1 0-4h.1A1.6 1.6 0 0 0 5.6 6.6L5.5 6.5a2 2 0 1 1 2.8-2.8l.1.1A1.6 1.6 0 0 0 11 4V4a2 2 0 1 1 4 0v.1a1.6 1.6 0 0 0 2.7 1.1l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.6 1.6 0 0 0 1.1 2.7H21a2 2 0 1 1 0 4h-.1a1.6 1.6 0 0 0-1.5 1z" /></svg> },
  ]},
]

async function doLogout() {
  try { await api('/logout', { method: 'POST' }) } catch { /* ignore */ }
  localStorage.clear()
  location.reload()
}

export default function Sidebar({ view, onNavigate }: { view: View; onNavigate: (v: View) => void }) {
  const { key } = useRefresh()
  const { data: orders } = useFetch<OnlineOrder[]>('/orders', [key])
  const { data: dash } = useFetch<Dashboard>('/dashboard', [key])
  const pending = orders?.length ?? 0
  const overdue = dash?.overdue_count ?? 0
  const role = currentRole()
  const name = currentName()

  return (
    <aside className="sidebar">
      <div className="logo" style={{ padding: '2px 6px 4px' }}>
        <img src="/logo-trim.png" alt="Laptops.az" style={{ height: 26, width: 'auto', display: 'block' }} />
      </div>

      <div className="navscroll">
        {NAV.map((g) => (
          <div key={g.label}>
            <div className="navlabel">{g.label}</div>
            {g.items.filter((it) => !it.adminOnly || role === 'admin').map((it) => (
              <button key={it.view} className={'nav' + (view === it.view ? ' active' : '')} onClick={() => onNavigate(it.view)}>
                {it.icon}
                {it.label}
                {it.view === 'onlayn' && pending > 0 && <span className="badge pill ink" style={{ padding: '1px 7px' }}>{pending}</span>}
                {it.view === 'kredit' && overdue > 0 && <span className="dot" />}
              </button>
            ))}
          </div>
        ))}
      </div>

      <div className="sfoot">
        <div className="sync"><span className="p" /><div><div className="t">Sayt ilə sinxron</div><div className="s">real vaxtda</div></div></div>
        <div className="me"><div className="ava">{name.charAt(0)}</div><div><div style={{ fontSize: 13, fontWeight: 700 }}>{name}</div><div className="tiny">{role === 'admin' ? 'Mağaza sahibi' : 'Satıcı'}</div></div></div>
        <button className="logout" onClick={doLogout}>
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><path d="M16 17l5-5-5-5M21 12H9" /></svg>
          Çıxış
        </button>
      </div>
    </aside>
  )
}
