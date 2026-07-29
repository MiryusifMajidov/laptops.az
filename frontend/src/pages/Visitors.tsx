import { useMemo } from 'react'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'
import { dateTimeAz, dateFullAz } from '../lib/format'

type Visit = { id: number; ip: string; path: string; user_agent: string; created_at: string }
type VisitsData = { total: number; unique_ips: number; today: number; visits: Visit[] }

// User-Agent → oxunaqlı brauzer/OS
function device(ua: string): { label: string; mobile: boolean } {
  const mobile = /Mobile|Android|iPhone|iPad/i.test(ua)
  let br = '—'
  if (/Edg/i.test(ua)) br = 'Edge'
  else if (/OPR|Opera/i.test(ua)) br = 'Opera'
  else if (/SamsungBrowser/i.test(ua)) br = 'Samsung'
  else if (/Chrome/i.test(ua)) br = 'Chrome'
  else if (/Firefox/i.test(ua)) br = 'Firefox'
  else if (/Safari/i.test(ua)) br = 'Safari'
  let os = ''
  if (/Android/i.test(ua)) os = 'Android'
  else if (/iPhone|iPad|iPod/i.test(ua)) os = 'iOS'
  else if (/Windows/i.test(ua)) os = 'Windows'
  else if (/Mac OS X|Macintosh/i.test(ua)) os = 'macOS'
  else if (/Linux/i.test(ua)) os = 'Linux'
  return { label: [br, os].filter(Boolean).join(' · ') || '—', mobile }
}

export default function Visitors() {
  const { key } = useRefresh()
  const { data } = useFetch<VisitsData>('/visits', [key])
  const visits = data?.visits ?? []

  // günlərə görə qruplaşdır (backend created_at desc qaytarır)
  const groups = useMemo(() => {
    const m = new Map<string, Visit[]>()
    for (const v of visits) {
      const day = (v.created_at || '').slice(0, 10)
      if (!m.has(day)) m.set(day, [])
      m.get(day)!.push(v)
    }
    return [...m.entries()]
  }, [visits])

  return (
    <div className="content">
      <div className="row">
        <div className="card kpi" style={{ flex: 1 }}><div className="eyebrow">Bu gün ziyarət</div><div className="v">{data?.today ?? 0}</div></div>
        <div className="card kpi" style={{ flex: 1 }}><div className="eyebrow">Unikal IP (nəfər)</div><div className="v" style={{ color: 'var(--good-ink)' }}>{data?.unique_ips ?? 0}</div></div>
        <div className="card kpi" style={{ flex: 1 }}><div className="eyebrow">Ümumi ziyarət</div><div className="v">{data?.total ?? 0}</div></div>
      </div>

      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><circle cx="12" cy="12" r="9" /><path d="M3 12h18M12 3a15 15 0 0 1 0 18M12 3a15 15 0 0 0 0 18" /></svg><div>Sayta girən <b>hər IP və vaxtı</b>. «Unikal IP» ≈ neçə nəfər. Yalnız real brauzerlər sayılır (JS icra edən) — botların çoxu süzülür. <b>Yalnız admin görür.</b></div></div>

      {groups.length === 0 && <div className="card"><div className="center-msg">Hələ ziyarət yoxdur</div></div>}

      {groups.map(([day, rows]) => {
        const uniq = new Set(rows.map((v) => v.ip)).size
        return (
          <div className="card" key={day} style={{ overflow: 'hidden' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '14px 18px', background: 'var(--surface)', borderBottom: '1px solid var(--line)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#5A6474" strokeWidth={2} strokeLinecap="round"><rect x="3" y="4.5" width="18" height="17" rx="2" /><path d="M3 9h18M8 2.5v4M16 2.5v4" /></svg>
                <span style={{ fontWeight: 800, fontSize: 14 }}>{dateFullAz(day)}</span>
                <span className="pill neut">{rows.length} ziyarət</span>
                <span className="pill ink">{uniq} IP</span>
              </div>
            </div>
            <table>
              <thead><tr><th style={{ width: 44, textAlign: 'center' }}>№</th><th>IP ünvan</th><th>Vaxt</th><th>Səhifə</th><th>Brauzer / cihaz</th></tr></thead>
              <tbody>
                {rows.map((v, idx) => {
                  const d = device(v.user_agent)
                  return (
                    <tr key={v.id}>
                      <td className="data" style={{ textAlign: 'center', color: 'var(--muted)', fontWeight: 700 }}>{idx + 1}</td>
                      <td className="data" style={{ fontWeight: 700 }}>{v.ip}</td>
                      <td className="tiny data">{dateTimeAz(v.created_at)}</td>
                      <td className="tiny" style={{ color: 'var(--ink2)' }}>{v.path || '/'}</td>
                      <td className="tiny">{d.mobile ? '📱 ' : '💻 '}{d.label}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )
      })}
    </div>
  )
}
