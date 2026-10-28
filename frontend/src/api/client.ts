import type { ListProductsResponse, AdjustStockRequest } from '../types'

const BASE = '/api'

export class ApiError extends Error {
  status: number
  code: string
  extras: Record<string, unknown>

  constructor(status: number, code: string, extras: Record<string, unknown>) {
    super(code)
    this.status = status
    this.code = code
    this.extras = extras
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })

  if (!res.ok) {
    let code = 'HTTP_ERROR'
    let extras: Record<string, unknown> = {}
    try {
      const body = await res.json()
      if (body?.error) {
        const { code: c, ...rest } = body.error
        code = c ?? code
        extras = rest
      }
    } catch {
      // body not JSON
      // leave it default
    }
    throw new ApiError(res.status, code, extras)
  }

  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export function listProducts(
  params: { search?: string; page?: number; limit?: number } = {},
) {
  const qs = new URLSearchParams()
  if (params.search) qs.set('search', params.search)
  if (params.page) qs.set('page', String(params.page))
  if (params.limit) qs.set('limit', String(params.limit))
  const suffix = qs.toString() ? `?${qs.toString()}` : ''
  return request<ListProductsResponse>(`/products${suffix}`)
}

export function adjustStock(productId: number, body: AdjustStockRequest) {
  return request<{
    product_id: number
    stock: number
    type: string
    qty: number
  }>(`/products/${productId}/stock`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}
