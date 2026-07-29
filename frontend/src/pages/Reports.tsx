import { type Reports as R } from '../api'
import { useFetch } from '../lib/hooks'
import { money } from '../lib/format'

const COLORS = ['#14213A', '#5A6474', '#98A0AC', '#C5CAD2', '#D8DCE2']

export default function Reports() {
  const { data } = useFetch<R>('/reports', [])
  const top = data?.top_models ?? []
  const perf = data?.branch_perf ?? []
  const dead = data?.dead_stock ?? []
  const topMax = top[0]?.count ?? 1
  const perfMax = Math.max(...perf.map((p) => p.profit), 1)
  const deadValue = dead.reduce((a, d) => a + d.cost, 0)

  return (
    <div className="content">
      <div className="grid4">
        <div className="card kpi"><div className="eyebrow">Ən çox satılan</div><div className="v" style={{ fontSize: 16 }}>{top[0]?.name?.slice(0, 18) ?? '—'}</div></div>
        <div className="card kpi"><div className="eyebrow">Model sayı (satış)</div><div className="v">{top.length}</div></div>
        <div className="card kpi"><div className="eyebrow" style={{ color: 'var(--warn)' }}>Ölü stok</div><div className="v" style={{ color: 'var(--warn)' }}>{dead.length}</div></div>
        <div className="card kpi"><div className="eyebrow" style={{ color: 'var(--warn)' }}>Bağlı kapital</div><div className="v" style={{ color: 'var(--warn)' }}>{money(deadValue)}</div></div>
      </div>

      <div className="two">
        <div className="card pad">
          <div className="h" style={{ marginBottom: 14 }}>Ən çox satılan modellər</div>
          {top.map((m, i) => (
            <div key={m.name} style={{ marginBottom: 14 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}><span style={{ fontSize: 13, fontWeight: 700 }}>{m.name}</span><span className="data tiny" style={{ color: 'var(--ink)' }}>{m.count} ədəd</span></div>
              <div className="bar"><span style={{ width: `${Math.round((m.count / topMax) * 100)}%`, background: COLORS[i % COLORS.length] }} /></div>
            </div>
          ))}
          {top.length === 0 && <div className="tiny">məlumat yoxdur</div>}
        </div>

        <div className="card pad">
          <div className="h" style={{ marginBottom: 6 }}>Ölü stok · 90+ gündür satılmayan</div>
          <div className="tiny" style={{ marginBottom: 12 }}>bağlı kapital — endirim və ya filiala transfer düşün</div>
          {dead.map((d) => (
            <div key={d.name} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '10px 0', borderTop: '1px solid var(--line3)' }}>
              <div><div style={{ fontSize: 13, fontWeight: 700 }}>{d.name}</div><div className="tiny" style={{ color: 'var(--warn)', fontWeight: 700 }}>{d.age_days} gün</div></div>
              <span className="data cost">{money(d.cost)}</span>
            </div>
          ))}
          {dead.length === 0 && <div className="tiny">ölü stok yoxdur — əla!</div>}
        </div>
      </div>

      <div className="card pad">
        <div className="h" style={{ marginBottom: 14 }}>Filial üzrə mənfəət</div>
        {perf.map((b, i) => (
          <div key={b.name} style={{ marginBottom: 14 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 6 }}><span style={{ fontSize: 13, fontWeight: 700 }}>{b.name}</span><span className="data tiny profit">+{money(b.profit)}</span></div>
            <div className="bar"><span style={{ width: `${Math.round((b.profit / perfMax) * 100)}%`, background: COLORS[i % COLORS.length] }} /></div>
          </div>
        ))}
        {perf.length === 0 && <div className="tiny">məlumat yoxdur</div>}
      </div>
    </div>
  )
}
