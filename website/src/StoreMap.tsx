import { useEffect, useRef, useState } from 'react'
import { useI18n } from './i18n'
import { onOpenMap } from './mapBus'
import { fetchSettings } from './api'

// İlkin koordinat (admin Site Settings-dən dəyişilir)
const DEFAULT = { lat: 40.3811247, lng: 49.8474406 }
const ZOOM = 16
const PHONE = '+994708151283'

/* eslint-disable @typescript-eslint/no-explicit-any */
declare global {
  interface Window { L: any }
}

export default function StoreMap() {
  const { t } = useI18n()
  const [open, setOpen] = useState(false)
  const [store, setStore] = useState(DEFAULT)
  const boxRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<any>(null)

  useEffect(() => onOpenMap(() => setOpen(true)), [])

  // koordinatı tənzimləmələrdən oxu
  useEffect(() => {
    fetchSettings().then((s) => {
      const lat = parseFloat(s.store_lat), lng = parseFloat(s.store_lng)
      if (!isNaN(lat) && !isNaN(lng)) setStore({ lat, lng })
    }).catch(() => {})
  }, [])

  // xəritəni modal açılanda qur
  useEffect(() => {
    if (!open || !boxRef.current || mapRef.current) return
    const L = window.L
    if (!L) return
    const map = L.map(boxRef.current, { zoomControl: true, scrollWheelZoom: false })
      .setView([store.lat, store.lng], ZOOM)
    // minimal açıq üslub (Carto Positron) — saytın təmiz dizaynına uyğun
    L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}.png', {
      subdomains: 'abcd', maxZoom: 20,
      attribution: '© OpenStreetMap · © CARTO',
    }).addTo(map)
    const icon = L.divIcon({ className: 'store-pin', html: '<span></span>', iconSize: [26, 26], iconAnchor: [13, 26] })
    L.marker([store.lat, store.lng], { icon }).addTo(map)
    mapRef.current = map
    setTimeout(() => map.invalidateSize(), 120) // modal ölçüsü oturandan sonra
  }, [open, store])

  // bağlananda təmizlə (yenidən açılanda düzgün qurulsun)
  useEffect(() => {
    if (!open && mapRef.current) { mapRef.current.remove(); mapRef.current = null }
  }, [open])

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false) }
    if (open) window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open])

  if (!open) return null
  const directions = `https://www.google.com/maps/dir/?api=1&destination=${store.lat},${store.lng}`

  return (
    <div className="map-scrim" onClick={() => setOpen(false)}>
      <div className="map-modal" onClick={(e) => e.stopPropagation()}>
        <div className="map-head">
          <div style={{ minWidth: 0 }}>
            <div className="map-title">{t('footer.storeLabel')}</div>
            <div className="map-sub" style={{ whiteSpace: 'pre-line' }}>{t('footer.address')}</div>
          </div>
          <button className="map-close" onClick={() => setOpen(false)} aria-label="Bağla">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.2} strokeLinecap="round"><path d="M6 6l12 12M18 6L6 18" /></svg>
          </button>
        </div>
        <div className="map-canvas" ref={boxRef} />
        <div className="map-foot">
          <a className="btn primary" href={directions} target="_blank" rel="noopener noreferrer">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round"><path d="M3 11l19-9-9 19-2-8-8-2z" /></svg>
            {t('map.directions')}
          </a>
          <a className="btn ghost" href={`tel:${PHONE}`}>{t('common.call')}</a>
        </div>
      </div>
    </div>
  )
}
