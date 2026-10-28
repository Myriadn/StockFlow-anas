export interface Product {
  id: number;
  sku: string;
  name: string;
  price: number;
  stock: number;
  is_active: boolean;
  last_movement_at: string | null;
}

export interface ListProductsResponse {
  data: Product[];
  page: number;
  limit: number;
  total: number;
}

export interface AdjustStockRequest {
  type: 'in' | 'out';
  qty: number;
  note: string;
}