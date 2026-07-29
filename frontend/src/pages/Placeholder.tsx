export default function Placeholder({ title, note }: { title: string; note: string }) {
  return (
    <div className="content">
      <div className="card" style={{ padding: '48px 30px', textAlign: 'center' }}>
        <div style={{ width: 46, height: 46, borderRadius: 12, background: 'var(--surface)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', marginBottom: 14 }}>
          <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#98A0AC" strokeWidth={1.9} strokeLinecap="round" strokeLinejoin="round"><path d="M12 8v4l3 2" /><circle cx="12" cy="12" r="9" /></svg>
        </div>
        <div className="h" style={{ fontSize: 17 }}>{title}</div>
        <div className="tiny" style={{ maxWidth: 420, margin: '8px auto 0', lineHeight: 1.6 }}>{note}</div>
      </div>
    </div>
  )
}
