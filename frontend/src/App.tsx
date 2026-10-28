import { useEffect, useState } from 'react'
import { listProducts } from './api/client'
import type { Product } from './types'
import { AdjustModal } from './components/AdjustModal'

export default function App() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [adjustTarget, setAdjustTarget] = useState<Product | null>(null)

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const resp = await listProducts({ search, page: 1, limit: 50 })
      setProducts(resp.data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load products')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [search])

  return (
    <div
      style={{
        fontFamily: 'sans-serif',
        maxWidth: 960,
        margin: '24px auto',
        padding: '0 16px',
      }}
    >
      <h1>StockFlow</h1>

      <input
        type="search"
        placeholder="Search SKU or name..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        style={{ padding: 8, width: 300, marginBottom: 16 }}
      />

      {loading && <p>Loading...</p>}
      {error && !loading && (
        <div style={{ color: 'crimson', marginBottom: 12 }}>
          {error} <button onClick={load}>Retry</button>
        </div>
      )}

      {!loading && !error && products.length === 0 && <p>No products found.</p>}

      {!loading && !error && products.length > 0 && (
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ textAlign: 'left', borderBottom: '1px solid #ccc' }}>
              <th>SKU</th>
              <th>Name</th>
              <th style={{ textAlign: 'right' }}>Price</th>
              <th style={{ textAlign: 'right' }}>Stock</th>
              <th>Status</th>
              <th>Last movement</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {products.map((p) => (
              <tr key={p.id} style={{ borderBottom: '1px solid #eee' }}>
                <td>{p.sku}</td>
                <td>{p.name}</td>
                <td style={{ textAlign: 'right' }}>
                  Rp {p.price.toLocaleString('id-ID')}
                </td>
                <td style={{ textAlign: 'right' }}>{p.stock}</td>
                <td>{p.is_active ? 'Active' : 'Inactive'}</td>
                <td>
                  {p.last_movement_at
                    ? new Date(p.last_movement_at).toLocaleString()
                    : 'None'}
                </td>
                <td>
                  <button onClick={() => setAdjustTarget(p)}>Adjust</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {adjustTarget && (
        <AdjustModal
          product={adjustTarget}
          onClose={() => setAdjustTarget(null)}
          onSuccess={() => {
            setAdjustTarget(null)
            load()
          }}
        />
      )}
    </div>
  )
}
