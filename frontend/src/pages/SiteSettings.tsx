import { useEffect, useMemo, useRef, useState } from 'react'
import { api, uploadFile, imgURL, type Language, type UiStringsData, type UiStringRow, type TermsData, type TermRow, type Item } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'

/* eslint-disable @typescript-eslint/no-explicit-any */
declare global { interface Window { L: any } }

export default function SiteSettings() {
  const [tab, setTab] = useState<'general' | 'langs' | 'texts' | 'terms'>('general')
  return (
    <div className="content">
      <div className="subtabs">
        <button className={'subtab' + (tab === 'general' ? ' on' : '')} onClick={() => setTab('general')}>Ümumi</button>
        <button className={'subtab' + (tab === 'langs' ? ' on' : '')} onClick={() => setTab('langs')}>Dillər</button>
        <button className={'subtab' + (tab === 'texts' ? ' on' : '')} onClick={() => setTab('texts')}>Sayt mətnləri</button>
        <button className={'subtab' + (tab === 'terms' ? ' on' : '')} onClick={() => setTab('terms')}>Kateqoriya & Xüsusiyyət</button>
      </div>
      {tab === 'general' && <GeneralTab />}
      {tab === 'langs' && <LanguagesTab />}
      {tab === 'texts' && <TextsTab />}
      {tab === 'terms' && <TermsTab />}
    </div>
  )
}

