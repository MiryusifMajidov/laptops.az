import { useEffect, useState } from 'react'
import { api, uploadFile, imgURL, type Branch, type Category, type Item, type Language, STATUS_AZ, isAdmin, currentBranchId, currentBranchName } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'

export default function NewProductModal({ onClose, edit }: { onClose: () => void; edit?: Item }) {
  const toast = useToast()
  const { bump } = useRefresh()
  const { data: cats } = useFetch<Category[]>('/categories', [])
  const { data: branches } = useFetch<Branch[]>('/branches', [])
  const { data: langs } = useFetch<Language[]>('/languages', [])

  const [catId, setCatId] = useState<number | null>(edit ? edit.category_id : null)
  const [serial, setSerial] = useState(edit?.serial ?? '')
  const [cost, setCost] = useState(edit ? String(edit.cost) : '')
  const [price, setPrice] = useState(edit ? String(edit.price) : '')
  const [wholesale, setWholesale] = useState(edit && edit.wholesale_price ? String(edit.wholesale_price) : '')
  const [discount, setDiscount] = useState(edit && edit.discount ? String(edit.discount) : '')
  const [quantity, setQuantity] = useState(edit ? String(edit.quantity ?? 1) : '1')
  const admin = isAdmin()
  // satıcı (admin deyil) yeni məhsulda öz filialına kilidlidir; admin istədiyi filialı seçir
  const [branchId, setBranchId] = useState<number | null>(edit ? edit.branch_id : (admin ? null : currentBranchId()))
  // adlar dil üzrə: {az:.., ru:.., ...}
  const [names, setNames] = useState<Record<string, string>>({})
  const [nameTab, setNameTab] = useState('')
  const [inited, setInited] = useState(false)
  const [site, setSite] = useState(edit ? edit.show_on_site : true)
  const [status, setStatus] = useState(edit?.status ?? 'in_stock')
  const [createdAt, setCreatedAt] = useState((edit?.created_at ?? '').slice(0, 10))
  const [soldAt, setSoldAt] = useState((edit?.sold_at ?? '').slice(0, 10))
  const [cardImage, setCardImage] = useState(edit?.card_image ?? '')
  const [gallery, setGallery] = useState<string[]>(() => {
    try { return edit?.gallery ? JSON.parse(edit.gallery) : [] } catch { return [] }
  })
  const [uploading, setUploading] = useState(false)
  const [dragImg, setDragImg] = useState<number | null>(null)
  // hər xüsusiyyət üçün seçilmiş dəyər(lər) — massiv (tək seçim = 1 element, multiselect = çox)
  const [values, setValues] = useState<Record<number, string[]>>(() => {
    const m: Record<number, string[]> = {}
    edit?.values?.forEach((v) => { (m[v.attribute_id] ||= []).push(v.value) })
    return m
  })
  const [saving, setSaving] = useState(false)
  const [translating, setTranslating] = useState(false)

  useEffect(() => { if (!edit && cats && cats.length && catId === null) setCatId(cats[0].id) }, [cats, catId, edit])
  useEffect(() => { if (!edit && admin && branches && branches.length && branchId === null) setBranchId(branches[0].id) }, [branches, branchId, edit, admin])
  // dillər yüklənəndə ad xanalarını qur
  useEffect(() => {
    if (inited || !langs) return
    const def = langs.find((l) => l.is_default)?.code ?? langs[0]?.code ?? 'az'
    const m: Record<string, string> = { [def]: edit?.name ?? '' }
    ;(edit?.translations ?? []).forEach((t) => { m[t.lang] = t.name })
    setNames(m)
    setNameTab(def)
    setInited(true)
  }, [langs, inited, edit])

  const cat = cats?.find((c) => c.id === catId)
  const activeLangs = (langs ?? []).filter((l) => l.enabled)
  const defCode = (langs ?? []).find((l) => l.is_default)?.code ?? 'az'

  async function onCard(e: React.ChangeEvent<HTMLInputElement>) {
    const f = e.target.files?.[0]
    if (!f) return
    setUploading(true)
    try { setCardImage(await uploadFile(f)) } catch { toast('şəkil yüklənmədi') } finally { setUploading(false) }
  }
  async function onGallery(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? [])
    setUploading(true)
    try { for (const f of files) { const url = await uploadFile(f); setGallery((g) => [...g, url]) } } catch { toast('şəkil yüklənmədi') } finally { setUploading(false) }
  }

  async function save() {
    const mainName = (names[defCode] ?? '').trim()
    if (!mainName || !catId || !branchId) { toast('Ad (əsas dil), kateqoriya və filial vacibdir'); return }
    setSaving(true)
    const translations = Object.entries(names)
      .filter(([lang]) => lang !== defCode)
      .map(([lang, n]) => ({ lang, name: (n ?? '').trim() }))
    const body: Record<string, unknown> = {
      name: mainName, serial, category_id: catId, branch_id: branchId,
      cost: Number(cost) || 0, price: Number(price) || 0,
      wholesale_price: Number(wholesale) || 0, discount: Number(discount) || 0,
      quantity: Math.max(1, Number(quantity) || 1),
      show_on_site: site, status, card_image: cardImage, gallery: JSON.stringify(gallery),
      translations,
      values: Object.entries(values).flatMap(([aid, vals]) => (vals ?? []).filter(Boolean).map((val) => ({ attribute_id: Number(aid), value: val }))),
    }
    if (createdAt) body.created_at = createdAt
    if (edit && status === 'sold' && soldAt) body.sold_at = soldAt
    try {
      if (edit) await api(`/items/${edit.id}`, { method: 'PUT', body: JSON.stringify(body) })
      else await api('/items', { method: 'POST', body: JSON.stringify(body) })
      bump(); toast(edit ? 'Məhsul yeniləndi' : 'Məhsul stoka əlavə edildi'); onClose()
    } catch (e) { toast((e as Error).message) } finally { setSaving(false) }
  }

  async function remove() {
    if (!edit || !window.confirm(`«${edit.name}» silinsin?`)) return
    try { await api(`/items/${edit.id}`, { method: 'DELETE' }); bump(); toast('Məhsul silindi'); onClose() }
    catch (e) { toast((e as Error).message) }
  }

  // əsas dildəki adı qalan dillərə AI ilə tərcümə et → tabları doldur
  async function aiTranslateName() {
    const src = (names[defCode] ?? '').trim()
    if (!src) { toast('Əvvəlcə əsas dildə (AZ) ad yazın'); return }
    const targets = activeLangs.filter((l) => l.code !== defCode).map((l) => l.code)
    if (targets.length === 0) return
    setTranslating(true)
    try {
      const r = await api<{ translations: Record<string, string> }>('/translate', { method: 'POST', body: JSON.stringify({ text: src, from: defCode, targets }) })
      setNames((n) => ({ ...n, ...r.translations }))
      toast('AI tərcümə əlavə olundu')
    } catch (e) { toast((e as Error).message) } finally { setTranslating(false) }
  }

  return (
    <>
      <div className="scrim" onClick={onClose} />
      <div className="modal">
        <header>
          <div><div className="h">{edit ? 'Məhsulu redaktə et' : 'Yeni məhsul'}</div><div className="tiny">{edit ? (edit.serial || '—') : 'kateqoriya seç → uyğun xüsusiyyətlər çıxacaq'}</div></div>
          <button className="x" onClick={onClose}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
        </header>
        <div className="body">
          <div className="field full">
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
              <label style={{ marginBottom: 0 }}>Məhsulun adı {activeLangs.length > 1 && <span className="tiny" style={{ fontWeight: 600 }}>(hər dil üçün)</span>}</label>
              {activeLangs.length > 1 && (
                <button type="button" className="btn ghost sm" onClick={aiTranslateName} disabled={translating} style={{ padding: '4px 10px', flex: 'none' }}>
                  {translating ? 'Tərcümə olunur…' : '✨ AI ilə tərcümə'}
                </button>
              )}
            </div>
            {activeLangs.length > 1 && (
              <div className="lang-tabs" style={{ marginTop: 7 }}>
                {activeLangs.map((l) => (
                  <button key={l.code} type="button" className={'lang-tab' + (nameTab === l.code ? ' on' : '')} onClick={() => setNameTab(l.code)}>
                    {l.code.toUpperCase()}{l.is_default ? ' ★' : (names[l.code]?.trim() ? ' ●' : '')}
                  </button>
                ))}
              </div>
            )}
            <input
              placeholder={nameTab === defCode ? 'məs. NOTEBOOK HP ProBook 450 G10' : (names[defCode]?.trim() ? `${names[defCode]} — ${nameTab.toUpperCase()} tərcümə` : 'tərcümə (boş qalsa əsas dil göstərilir)')}
              value={names[nameTab] ?? ''}
              onChange={(e) => setNames((n) => ({ ...n, [nameTab]: e.target.value }))}
            />
          </div>
          <div className="field"><label>Kateqoriya</label>
            <select value={catId ?? ''} onChange={(e) => { setCatId(Number(e.target.value)); setValues({}) }}>
              {cats?.map((c) => <option key={c.id} value={c.id}>{c.name}</option>)}
            </select>
          </div>
          <div className="field"><label>Seriya nömrəsi</label><input placeholder="məs. 5cd5232k5c" value={serial} onChange={(e) => setSerial(e.target.value)} /></div>

          {edit && (
            <div className="field"><label>Status</label>
              <select value={status} onChange={(e) => setStatus(e.target.value)}>
                {['in_stock', 'reserved', 'sold', 'returned'].map((s) => <option key={s} value={s}>{STATUS_AZ[s]}</option>)}
              </select>
            </div>
          )}

          {cat?.attributes?.map((a) => (
            <div className="field" key={a.id}>
              <label>{a.name}{a.multiselect && <span className="tiny" style={{ fontWeight: 600 }}> (çox seçim)</span>}</label>
              {a.multiselect ? (
                <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, paddingTop: 2 }}>
                  {a.options?.map((o) => {
                    const on = (values[a.id] ?? []).includes(o.value)
                    return (
                      <span key={o.id} className={'chip' + (on ? ' on' : '')} style={{ cursor: 'pointer' }}
                        onClick={() => setValues((v) => {
                          const cur = v[a.id] ?? []
                          return { ...v, [a.id]: on ? cur.filter((x) => x !== o.value) : [...cur, o.value] }
                        })}>
                        {o.value}
                      </span>
                    )
                  })}
                  {!a.options?.length && <span className="tiny" style={{ color: 'var(--muted)' }}>seçim yoxdur</span>}
                </div>
              ) : (
                <select value={values[a.id]?.[0] ?? ''} onChange={(e) => setValues((v) => ({ ...v, [a.id]: e.target.value ? [e.target.value] : [] }))}>
                  <option value="">— seç —</option>
                  {/* cari dəyər seçimlər siyahısında yoxdursa da göstər (seçili qalsın) */}
                  {(() => { const cur = values[a.id]?.[0]; return cur && !a.options?.some((o) => o.value === cur) ? <option value={cur}>{cur}</option> : null })()}
                  {a.options?.map((o) => <option key={o.id} value={o.value}>{o.value}</option>)}
                </select>
              )}
            </div>
          ))}

          <div className="field"><label>Alış qiyməti (₼) <span className="tiny" style={{ fontWeight: 600 }}>daxili</span></label><input className="data" type="number" placeholder="0" value={cost} onChange={(e) => setCost(e.target.value)} /></div>
          <div className="field"><label>Topdan / optavoy (₼) <span className="tiny" style={{ fontWeight: 600 }}>daxili · optional</span></label><input className="data" type="number" placeholder="0" value={wholesale} onChange={(e) => setWholesale(e.target.value)} /></div>
          <div className="field"><label>Satış qiyməti (₼) <span className="tiny" style={{ fontWeight: 600 }}>saytda görünür</span></label><input className="data" type="number" placeholder="0" value={price} onChange={(e) => setPrice(e.target.value)} /></div>
          <div className="field"><label>Endirim (₼) <span className="tiny" style={{ fontWeight: 600 }}>optional · manatla</span></label><input className="data" type="number" placeholder="0" value={discount} onChange={(e) => setDiscount(e.target.value)} />
            {Number(discount) > 0 && Number(price) > 0 && <div className="tiny" style={{ marginTop: 4, color: 'var(--good-ink)' }}>Saytda: ₼{Math.max(0, Number(price) - Number(discount))} <span style={{ textDecoration: 'line-through', color: 'var(--muted)' }}>₼{Number(price)}</span></div>}
          </div>
          <div className="field"><label>Say (ədəd) <span className="tiny" style={{ fontWeight: 600 }}>serialı mal = 1 · aksesuar = çox</span></label>
            <input className="data" type="number" min={1} placeholder="1" value={quantity} onChange={(e) => setQuantity(e.target.value)} />
          </div>
          <div className="field"><label>Filial {!admin && <span className="tiny" style={{ fontWeight: 600 }}>(öz filialınız — dəyişilməz)</span>}</label>
            {admin ? (
              <select value={branchId ?? ''} onChange={(e) => setBranchId(Number(e.target.value))}>
                {branches?.map((b) => <option key={b.id} value={b.id}>{b.name}</option>)}
              </select>
            ) : (
              <input readOnly disabled value={branches?.find((b) => b.id === branchId)?.name || currentBranchName() || '—'} style={{ background: 'var(--surface)', color: 'var(--muted)', cursor: 'not-allowed' }} />
            )}
          </div>

          <div className="field"><label>Alınma tarixi <span className="tiny" style={{ fontWeight: 600 }}>(gəldiyi gün)</span></label>
            <input type="date" value={createdAt} onChange={(e) => setCreatedAt(e.target.value)} />
          </div>
          {edit && status === 'sold' && (
            <div className="field"><label>Satılma tarixi <span className="tiny" style={{ fontWeight: 600 }}>(satıldığı gün)</span></label>
              <input type="date" value={soldAt} onChange={(e) => setSoldAt(e.target.value)} />
            </div>
          )}

          {/* Kart şəkli — saytda kartlarda görünən */}
          <div className="field"><label>Kart şəkli <span className="tiny" style={{ fontWeight: 600 }}>(saytda kartda)</span></label>
            <label className="uploadbox" style={{ cursor: 'pointer', padding: cardImage ? 8 : 22, display: 'block' }}>
              {cardImage
                ? <img src={imgURL(cardImage)} alt="" style={{ maxHeight: 90, maxWidth: '100%', borderRadius: 8, display: 'block', margin: '0 auto' }} />
                : 'Kart şəkli seç'}
              <input type="file" accept="image/*" hidden onChange={onCard} />
            </label>
            {cardImage && <button className="btn ghost sm" style={{ marginTop: 6, color: 'var(--bad)' }} onClick={() => setCardImage('')}>Sil</button>}
          </div>

          {/* Qalereya — məhsul səhifəsində görünən çoxlu şəkil */}
          <div className="field"><label>Qalereya <span className="tiny" style={{ fontWeight: 600 }}>(məhsul səhifəsi — çox şəkil · sürüşdürüb sırala)</span></label>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {gallery.map((g, i) => (
                <div key={g + i} draggable
                  onDragStart={() => setDragImg(i)}
                  onDragOver={(e) => e.preventDefault()}
                  onDrop={(e) => { e.preventDefault(); setGallery((arr) => { if (dragImg === null || dragImg === i) return arr; const a = [...arr]; const [m] = a.splice(dragImg, 1); a.splice(i, 0, m); return a }); setDragImg(null) }}
                  onDragEnd={() => setDragImg(null)}
                  style={{ position: 'relative', width: 64, height: 64, cursor: 'grab', opacity: dragImg === i ? 0.4 : 1 }} title="Sürüşdürüb sırala">
                  <img src={imgURL(g)} alt="" style={{ width: 64, height: 64, objectFit: 'cover', borderRadius: 8, border: '1px solid var(--line)', pointerEvents: 'none' }} />
                  {i === 0 && <span style={{ position: 'absolute', bottom: -6, left: '50%', transform: 'translateX(-50%)', background: 'var(--ink)', color: '#fff', fontSize: 8.5, fontWeight: 700, padding: '1px 5px', borderRadius: 6, whiteSpace: 'nowrap' }}>əsas</span>}
                  <button onClick={() => setGallery((arr) => arr.filter((_, x) => x !== i))} style={{ position: 'absolute', top: -6, right: -6, width: 18, height: 18, borderRadius: '50%', background: 'var(--bad)', color: '#fff', fontSize: 11, lineHeight: '18px', textAlign: 'center' }}>×</button>
                </div>
              ))}
              <label className="uploadbox" style={{ cursor: 'pointer', width: 64, height: 64, padding: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 22 }}>
                +<input type="file" accept="image/*" multiple hidden onChange={onGallery} />
              </label>
            </div>
          </div>

          <div className="field full" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: 'var(--card)', border: '1px solid var(--line)', borderRadius: 10, padding: '12px 14px' }}>
            <div><div style={{ fontSize: 13, fontWeight: 700 }}>Saytda göstər</div><div className="tiny">{uploading ? 'şəkil yüklənir…' : 'zədəli/qutusuz malları söndür'}</div></div>
            <div className={'tg' + (site ? ' on' : '')} onClick={() => setSite((v) => !v)} />
          </div>
        </div>
        <footer>
          {edit && <button className="btn ghost" style={{ color: 'var(--bad)' }} onClick={remove}>Sil</button>}
          <button className="btn ghost" style={{ flex: 1 }} onClick={onClose}>İmtina</button>
          <button className="btn primary" style={{ flex: 1.6 }} onClick={save} disabled={saving || uploading}>{saving ? 'Saxlanır…' : (edit ? 'Yadda saxla' : 'Stoka əlavə et')}</button>
        </footer>
      </div>
    </>
  )
}
