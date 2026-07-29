import { useState } from 'react'
import { api } from '../api'

export default function Login({ onLogin }: { onLogin: () => void }) {
  const [u, setU] = useState('')
  const [p, setP] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    setErr('')
    try {
      const res = await api<{ token: string; role: string; name: string }>('/login', { method: 'POST', body: JSON.stringify({ username: u, password: p }) })
      localStorage.setItem('auth', '1')
      localStorage.setItem('token', res.token)
      localStorage.setItem('role', res.role)
      localStorage.setItem('name', res.name)
      onLogin()
    } catch (ex) {
      setErr((ex as Error).message)
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="login-wrap">
      <form className="login-card" onSubmit={submit}>
        <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 6 }}>
          <img src="/logo-trim.png" alt="Laptops.az" style={{ height: 34, width: 'auto' }} />
        </div>
        <div className="tiny" style={{ textAlign: 'center', marginBottom: 18 }}>Admin panelə giriş</div>

        <div className="login-field">
          <label>İstifadəçi adı</label>
          <input value={u} onChange={(e) => setU(e.target.value)} placeholder="admin" autoFocus autoComplete="username" />
        </div>
        <div className="login-field">
          <label>Parol</label>
          <input type="password" value={p} onChange={(e) => setP(e.target.value)} placeholder="••••••" autoComplete="current-password" />
        </div>

        {err && <div className="login-err">{err}</div>}

        <button className="btn primary" type="submit" disabled={busy} style={{ width: '100%', justifyContent: 'center', marginTop: 18 }}>
          {busy ? 'Yoxlanılır…' : 'Daxil ol'}
        </button>
      </form>
    </div>
  )
}