// ---------------- Ümumi (xəritə yeri + seçilmiş məhsul) ----------------
function GeneralTab() {
  const toast = useToast()
  const [tick, setTick] = useState(0)
  const { data: settings } = useFetch<Record<string, string>>('/settings', [tick])
  const { data: items } = useFetch<Item[]>('/items?status=in_stock', [])
  const [lat, setLat] = useState(40.3811247)
  const [lng, setLng] = useState(49.8474406)
  const [featuredId, setFeaturedId] = useState('')
  const [tileNb, setTileNb] = useState('')
  const [tilePc, setTilePc] = useState('')
  const [q, setQ] = useState('')
  const [busy, setBusy] = useState(false)
  const [inited, setInited] = useState(false)
  const boxRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<any>(null)
  const markerRef = useRef<any>(null)

  useEffect(() => {
    if (inited || !settings) return
    const la = parseFloat(settings.store_lat), lo = parseFloat(settings.store_lng)
    if (!isNaN(la)) setLat(la)
    if (!isNaN(lo)) setLng(lo)
    setFeaturedId(settings.featured_item_id || '')
    setTileNb(settings.tile_notebook_img || '')
    setTilePc(settings.tile_desktop_img || '')
    setInited(true)
  }, [settings, inited])

  // xəritəni qur (settings yüklənəndən sonra bir dəfə)
  useEffect(() => {
    if (!inited || !boxRef.current || mapRef.current || !window.L) return
    const L = window.L
    const map = L.map(boxRef.current).setView([lat, lng], 16)
    L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png', { subdomains: 'abcd', maxZoom: 20, attribution: '© OpenStreetMap · © CARTO' }).addTo(map)
    const icon = L.divIcon({ className: 'store-pin', html: '<span></span>', iconSize: [26, 26], iconAnchor: [13, 26] })
    const marker = L.marker([lat, lng], { icon, draggable: true }).addTo(map)
    marker.on('dragend', () => { const p = marker.getLatLng(); setLat(p.lat); setLng(p.lng) })
    map.on('click', (e: any) => { marker.setLatLng(e.latlng); setLat(e.latlng.lat); setLng(e.latlng.lng) })
    mapRef.current = map; markerRef.current = marker
    setTimeout(() => map.invalidateSize(), 150)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [inited])

  useEffect(() => () => { if (mapRef.current) { mapRef.current.remove(); mapRef.current = null } }, [])

  function pasteLink(url: string) {
    const m = url.match(/@(-?\d+\.\d+),(-?\d+\.\d+)/) || url.match(/(-?\d+\.\d+),\s*(-?\d+\.\d+)/)
    if (!m) { if (url.trim()) toast('Linkdən koordinat tapılmadı — tam Google Maps linki yapışdırın'); return }
    const la = parseFloat(m[1]), lo = parseFloat(m[2])
    setLat(la); setLng(lo)
    if (mapRef.current && markerRef.current) { markerRef.current.setLatLng([la, lo]); mapRef.current.setView([la, lo], 16) }
  }

  async function uploadTile(which: 'nb' | 'pc', file?: File) {
    if (!file) return
    try { const url = await uploadFile(file); if (which === 'nb') setTileNb(url); else setTilePc(url) }
    catch { toast('şəkil yüklənmədi') }
  }

  async function save() {
    setBusy(true)
    try {
      await api('/settings', { method: 'PUT', body: JSON.stringify({ store_lat: String(lat), store_lng: String(lng), featured_item_id: featuredId, tile_notebook_img: tileNb, tile_desktop_img: tilePc }) })
      toast('Tənzimləmələr yadda saxlanıldı'); setTick((t) => t + 1)
    } catch (e) { toast((e as Error).message) } finally { setBusy(false) }
  }

  const list = items ?? []
  const results = q ? list.filter((i) => i.name.toLowerCase().includes(q.toLowerCase()) || i.serial.toLowerCase().includes(q.toLowerCase())).slice(0, 8) : []
  const featuredItem = list.find((i) => String(i.id) === featuredId)

  return (
    <>
      <div className="card pad" style={{ marginBottom: 16 }}>
        <div className="h" style={{ fontSize: 15, marginBottom: 4 }}>Mağazanın xəritədə yeri</div>
        <div className="tiny" style={{ marginBottom: 12 }}>Xəritədə klikləyin və ya pini sürüşdürün — yaxud Google Maps linkini yapışdırın. Müştəri saytında «Mağazaya gəl» bu nöqtəni göstərəcək.</div>
        <div ref={boxRef} style={{ width: '100%', height: 320, borderRadius: 12, overflow: 'hidden', border: '1px solid var(--line)', background: '#EDEDE8' }} />
        <div style={{ display: 'flex', gap: 10, marginTop: 10, flexWrap: 'wrap', alignItems: 'center' }}>
          <input placeholder="Google Maps linkini yapışdır…" onChange={(e) => pasteLink(e.target.value)} style={{ flex: 1, minWidth: 220 }} />
          <span className="tiny data" style={{ color: 'var(--muted)' }}>{lat.toFixed(6)}, {lng.toFixed(6)}</span>
        </div>
      </div>

      <div className="card pad" style={{ marginBottom: 16 }}>
        <div className="h" style={{ fontSize: 15, marginBottom: 4 }}>Ana səhifədə seçilmiş məhsul</div>
        <div className="tiny" style={{ marginBottom: 12 }}>Ana səhifənin yuxarısında böyük göstəriləcək məhsul. Boş olsa — ən yeni məhsul avtomatik seçilir.</div>
        {featuredId
          ? (featuredItem
            ? <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 10 }}><div style={{ fontWeight: 700 }}>{featuredItem.name}</div><button className="btn ghost sm" onClick={() => setFeaturedId('')}>Təmizlə (avtomatik)</button></div>
            : <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 10 }}><div className="tiny" style={{ color: 'var(--warn)' }}>Seçilmiş #{featuredId} — stokda deyil</div><button className="btn ghost sm" onClick={() => setFeaturedId('')}>Təmizlə</button></div>)
          : <div className="tiny" style={{ marginBottom: 10, color: 'var(--muted)' }}>Seçilməyib — avtomatik (ən yeni məhsul)</div>}
        <input placeholder="Məhsul axtar (ad və ya seriya)…" value={q} onChange={(e) => setQ(e.target.value)} style={{ width: '100%' }} />
        {results.length > 0 && (
          <div className="res" style={{ marginTop: 8 }}>
            {results.map((i) => (
              <div key={i.id} className="it" style={{ cursor: 'pointer' }} onClick={() => { setFeaturedId(String(i.id)); setQ('') }}>
                <div><div style={{ fontWeight: 600, fontSize: 13 }}>{i.name}</div><div className="ser">{i.serial || '—'}</div></div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="card pad" style={{ marginBottom: 16 }}>
        <div className="h" style={{ fontSize: 15, marginBottom: 4 }}>Plitə fon şəkilləri</div>
        <div className="tiny" style={{ marginBottom: 12 }}>Ana səhifədəki «Notebooklar» və «Stolüstü PC» plitələrinin arxa fon şəkli. Silsən — sadə tünd kart görünəcək (default illüstrasiya).</div>
        <div className="row" style={{ gap: 14, flexWrap: 'wrap' }}>
          {([['nb', 'Notebooklar plitəsi', tileNb], ['pc', 'Stolüstü PC plitəsi', tilePc]] as const).map(([k, label, val]) => (
            <div key={k} style={{ flex: 1, minWidth: 220 }}>
              <div className="tiny" style={{ fontWeight: 700, marginBottom: 6 }}>{label}</div>
              <label className="uploadbox" style={{ cursor: 'pointer', padding: val ? 8 : 22, display: 'block', background: '#14213A', borderRadius: 10, textAlign: 'center' }}>
                {val
                  ? <img src={imgURL(val)} alt="" style={{ maxHeight: 96, maxWidth: '100%', display: 'block', margin: '0 auto' }} />
                  : <span style={{ color: '#fff', fontSize: 13 }}>Şəkil seç</span>}
                <input type="file" accept="image/*" hidden onChange={(e) => uploadTile(k, e.target.files?.[0])} />
              </label>
              {val && <button className="btn ghost sm" style={{ marginTop: 6, color: 'var(--bad)' }} onClick={() => (k === 'nb' ? setTileNb('') : setTilePc(''))}>Sil (sadə görünüş)</button>}
            </div>
          ))}
        </div>
      </div>

      <button className="btn primary" onClick={save} disabled={busy}>{busy ? 'Saxlanır…' : 'Yadda saxla'}</button>
    </>
  )
}

// ---------------- Dillər ----------------
function LanguagesTab() {
  const toast = useToast()
  const [tick, setTick] = useState(0)
  const { data: langs } = useFetch<Language[]>('/languages', [tick])
  const refresh = () => setTick((t) => t + 1)
  const [code, setCode] = useState('')
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)

  async function add() {
    if (!code.trim() || !name.trim()) { toast('Kod (məs. de) və ad vacibdir'); return }
    setBusy(true)
    try {
      await api('/languages', { method: 'POST', body: JSON.stringify({ code: code.trim(), name: name.trim(), enabled: true }) })
      setCode(''); setName(''); refresh()
    } catch (e) { toast((e as Error).message) } finally { setBusy(false) }
  }
  async function patch(id: number, body: Record<string, unknown>) {
    try { await api('/languages/' + id, { method: 'PUT', body: JSON.stringify(body) }); refresh() }
    catch (e) { toast((e as Error).message) }
  }
  async function del(l: Language) {
    if (l.is_default) { toast('Əsas dili silmək olmaz'); return }
    if (!window.confirm(`«${l.name}» dilini silmək? Bu dilin bütün tərcümələri də silinəcək.`)) return
    try { await api('/languages/' + l.id, { method: 'DELETE' }); refresh() }
    catch (e) { toast((e as Error).message) }
  }
  async function pickIcon(l: Language, file?: File) {
    if (!file) return
    try { const url = await uploadFile(file); await patch(l.id, { icon: url }) }
    catch { toast('Şəkil yüklənmədi') }
  }

  return (
    <>
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><circle cx="12" cy="12" r="9" /><path d="M2.5 12h19M12 2.5a15 15 0 0 1 0 19M12 2.5a15 15 0 0 0 0 19" /></svg><div>Saytda görünəcək dillər. <b>Əsas dil</b> — tərcümə olmayanda müştəri bu dili görür. Hər dilə kiçik bayraq şəkli yükləyin (yoxdursa kod göstərilir).</div></div>

      <div className="card" style={{ overflow: 'hidden', marginTop: 14 }}>
        <table>
          <thead><tr><th>İkon</th><th>Kod</th><th>Ad</th><th>Aktiv</th><th>Əsas</th><th></th></tr></thead>
          <tbody>
            {(langs ?? []).map((l) => (
              <tr key={l.id}>
                <td>
                  <label className="flag-up" title="Bayraq yüklə">
                    {l.icon ? <img src={imgURL(l.icon)} alt="" style={{ width: 26, height: 18, objectFit: 'cover', borderRadius: 3 }} /> : <span className="pill neut" style={{ fontWeight: 800 }}>{l.code.toUpperCase()}</span>}
                    <input type="file" accept="image/*" style={{ display: 'none' }} onChange={(e) => pickIcon(l, e.target.files?.[0])} />
                  </label>
                </td>
                <td className="data">{l.code}</td>
                <td>
                  <input key={l.id + l.name} defaultValue={l.name} onBlur={(e) => { if (e.target.value.trim() && e.target.value !== l.name) patch(l.id, { name: e.target.value.trim() }) }} style={{ maxWidth: 180 }} />
                </td>
                <td>
                  <div className={'tg' + (l.enabled ? ' on' : '')} onClick={() => !l.is_default && patch(l.id, { enabled: !l.enabled })} style={{ opacity: l.is_default ? 0.5 : 1, cursor: l.is_default ? 'not-allowed' : 'pointer' }} />
                </td>
                <td>
                  <input type="radio" name="deflang" checked={l.is_default} onChange={() => patch(l.id, { is_default: true })} style={{ cursor: 'pointer' }} />
                </td>
                <td>
                  {!l.is_default && <button className="btn ghost sm" style={{ color: 'var(--bad)' }} onClick={() => del(l)}>Sil</button>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="card pad" style={{ marginTop: 14, maxWidth: 480 }}>
        <div className="h" style={{ fontSize: 14, marginBottom: 12 }}>Yeni dil əlavə et</div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'flex-end', flexWrap: 'wrap' }}>
          <div className="field" style={{ width: 100 }}><label>Kod</label><input value={code} onChange={(e) => setCode(e.target.value)} placeholder="de" /></div>
          <div className="field" style={{ flex: 1, minWidth: 160 }}><label>Ad (öz dilində)</label><input value={name} onChange={(e) => setName(e.target.value)} placeholder="Deutsch" /></div>
          <button className="btn primary" onClick={add} disabled={busy}>{busy ? 'Əlavə olunur…' : 'Əlavə et'}</button>
        </div>
        <div className="tiny" style={{ marginTop: 8 }}>Kod ISO qısaltmasıdır: az, ru, tr, en, de, fr, ar… Əlavə etdikdən sonra «Sayt mətnləri»ndə tərcümələri doldura (və ya AI ilə avtomatik tərcümə edə) bilərsiniz.</div>
      </div>
    </>
  )
}

// ---------------- Sayt mətnləri ----------------
function TextsTab() {
  const toast = useToast()
  const [tick, setTick] = useState(0)
  const { data } = useFetch<UiStringsData>('/ui-strings', [tick])
  const [edits, setEdits] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState(false)
  const [translating, setTranslating] = useState('')

  const groups = useMemo(() => {
    if (!data) return [] as { section: string; rows: UiStringRow[] }[]
    const m = new Map<string, UiStringRow[]>()
    for (const s of data.strings) {
      if (!m.has(s.section)) m.set(s.section, [])
      m.get(s.section)!.push(s)
    }
    return [...m.entries()].map(([section, rows]) => ({ section, rows }))
  }, [data])

  if (!data) return <div className="card"><div className="center-msg">Yüklənir…</div></div>
  const langs = data.languages.filter((l) => l.enabled)
  const def = data.languages.find((l) => l.is_default)?.code || 'az'

  const valOf = (row: UiStringRow, lang: string) => {
    const k = row.key + '::' + lang
    return edits[k] ?? row.values[lang] ?? ''
  }
  const setVal = (key: string, lang: string, value: string) =>
    setEdits((p) => ({ ...p, [key + '::' + lang]: value }))

  async function save() {
    const entries = Object.entries(edits)
      .map(([k, value]) => { const i = k.indexOf('::'); return { key: k.slice(0, i), lang: k.slice(i + 2), value } })
    if (entries.length === 0) { toast('Dəyişiklik yoxdur'); return }
    setBusy(true)
    try { await api('/ui-strings', { method: 'PUT', body: JSON.stringify({ entries }) }); setEdits({}); setTick((t) => t + 1); toast(`${entries.length} mətn yadda saxlanıldı`) }
    catch (e) { toast((e as Error).message) } finally { setBusy(false) }
  }
  async function translate(to: string) {
    setTranslating(to)
    try {
      const r = await api<{ translated: number; note?: string }>('/ui-strings/translate', { method: 'POST', body: JSON.stringify({ to }) })
      toast(r.note ? r.note : `${r.translated} mətn ${to.toUpperCase()} dilinə tərcümə olundu`)
      setEdits({}); setTick((t) => t + 1)
    } catch (e) { toast((e as Error).message) } finally { setTranslating('') }
  }

  const dirty = Object.keys(edits).length

  return (
    <>
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M4 7V5h16v2M9 5v14M7 19h4" /><path d="M14 12h6M14 16h6M17 12v8" /></svg><div>Saytdakı bütün statik mətnlər. Hər dil üçün ayrıca doldurun. Boş qalan xana avtomatik <b>əsas dilə</b> düşür. <b>«AI ilə tərcümə»</b> əsas dildən doldurur — sonra əl ilə düzəldə bilərsiniz.</div></div>

      <div className="toolbar" style={{ position: 'sticky', top: 0, zIndex: 5, background: 'var(--bg)' }}>
        <div className="h" style={{ marginRight: 'auto' }}>Sayt mətnləri · {data.strings.length} açar</div>
        {langs.filter((l) => l.code !== def).map((l) => (
          <button key={l.code} className="btn ghost sm" disabled={!!translating} onClick={() => translate(l.code)}>
            {translating === l.code ? '…' : `✨ ${l.code.toUpperCase()} AI tərcümə`}
          </button>
        ))}
        <button className="btn primary sm" onClick={save} disabled={busy || dirty === 0}>{busy ? 'Saxlanır…' : `Yadda saxla${dirty ? ` (${dirty})` : ''}`}</button>
      </div>

      {groups.map((g) => (
        <div className="card" key={g.section} style={{ overflow: 'hidden', marginBottom: 14 }}>
          <div style={{ padding: '10px 16px', background: 'var(--surface)', borderBottom: '1px solid var(--line)', fontWeight: 800, fontSize: 13, textTransform: 'uppercase', letterSpacing: '.04em' }}>{g.section}</div>
          <div style={{ overflowX: 'auto' }}>
            <table className="i18n-table">
              <thead><tr><th style={{ minWidth: 150 }}>Açar</th>{langs.map((l) => <th key={l.code} style={{ minWidth: 220 }}>{l.name} {l.is_default && <span className="tiny">(əsas)</span>}</th>)}</tr></thead>
              <tbody>
                {g.rows.map((row) => (
                  <tr key={row.key}>
                    <td className="tiny" style={{ fontFamily: 'var(--mono, monospace)', color: 'var(--muted)', verticalAlign: 'top', paddingTop: 12 }}>{row.key.split('.').slice(1).join('.')}</td>
                    {langs.map((l) => (
                      <td key={l.code} style={{ verticalAlign: 'top' }}>
                        <textarea rows={2} className="i18n-cell" value={valOf(row, l.code)} placeholder={l.code === def ? '' : (row.values[def] || '')}
                          onChange={(e) => setVal(row.key, l.code, e.target.value)} />
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ))}
    </>
  )
}

// ---------------- Kateqoriya & Xüsusiyyət (kataloq terminləri) ----------------
function TermsTab() {
  const toast = useToast()
  const [tick, setTick] = useState(0)
  const { data } = useFetch<TermsData>('/terms', [tick])
  const [edits, setEdits] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState(false)
  const [translating, setTranslating] = useState('')

  if (!data) return <div className="card"><div className="center-msg">Yüklənir…</div></div>
  const langs = data.languages.filter((l) => l.enabled)
  const def = data.languages.find((l) => l.is_default)?.code || 'az'

  const ek = (kind: string, id: number, lang: string) => `${kind}:${id}::${lang}`
  const valOf = (row: TermRow, lang: string) => edits[ek(row.kind, row.id, lang)] ?? row.values[lang] ?? ''
  const setVal = (kind: string, id: number, lang: string, value: string) =>
    setEdits((p) => ({ ...p, [ek(kind, id, lang)]: value }))

  async function save() {
    const entries = Object.entries(edits).map(([k, value]) => {
      const [left, lang] = k.split('::')
      const [kind, id] = left.split(':')
      return { kind, ref_id: Number(id), lang, value }
    })
    if (entries.length === 0) { toast('Dəyişiklik yoxdur'); return }
    setBusy(true)
    try { await api('/terms', { method: 'PUT', body: JSON.stringify({ entries }) }); setEdits({}); setTick((t) => t + 1); toast(`${entries.length} tərcümə yadda saxlanıldı`) }
    catch (e) { toast((e as Error).message) } finally { setBusy(false) }
  }
  async function translate(to: string) {
    setTranslating(to)
    try {
      const r = await api<{ translated: number; note?: string }>('/terms/translate', { method: 'POST', body: JSON.stringify({ to }) })
      toast(r.note ? r.note : `${r.translated} termin ${to.toUpperCase()} dilinə tərcümə olundu`)
      setEdits({}); setTick((t) => t + 1)
    } catch (e) { toast((e as Error).message) } finally { setTranslating('') }
  }

  const dirty = Object.keys(edits).length
  const groups: { label: string; rows: TermRow[] }[] = [
    { label: 'Kateqoriyalar', rows: data.categories },
    { label: 'Xüsusiyyət adları', rows: data.attributes },
    { label: 'Dəyərlər (seçimlər)', rows: data.options },
  ]

  return (
    <>
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M4 7h16M4 12h16M4 17h10" /></svg><div>Kateqoriya adları, xüsusiyyət adları və dəyərlər. Boş qalan avtomatik <b>əsas dilə</b> düşür — markaları (Apple, HP) və ölçüləri (16 GB) tərcümə etmək lazım deyil. <b>«AI tərcümə»</b> markaları toxunmadan doldurur.</div></div>

      <div className="toolbar" style={{ position: 'sticky', top: 0, zIndex: 5, background: 'var(--bg)' }}>
        <div className="h" style={{ marginRight: 'auto' }}>Kataloq terminləri</div>
        {langs.filter((l) => l.code !== def).map((l) => (
          <button key={l.code} className="btn ghost sm" disabled={!!translating} onClick={() => translate(l.code)}>
            {translating === l.code ? '…' : `✨ ${l.code.toUpperCase()} AI tərcümə`}
          </button>
        ))}
        <button className="btn primary sm" onClick={save} disabled={busy || dirty === 0}>{busy ? 'Saxlanır…' : `Yadda saxla${dirty ? ` (${dirty})` : ''}`}</button>
      </div>

      {groups.map((g) => (
        <div className="card" key={g.label} style={{ overflow: 'hidden', marginBottom: 14 }}>
          <div style={{ padding: '10px 16px', background: 'var(--surface)', borderBottom: '1px solid var(--line)', fontWeight: 800, fontSize: 13, textTransform: 'uppercase', letterSpacing: '.04em' }}>{g.label} · {g.rows.length}</div>
          <div style={{ overflowX: 'auto' }}>
            <table className="i18n-table">
              <thead><tr><th style={{ minWidth: 150 }}>Əsas (AZ)</th>{langs.filter((l) => l.code !== def).map((l) => <th key={l.code} style={{ minWidth: 200 }}>{l.name}</th>)}</tr></thead>
              <tbody>
                {g.rows.map((row) => (
                  <tr key={row.id}>
                    <td style={{ verticalAlign: 'top', paddingTop: 12, fontWeight: 600 }}>{row.az}</td>
                    {langs.filter((l) => l.code !== def).map((l) => (
                      <td key={l.code} style={{ verticalAlign: 'top' }}>
                        <input className="i18n-cell" value={valOf(row, l.code)} placeholder={row.az}
                          onChange={(e) => setVal(row.kind, row.id, l.code, e.target.value)} />
                      </td>
                    ))}
                  </tr>
                ))}
                {g.rows.length === 0 && <tr><td colSpan={langs.length} className="tiny" style={{ padding: 14, color: 'var(--muted)' }}>yoxdur</td></tr>}
              </tbody>
            </table>
          </div>
        </div>
      ))}
    </>
  )
}
