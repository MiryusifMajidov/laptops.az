import { createContext, useCallback, useContext, useState, type ReactNode } from 'react'

// Sadə refresh mexanizmi: satış/məhsul əlavə olunanda səhifələr yenidən yüklənsin
const RefreshCtx = createContext<{ key: number; bump: () => void }>({ key: 0, bump: () => {} })
export const useRefresh = () => useContext(RefreshCtx)

export function RefreshProvider({ children }: { children: ReactNode }) {
  const [key, setKey] = useState(0)
  const bump = useCallback(() => setKey((k) => k + 1), [])
  return <RefreshCtx.Provider value={{ key, bump }}>{children}</RefreshCtx.Provider>
}
