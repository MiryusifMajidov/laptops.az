import { useState } from 'react'
import { api, type Attribute, type Category } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import FormModal from '../components/FormModal'

export default function Categories() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: cats } = useFetch<Category[]>('/categories', [key])
  const { data: attrs } = useFetch<Attribute[]>('/attributes', [key])
  const [catForm, setCatForm] = useState(false)
  const [catEdit, setCatEdit] = useState<Category | null>(null)
  const [attrForm, setAttrForm] = useState(false)
  const [attrEdit, setAttrEdit] = useState<Attribute | null>(null)
  const [optFor, setOptFor] = useState<Attribute | null>(null)

  async function del(path: string, msg: string) {
    try { await api(path, { method: 'DELETE' }); bump(); toast(msg) } catch (e) { toast((e as Error).message) }
  }

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M4 6h16M4 12h16M4 18h10" /></svg><div>Kateqoriya və xüsusiyyətləri buradan <b>yarat, redaktə et və sil</b>. Məhsul əlavə edəndə yalnız seçdiyin kateqoriyanın xüsusiyyətləri çıxır; saytda filtrlər avtomatik bunlardan yaranır.</div></div>

      <div className="two">
        <div className="card" style={{ overflow: 'hidden' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '16px 18px 10px' }}><div className="h">Kateqoriyalar</div><button className="btn ghost sm" onClick={() => setCatForm(true)}>+ Kateqoriya</button></div>
          <div className="catlist">
            {cats?.map((c) => (
              <div key={c.id} className="catitem" style={{ cursor: 'default' }}>
                <div><div style={{ fontWeight: 700, fontSize: 13.5 }}>{c.name}</div><div className="tiny" style={{ marginTop: 2 }}>{c.attributes?.length ?? 0} xüsusiyyət: {c.attributes?.map((a) => a.name).join(', ')}</div></div>
                <div style={{ display: 'flex', gap: 6, flex: 'none' }}>
                  <button className="btn ghost sm" onClick={() => setCatEdit(c)}>Redaktə</button>
                  <button className="btn ghost sm" style={{ color: 'var(--bad)' }} onClick={() => window.confirm(`«${c.name}» silinsin?`) && del(`/categories/${c.id}`, 'Kateqoriya silindi')}>Sil</button>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="card pad">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}><div className="h">Xüsusiyyətlər</div><button className="btn ghost sm" onClick={() => setAttrForm(true)}>+ Xüsusiyyət</button></div>
          <div className="tiny" style={{ marginBottom: 14 }}>hər xüsusiyyət və onun seçimləri (option-ları)</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            {attrs?.map((a) => (
              <div key={a.id}>
                <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 7 }}>
                  <div style={{ fontSize: 12.5, fontWeight: 700 }}>{a.name}</div>
                  <div style={{ display: 'flex', gap: 6 }}>
                    <button className="btn ghost sm" onClick={() => setAttrEdit(a)}>Redaktə</button>
                    <button className="btn ghost sm" style={{ color: 'var(--bad)' }} onClick={() => window.confirm(`«${a.name}» və bütün seçimləri silinsin?`) && del(`/attributes/${a.id}`, 'Xüsusiyyət silindi')}>Sil</button>
                  </div>
                </div>
                <div className="optchips">
                  {a.options?.map((o) => (
                    <span key={o.id} className="optchip">{o.value}<span onClick={() => window.confirm('Option silinsin?') && del(`/attribute-options/${o.id}`, 'Option silindi')} style={{ marginLeft: 7, color: 'var(--bad)', cursor: 'pointer', fontWeight: 800 }}>×</span></span>
                  ))}
                  <span className="optchip add" onClick={() => setOptFor(a)}>+ option</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {catForm && (
        <FormModal
          title="Yeni kateqoriya" subtitle="hansı xüsusiyyətlər aiddir seç" submitLabel="Əlavə et" onClose={() => setCatForm(false)}
          fields={[
            { name: 'name', label: 'Kateqoriya adı', required: true, full: true, placeholder: 'məs. Planşet' },
            { name: 'attribute_ids', label: 'Xüsusiyyətlər', type: 'multiselect', options: (attrs ?? []).map((a) => ({ value: a.id, label: a.name })) },
          ]}
          onSubmit={async (v) => { await api('/categories', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Kateqoriya əlavə edildi'); setCatForm(false) }}
        />
      )}
      {catEdit && (
        <FormModal
          title="Kateqoriyanı redaktə et" subtitle="ad + aid xüsusiyyətlər" submitLabel="Yadda saxla" onClose={() => setCatEdit(null)}
          initial={{ name: catEdit.name, attribute_ids: catEdit.attributes?.map((a) => a.id) ?? [] }}
          fields={[
            { name: 'name', label: 'Kateqoriya adı', required: true, full: true },
            { name: 'attribute_ids', label: 'Xüsusiyyətlər', type: 'multiselect', options: (attrs ?? []).map((a) => ({ value: a.id, label: a.name })) },
          ]}
          onSubmit={async (v) => { await api(`/categories/${catEdit.id}`, { method: 'PUT', body: JSON.stringify(v) }); bump(); toast('Kateqoriya yeniləndi'); setCatEdit(null) }}
        />
      )}

      {attrForm && (
        <FormModal
          title="Yeni xüsusiyyət" subtitle="seçimləri vergüllə ayır" submitLabel="Əlavə et" onClose={() => setAttrForm(false)}
          fields={[
            { name: 'name', label: 'Xüsusiyyət adı', required: true, full: true, placeholder: 'məs. Klaviatura' },
            { name: 'options', label: 'Seçimlər (vergüllə)', full: true, placeholder: 'ENG, RUS, AZ' },
          ]}
          onSubmit={async (v) => {
            const opts = String(v.options ?? '').split(',').map((x) => x.trim()).filter(Boolean)
            await api('/attributes', { method: 'POST', body: JSON.stringify({ name: v.name, options: opts }) })
            bump(); toast('Xüsusiyyət əlavə edildi'); setAttrForm(false)
          }}
        />
      )}
      {attrEdit && (
        <FormModal
          title="Xüsusiyyəti redaktə et" submitLabel="Yadda saxla" onClose={() => setAttrEdit(null)}
          initial={{ name: attrEdit.name }}
          fields={[{ name: 'name', label: 'Xüsusiyyət adı', required: true, full: true }]}
          onSubmit={async (v) => { await api(`/attributes/${attrEdit.id}`, { method: 'PUT', body: JSON.stringify(v) }); bump(); toast('Xüsusiyyət yeniləndi'); setAttrEdit(null) }}
        />
      )}

      {optFor && (
        <FormModal
          title={`«${optFor.name}» — yeni option`} submitLabel="Əlavə et" onClose={() => setOptFor(null)}
          fields={[{ name: 'value', label: 'Dəyər', required: true, full: true, placeholder: 'məs. 128 GB' }]}
          onSubmit={async (v) => { await api(`/attributes/${optFor.id}/options`, { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Option əlavə edildi'); setOptFor(null) }}
        />
      )}
    </div>
  )
}
