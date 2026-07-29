import { money, hasDiscount, finalPrice, type Product } from './api'

// Qiymət — endirim varsa: yeni qiymət + üstündən xətli köhnə qiymət
export function Price({ p, className = 'price' }: { p: Product; className?: string }) {
  if (hasDiscount(p)) {
    return (
      <span className="price-wrap">
        <span className={className}>{money(finalPrice(p))}</span>
        <span className="old-price">{money(p.price)}</span>
      </span>
    )
  }
  return <span className={className}>{money(p.price)}</span>
}

// Endirim nişanı (şəkil üzərində) — «−50 ₼»
export function DiscBadge({ p }: { p: Product }) {
  if (!hasDiscount(p)) return null
  return <span className="disc-badge">−{money(p.discount)}</span>
}
