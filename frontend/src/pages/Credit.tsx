import { useState } from 'react'
import { api, type Credit as C } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money, dateAz, dueLabel, dateTimeAz } from '../lib/format'
import FormModal from '../components/FormModal'

export default function Credit() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: credits } = useFetch<C[]>('/credits', [key])
  const [form, setForm] = useState(false)
  const [pay, setPay] = useState<C | null>(null)
  const [detail, setDetail] = useState<C | null>(null)

  const list = credits ?? []
  const open = list.reduce((a, c) => a + Math.max(0, c.total - c.paid), 0)
  const overdue = list.filter((c) => c.status === 'overdue').length

  // "Biz bağladıq" — seçilsə satış Satışlar/hesabata düşür (borc yenə qalır)
  async function toggleClosed(c: C) {
    try { await api(`/credits/${c.id}`, { method: 'PUT', body: JSON.stringify({ closed: !c.closed }) }); bump() }
    catch (e) { toast((e as Error).message) }
  }

  return (
    <div className="content">
      <div className="grid4">
        <div className="card kpi"><div className="eyebrow">Açıq borc (qalıq)</div><div className="v">{money(open)}</div></div>
        <div className="card kpi"><div className="eyebrow" style={{ color: 'var(--bad)' }}>Gecikmiş</div><div className="v" style={{ color: 'var(--bad)' }}>{overdue}</div></div>
        <div className="card kpi"><div className="eyebrow">Aktiv müqavilə</div><div className="v">{list.filter((c) => c.status !== 'paid').length}</div></div>
        <div className="card kpi"><div className="eyebrow">Ödənilib (cəmi)</div><div className="v" style={{ color: 'var(--good-ink)' }}>{money(list.reduce((a, c) => a + c.paid, 0))}</div></div>
      </div>

      <div className="toolbar">
        <div className="h" style={{ marginRight: 'auto' }}>Kredit / taksit müqavilələri</div>
        <button className="btn primary sm" onClick={() => setForm(true)}><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth={2.6} strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>Yeni kredit</button>
      </div>

      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M9 12l2 2 4-4" /><circle cx="12" cy="12" r="9" /></svg><div><b>«Biz bağladıq» seçimi.</b> Kreditə verilən mal əlimizə pul gəlmədiyi üçün satılanlara düşmür. Borc tam ödənəndə <b>avtomatik</b> düşür. Bunu özün bağlasan (məsələn ödənişini biz özümüz örtdük), mal da satılanlara/hesabata düşür — <b>borc yenə qalır</b>, ödənişi müştəridən adi qaydada alırsan.</div></div>

      <div className="card" style={{ overflow: 'hidden' }}>
        <table>
          <thead><tr><th>Müştəri</th><th>Telefon</th><th>Məhsul</th><th className="tright">Ümumi</th><th className="tright">Ödənilib</th><th className="tright">Qalıq</th><th>Növbəti ödəniş</th><th style={{ textAlign: 'center' }}>Biz bağladıq</th><th></th></tr></thead>
          <tbody>
            {list.map((c) => {
              const paidOff = c.paid >= c.total
              const l = dueLabel(c.next_due)
              return (
                <tr key={c.id}>
                  <td className="prod">{c.customer?.name}</td>
                  <td className="tiny">{c.customer?.phone}</td>
                  <td>{c.item_name}</td>
                  <td className="tright data cost">{money(c.total)}</td>
                  <td className="tright data" style={{ color: 'var(--good-ink)' }}>{money(c.paid)}</td>
                  <td className="tright data" style={{ fontWeight: 700 }}>{money(Math.max(0, c.total - c.paid))}</td>
                  <td>{paidOff ? <span className="pill good">Ödənilib</span> : <span className={'pill ' + l.tag}>{dateAz(c.next_due)} · {l.text}</span>}</td>
                  <td style={{ textAlign: 'center' }} title={paidOff ? 'Tam ödənilib — onsuz da satılanlarda' : 'Satılanlara/hesabata düşsün?'}>
                    <div className={'tg' + (paidOff || c.closed ? ' on' : '')} style={paidOff ? { opacity: .5, pointerEvents: 'none' } : { margin: '0 auto' }} onClick={() => !paidOff && toggleClosed(c)} />
                  </td>
                  <td className="tright"><button className="btn ghost sm" onClick={() => setDetail(c)}>Tarixçə</button>{!paidOff && <button className="btn ghost sm" style={{ marginLeft: 6 }} onClick={() => setPay(c)}>Ödəniş</button>}</td>
                </tr>
              )
            })}
            {list.length === 0 && <tr><td colSpan={9} className="center-msg">Kredit müqaviləsi yoxdur</td></tr>}
          </tbody>
        </table>
      </div>

      {form && (
        <FormModal
          title="Yeni kredit / taksit" submitLabel="Yarat" onClose={() => setForm(false)}
          fields={[
            { name: 'customer_id', label: 'Müştəri', type: 'customer', required: true, full: true, placeholder: 'müştəri axtar və ya yeni yarat…' },
            { name: 'item_name', label: 'Məhsul', full: true, placeholder: 'məs. Lenovo Legion 5' },
            { name: 'total', label: 'Ümumi məbləğ (₼)', type: 'number', required: true },
            { name: 'paid', label: 'İlkin ödəniş (₼)', type: 'number' },
            { name: 'next_due', label: 'Növbəti ödəniş tarixi', type: 'date' },
          ]}
          onSubmit={async (v) => { await api('/credits', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Kredit müqaviləsi yaradıldı'); setForm(false) }}
        />
      )}

      {pay && (
        <FormModal
          title={`Ödəniş — ${pay.customer?.name}`} subtitle={`qalıq: ${money(pay.total - pay.paid)}`} submitLabel="Ödənişi qeyd et" onClose={() => setPay(null)}
          fields={[
            { name: 'amount', label: 'Ödəniş məbləği (₼)', type: 'number', required: true, full: true },
            { name: 'next_due', label: 'Növbəti ödəniş tarixi', type: 'date', full: true },
          ]}
          onSubmit={async (v) => { await api(`/credits/${pay.id}/pay`, { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Ödəniş qeyd edildi'); setPay(null) }}
        />
      )}

      {detail && (
        <>
          <div className="scrim" onClick={() => setDetail(null)} />
          <div className="modal" style={{ width: 480 }}>
            <header>
              <div><div className="h">Ödəniş tarixçəsi</div><div className="tiny">{detail.customer?.name} · {detail.item_name}</div></div>
              <button className="x" onClick={() => setDetail(null)}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
            </header>
            <div className="body" style={{ display: 'block' }}>
              <div className="row" style={{ marginBottom: 16 }}>
                <div className="card pad" style={{ flex: 1, boxShadow: 'none' }}><div className="eyebrow">Ümumi</div><div className="data" style={{ fontSize: 19, fontWeight: 600, marginTop: 4 }}>{money(detail.total)}</div></div>
                <div className="card pad" style={{ flex: 1, boxShadow: 'none' }}><div className="eyebrow">Ödənilib</div><div className="data" style={{ fontSize: 19, fontWeight: 600, marginTop: 4, color: 'var(--good-ink)' }}>{money(detail.paid)}</div></div>
                <div className="card pad" style={{ flex: 1, boxShadow: 'none' }}><div className="eyebrow">Qalıq</div><div className="data" style={{ fontSize: 19, fontWeight: 600, marginTop: 4, color: 'var(--bad)' }}>{money(Math.max(0, detail.total - detail.paid))}</div></div>
              </div>
              <div className="h" style={{ fontSize: 14, marginBottom: 4 }}>Ödənişlər</div>
              {(detail.payments ?? []).slice().sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()).map((p) => (
                <div key={p.id} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '11px 0', borderTop: '1px solid var(--line3)' }}>
                  <span className="tiny data">{dateTimeAz(p.created_at)}</span>
                  <span className="data profit">+{money(p.amount)}</span>
                </div>
              ))}
              {(!detail.payments || detail.payments.length === 0) && <div className="tiny" style={{ marginTop: 8 }}>Hələ ödəniş yoxdur</div>}
            </div>
            <footer>
              <button className="btn ghost" style={{ flex: 1 }} onClick={() => setDetail(null)}>Bağla</button>
              {detail.paid < detail.total && <button className="btn primary" style={{ flex: 1 }} onClick={() => { setPay(detail); setDetail(null) }}>Ödəniş et</button>}
            </footer>
          </div>
        </>
      )}
    </div>
  )
}
