import { useState } from 'react'
import { api } from '../api'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'
import { dateTimeAz } from '../lib/format'

type ChatRow = { id: string; source: string; lang: string; count: number; created_at: string; updated_at: string; preview: string }
type Msg = { role: string; text: string }
type Detail = { id: string; source: string; lang: string; ip: string; created_at: string; updated_at: string; messages: Msg[] }

const SRC: Record<string, [string, string]> = { web: ['Sayt', 'ink'], app: ['Tətbiq', 'good'] }
const clean = (t: string) => t.replace(/\[\[product:\d+\]\]/g, '').trim()

export default function AiChats() {
  const { key } = useRefresh()
  const { data } = useFetch<ChatRow[]>('/ai-chats', [key])
  const [sel, setSel] = useState<string | null>(null)
  const [detail, setDetail] = useState<Detail | null>(null)
  const [loading, setLoading] = useState(false)
  const [filter, setFilter] = useState<'all' | 'web' | 'app'>('all')

  const list = (data ?? []).filter((c) => filter === 'all' || c.source === filter)

  async function open(id: string) {
    setSel(id); setDetail(null); setLoading(true)
    try { setDetail(await api<Detail>('/ai-chats/' + id)) } catch { /* ignore */ } finally { setLoading(false) }
  }

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round"><path d="M21 11.5a8.38 8.38 0 0 1-8.5 8.5 8.5 8.5 0 0 1-3.8-.9L3 21l1.9-5.7a8.5 8.5 0 0 1-.9-3.8A8.38 8.38 0 0 1 12.5 3 8.38 8.38 0 0 1 21 11.5z" /></svg><div>Sayt və tətbiqdəki AI köməkçi ilə <b>bütün müştəri yazışmaları</b>. Söhbətə klik → tam yazışma. <b>Yalnız admin görür.</b></div></div>

      <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
        {(['all', 'web', 'app'] as const).map((f) => (
          <span key={f} className={'chip' + (filter === f ? ' on' : '')} onClick={() => setFilter(f)}>
            {f === 'all' ? 'Hamısı' : SRC[f][0]}
          </span>
        ))}
        <span className="tiny" style={{ marginLeft: 'auto', alignSelf: 'center', color: 'var(--muted)' }}>{list.length} söhbət</span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '340px 1fr', gap: 14, alignItems: 'start' }}>
        {/* siyahı */}
        <div className="card" style={{ overflow: 'hidden', maxHeight: '72vh', overflowY: 'auto' }}>
          {list.length === 0 && <div className="center-msg" style={{ padding: 30 }}>Hələ söhbət yoxdur</div>}
          {list.map((c) => (
            <button key={c.id} onClick={() => open(c.id)}
              style={{ display: 'block', width: '100%', textAlign: 'left', padding: '11px 13px', border: 'none', borderBottom: '1px solid var(--line)', background: sel === c.id ? 'var(--card2, #f2f1ec)' : 'transparent', cursor: 'pointer' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 7, marginBottom: 4 }}>
                <span className={'pill ' + (SRC[c.source]?.[1] ?? 'neut')} style={{ fontSize: 10.5 }}>{SRC[c.source]?.[0] ?? c.source}</span>
                <span className="tiny" style={{ color: 'var(--muted)' }}>{c.lang?.toUpperCase()}</span>
                <span className="tiny" style={{ marginLeft: 'auto', color: 'var(--muted)' }}>{c.count} mesaj</span>
              </div>
              <div style={{ fontSize: 13, fontWeight: 600, lineHeight: 1.4, display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical', overflow: 'hidden' }}>{clean(c.preview) || '—'}</div>
              <div className="tiny data" style={{ color: 'var(--muted)', marginTop: 3 }}>{dateTimeAz(c.updated_at)}</div>
            </button>
          ))}
        </div>

        {/* yazışma */}
        <div className="card" style={{ padding: 0, minHeight: 300, maxHeight: '72vh', overflowY: 'auto' }}>
          {!sel && <div className="center-msg" style={{ padding: 60 }}>Yazışmanı görmək üçün soldan söhbət seçin</div>}
          {sel && loading && <div className="center-msg" style={{ padding: 60 }}>Yüklənir…</div>}
          {detail && (
            <div>
              <div style={{ position: 'sticky', top: 0, background: 'var(--bg, #fff)', borderBottom: '1px solid var(--line)', padding: '11px 16px', display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap' }}>
                <span className={'pill ' + (SRC[detail.source]?.[1] ?? 'neut')}>{SRC[detail.source]?.[0] ?? detail.source}</span>
                <span className="tiny">Dil: <b>{detail.lang?.toUpperCase() || '—'}</b></span>
                <span className="tiny">IP: <b className="data">{detail.ip || '—'}</b></span>
                <span className="tiny" style={{ marginLeft: 'auto', color: 'var(--muted)' }}>{dateTimeAz(detail.created_at)}</span>
              </div>
              <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 10 }}>
                {detail.messages.map((m, i) => {
                  const mine = m.role === 'user'
                  return (
                    <div key={i} style={{ alignSelf: mine ? 'flex-end' : 'flex-start', maxWidth: '78%' }}>
                      <div className="tiny" style={{ color: 'var(--muted)', marginBottom: 3, textAlign: mine ? 'right' : 'left' }}>{mine ? 'Müştəri' : 'AI köməkçi'}</div>
                      <div style={{ padding: '9px 13px', borderRadius: 14, whiteSpace: 'pre-wrap', lineHeight: 1.5, fontSize: 13.5,
                        background: mine ? 'var(--ink, #14213A)' : 'var(--card2, #f2f1ec)',
                        color: mine ? '#fff' : 'var(--ink)',
                        borderTopRightRadius: mine ? 4 : 14, borderTopLeftRadius: mine ? 14 : 4 }}>
                        {clean(m.text) || '…'}
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
