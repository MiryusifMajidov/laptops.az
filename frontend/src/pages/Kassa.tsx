import { api, type KassaData } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money, dateAz } from '../lib/format'

export default function Kassa() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data } = useFetch<KassaData>('/kassa', [key])
  const t = data?.today

  const tiles: [string, number][] = [
    ['Nağd', t?.cash ?? 0], ['Kart · POS', t?.card ?? 0],
    ['Taksit', t?.installment ?? 0], ['Kredit (nisyə)', t?.credit ?? 0],
  ]

  async function close() {
    try {
      await api('/kassa/close', { method: 'POST' })
      bump()
      toast('Gün bağlandı · kassa hesabatı saxlandı')
    } catch (e) {
      toast((e as Error).message)
    }
  }

  return (
    <div className="content">
      <div className="row" style={{ alignItems: 'flex-start' }}>
        <div className="card pad" style={{ flex: 1.2 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}><div className="h">Bu günün kassası</div><span className="tiny">açıq</span></div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginTop: 16 }}>
            {tiles.map(([l, v]) => (
              <div key={l} className="card pad" style={{ boxShadow: 'none' }}><div className="eyebrow">{l}</div><div className="data" style={{ fontSize: 22, fontWeight: 600, marginTop: 4 }}>{money(v)}</div></div>
            ))}
          </div>
          <div className="sep" style={{ margin: '18px 0' }} />
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div><div className="eyebrow">Kassada gözlənilən nağd</div><div className="data" style={{ fontSize: 28, fontWeight: 600, marginTop: 4 }}>{money(t?.cash ?? 0)}</div></div>
            <button className="btn primary" onClick={close}>Günü bağla</button>
          </div>
        </div>
        <div className="card pad" style={{ flex: 1 }}>
          <div className="h">Son bağlanışlar</div>
          <div style={{ marginTop: 8 }}>
            {(data?.closes ?? []).map((c) => (
              <div key={c.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '11px 0', borderTop: '1px solid var(--line3)' }}><span style={{ fontSize: 13, fontWeight: 600 }}>{dateAz(c.date)}</span><span className="data" style={{ fontWeight: 600 }}>{money(c.total)}</span></div>
            ))}
            {(!data?.closes || data.closes.length === 0) && <div className="tiny" style={{ marginTop: 10 }}>Hələ bağlanış yoxdur</div>}
          </div>
        </div>
      </div>
    </div>
  )
}
