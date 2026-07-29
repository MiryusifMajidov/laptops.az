import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { fetchLanguages, fetchI18n, fetchTerms, type Language } from './api'

type Dict = Record<string, string>

interface I18nCtx {
  lang: string
  setLang: (code: string) => void
  langs: Language[]
  t: (key: string) => string
  tt: (azText: string) => string // kataloq terminləri (kateqoriya/xüsusiyyət/dəyər)
  ready: boolean
}

const Ctx = createContext<I18nCtx>({
  lang: 'az', setLang: () => {}, langs: [], t: (k) => k, tt: (s) => s, ready: false,
})

export function useI18n() {
  return useContext(Ctx)
}

const stored = () => localStorage.getItem('lang') || ''

export function I18nProvider({ children }: { children: ReactNode }) {
  const [langs, setLangs] = useState<Language[]>([])
  const [lang, setLangState] = useState(stored() || 'az')
  const [dict, setDict] = useState<Dict>({})
  const [terms, setTerms] = useState<Dict>({})
  const [ready, setReady] = useState(false)

  // aktiv dilləri yüklə → seçici üçün + cari dili doğrula
  useEffect(() => {
    fetchLanguages().then((ls) => {
      setLangs(ls)
      const codes = ls.map((l) => l.code)
      const cur = stored()
      if (!cur || !codes.includes(cur)) {
        const def = ls.find((l) => l.is_default) ?? ls[0]
        if (def?.code) setLangState(def.code)
      }
    }).catch(() => {})
  }, [])

  // dil dəyişəndə mətn lüğətini + kataloq terminlərini yenilə
  useEffect(() => {
    if (!lang) return
    localStorage.setItem('lang', lang)
    document.documentElement.lang = lang
    Promise.all([
      fetchI18n(lang).catch(() => ({} as Dict)),
      fetchTerms(lang).catch(() => ({} as Dict)),
    ]).then(([d, tm]) => { setDict(d); setTerms(tm); setReady(true) })
  }, [lang])

  const t = (key: string) => dict[key] ?? key
  const tt = (s: string) => (s ? (terms[s] ?? s) : s)
  const setLang = (code: string) => setLangState(code)

  // ilk lüğət yüklənənə qədər boş saxla (xam açarların görünməməsi üçün)
  if (!ready) return null

  return <Ctx.Provider value={{ lang, setLang, langs, t, tt, ready }}>{children}</Ctx.Provider>
}
