export const money = (n: number): string =>
  '₼' + Math.round(n).toLocaleString('en-US')

export const money0 = (n: number): string =>
  Math.round(n).toLocaleString('en-US')

export function dateAz(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '—'
  const months = ['Yan', 'Fev', 'Mar', 'Apr', 'May', 'İyun', 'İyul', 'Avq', 'Sen', 'Okt', 'Noy', 'Dek']
  return `${d.getDate()} ${months[d.getMonth()]}`
}

// neçə gün əvvəl / sonra
export function dueLabel(iso: string): { text: string; tag: string } {
  const d = new Date(iso)
  const days = Math.round((d.getTime() - Date.now()) / 86400000)
  if (days < 0) return { text: `${-days} gün gecikib`, tag: 'bad' }
  if (days === 0) return { text: 'bu gün', tag: 'warn' }
  if (days === 1) return { text: 'sabah', tag: 'neut' }
  return { text: `${days} gün qalıb`, tag: 'neut' }
}

export function ageDays(iso: string): number {
  const d = new Date(iso)
  return Math.round((Date.now() - d.getTime()) / 86400000)
}

export function dateFullAz(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '—'
  const days = ['Bazar', 'Bazar ertəsi', 'Çərşənbə axşamı', 'Çərşənbə', 'Cümə axşamı', 'Cümə', 'Şənbə']
  const months = ['Yanvar', 'Fevral', 'Mart', 'Aprel', 'May', 'İyun', 'İyul', 'Avqust', 'Sentyabr', 'Oktyabr', 'Noyabr', 'Dekabr']
  return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}, ${days[d.getDay()]}`
}

export function dateTimeAz(iso: string): string {
  const d = new Date(iso)
  if (isNaN(d.getTime())) return '—'
  const p = (n: number) => String(n).padStart(2, '0')
  return `${p(d.getDate())}.${p(d.getMonth() + 1)}.${d.getFullYear()} ${p(d.getHours())}:${p(d.getMinutes())}`
}

export function remainLabel(iso: string): { text: string; tag: string } {
  const ms = new Date(iso).getTime() - Date.now()
  if (ms <= 0) return { text: 'vaxt bitdi', tag: 'bad' }
  const h = Math.floor(ms / 3600000)
  const m = Math.floor((ms % 3600000) / 60000)
  const tag = h < 6 ? 'bad' : h < 18 ? 'warn' : 'neut'
  return { text: `${h}s ${m}d qalıb`, tag }
}
