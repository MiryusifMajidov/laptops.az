import { useState } from 'react'
import { api, type PartnerApplication, type AppUser } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { dateTimeAz } from '../lib/format'

export default function Partners() {
  const toast = useToast()
  const [tick, setTick] = useState(0)
  const { data: apps } = useFetch<PartnerApplication[]>('/partner-applications', [tick])
  const { data: users } = useFetch<AppUser[]>('/users', [tick])
  const refresh = () => setTick((t) => t + 1)

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)

  async function setStatus(id: number, status: string) {
    try { await api(`/partner-applications/${id}`, { method: 'PUT', body: JSON.stringify({ status }) }); refresh() }
    catch (e) { toast((e as Error).message) }
  }
  async function createPartner() {
    if (!username.trim() || password.length < 4) { toast('İstifadəçi adı və ən az 4 simvol parol vacibdir'); return }
    setBusy(true)
    try {
      await api('/users', { method: 'POST', body: JSON.stringify({ username: username.trim(), password, name: name.trim(), role: 'partner' }) })
      toast('Partner yaradıldı: ' + username.trim())
      setUsername(''); setPassword(''); setName(''); refresh()
    } catch (e) { toast((e as Error).message) } finally { setBusy(false) }
  }
  async function delUser(u: AppUser) {
    if (!window.confirm(`«${u.username}» partner hesabı silinsin?`)) return
    try { await api(`/users/${u.id}`, { method: 'DELETE' }); refresh() }
    catch (e) { toast((e as Error).message) }
  }
  function prefill(a: PartnerApplication) {
    setName(a.name)
    setUsername((a.store_name || a.name).toLowerCase().replace(/[^a-z0-9]+/g, '').slice(0, 16))
  }

  const partners = (users ?? []).filter((u) => u.role === 'partner')

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" /></svg><div>Mobil tətbiqdən gələn <b>tərəfdaşlıq müraciətləri</b>. Müştəri ilə əlaqə saxlayıb, ona <b>partner hesabı</b> yaradın — o hesabla tətbiqdə <b>optavoy qiymətləri</b> görəcək.</div></div>

      <div className="toolbar"><div className="h" style={{ marginRight: 'auto' }}>Müraciətlər · {apps?.length ?? 0}</div></div>
      <div className="card" style={{ overflow: 'hidden' }}>
        <table>
          <thead><tr><th>Tarix</th><th>Ad Soyad</th><th>Mağaza</th><th>Telefon</th><th>Status</th><th></th></tr></thead>
          <tbody>
            {(apps ?? []).map((a) => (
              <tr key={a.id}>
                <td className="tiny data">{dateTimeAz(a.created_at)}</td>
                <td className="prod">{a.name}</td>
                <td>{a.store_name || '—'}</td>
                <td><a className="sg" href={`tel:${a.phone.replace(/[^\d+]/g, '')}`}>{a.phone}</a></td>
                <td>
                  <select className="select" value={a.status} onChange={(e) => setStatus(a.id, e.target.value)}>
                    <option value="pending">Yeni</option>
                    <option value="contacted">Əlaqə saxlanıldı</option>
                    <option value="done">Tamamlandı</option>
                  </select>
                </td>
                <td><button className="btn ghost sm" onClick={() => prefill(a)}>Partner yarat →</button></td>
              </tr>
            ))}
            {(apps?.length ?? 0) === 0 && <tr><td colSpan={6} className="center-msg">Hələ müraciət yoxdur</td></tr>}
          </tbody>
        </table>
      </div>

      <div className="row" style={{ alignItems: 'flex-start', marginTop: 16 }}>
        <div className="card pad" style={{ flex: 1, minWidth: 300 }}>
          <div className="h" style={{ fontSize: 14, marginBottom: 4 }}>Yeni partner hesabı</div>
          <div className="tiny" style={{ marginBottom: 14 }}>Bu hesabla partner mobil tətbiqə daxil olub optavoy qiymətləri görəcək.</div>
          <div className="field" style={{ marginBottom: 10 }}><label>Mağaza / ad</label><input value={name} onChange={(e) => setName(e.target.value)} placeholder="məs. Salman Computers" /></div>
          <div className="field" style={{ marginBottom: 10 }}><label>İstifadəçi adı (login)</label><input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="salman" /></div>
          <div className="field" style={{ marginBottom: 14 }}><label>Parol</label><input value={password} onChange={(e) => setPassword(e.target.value)} placeholder="ən az 4 simvol" /></div>
          <button className="btn primary" onClick={createPartner} disabled={busy}>{busy ? 'Yaradılır…' : 'Partner yarat'}</button>
        </div>

        <div className="card" style={{ flex: 1.2, minWidth: 300, overflow: 'hidden' }}>
          <div style={{ padding: '12px 16px', borderBottom: '1px solid var(--line)', fontWeight: 800, fontSize: 13 }}>Partner hesabları · {partners.length}</div>
          <table>
            <thead><tr><th>Ad</th><th>Login</th><th></th></tr></thead>
            <tbody>
              {partners.map((u) => (
                <tr key={u.id}>
                  <td className="prod">{u.name}</td>
                  <td className="data">{u.username}</td>
                  <td><button className="btn ghost sm" style={{ color: 'var(--bad)' }} onClick={() => delUser(u)}>Sil</button></td>
                </tr>
              ))}
              {partners.length === 0 && <tr><td colSpan={3} className="center-msg">Hələ partner yoxdur</td></tr>}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
