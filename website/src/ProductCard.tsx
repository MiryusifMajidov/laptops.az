import { Link } from 'react-router-dom'
import { thumbURL, specLine, type Product } from './api'
import { useI18n } from './i18n'
import { Price, DiscBadge } from './Price'

export function PhIcon() {
  return <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.4} strokeLinejoin="round"><path d="M4 5h16v10H4z" /><path d="M2 18h20" /></svg>
}

export default function ProductCard({ p }: { p: Product }) {
  const { t, tt } = useI18n()
  return (
    <Link to={`/mehsul/${p.id}`} className="pcard">
      <div className="imgwrap">
        {p.card_image ? <img src={thumbURL(p.card_image, 400)} alt={p.name} loading="lazy" decoding="async" /> : <span className="ph"><PhIcon /></span>}
        <span className="avail good"><span className="dot" />{t('common.available')}</span>
        <DiscBadge p={p} />
      </div>
      <div className="body">
        <div className="name">{p.name}</div>
        <div className="spec">{specLine(p) || tt(p.category?.name ?? '')}</div>
        <div className="row">
          <Price p={p} />
          <span className="btn primary sm">{t('common.details')}</span>
        </div>
      </div>
    </Link>
  )
}
