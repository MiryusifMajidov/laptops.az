import { useEffect, useState } from 'react'
import { api } from '../api'

export function useFetch<T>(path: string | null, deps: unknown[] = []) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (path === null) return
    let on = true
    setLoading(true)
    api<T>(path)
      .then((d) => { if (on) { setData(d); setError('') } })
      .catch((e) => { if (on) setError(e.message) })
      .finally(() => { if (on) setLoading(false) })
    return () => { on = false }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps)

  return { data, loading, error }
}
