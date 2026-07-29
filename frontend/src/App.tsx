import { useEffect, useState } from 'react'
import Sidebar, { type View } from './components/Sidebar'
import Topbar from './components/Topbar'
import NewSaleDrawer from './components/NewSaleDrawer'
import { ToastProvider } from './lib/toast'
import { RefreshProvider } from './lib/refresh'

import Dashboard from './pages/Dashboard'
import Stock from './pages/Stock'
import Sales from './pages/Sales'
import Credit from './pages/Credit'
import Realizasiya from './pages/Realizasiya'
import Categories from './pages/Categories'
import Expenses from './pages/Expenses'
import Customers from './pages/Customers'
import Branches from './pages/Branches'
import OnlineOrders from './pages/OnlineOrders'
import Kassa from './pages/Kassa'
import Supplies from './pages/Supplies'
import Reports from './pages/Reports'
import Calculator from './pages/Calculator'
import Login from './pages/Login'
import Audit from './pages/Audit'
import MailExcel from './pages/MailExcel'
import Settings from './pages/Settings'
import SiteSettings from './pages/SiteSettings'
import Partners from './pages/Partners'
import AiChats from './pages/AiChats'
import Visitors from './pages/Visitors'

function todayAz(): string {
  const days = ['Bazar', 'Bazar ertəsi', 'Çərşənbə axşamı', 'Çərşənbə', 'Cümə axşamı', 'Cümə', 'Şənbə']
  const months = ['Yanvar', 'Fevral', 'Mart', 'Aprel', 'May', 'İyun', 'İyul', 'Avqust', 'Sentyabr', 'Oktyabr', 'Noyabr', 'Dekabr']
  const d = new Date()
  return `${days[d.getDay()]}, ${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}`
}

const META: Record<View, { title: string; sub: string }> = {
  dashboard: { title: 'Ana Panel', sub: todayAz() },
  satislar: { title: 'Satışlar', sub: 'Excel-i əvəz edən əsas jurnal' },
  stok: { title: 'Stok / Anbar', sub: 'bir mal = bir qeyd · stok və satış eyni bazadan' },
  onlayn: { title: 'Onlayn Sifarişlər', sub: 'saytdan gələn · 24 saat rezerv' },
  kredit: { title: 'Kredit / Borclar', sub: 'ödəniş izləməsi · gecikmə xəbərdarlığı' },
  kassa: { title: 'Kassa', sub: 'gün bağlama · nağd/kart/taksit' },
  xercler: { title: 'Xərclər', sub: 'brüt → xalis mənfəət' },
  musteriler: { title: 'Müştərilər', sub: 'profil, alış və borc keçmişi' },
  filiallar: { title: 'Filiallar', sub: 'Mərkəz · Elçin & Rəşid · Zaur' },
  realizasiya: { title: 'Realizasiya', sub: 'başqa mağazalara topdan verilən mallar' },
  techizat: { title: 'Təchizat', sub: 'gələn mal partiyaları' },
  kateqoriya: { title: 'Kateqoriyalar', sub: 'kateqoriya + xüsusiyyət idarəsi' },
  partnyor: { title: 'Tərəfdaşlıq', sub: 'partner müraciətləri · optavoy hesablar' },
  hesabatlar: { title: 'Hesabatlar', sub: 'İyul 2026 · analitika' },
  kalkulyator: { title: 'Kalkulyator', sub: 'aralıq + filial + kateqoriya → dövriyyə / net qazanc' },
  mail: { title: 'Mail & Excel', sub: 'cari stok Excel-i şablonlara göndər' },
  sayt: { title: 'Sayt Tənzimləmələri', sub: 'dillər · sayt mətnləri (çoxdilli)' },
  aisohbet: { title: 'AI söhbətləri', sub: 'sayt + app köməkçi ilə bütün yazışmalar' },
  ziyaretchi: { title: 'Ziyarətçilər', sub: 'sayta girən IP-lər · vaxt · səhifə' },
  audit: { title: 'Audit jurnalı', sub: 'kim nə etdi — yalnız admin' },
  ayarlar: { title: 'Tənzimləmələr', sub: 'profil, parol, backup' },
}

