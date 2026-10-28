import { useState } from 'react'
import type { Product } from '../types'
import { adjustStock, ApiError } from '../api/client'

export function AdjustModal({
  product,
  onClose,
  onSuccess,
}: {
  product: Product
  onClose: () => void
  onSuccess: () => void
}) {
  const [type, setType] = useState<'in' | 'out'>('in')
  const [qty, setQty] = useState('1')
  const [note, setNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [errMsg, setErrMsg] = useState<string | null>(null)

  async function submit() {
    setErrMsg(null)
    const n = Number(qty)
    if (!Number.isInteger(n) || n <= 0) {
      setErrMsg('Qty must be a positive integer.')
      return
    }
    setSaving(true)
    try {
      await adjustStock(product.id, { type, qty: n, note })
      onSuccess()
    } catch (e) {
      if (e instanceof ApiError && e.code === 'INSUFFICIENT_STOCK') {
        const available = e.extras.available
        setErrMsg(`Insufficient stock. Available: ${available}`)
      } else if (e instanceof ApiError) {
        setErrMsg(`${e.code}: ${JSON.stringify(e.extras)}`)
      } else {
        setErrMsg(e instanceof Error ? e.message : 'Failed')
      }
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.4)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
      onClick={onClose}
    >
      <div
        style={{ background: '#fff', padding: 24, borderRadius: 8, width: 360 }}
        onClick={(e) => e.stopPropagation()}
      >
        <h3 style={{ marginTop: 0 }}>Adjust stock — {product.sku}</h3>
        <p style={{ margin: '0 0 12px', color: '#666' }}>
          Current stock: {product.stock}
        </p>

        <label style={{ display: 'block', marginBottom: 8 }}>
          Direction
          <select
            value={type}
            onChange={(e) => setType(e.target.value as 'in' | 'out')}
            style={{ display: 'block', padding: 6, width: '100%' }}
          >
            <option value="in">in</option>
            <option value="out">out</option>
          </select>
        </label>

        <label style={{ display: 'block', marginBottom: 8 }}>
          Qty
          <input
            type="number"
            value={qty}
            onChange={(e) => setQty(e.target.value)}
            style={{ display: 'block', padding: 6, width: '100%' }}
          />
        </label>

        <label style={{ display: 'block', marginBottom: 12 }}>
          Note
          <input
            type="text"
            value={note}
            onChange={(e) => setNote(e.target.value)}
            style={{ display: 'block', padding: 6, width: '100%' }}
          />
        </label>

        {errMsg && <p style={{ color: 'crimson' }}>{errMsg}</p>}

        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
          <button onClick={onClose} disabled={saving}>
            Cancel
          </button>
          <button onClick={submit} disabled={saving}>
            {saving ? 'Saving...' : 'Submit'}
          </button>
        </div>
      </div>
    </div>
  )
}
