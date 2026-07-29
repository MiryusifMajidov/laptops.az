export default function Topbar({ title, sub, onNewSale, search, onSearch, onSearchEnter }: {
  title: string
  sub: string
  onNewSale: () => void
  search: string
  onSearch: (v: string) => void
  onSearchEnter: () => void
}) {
  return (
    <header className="topbar">
      <div>
        <h1>{title}</h1>
        <div className="sub">{sub}</div>
      </div>
      <div className="tb-actions">
        <div className="search">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#B0AEA8" strokeWidth={2} strokeLinecap="round"><circle cx="11" cy="11" r="7" /><path d="M21 21l-4-4" /></svg>
          <input placeholder="Məhsul, seriya…" value={search} onChange={(e) => onSearch(e.target.value)} onKeyDown={(e) => { if (e.key === 'Enter') onSearchEnter() }} />
        </div>
        <button className="btn primary" onClick={onNewSale}>
          <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="#fff" strokeWidth={2.4} strokeLinecap="round"><path d="M12 5v14M5 12h14" /></svg>
          Yeni satış
        </button>
      </div>
    </header>
  )
}
