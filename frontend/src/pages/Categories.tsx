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
  const [optEdit, setOptEdit] = useState<{ id: number; value: string } | null>(null)
  const [drag, setDrag] = useState<{ attrId: number; idx: number } | null>(null)

  async function del(path: string, msg: string) {
    try { await api(path, { method: 'DELETE' }); bump(); toast(msg) } catch (e) { toast((e as Error).message) }
  }

  // option sıralaması — verilmiş ardıcıllığı serverdə saxla
  async function reorder(a: Attribute, ids: number[]) {
    try { await api(`/attributes/${a.id}/options/order`, { method: 'PUT', body: JSON.stringify({ ids }) }); bump() }
    catch (e) { toast((e as Error).message) }
  }
  // A→Z (rəqəm nəzərə alınır: 8 GB < 16 GB < 32 GB)
  function sortAZ(a: Attribute) {
    const ids = [...(a.options ?? [])].sort((x, y) => x.value.localeCompare(y.value, undefined, { numeric: true })).map((o) => o.id)
    if (ids.length) reorder(a, ids)
  }
  // bir option-u sola/sağa (əvvələ/sona) sürüşdür
  function move(a: Attribute, idx: number, dir: -1 | 1) {
    const opts = [...(a.options ?? [])]
    const j = idx + dir
    if (j < 0 || j >= opts.length) return
    ;[opts[idx], opts[j]] = [opts[j], opts[idx]]
    reorder(a, opts.map((o) => o.id))
  }
  // drag ilə: from mövqeyindən to mövqeyinə köçür
  function dropReorder(a: Attribute, from: number, to: number) {
    if (from === to) return
    const opts = [...(a.options ?? [])]
    const [moved] = opts.splice(from, 1)
    opts.splice(to, 0, moved)
    reorder(a, opts.map((o) => o.id))
  }
  // saytda göstər/gizlət (default: göstərilir)
  async function toggleSite(a: Attribute) {
    const cur = a.show_on_site !== false
    try { await api(`/attributes/${a.id}`, { method: 'PUT', body: JSON.stringify({ name: a.name, show_on_site: !cur }) }); bump(); toast(!cur ? 'Saytda göstərilir' : 'Saytda gizləndi') }
    catch (e) { toast((e as Error).message) }
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
                  <div style={{ fontSize: 12.5, fontWeight: 700 }}>{a.name}{a.multiselect && <span className="pill good" style={{ marginLeft: 6, fontSize: 9.5 }}>çox seçim</span>}{a.show_on_site === false && <span className="pill" style={{ marginLeft: 6, fontSize: 9.5, background: 'var(--card)', border: '1px solid var(--line)', color: 'var(--muted)' }}>saytda gizli</span>}</div>
                  <div style={{ display: 'flex', gap: 6 }}>
                    <button className="btn ghost sm" onClick={() => toggleSite(a)} title="Saytda göstərilsin?" style={{ color: a.show_on_site === false ? 'var(--muted)' : 'var(--good)', fontWeight: 700 }}>{a.show_on_site === false ? '⊘ Saytda gizli' : '◉ Saytda'}</button>
                    <button className="btn ghost sm" onClick={() => sortAZ(a)} title="Seçimləri A→Z (rəqəm nəzərə alınır) sırala">↕ Sırala</button>
                    <button className="btn ghost sm" onClick={() => setAttrEdit(a)}>Redaktə</button>
                    <button className="btn ghost sm" style={{ color: 'var(--bad)' }} onClick={() => window.confirm(`«${a.name}» və bütün seçimləri silinsin?`) && del(`/attributes/${a.id}`, 'Xüsusiyyət silindi')}>Sil</button>
                  </div>
                </div>
                <div className="optchips">
                  {a.options?.map((o, oi) => (
                    <span key={o.id} className="optchip" draggable
                      onDragStart={() => setDrag({ attrId: a.id, idx: oi })}
                      onDragOver={(e) => { if (drag?.attrId === a.id) e.preventDefault() }}
                      onDrop={(e) => { e.preventDefault(); if (drag && drag.attrId === a.id) dropReorder(a, drag.idx, oi); setDrag(null) }}
                      onDragEnd={() => setDrag(null)}
                      style={{ opacity: drag?.attrId === a.id && drag.idx === oi ? 0.4 : 1 }}
                      title="Sürüşdürüb yerini dəyiş">
                      <span onClick={() => move(a, oi, -1)} style={{ cursor: 'pointer', color: 'var(--muted)', fontWeight: 800, marginRight: 5 }} title="Əvvələ">‹</span>
                      <span onClick={() => setOptEdit({ id: o.id, value: o.value })} style={{ cursor: 'pointer' }} title="Redaktə et (klik)">{o.value}</span>
                      <span onClick={() => move(a, oi, 1)} style={{ cursor: 'pointer', color: 'var(--muted)', fontWeight: 800, marginLeft: 5 }} title="Sona">›</span>
                      <span onClick={() => window.confirm('Option silinsin?') && del(`/attribute-options/${o.id}`, 'Option silindi')} style={{ marginLeft: 6, color: 'var(--bad)', cursor: 'pointer', fontWeight: 800 }}>×</span>
                    </span>
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
            { name: 'multiselect', label: 'Seçim növü', type: 'select', options: [{ value: '', label: 'Tək seçim (adi)' }, { value: 'multi', label: 'Çox seçimli — bir neçə option seçilə bilər' }] },
            { name: 'options', label: 'Seçimlər (vergüllə)', full: true, placeholder: 'ENG, RUS, AZ' },
          ]}
          onSubmit={async (v) => {
            const opts = String(v.options ?? '').split(',').map((x) => x.trim()).filter(Boolean)
            await api('/attributes', { method: 'POST', body: JSON.stringify({ name: v.name, multiselect: v.multiselect === 'multi', options: opts }) })
            bump(); toast('Xüsusiyyət əlavə edildi'); setAttrForm(false)
          }}
        />
      )}
      {attrEdit && (
        <FormModal
          title="Xüsusiyyəti redaktə et" submitLabel="Yadda saxla" onClose={() => setAttrEdit(null)}
          initial={{ name: attrEdit.name, multiselect: attrEdit.multiselect ? 'multi' : '' }}
          fields={[
            { name: 'name', label: 'Xüsusiyyət adı', required: true, full: true },
            { name: 'multiselect', label: 'Seçim növü', type: 'select', options: [{ value: '', label: 'Tək seçim (adi)' }, { value: 'multi', label: 'Çox seçimli (multiselect)' }] },
          ]}
          onSubmit={async (v) => { await api(`/attributes/${attrEdit.id}`, { method: 'PUT', body: JSON.stringify({ name: v.name, multiselect: v.multiselect === 'multi' }) }); bump(); toast('Xüsusiyyət yeniləndi'); setAttrEdit(null) }}
        />
      )}

      {optFor && (
        <FormModal
          title={`«${optFor.name}» — yeni option`} submitLabel="Əlavə et" onClose={() => setOptFor(null)}
          fields={[{ name: 'value', label: 'Dəyər', required: true, full: true, placeholder: 'məs. 128 GB' }]}
          onSubmit={async (v) => { await api(`/attributes/${optFor.id}/options`, { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Option əlavə edildi'); setOptFor(null) }}
        />
      )}
      {optEdit && (
        <FormModal
          title="Option redaktə et" subtitle="dəyəri dəyiş — məhsullardakı dəyər də avtomatik yenilənir. Mövcud başqa dəyəri yazsan → birləşir" submitLabel="Yadda saxla" onClose={() => setOptEdit(null)}
          initial={{ value: optEdit.value }}
          fields={[{ name: 'value', label: 'Dəyər', required: true, full: true }]}
          onSubmit={async (v) => { const r = await api<{ merged?: boolean; items_updated?: number }>(`/attribute-options/${optEdit.id}`, { method: 'PUT', body: JSON.stringify(v) }); bump(); toast(r.merged ? `Birləşdirildi · ${r.items_updated ?? 0} məhsul yeniləndi` : 'Option yeniləndi'); setOptEdit(null) }}
        />
      )}
    </div>
  )
}
