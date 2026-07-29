import { type AuditLog } from '../api'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'
import { dateTimeAz } from '../lib/format'

const ACT: Record<string, [string, string]> = {
  POST: ['Yaratdı', 'good'], PUT: ['Yenilədi', 'warn'], DELETE: ['Sildi', 'bad'],
}
// yol → oxunaqlı ad (+ varsa obyekt ID-si)
function human(path: string): string {
  const segs = path.replace('/api/', '').split('/')
  const map: Record<string, string> = {
    items: 'Məhsul', sales: 'Satış', customers: 'Müştəri', expenses: 'Xərc',
    categories: 'Kateqoriya', attributes: 'Xüsusiyyət', branches: 'Filial',
    transfers: 'Transfer', supplies: 'Təchizat', consignments: 'Realizasiya',
    credits: 'Kredit', orders: 'Onlayn sifariş', kassa: 'Kassa', backup: 'Backup',
    'email-templates': 'Mail şablonu', email: 'Mail göndərmə', upload: 'Şəkil',
    languages: 'Dil', 'ui-strings': 'Sayt mətni', terms: 'Termin tərcümə',
    translate: 'AI tərcümə', 'change-password': 'Parol', 'attribute-options': 'Seçim',
  }
  const base = map[segs[0]] || segs[0]
  const id = segs.find((s, i) => i > 0 && /^\d+$/.test(s))
  return id ? `${base} #${id}` : base
}

export default function Audit() {
  const { key } = useRefresh()
  const { data } = useFetch<AuditLog[]>('/audit', [key])
  const list = data ?? []

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M9 11l3 3 8-8" /><path d="M20 12v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h9" /></svg><div>Bu jurnalı <b>yalnız admin görür</b> (satıcı görə bilməz). Kim, nə vaxt, nəyi yaratdı/yenilədi/sildi — hamısı burada qeydə alınır.</div></div>

      <div className="card" style={{ overflow: 'hidden' }}>
        <table>
          <thead><tr><th>Tarix / saat</th><th>İstifadəçi</th><th>Rol</th><th>Əməliyyat</th><th>Obyekt</th><th>Detal</th></tr></thead>
          <tbody>
            {list.map((l) => {
              const a = ACT[l.action] ?? [l.action, 'neut']
              return (
                <tr key={l.id}>
                  <td className="tiny data">{dateTimeAz(l.created_at)}</td>
                  <td className="prod">{l.username}</td>
                  <td><span className={'pill ' + (l.role === 'admin' ? 'ink' : 'neut')}>{l.role === 'admin' ? 'Admin' : 'Satıcı'}</span></td>
                  <td><span className={'pill ' + a[1]}>{a[0]}</span></td>
                  <td style={{ whiteSpace: 'nowrap' }}>{human(l.path)}</td>
                  <td className="tiny" style={{ color: 'var(--ink2)', maxWidth: 420, lineHeight: 1.5 }}>{l.detail || '—'}</td>
                </tr>
              )
            })}
            {list.length === 0 && <tr><td colSpan={6} className="center-msg">Hələ qeyd yoxdur</td></tr>}
          </tbody>
        </table>
      </div>
    </div>
  )
}
