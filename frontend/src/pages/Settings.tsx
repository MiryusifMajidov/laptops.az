import { useState } from 'react'
import { api, currentName, currentRole } from '../api'
import { useToast } from '../lib/toast'

export default function Settings() {
  const toast = useToast()
  const role = currentRole()
  const [oldP, setOldP] = useState('')
  const [newP, setNewP] = useState('')
  const [busy, setBusy] = useState(false)

  async function changePass() {
    if (!oldP || newP.length < 3) { toast('Yeni parol ən az 3 simvol olmalıdır'); return }
    setBusy(true)
    try {
      await api('/change-password', { method: 'POST', body: JSON.stringify({ old: oldP, new: newP }) })
      toast('Parol dəyişdirildi')
      setOldP(''); setNewP('')
    } catch (e) { toast((e as Error).message) } finally { setBusy(false) }
  }

  async function backup() {
    try {
      const r = await api<{ file: string }>('/backup', { method: 'POST', body: '{}' })
      toast('Backup yaradıldı: ' + r.file)
    } catch (e) { toast((e as Error).message) }
  }

  return (
    <div className="content">
      <div className="row" style={{ alignItems: 'flex-start' }}>
        <div className="card pad" style={{ flex: 1 }}>
          <div className="h" style={{ marginBottom: 4 }}>Profil</div>
          <div className="tiny" style={{ marginBottom: 16 }}>cari istifadəçi</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div className="ava" style={{ width: 42, height: 42, fontSize: 16 }}>{currentName().charAt(0)}</div>
            <div><div style={{ fontWeight: 700 }}>{currentName()}</div><div className="tiny">{role === 'admin' ? 'Admin — Mağaza sahibi' : 'Satıcı'}</div></div>
          </div>
          <div className="sep" style={{ margin: '18px 0' }} />
          <div className="h" style={{ fontSize: 14, marginBottom: 12 }}>Parolu dəyiş</div>
          <div className="field" style={{ marginBottom: 10 }}><label>Köhnə parol</label><input type="password" value={oldP} onChange={(e) => setOldP(e.target.value)} /></div>
          <div className="field" style={{ marginBottom: 14 }}><label>Yeni parol</label><input type="password" value={newP} onChange={(e) => setNewP(e.target.value)} /></div>
          <button className="btn primary" onClick={changePass} disabled={busy}>{busy ? 'Saxlanır…' : 'Parolu dəyiş'}</button>
        </div>

        <div className="card pad" style={{ flex: 1 }}>
          <div className="h" style={{ marginBottom: 4 }}>Məlumat təhlükəsizliyi</div>
          <div className="tiny" style={{ marginBottom: 16 }}>bazanın ehtiyat nüsxəsi</div>
          <div className="banner" style={{ marginBottom: 16 }}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M21 12a9 9 0 1 1-3-6.7L21 8" /><path d="M21 3v5h-5" /></svg><div>Bütün datanın təmiz nüsxəsi <b>backups/</b> qovluğunda saxlanılır. İnternetə çıxanda PostgreSQL + avtomatik gündəlik backup qoşulacaq.</div></div>
          <button className="btn ghost" onClick={backup}>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M12 3v12M7 10l5 5 5-5M4 21h16" /></svg>İndi backup al
          </button>
        </div>
      </div>
    </div>
  )
}
