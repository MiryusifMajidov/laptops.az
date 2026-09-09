import { useState } from 'react'
import { api } from '../api'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'
import { useToast } from '../lib/toast'
import { dateTimeAz } from '../lib/format'

type Att = { name: string; content: string; type: string }
type InMail = { id: number; box: string; from: string; to: string; subject: string; body: string; attach: string; received_at: string; seen: boolean }
type Compose = { to: string; subject: string; body: string; files: Att[] }

const addrOnly = (s: string) => s.match(/<(.+?)>/)?.[1] || s.trim()
const nameOf = (s: string) => (s.match(/^\s*"?([^"<]+?)"?\s*</)?.[1] || addrOnly(s) || s).trim()
const initial = (s: string) => (nameOf(s).replace(/[^A-Za-zĞÜŞİÖÇƏ0-9]/g, '')[0] || '?').toUpperCase()
const parseAttach = (s?: string): string[] => { try { return s ? JSON.parse(s) : [] } catch { return [] } }

export default function Mail() {
  const { key, bump } = useRefresh()
  const toast = useToast()
  const [box, setBox] = useState<'inbox' | 'sent'>('inbox')
  const { data: mails, loading } = useFetch<InMail[]>('/mail/inbox?box=' + box, [key, box])
  const [sel, setSel] = useState<InMail | null>(null)
  const [compose, setCompose] = useState<Compose | null>(null)
  const [sending, setSending] = useState(false)

  const list = mails ?? []

  async function open(m: InMail) {
    try { setSel(await api<InMail>('/mail/inbox/' + m.id)); if (box === 'inbox' && !m.seen) bump() } catch { setSel(m) }
  }
  function newMsg() { setCompose({ to: '', subject: '', body: '', files: [] }) }
  function reply(m: InMail) { setCompose({ to: addrOnly(m.from), subject: /^re:/i.test(m.subject) ? m.subject : 'Re: ' + m.subject, body: '', files: [] }) }
  function addFiles(fl: FileList | null) {
    if (!fl) return
    Array.from(fl).forEach((f) => {
      if (f.size > 15 * 1024 * 1024) { toast(`${f.name} çox böyükdür (max 15MB)`); return }
      const reader = new FileReader()
      reader.onload = () => setCompose((c) => (c ? { ...c, files: [...c.files, { name: f.name, content: String(reader.result), type: f.type || 'application/octet-stream' }] } : c))
      reader.readAsDataURL(f)
    })
  }
  async function send() {
    if (!compose || !compose.to.trim()) { toast('Alıcı ünvanını yazın'); return }
    setSending(true)
    try {
      await api('/mail/send', { method: 'POST', body: JSON.stringify({ to: compose.to, subject: compose.subject, body: compose.body, attachments: compose.files }) })
      toast('Göndərildi ✅'); setCompose(null); setBox('sent')
    } catch (e) { toast((e as Error).message) } finally { setSending(false) }
  }

  return (
    <div className="content">
      {/* toolbar: tablar + yeni mesaj */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <div style={{ display: 'inline-flex', background: 'var(--line2)', borderRadius: 9, padding: 3 }}>
          {(['inbox', 'sent'] as const).map((v) => (
            <span key={v} className={'chip' + (box === v ? ' on' : '')} onClick={() => { setBox(v); setSel(null) }}>
              {v === 'inbox' ? 'Gələnlər' : 'Göndərilənlər'}
            </span>
          ))}
        </div>
        <span className="tiny" style={{ color: 'var(--muted)' }}>{list.length} məktub</span>
        <button className="btn primary sm" style={{ marginLeft: 'auto' }} onClick={newMsg}>
          <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>
          Yeni mesaj
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '380px 1fr', gap: 14, alignItems: 'start' }}>
        {/* siyahı */}
        <div className="card" style={{ overflow: 'hidden', maxHeight: '72vh', overflowY: 'auto' }}>
          {loading && list.length === 0 && <div className="center-msg">Yüklənir…</div>}
          {!loading && list.length === 0 && (
            <div className="center-msg" style={{ lineHeight: 1.6 }}>
              {box === 'inbox' ? 'Gələn məktub yoxdur' : 'Göndərilmiş məktub yoxdur'}
            </div>
          )}
          {list.map((m) => {
            const who = box === 'sent' ? m.to : m.from
            const unread = box === 'inbox' && !m.seen
            return (
              <button key={m.id} onClick={() => open(m)}
                style={{ display: 'flex', gap: 11, width: '100%', textAlign: 'left', padding: '12px 14px', border: 'none', borderBottom: '1px solid var(--line3)', background: sel?.id === m.id ? 'var(--surface)' : 'transparent', cursor: 'pointer' }}>
                <div style={{ width: 34, height: 34, borderRadius: '50%', flex: 'none', background: unread ? 'var(--ink)' : 'var(--line2)', color: unread ? '#fff' : 'var(--ink2)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 700, fontSize: 13 }}>{initial(who)}</div>
                <div style={{ minWidth: 0, flex: 1 }}>
                  <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
                    <span style={{ fontSize: 13, fontWeight: unread ? 800 : 600, flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{box === 'sent' ? nameOf(who) : (nameOf(who) || '—')}</span>
                    <span className="tiny" style={{ flex: 'none' }}>{dateTimeAz(m.received_at)}</span>
                  </div>
                  <div style={{ fontSize: 12.5, fontWeight: unread ? 700 : 600, color: 'var(--ink)', marginTop: 2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {parseAttach(m.attach).length > 0 && '📎 '}{m.subject || '(mövzusuz)'}
                  </div>
                  <div className="tiny" style={{ marginTop: 2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap', fontWeight: 500 }}>{m.body || '…'}</div>
                </div>
                {unread && <span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--good)', flex: 'none', marginTop: 4 }} />}
              </button>
            )
          })}
        </div>

        {/* oxu */}
        <div className="card" style={{ padding: 0, minHeight: 360, maxHeight: '72vh', overflowY: 'auto' }}>
          {!sel && <div className="center-msg" style={{ padding: '80px 20px' }}>Oxumaq üçün soldan məktub seçin</div>}
          {sel && (
            <div>
              <div style={{ position: 'sticky', top: 0, background: 'var(--card)', borderBottom: '1px solid var(--line)', padding: '16px 20px', zIndex: 1 }}>
                <div style={{ display: 'flex', alignItems: 'flex-start', gap: 12 }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 17, fontWeight: 800, marginBottom: 6, lineHeight: 1.3 }}>{sel.subject || '(mövzusuz)'}</div>
                    <div className="tiny" style={{ color: 'var(--ink2)' }}><b>{box === 'sent' ? 'Kimə: ' : 'Kimdən: '}</b>{box === 'sent' ? sel.to : sel.from}</div>
                    <div className="tiny" style={{ marginTop: 2 }}>{dateTimeAz(sel.received_at)}</div>
                  </div>
                  {box === 'inbox' && <button className="btn ghost sm" style={{ flex: 'none' }} onClick={() => reply(sel)}>↩ Cavabla</button>}
                </div>
                {parseAttach(sel.attach).length > 0 && (
                  <div style={{ marginTop: 10, display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                    {parseAttach(sel.attach).map((n, i) => <span key={i} className="pill neut">📎 {n}</span>)}
                  </div>
                )}
              </div>
              <div style={{ padding: 20, whiteSpace: 'pre-wrap', lineHeight: 1.65, fontSize: 14, color: 'var(--ink)' }}>{sel.body || '(boş)'}</div>
            </div>
          )}
        </div>
      </div>

      {compose && (
        <>
          <div className="scrim" onClick={() => setCompose(null)} />
          <div className="modal" style={{ width: 600 }}>
            <header>
              <div><div className="h" style={{ fontSize: 16, fontWeight: 800 }}>Yeni mesaj</div><div className="tiny">info@laptops.az-dan</div></div>
              <button className="x" onClick={() => setCompose(null)}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
            </header>
            <div className="body">
              <div className="field full"><label>Kimə</label><input placeholder="ad@example.com  (vergüllə çox ünvan)" value={compose.to} onChange={(e) => setCompose({ ...compose, to: e.target.value })} /></div>
              <div className="field full"><label>Mövzu</label><input placeholder="mövzu" value={compose.subject} onChange={(e) => setCompose({ ...compose, subject: e.target.value })} /></div>
              <div className="field full"><label>Mətn</label><textarea rows={9} placeholder="mesajınızı yazın…" value={compose.body} onChange={(e) => setCompose({ ...compose, body: e.target.value })} /></div>
              <div className="field full">
                <label>Fayl əlavəsi</label>
                <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
                  {compose.files.map((f, i) => (
                    <span key={i} className="pill neut" style={{ paddingRight: 5 }}>📎 {f.name}
                      <span onClick={() => setCompose({ ...compose, files: compose.files.filter((_, x) => x !== i) })} style={{ marginLeft: 6, color: 'var(--bad)', cursor: 'pointer', fontWeight: 800 }}>×</span>
                    </span>
                  ))}
                  <label className="btn ghost sm" style={{ cursor: 'pointer' }}>+ Fayl seç<input type="file" multiple hidden onChange={(e) => { addFiles(e.target.files); e.target.value = '' }} /></label>
                </div>
              </div>
            </div>
            <footer>
              <button className="btn ghost" style={{ flex: 1 }} onClick={() => setCompose(null)}>İmtina</button>
              <button className="btn primary" style={{ flex: 1.6 }} onClick={send} disabled={sending}>{sending ? 'Göndərilir…' : 'Göndər ➤'}</button>
            </footer>
          </div>
        </>
      )}
    </div>
  )
}
