import { type Credit, type Dashboard as Dash, type Item, type Sale } from '../api'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'
import { money, dateAz, dueLabel } from '../lib/format'

const COLORS = ['#14213A', '#5A6474', '#98A0AC', '#C5CAD2']

export default function Dashboard() {
  const { key } = useRefresh()
  const { data: d } = useFetch<Dash>('/dashboard', [key])
  const { data: sales } = useFetch<Sale[]>('/sales?limit=6', [key])
  const { data: credits } = useFetch<Credit[]>('/credits', [key])
  const { data: reserved } = useFetch<Item[]>('/items?status=reserved', [key])

  // kateqoriya bölgüsü — backend-dən (bütün satışlar üzrə)
  const catRows = (d?.by_category ?? []).slice(0, 4)
  const catMax = catRows[0]?.turnover ?? 1

  const margin = d && d.turnover > 0 ? ((d.gross_profit / d.turnover) * 100).toFixed(1) : '0.0'

  return (
    <div className="content">
      <div className="row">
        <div className="card" style={{ flex: 1.55, padding: '24px 28px' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}><span style={{ width: 8, height: 8, borderRadius: '50%', background: 'var(--good)', animation: 'lp 2.2s infinite' }} /><span className="eyebrow">Ümumi · mənfəət</span></div>
            <div className="data" style={{ fontSize: 52, fontWeight: 600, letterSpacing: '-.03em', lineHeight: 1.05, marginTop: 8 }}>{money(d?.gross_profit ?? 0)}</div>
            <div className="stat-inline">
              <div><div className="l">Dövriyyə</div><div className="n">{money(d?.turnover ?? 0)}</div></div>
              <div><div className="l">Satış</div><div className="n">{d?.sales_count ?? 0}</div></div>
              <div><div className="l">Orta marja</div><div className="n" style={{ color: 'var(--good-ink)' }}>{margin}%</div></div>
            </div>
          </div>
        </div>
        <div className="card" style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', padding: '4px 24px' }}>
          <div style={{ padding: '14px 0' }}><div className="eyebrow">Stokda</div><div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginTop: 4 }}><span className="data" style={{ fontSize: 25, fontWeight: 600 }}>{d?.stock_count ?? 0}</span><span className="tiny">cihaz · dəyər {money(d?.stock_value ?? 0)}</span></div></div>
          <div className="sep" />
          <div style={{ padding: '14px 0' }}><div className="eyebrow">Açıq kredit</div><div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginTop: 4 }}><span className="data" style={{ fontSize: 25, fontWeight: 600 }}>{money(d?.open_credit ?? 0)}</span>{!!d?.overdue_count && <span className="tiny" style={{ color: 'var(--bad)', fontWeight: 700 }}>{d.overdue_count} gecikmiş</span>}</div></div>
          <div className="sep" />
          <div style={{ padding: '14px 0' }}><div className="eyebrow">Xalis mənfəət (xərc çıxılmış)</div><div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginTop: 4 }}><span className="data" style={{ fontSize: 25, fontWeight: 600 }}>{money(d?.net_profit ?? 0)}</span></div></div>
        </div>
      </div>

      <div className="row" style={{ alignItems: 'flex-start' }}>
        <div className="card" style={{ flex: 1.55, overflow: 'hidden' }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '18px 22px 12px' }}><div className="h">Son satışlar</div></div>
          <table>
            <thead><tr><th>Məhsul</th><th className="tright">Alış</th><th className="tright">Satış</th><th className="tright">Mənfəət</th><th className="tright">Kanal</th></tr></thead>
            <tbody>
              {(sales ?? []).slice(0, 6).map((s) => (
                <tr key={s.id}>
                  <td><div className="prod" style={{ fontSize: 13 }}>{s.item?.name}</div><div className="ser">{s.item?.serial || '—'}</div></td>
                  <td className="tright cost data">{money(s.item?.cost ?? 0)}</td>
                  <td className="tright data" style={{ fontWeight: 600 }}>{money(s.sale_price)}</td>
                  <td className="tright profit data">+{money(s.profit)}</td>
                  <td className="tright tiny">{s.channel}</td>
                </tr>
              ))}
              {(!sales || sales.length === 0) && <tr><td colSpan={5} className="center-msg">Hələ satış yoxdur — «Yeni satış» ilə başla</td></tr>}
            </tbody>
          </table>
        </div>

        <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div className="card pad">
            <div className="h" style={{ fontSize: 14.5, marginBottom: 14 }}>Kateqoriya üzrə (dövriyyə)</div>
            {catRows.map((c, i) => (
              <div key={c.name} style={{ marginBottom: 14 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}><span style={{ fontSize: 13, fontWeight: 700 }}>{c.name}</span><span className="data tiny" style={{ color: 'var(--ink)' }}>{money(c.turnover)}</span></div>
                <div className="bar"><span style={{ width: `${Math.round((c.turnover / catMax) * 100)}%`, background: COLORS[i] }} /></div>
              </div>
            ))}
            {catRows.length === 0 && <div className="tiny">məlumat yoxdur</div>}
          </div>
          <div className="card pad">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}><div className="h" style={{ fontSize: 14.5 }}>Kredit — vaxtı çatan</div>{!!d?.overdue_count && <span className="pill bad">{d.overdue_count}</span>}</div>
            {(credits ?? []).slice(0, 3).map((c) => {
              const l = dueLabel(c.next_due)
              return (
                <div key={c.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '9px 0', borderTop: '1px solid var(--line3)' }}>
                  <div><div style={{ fontSize: 13, fontWeight: 700 }}>{c.customer?.name}</div><div className="tiny" style={{ color: `var(--${l.tag === 'neut' ? 'muted' : l.tag})`, fontWeight: 700 }}>{l.text}</div></div>
                  <span className="data" style={{ fontWeight: 600 }}>{money(c.total - c.paid)}</span>
                </div>
              )
            })}
          </div>
        </div>
      </div>

      {!!(reserved && reserved.length) && (
        <div className="card pad">
          <div className="h" style={{ fontSize: 14.5, marginBottom: 10 }}>Rezervdə olan cihazlar (24 saat)</div>
          {reserved.map((i) => (
            <div key={i.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '9px 0', borderTop: '1px solid var(--line3)' }}>
              <span style={{ fontSize: 13, fontWeight: 700 }}>{i.name}</span><span className="tiny">{dateAz(i.created_at)} · {i.branch?.name}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
