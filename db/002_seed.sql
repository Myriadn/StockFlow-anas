INSERT INTO products (sku, name, price, stock, is_active) VALUES
  ('KOPI-ARB-1K',   'Kopi Arabica 1kg',     150000,  48, TRUE),
  ('GULA-PSR-1K',   'Gula Pasir 1kg',        18000,  38, TRUE),
  ('TEH-CLP-25',    'Teh Celup 25pcs',       12000,  25, TRUE),
  ('BERAS-PRM-5K',  'Beras Premium 5kg',     75000,  60, TRUE),
  ('MINYAK-GRG-2L', 'Minyak Goreng 2L',      35000,  45, TRUE),
  ('TEPUNG-TRG-1K', 'Tepung Terigu 1kg',     14000,  80, TRUE),
  ('SUSU-UB-1L',    'Susu UHT 1L',           22000,  30, TRUE),
  ('TELUR-AYM-1K',  'Telur Ayam 1kg',        28000,  25, TRUE),
  ('SABUN-MND-100', 'Sabun Mandi 100g',       5000, 120, TRUE),
  ('SHAMPOO-200',   'Shampoo 200ml',         32000,  40, TRUE),
  ('PASTA-GIG-150', 'Pasta Gigi 150g',       18000,  55, TRUE),
  ('DETERJEN-1K',   'Deterjen 1kg',          25000,  35, TRUE),
  ('AIR-MIN-600',   'Air Mineral 600ml',      4000, 200, TRUE),
  ('SNACK-KRIP-50', 'Keripik Kentang 50g',   12000,   0, FALSE),  -- inactive
  ('MADU-MRN-250',  'Madu Murni 250ml',      85000,   8, TRUE);   -- low stock

-- Movements: "in" = stok awal. Untuk produk yang sudah kepotong order, tambah "out".
INSERT INTO stock_movements (product_id, type, qty, note, created_at)
SELECT id, 'in', 50, 'Initial stock', now() - interval '7 days' FROM products WHERE sku = 'KOPI-ARB-1K'
UNION ALL SELECT id, 'out', 2, 'Order #1', now() - interval '2 hours' FROM products WHERE sku = 'KOPI-ARB-1K'
UNION ALL SELECT id, 'in', 40, 'Initial stock', now() - interval '6 days' FROM products WHERE sku = 'GULA-PSR-1K'
UNION ALL SELECT id, 'out', 2, 'Order #1', now() - interval '2 hours' FROM products WHERE sku = 'GULA-PSR-1K'
UNION ALL SELECT id, 'in', 25, 'Initial stock', now() - interval '5 days' FROM products WHERE sku = 'TEH-CLP-25'
UNION ALL SELECT id, 'in', 60, 'Initial stock', now() - interval '10 days' FROM products WHERE sku = 'BERAS-PRM-5K'
UNION ALL SELECT id, 'in', 45, 'Initial stock', now() - interval '10 days' FROM products WHERE sku = 'MINYAK-GRG-2L'
UNION ALL SELECT id, 'in', 80, 'Initial stock', now() - interval '10 days' FROM products WHERE sku = 'TEPUNG-TRG-1K'
UNION ALL SELECT id, 'in', 30, 'Initial stock', now() - interval '4 days' FROM products WHERE sku = 'SUSU-UB-1L'
UNION ALL SELECT id, 'in', 25, 'Initial stock', now() - interval '4 days' FROM products WHERE sku = 'TELUR-AYM-1K'
UNION ALL SELECT id, 'in', 120, 'Initial stock', now() - interval '20 days' FROM products WHERE sku = 'SABUN-MND-100'
UNION ALL SELECT id, 'in', 40, 'Initial stock', now() - interval '3 days' FROM products WHERE sku = 'SHAMPOO-200'
UNION ALL SELECT id, 'in', 55, 'Initial stock', now() - interval '3 days' FROM products WHERE sku = 'PASTA-GIG-150'
UNION ALL SELECT id, 'in', 35, 'Initial stock', now() - interval '8 days' FROM products WHERE sku = 'DETERJEN-1K'
UNION ALL SELECT id, 'in', 200, 'Initial stock', now() - interval '30 days' FROM products WHERE sku = 'AIR-MIN-600'
UNION ALL SELECT id, 'in', 8, 'Initial stock', now() - interval '15 days' FROM products WHERE sku = 'MADU-MRN-250';

INSERT INTO sales_orders (customer, status, total, created_at) VALUES
  ('Toko Maju Jaya', 'CREATED',   336000, now() - interval '2 hours'),
  ('CV Sentosa',     'CANCELLED',  36000, now() - interval '1 day');

INSERT INTO sales_order_items (order_id, sku, qty, unit_price)
SELECT o.id, 'KOPI-ARB-1K', 2, 150000 FROM sales_orders o WHERE o.customer = 'Toko Maju Jaya'
UNION ALL SELECT o.id, 'GULA-PSR-1K', 2, 18000 FROM sales_orders o WHERE o.customer = 'Toko Maju Jaya'
UNION ALL SELECT o.id, 'TEH-CLP-25',  3, 12000 FROM sales_orders o WHERE o.customer = 'CV Sentosa';

UPDATE sales_orders SET cancelled_at = now() - interval '12 hours' WHERE customer = 'CV Sentosa';