import { useState } from 'react'
import { api, type Branch, type Item } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import { money } from '../lib/format'
import FormModal from '../components/FormModal'

export default function Branches() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: branches } = useFetch<Branch[]>('/branches', [key])
  const { data: items } = useFetch<Item[]>('/items', [key])
  const [branchForm, setBranchForm] = useState(false)
  const [transferForm, setTransferForm] = useState(false)
  const all = items ?? []

  const stat = (bid: number) => {
    const inStock = all.filter((i) => i.branch_id === bid && i.status === 'in_stock')
    return { count: inStock.length, value: inStock.reduce((a, i) => a + i.cost, 0) }
  }
  const inStockItems = all.filter((i) => i.status === 'in_stock')

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M3 21V9l9-6 9 6v12" /></svg><div>Excel-dəki «Elçin & Rəşid» və «Zaur» əslində <b>filiallarımızdır</b>. Hansı cihazı hansı filiala göndərdiyimizi buradan izləyirik.</div></div>

      <div className="row" style={{ flexWrap: 'wrap' }}>
        {branches?.map((b) => {
          const st = stat(b.id)
          return (
            <div key={b.id} className="card branchcard" style={{ minWidth: 240 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}><div className="nm">{b.name}</div><span className={'pill ' + (b.is_main ? 'ink' : 'neut')}>{b.is_main ? 'Əsas' : 'Filial'}</span></div>
              <div className="tiny">{b.address}</div>
              <div className="stat-inline"><div><div className="l">Stok</div><div className="n">{st.count}</div></div><div><div className="l">Dəyər</div><div className="n">{money(st.value)}</div></div></div>
            </div>
          )
        })}
        <div className="card branchcard" style={{ minWidth: 200, display: 'flex', alignItems: 'center', justifyContent: 'center', flexDirection: 'column', gap: 8, color: 'var(--muted)', borderStyle: 'dashed', cursor: 'pointer' }} onClick={() => setBranchForm(true)}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>
          <div style={{ fontSize: 12.5, fontWeight: 700 }}>Yeni filial</div>
        </div>
      </div>

      <div className="card pad">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}><div className="h">Filiala transfer</div><button className="btn primary sm" onClick={() => setTransferForm(true)}>+ Transfer</button></div>
        <div className="tiny">Cihazı filiala göndərdikdə buradan qeyd et; həmin cihazın filialı dəyişir və Stok/Anbarda görünür.</div>
      </div>

      {branchForm && (
        <FormModal
          title="Yeni filial" submitLabel="Əlavə et" onClose={() => setBranchForm(false)}
          fields={[
            { name: 'name', label: 'Filial adı', required: true, full: true, placeholder: 'məs. Gənclik filialı' },
            { name: 'address', label: 'Ünvan', full: true },
          ]}
          onSubmit={async (v) => { await api('/branches', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Filial əlavə edildi'); setBranchForm(false) }}
        />
      )}

      {transferForm && (
        <FormModal
          title="Filiala transfer" subtitle="cihazı seç → hansı filiala" submitLabel="Transfer et" onClose={() => setTransferForm(false)}
          fields={[
            { name: 'item_id', label: 'Cihaz (stokda)', type: 'searchselect', required: true, full: true, placeholder: 'ad, seriya və ya marka yaz…', options: inStockItems.map((i) => ({ value: i.id, label: `${i.name}${i.serial ? ' · ' + i.serial : ''} — ${i.branch?.name}` })) },
            { name: 'branch_id', label: 'Hansı filiala', type: 'select', required: true, full: true, options: (branches ?? []).map((b) => ({ value: b.id, label: b.name })) },
          ]}
          onSubmit={async (v) => { await api('/transfers', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Cihaz filiala köçürüldü'); setTransferForm(false) }}
        />
      )}
    </div>
  )
}
