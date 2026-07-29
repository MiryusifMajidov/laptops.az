import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { aiChat, thumbURL, money, specLine, hasDiscount, finalPrice, type AiChatMsg, type Product } from './api'
import { useI18n } from './i18n'
import { onOpenAi } from './aiBus'

const Spark = ({ size = 15 }: { size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="#4C86FF"><path d="M12 2.5l2.2 5.8 5.8 2.2-5.8 2.2L12 18.5l-2.2-5.8L4 10.5l5.8-2.2z" /></svg>
)

// [etiket açarı, sual açarı]
const QUICK: [string, string][] = [
  ['ai.quickGameLabel', 'ai.quickGameQ'],
  ['ai.quickOfficeLabel', 'ai.quickOfficeQ'],
  ['ai.quickStudentLabel', 'ai.quickStudentQ'],
  ['ai.quickDesignLabel', 'ai.quickDesignQ'],
]

type Msg = { role: 'user' | 'assistant'; text: string; products?: Record<string, Product> }

function ProdCard({ p, onClick, available }: { p: Product; onClick: () => void; available: string }) {
  return (
    <button className="ai-prod" onClick={onClick}>
      <div className="ai-prod-img">{p.card_image ? <img src={thumbURL(p.card_image, 200)} alt={p.name} loading="lazy" /> : <span className="ph" />}</div>
      <div className="ai-prod-info">
        <div className="ai-prod-name">{p.name}</div>
        <div className="ai-prod-spec">{specLine(p) || p.category?.name}</div>
        <div className="ai-prod-row">
          <span className="sg ai-prod-price">{money(finalPrice(p))}{hasDiscount(p) && <span className="old-price" style={{ marginLeft: 5 }}>{money(p.price)}</span>}</span>
          <span className="ai-prod-avail"><span className="dot" />{available}</span>
        </div>
      </div>
    </button>
  )
}

function Bubble({ m, onProduct, available }: { m: Msg; onProduct: (id: string) => void; available: string }) {
  if (m.role === 'user') {
    return <div className="ai-msg ai-out"><div className="ai-user">{m.text}</div></div>
  }
  const parts = m.text.split(/(\[\[product:\d+\]\])/g)
  return (
    <div className="ai-msg ai-in">
      <div className="ai-mini"><Spark size={14} /></div>
      <div className="ai-bubble-wrap">
        {parts.map((part, i) => {
          const mt = part.match(/\[\[product:(\d+)\]\]/)
          if (mt) {
            const p = m.products?.[mt[1]]
            return p ? <ProdCard key={i} p={p} onClick={() => onProduct(String(p.id))} available={available} /> : null
          }
          const txt = part
            .replace(/\*\*/g, '')
            .replace(/^#{1,6}\s*/gm, '')
            .replace(/^\s*[-*]\s+/gm, '• ')
            .trim()
          return txt ? <div className="ai-bubble" key={i}>{txt}</div> : null
        })}
      </div>
    </div>
  )
}

export default function AiAssistant() {
  const { t, lang } = useI18n()
  const [open, setOpen] = useState(false)
  // Söhbət sessiya boyu qalır (səhifə yenilənsə də) — sessiya bitəndə (tab bağlananda) silinir.
  const [msgs, setMsgs] = useState<Msg[]>(() => {
    try {
      const raw = sessionStorage.getItem('ai_chat')
      return raw ? (JSON.parse(raw) as Msg[]) : []
    } catch { return [] }
  })
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const bodyRef = useRef<HTMLDivElement>(null)
  const nav = useNavigate()

  useEffect(() => onOpenAi(() => setOpen(true)), [])
  useEffect(() => {
    try { sessionStorage.setItem('ai_chat', JSON.stringify(msgs)) } catch { /* storage dolu ola bilər */ }
  }, [msgs])
  useEffect(() => { bodyRef.current?.scrollTo({ top: bodyRef.current.scrollHeight, behavior: 'smooth' }) }, [msgs, loading, open])

  const send = async (text: string) => {
    const t2 = text.trim()
    if (!t2 || loading) return
    setInput('')
    const next: Msg[] = [...msgs, { role: 'user', text: t2 }]
    setMsgs(next)
    setLoading(true)
    try {
      const history: AiChatMsg[] = next.map((m) => ({ role: m.role, text: m.text }))
      const r = await aiChat(history, lang)
      setMsgs((m) => [...m, { role: 'assistant', text: r.reply, products: r.products }])
    } catch (e) {
      // backend mesajı varsa (limit və s.) onu göstər; şəbəkə xətasında generic
      const err = e as Error & { status?: number }
      setMsgs((m) => [...m, { role: 'assistant', text: err.status ? err.message : t('ai.error') }])
    } finally {
      setLoading(false)
    }
  }

  const goProduct = (id: string) => { setOpen(false); nav('/mehsul/' + id) }
  const available = t('common.available')

  return (
    <>
      <button className="ai-fab" onClick={() => setOpen(true)} aria-label={t('ai.button')}>
        <span className="ai-fab-ic"><Spark size={19} /><span className="ai-dot" /></span>
        <span className="ai-fab-tx">{t('ai.button')}</span>
      </button>

      {open && (
        <div className="ai-scrim" onClick={() => setOpen(false)}>
          <div className="ai-panel" onClick={(e) => e.stopPropagation()}>
            <div className="ai-head">
              <div className="ai-ava"><Spark size={21} /><span className="ai-dot" /></div>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div className="ai-title">{t('ai.button')}</div>
                <div className="ai-status">{t('ai.status')}</div>
              </div>
              <button onClick={() => setOpen(false)} aria-label="Bağla" className="ai-close">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#98A0AC" strokeWidth={2.2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
              </button>
            </div>

            <div className="ai-body" ref={bodyRef}>
              <Bubble m={{ role: 'assistant', text: t('ai.welcome') }} onProduct={goProduct} available={available} />
              {msgs.map((m, i) => <Bubble key={i} m={m} onProduct={goProduct} available={available} />)}
              {msgs.length === 0 && !loading && (
                <div className="ai-quick">
                  {QUICK.map(([labelKey, qKey]) => <button key={labelKey} onClick={() => send(t(qKey))}>{t(labelKey)}</button>)}
                </div>
              )}
              {loading && (
                <div className="ai-msg ai-in">
                  <div className="ai-mini"><Spark size={14} /></div>
                  <div className="ai-bubble ai-typing"><span /><span /><span /></div>
                </div>
              )}
            </div>

            <form className="ai-input" onSubmit={(e) => { e.preventDefault(); send(input) }}>
              <input value={input} onChange={(e) => setInput(e.target.value)} placeholder={t('ai.inputPlaceholder')} />
              <button type="submit" aria-label="Göndər" disabled={loading || !input.trim()}>
                <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth={2.2} strokeLinecap="round" strokeLinejoin="round"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4z" /></svg>
              </button>
            </form>
          </div>
        </div>
      )}
    </>
  )
}