// slug URL sistemi — hər ekranın öz linki (refresh + paylaşım işləyir)
const SLUGS: Record<View, string> = {
  dashboard: '/', satislar: '/satislar', stok: '/stok', onlayn: '/onlayn-sifarisler',
  kredit: '/kredit', kassa: '/kassa', xercler: '/xercler',
  musteriler: '/musteriler', filiallar: '/filiallar', realizasiya: '/realizasiya',
  techizat: '/techizat', kateqoriya: '/kateqoriyalar', partnyor: '/terefdasliq', hesabatlar: '/hesabatlar',
  kalkulyator: '/kalkulyator', mail: '/mail', sayt: '/sayt', aisohbet: '/ai-sohbetler', ziyaretchi: '/ziyaretchiler', audit: '/audit', ayarlar: '/tenzimlemeler',
}
// Prod-da panel /admin altında verilir (Vite base '/admin/'); dev-də kökdə.
// BASE_URL = '/admin/' (prod) və ya '/' (dev). Sonundakı '/' atılır → '/admin' / ''.
const BASE = import.meta.env.BASE_URL.replace(/\/+$/, '')
function viewFromPath(): View {
  let p = window.location.pathname
  if (BASE && p.startsWith(BASE)) p = p.slice(BASE.length)
  if (p === '') p = '/'
  const found = (Object.keys(SLUGS) as View[]).find((v) => SLUGS[v] === p)
  return found ?? 'dashboard'
}

function Page({ view, search }: { view: View; search: string }) {
  switch (view) {
    case 'dashboard': return <Dashboard />
    case 'stok': return <Stock search={search} />
    case 'satislar': return <Sales />
    case 'kredit': return <Credit />
    case 'realizasiya': return <Realizasiya />
    case 'kateqoriya': return <Categories />
    case 'partnyor': return <Partners />
    case 'xercler': return <Expenses />
    case 'musteriler': return <Customers />
    case 'filiallar': return <Branches />
    case 'onlayn': return <OnlineOrders />
    case 'kassa': return <Kassa />
    case 'techizat': return <Supplies />
    case 'hesabatlar': return <Reports />
    case 'kalkulyator': return <Calculator />
    case 'mail': return <MailExcel />
    case 'sayt': return <SiteSettings />
    case 'aisohbet': return <AiChats />
    case 'ziyaretchi': return <Visitors />
    case 'audit': return <Audit />
    case 'ayarlar': return <Settings />
    default: return <Dashboard />
  }
}

export default function App() {
  const [authed, setAuthed] = useState(() => localStorage.getItem('auth') === '1')
  const [view, setViewState] = useState<View>(viewFromPath)
  const [saleOpen, setSaleOpen] = useState(false)
  const [search, setSearch] = useState('')
  const meta = META[view]

  // brauzerin geri/irəli düymələri + ilk yüklənmə
  useEffect(() => {
    const onPop = () => setViewState(viewFromPath())
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  const setView = (v: View) => {
    const target = BASE + SLUGS[v] // '/admin' + '/stok' = '/admin/stok'; kök üçün '/admin/'
    if (window.location.pathname !== target) window.history.pushState(null, '', target)
    setViewState(v)
  }

  return (
    <ToastProvider>
      {authed ? (
        <RefreshProvider>
          <div className="app">
            <Sidebar view={view} onNavigate={setView} />
            <div className="main">
              <Topbar title={meta.title} sub={meta.sub} onNewSale={() => setSaleOpen(true)}
                search={search} onSearch={setSearch} onSearchEnter={() => setView('stok')} />
              <Page view={view} search={search} />
            </div>
          </div>
          {saleOpen && <NewSaleDrawer onClose={() => setSaleOpen(false)} />}
        </RefreshProvider>
      ) : (
        <Login onLogin={() => setAuthed(true)} />
      )}
    </ToastProvider>
  )
}
