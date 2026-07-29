import { createContext, useCallback, useContext, useState, type ReactNode } from 'react'

const ToastCtx = createContext<(m: string) => void>(() => {})
export const useToast = () => useContext(ToastCtx)

export function ToastProvider({ children }: { children: ReactNode }) {
  const [msg, setMsg] = useState('')
  const [show, setShow] = useState(false)

  const toast = useCallback((m: string) => {
    setMsg(m)
    setShow(true)
    window.setTimeout(() => setShow(false), 2600)
  }, [])

  return (
    <ToastCtx.Provider value={toast}>
      {children}
      <div className={'toast' + (show ? ' show' : '')}>
        <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="#5FE0AE" strokeWidth="2.5" strokeLinecap="round"><path d="M20 6L9 17l-5-5" /></svg>
        {msg}
      </div>
    </ToastCtx.Provider>
  )
}
