import { useState } from 'react'
import { api, thumbURL, money, type Product } from './api'
import { useI18n } from './i18n'
import { PhIcon } from './ProductCard'

export default function OrderModal({ p, onClose }: { p: Product; onClose: () => void }) {
  const { t, lang } = useI18n()
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [note, setNote] = useState('')
  const [busy, setBusy] = useState(false)
  const [ref, setRef] = useState('')
  const [err, setErr] = useState('')

  async function submit() {
    if (!name || !phone) { setErr(t('order.validation')); return }
    setBusy(true); setErr('')
    try {
      const r = await api<{ ref: string }>('/orders?lang=' + lang, { method: 'POST', body: JSON.stringify({ customer_name: name, phone, item_id: p.id, note }) })
      setRef(r.ref)
    } catch (e) { setErr((e as Error).message) } finally { setBusy(false) }
  }

  return (
    <div className="scrim" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        {ref ? (
          <div className="mbody" style={{ textAlign: 'center', padding: '34px 24px' }}>
            <div style={{ width: 58, height: 58, borderRadius: '50%', background: 'var(--good-bg)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}>
              <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="#0E9F6E" strokeWidth={2.4} strokeLinecap="round" strokeLinejoin="round"><path d="M20 6L9 17l-5-5" /></svg>
            </div>
            <div className="sg" style={{ fontSize: 20, fontWeight: 600, marginTop: 14 }}>{t('order.received')}</div>
            <div style={{ fontSize: 13.5, color: 'var(--ink2)', lineHeight: 1.6, marginTop: 8, maxWidth: 360, margin: '8px auto 0' }}>
              <b className="sg">#{ref}</b> — {t('order.receivedText')}
            </div>
            <button className="btn primary" style={{ marginTop: 20 }} onClick={onClose}>{t('order.close')}</button>
          </div>
        ) : (
          <>
            <div className="mhead">
              <div style={{ fontSize: 16, fontWeight: 800 }}>{t('order.title')}</div>
              <button onClick={onClose}><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#98A0AC" strokeWidth={2.2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg></button>
            </div>
            <div className="mbody">
              <div className="orditem">
                <div className="th">{p.card_image ? <img src={thumbURL(p.card_image, 200)} alt="" loading="lazy" style={{ width: '100%', height: '100%', objectFit: 'cover' }} /> : <PhIcon />}</div>
                <div style={{ flex: 1 }}><div style={{ fontWeight: 700, fontSize: 14 }}>{p.name}</div><div className="muted" style={{ fontSize: 12 }}>{t('order.qty')}</div></div>
                <span className="sg" style={{ fontSize: 17, fontWeight: 600 }}>{money(p.price)}</span>
              </div>
              <div className="field" style={{ marginBottom: 12 }}><label>{t('order.nameLabel')}</label><input value={name} onChange={(e) => setName(e.target.value)} placeholder={t('order.namePlaceholder')} /></div>
              <div className="field" style={{ marginBottom: 12 }}><label>{t('order.phoneLabel')}</label><input className="sg" value={phone} onChange={(e) => setPhone(e.target.value)} placeholder="+994 __ ___ __ __" /></div>
              <div className="field"><label>{t('order.noteLabel')}</label><textarea rows={2} value={note} onChange={(e) => setNote(e.target.value)} placeholder={t('order.notePlaceholder')} /></div>
              {err && <div style={{ color: 'var(--warn)', fontSize: 13, fontWeight: 700, marginTop: 10 }}>{err}</div>}
              <div className="note" style={{ marginTop: 14 }}>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#0A7E57" strokeWidth={2} strokeLinecap="round" style={{ flex: 'none', marginTop: 1 }}><path d="M9 12l2 2 4-4" /><circle cx="12" cy="12" r="9" /></svg>
                {t('order.paymentNote')}
              </div>
              <button className="btn primary" style={{ width: '100%', marginTop: 14 }} onClick={submit} disabled={busy}>{busy ? t('order.submitting') : t('order.submit')}</button>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
