import { useEffect, useState } from 'react'
import { api, type OnlineOrder } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money, remainLabel } from '../lib/format'
import NewSaleDrawer from '../components/NewSaleDrawer'

export default function OnlineOrders() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const [tick, setTick] = useState(0)
  const [sellOrder, setSellOrder] = useState<OnlineOrder | null>(null)
  useEffect(() => { const t = setInterval(() => setTick((x) => x + 1), 15000); return () => clearInterval(t) }, [])
  const { data } = useFetch<OnlineOrder[]>('/orders', [key, tick])
  const list = data ?? []

  async function cancel(id: number) {
    try {
      await api(`/orders/${id}/cancel`, { method: 'POST' })
      bump()
      toast('Sifariş ləğv edildi')
    } catch (e) {
      toast((e as Error).message)
    }
  }

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><circle cx="9" cy="21" r="1.5" /><circle cx="18" cy="21" r="1.5" /><path d="M2 3h3l2.5 13h11l2-9H6" /></svg><div>Saytdan/appdan gələn sifarişlər. <b>Onlayn ödəniş yoxdur</b> — mağaza zəng edib təsdiqləyir, cihaz <b>24 saat</b> rezerv olunur. Siyahı avtomatik yenilənir.</div></div>

      <div className="toolbar">
        <div className="h" style={{ marginRight: 'auto' }}>Gözləyən sifarişlər · {list.length}</div>
        <button className="btn ghost sm" onClick={() => setTick((x) => x + 1)}><svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M21 12a9 9 0 1 1-3-6.7L21 8" /><path d="M21 3v5h-5" /></svg>Yenilə</button>
      </div>

      <div className="row" style={{ flexWrap: 'wrap' }}>
        {list.map((o) => {
          const rl = remainLabel(o.expires_at)
          return (
            <div key={o.id} className="card pad" style={{ flex: 1, minWidth: 300 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
                <div><div className="prod">{o.customer_name}</div><div className="tiny">{o.phone}</div></div>
                <span className={'pill ' + rl.tag}>⏱ {rl.text}</span>
              </div>
              <div className="sep" style={{ margin: '10px 0' }} />
              <div style={{ fontWeight: 700, fontSize: 13.5 }}>{o.item?.name}</div>
              <div className="ser">{o.item?.serial || '—'} · {money(o.item?.cost ?? 0)}</div>
              <div style={{ display: 'flex', gap: 8, marginTop: 14 }}>
                <button className="btn primary sm" onClick={() => setSellOrder(o)} disabled={!o.item}>Satışa çevir</button>
                <a className="btn ghost sm" href={`tel:${o.phone.replace(/[^\d+]/g, '')}`}>Zəng et</a>
                <button className="btn ghost sm" style={{ color: 'var(--bad)' }} onClick={() => cancel(o.id)}>Ləğv et</button>
              </div>
            </div>
          )
        })}
        {list.length === 0 && <div className="card pad" style={{ flex: 1 }}><div className="center-msg">Yeni onlayn sifariş yoxdur</div></div>}
      </div>

      {sellOrder && (
        <NewSaleDrawer
          prefill={sellOrder.item}
          orderId={sellOrder.id}
          orderInfo={{ name: sellOrder.customer_name, phone: sellOrder.phone }}
          onClose={() => setSellOrder(null)}
        />
      )}
    </div>
  )
}
