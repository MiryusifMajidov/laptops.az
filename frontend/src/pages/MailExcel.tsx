import { useState } from 'react'
import { api, downloadFile, type EmailTemplate } from '../api'
import { useFetch } from '../lib/hooks'
import { useToast } from '../lib/toast'
import { useRefresh } from '../lib/refresh'
import FormModal from '../components/FormModal'

export default function MailExcel() {
  const toast = useToast()
  const { key, bump } = useRefresh()
  const { data: tpls } = useFetch<EmailTemplate[]>('/email-templates', [key])
  const [sel, setSel] = useState<number[]>([])
  const [form, setForm] = useState(false)
  const [edit, setEdit] = useState<EmailTemplate | null>(null)
  const [sending, setSending] = useState(false)
  const list = tpls ?? []

  const toggle = (id: number) => setSel((s) => (s.includes(id) ? s.filter((x) => x !== id) : [...s, id]))

  async function send() {
    if (!sel.length) { toast('Ən azı bir şablon seç'); return }
    setSending(true)
    try {
      const r = await api<{ sent: boolean; recipients: string[]; note?: string }>('/email/send', { method: 'POST', body: JSON.stringify({ template_ids: sel }) })
      if (r.sent) toast('Excel göndərildi: ' + r.recipients.join(', '))
      else toast(r.note || 'SMTP hələ qoşulmayıb')
    } catch (e) { toast((e as Error).message) } finally { setSending(false) }
  }

  async function del(id: number) {
    if (!window.confirm('Şablon silinsin?')) return
    try { await api(`/email-templates/${id}`, { method: 'DELETE' }); bump(); setSel((s) => s.filter((x) => x !== id)) } catch (e) { toast((e as Error).message) }
  }

  return (
    <div className="content">
      <div className="banner"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><rect x="3" y="5" width="18" height="14" rx="2" /><path d="M3 7l9 6 9-6" /></svg><div>Hər gün mağaza bağlananda <b>cari stokun Excel cədvəlini</b> seçdiyin mail şablonlarına göndər (özümüzə + digər mağazalara). Şablon = ad + ünvanlar. <b>Excel avtomatik yaradılır.</b> <i>(Real göndərmə üçün serverdə SMTP qoşulmalıdır — mexanizm hazırdır.)</i></div></div>

      <div className="row">
        <div className="card pad" style={{ flex: 1 }}>
          <div className="eyebrow">Cari stok Excel</div>
          <div className="tiny" style={{ margin: '4px 0 12px' }}>indiki stokdakı bütün cihazlar</div>
          <button className="btn ghost" onClick={() => downloadFile('/export/stock', 'laptops-cari-stok.xlsx').catch(() => toast('endirilmədi'))}>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round"><path d="M12 3v12M7 10l5 5 5-5M4 21h16" /></svg>Excel yüklə
          </button>
        </div>
        <div className="card pad" style={{ flex: 2 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div><div className="eyebrow">Seçilmiş şablonlara göndər</div><div className="tiny" style={{ marginTop: 4 }}>{sel.length} şablon seçildi</div></div>
            <button className="btn primary" onClick={send} disabled={sending}>{sending ? 'Göndərilir…' : 'Excel-i göndər'}</button>
          </div>
        </div>
      </div>

      <div className="card" style={{ overflow: 'hidden' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '16px 20px 6px' }}><div className="h">Mail şablonları</div><button className="btn ghost sm" onClick={() => setForm(true)}>+ Yeni şablon</button></div>
        <table>
          <thead><tr><th style={{ width: 40 }}></th><th>Şablon</th><th>Ünvanlar</th><th></th></tr></thead>
          <tbody>
            {list.map((t) => (
              <tr key={t.id}>
                <td><div className={'tg' + (sel.includes(t.id) ? ' on' : '')} onClick={() => toggle(t.id)} /></td>
                <td className="prod">{t.name}</td>
                <td className="tiny">{t.emails}</td>
                <td className="tright"><button className="btn ghost sm" onClick={() => setEdit(t)}>Redaktə</button> <button className="btn ghost sm" style={{ color: 'var(--bad)' }} onClick={() => del(t.id)}>Sil</button></td>
              </tr>
            ))}
            {list.length === 0 && <tr><td colSpan={4} className="center-msg">Şablon yoxdur</td></tr>}
          </tbody>
        </table>
      </div>

      {form && (
        <FormModal
          title="Yeni mail şablonu" subtitle="ünvanları vergüllə ayır" submitLabel="Əlavə et" onClose={() => setForm(false)}
          fields={[
            { name: 'name', label: 'Şablon adı', required: true, full: true, placeholder: 'məs. Digər filiallar' },
            { name: 'emails', label: 'Mail ünvanları (vergüllə)', full: true, placeholder: 'a@mail.com, b@mail.com' },
          ]}
          onSubmit={async (v) => { await api('/email-templates', { method: 'POST', body: JSON.stringify(v) }); bump(); toast('Şablon əlavə edildi'); setForm(false) }}
        />
      )}
      {edit && (
        <FormModal
          title="Şablonu redaktə et" submitLabel="Yadda saxla" onClose={() => setEdit(null)}
          initial={{ name: edit.name, emails: edit.emails }}
          fields={[
            { name: 'name', label: 'Şablon adı', required: true, full: true },
            { name: 'emails', label: 'Mail ünvanları (vergüllə)', full: true },
          ]}
          onSubmit={async (v) => { await api(`/email-templates/${edit.id}`, { method: 'PUT', body: JSON.stringify(v) }); bump(); toast('Şablon yeniləndi'); setEdit(null) }}
        />
      )}
    </div>
  )
}
