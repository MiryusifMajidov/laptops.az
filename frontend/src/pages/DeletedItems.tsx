import { api, type Item } from '../api'
import { useFetch } from '../lib/hooks'
import { useRefresh } from '../lib/refresh'
import { useToast } from '../lib/toast'
import { money, dateFullAz } from '../lib/format'

// Silinmiş məhsullar — soft-delete olunmuş cihazlar. Baxmaq + bərpa etmək.
export default function DeletedItems() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: items, loading } = useFetch<Item[]>('/items/deleted', [key])
  const list = items ?? []

  async function restore(i: Item) {
    if (!window.confirm(`«${i.name}» bərpa olunsun? Yenidən stoka qayıdacaq.`)) return
    try {
      await api(`/items/${i.id}/restore`, { method: 'POST' })
      bump()
      toast('Məhsul bərpa olundu')
    } catch (e) {
      toast((e as Error).message)
    }
  }

  return (
    <div className="content">
      <div className="banner">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M3 6h18M8 6V4h8v2M6 6l1 14h10l1-14" /></svg>
        <div><b>Silinmiş məhsullar burada saxlanılır.</b> Stokdan silinən hər cihaz bu siyahıya düşür — istənilən vaxt <b>Bərpa et</b> ilə geri qaytara bilərsiniz.</div>
      </div>

      {loading && <div className="card"><div className="center-msg">Yüklənir…</div></div>}
      {!loading && list.length === 0 && <div className="card"><div className="center-msg">Silinmiş məhsul yoxdur</div></div>}

      {list.length > 0 && (
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '14px 18px', background: 'var(--surface)', borderBottom: '1px solid var(--line)' }}>
            <span style={{ fontWeight: 800, fontSize: 14 }}>Silinmiş cihazlar</span>
            <span className="pill neut">{list.length}</span>
          </div>
          <table>
            <thead><tr><th>Məhsul</th><th>Kateqoriya</th><th>Filial</th><th className="tright">Alış</th><th className="tright">Satış</th><th>Silinmə tarixi</th><th className="tright">Əməliyyat</th></tr></thead>
            <tbody>
              {list.map((i) => (
                <tr key={i.id}>
                  <td>
                    <div className="prod">{i.name}</div>
                    <div className="ser">{i.serial || '—'}</div>
                    <div className="attrchips">{i.values?.map((v) => <span key={v.id} className="attr">{v.value}</span>)}</div>
                  </td>
                  <td>{i.category?.name ?? '—'}</td>
                  <td>{i.branch?.name ?? '—'}</td>
                  <td className="tright cost data">{money(i.cost)}</td>
                  <td className="tright data" style={{ fontWeight: 600 }}>{i.price ? money(i.price) : '—'}</td>
                  <td className="tiny">{i.deleted_at ? dateFullAz(i.deleted_at.slice(0, 10)) : '—'}</td>
                  <td className="tright"><button className="btn ghost sm" onClick={() => restore(i)}>↺ Bərpa et</button></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
