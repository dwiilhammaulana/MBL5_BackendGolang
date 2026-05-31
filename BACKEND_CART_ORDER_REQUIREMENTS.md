# Backend Requirements: Cart, Checkout, Payment, Order

## Context

Frontend Flutter sudah menambahkan fitur keranjang, checkout, pembayaran, dan daftar pesanan. Saat ini backend hanya merespons endpoint produk, sedangkan endpoint cart/order masih `404 Not Found`.

Backend menggunakan Golang Gin, GORM, MySQL, Firebase auth, dan route base:

```text
http://<host>:8081/v1
```

Contoh hasil pengecekan saat ini:

```text
GET  /v1/health            -> 200 OK
GET  /v1/products          -> 401 Unauthorized, route ada
POST /v1/cart              -> 404 Not Found
POST /v1/cart_items        -> 404 Not Found
GET  /v1/orders            -> 404 Not Found
POST /v1/orders/checkout   -> 404 Not Found
```

Jadi kebutuhan utamanya adalah menambahkan API backend untuk cart dan order agar sesuai dengan frontend Flutter.

## Existing Database Tables

Database sudah memiliki tabel berikut:

```text
users
products
cart_items
orders
order_items
payments
```

Kolom penting:

```text
cart_items:
- id
- user_id
- product_id
- quantity
- created_at
- updated_at

orders:
- id
- user_id
- total_amount
- status
- shipping_address
- notes
- payment_method
- payment_status
- va_number
- gopay_deeplink
- paid_at
- created_at
- updated_at

order_items:
- id
- order_id
- product_id
- product_name
- price
- quantity
- subtotal
- created_at

payments:
- id
- order_id
- payment_method
- payment_status
- amount
- va_number
- gopay_deeplink
- paid_at
- created_at
- updated_at

products:
- id
- name
- description
- price
- stock
- category
- image_url
- is_active
```

## Required Backend Files

Tambahkan file atau struktur setara sesuai pattern backend yang sudah ada:

```text
models/cart_item.go
models/order.go
models/order_item.go
models/payment.go

repositories/cart_repository.go
repositories/order_repository.go

services/cart_service.go
services/order_service.go

handlers/cart_handler.go
handlers/order_handler.go
```

Daftarkan handler baru di:

```text
routes/router.go
```

## Required Routes

Semua endpoint berikut harus berada di bawah `/v1` dan memakai Bearer Token auth seperti endpoint `/v1/products`.

```text
GET    /v1/cart
POST   /v1/cart
PUT    /v1/cart/:id
DELETE /v1/cart/:id
DELETE /v1/cart

POST   /v1/orders/checkout
GET    /v1/orders
GET    /v1/orders/:id
```

Contoh route Gin:

```go
cartHandler := handlers.NewCartHandler(...)
orderHandler := handlers.NewOrderHandler(...)

protected := v1.Group("")
protected.Use(middleware.AuthMiddleware())

cart := protected.Group("/cart")
{
    cart.GET("", cartHandler.GetCart)
    cart.POST("", cartHandler.AddToCart)
    cart.PUT("/:id", cartHandler.UpdateItem)
    cart.DELETE("/:id", cartHandler.RemoveItem)
    cart.DELETE("", cartHandler.ClearCart)
}

orders := protected.Group("/orders")
{
    orders.POST("/checkout", orderHandler.Checkout)
    orders.GET("", orderHandler.GetMyOrders)
    orders.GET("/:id", orderHandler.GetOrderDetail)
}
```

Sesuaikan constructor handler/service/repository dengan dependency injection yang sudah dipakai di project backend.

## API Contracts

### GET /v1/cart

Ambil isi keranjang user yang sedang login.

Response:

```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "product_id": 3,
        "product": {
          "id": 3,
          "name": "APAR 3kg",
          "price": 300000,
          "image_url": "https://example.com/apar.png",
          "category": "APAR"
        },
        "quantity": 2,
        "subtotal": 600000
      }
    ],
    "total": 600000,
    "item_count": 2
  }
}
```

Catatan:
- `subtotal` = `product.price * quantity`.
- `total` = jumlah semua subtotal.
- `item_count` = jumlah semua quantity, bukan jumlah baris item.

### POST /v1/cart

Tambah produk ke keranjang.

Request:

```json
{
  "product_id": 1,
  "quantity": 2
}
```

Expected behavior:
- Jika item produk belum ada di cart user, buat row baru di `cart_items`.
- Jika item produk sudah ada, tambahkan quantity ke row yang ada.
- Validasi product harus ada dan aktif.
- Validasi stock cukup.

Response minimal:

```json
{
  "message": "Produk ditambahkan ke keranjang"
}
```

### PUT /v1/cart/:id

Update quantity item cart berdasarkan `cart_items.id`.

Request:

```json
{
  "quantity": 3
}
```

Expected behavior:
- Hanya boleh update item milik user yang sedang login.
- Quantity minimal `1`.
- Validasi stock cukup.

Response minimal:

```json
{
  "message": "Keranjang diperbarui"
}
```

### DELETE /v1/cart/:id

Hapus satu item cart berdasarkan `cart_items.id`.

Expected behavior:
- Hanya boleh hapus item milik user yang sedang login.

Response minimal:

```json
{
  "message": "Item dihapus dari keranjang"
}
```

### DELETE /v1/cart

Kosongkan seluruh keranjang user yang sedang login.

Response minimal:

```json
{
  "message": "Keranjang dikosongkan"
}
```

### POST /v1/orders/checkout

Buat order dari semua item cart user.

Request:

```json
{
  "shipping_address": "Jl. Pasar Malam No. 1",
  "notes": "Catatan opsional",
  "payment_method": "virtual_account"
}
```

Allowed `payment_method`:

```text
gopay
bank_transfer
virtual_account
```

Expected behavior:
- Validasi cart tidak kosong.
- Validasi stock semua produk cukup.
- Buat row `orders`.
- Buat row `order_items` dari snapshot cart:
  - `product_id`
  - `product_name`
  - `price`
  - `quantity`
  - `subtotal`
- Buat row `payments`.
- Kurangi stock produk.
- Kosongkan cart user setelah checkout sukses.
- Gunakan transaction database agar semua proses atomic.

Status default:

```text
orders.status = pending
orders.payment_status = pending
payments.payment_status = pending
```

Payment data:
- Jika `payment_method = virtual_account`, generate `va_number`.
- Jika `payment_method = gopay`, generate atau isi `gopay_deeplink`.
- Jika belum ada payment gateway asli, boleh pakai dummy value yang konsisten untuk development.

Response:

```json
{
  "data": {
    "id": 10,
    "total_amount": 600000,
    "status": "pending",
    "shipping_address": "Jl. Pasar Malam No. 1",
    "notes": "Catatan opsional",
    "payment_method": "virtual_account",
    "payment_status": "pending",
    "va_number": "880812345678",
    "gopay_deeplink": null,
    "paid_at": null,
    "items": [
      {
        "product_id": 3,
        "product_name": "APAR 3kg",
        "price": 300000,
        "quantity": 2,
        "subtotal": 600000
      }
    ],
    "created_at": "2026-05-31T10:00:00Z"
  }
}
```

### GET /v1/orders

Ambil daftar order user yang sedang login.

Query optional:

```text
page=1
limit=10
```

Response boleh salah satu dari dua bentuk berikut.

Bentuk sederhana:

```json
{
  "data": [
    {
      "id": 10,
      "total_amount": 600000,
      "status": "pending",
      "shipping_address": "Jl. Pasar Malam No. 1",
      "notes": "Catatan opsional",
      "payment_method": "virtual_account",
      "payment_status": "pending",
      "va_number": "880812345678",
      "gopay_deeplink": null,
      "paid_at": null,
      "items": [],
      "created_at": "2026-05-31T10:00:00Z"
    }
  ]
}
```

Atau bentuk pagination:

```json
{
  "data": {
    "items": [],
    "page": 1,
    "limit": 10,
    "total": 1
  }
}
```

Frontend sudah bisa membaca keduanya.

### GET /v1/orders/:id

Ambil detail satu order user.

Expected behavior:
- Hanya boleh mengambil order milik user yang sedang login.
- Sertakan `items`.
- Sertakan payment fields: `payment_status`, `va_number`, `gopay_deeplink`, `paid_at`.

Response:

```json
{
  "data": {
    "id": 10,
    "total_amount": 600000,
    "status": "pending",
    "shipping_address": "Jl. Pasar Malam No. 1",
    "notes": "Catatan opsional",
    "payment_method": "virtual_account",
    "payment_status": "pending",
    "va_number": "880812345678",
    "gopay_deeplink": null,
    "paid_at": null,
    "items": [
      {
        "product_id": 3,
        "product_name": "APAR 3kg",
        "price": 300000,
        "quantity": 2,
        "subtotal": 600000
      }
    ],
    "created_at": "2026-05-31T10:00:00Z"
  }
}
```

## Response Field Compatibility

Frontend Flutter menerima field berikut:

```text
Product:
- id atau ID
- name
- price
- image_url
- category

Cart item:
- id atau ID
- product_id
- product
- quantity atau quantit
- subtotal

Order:
- id atau ID
- total_amount atau total
- status
- shipping_address
- notes
- payment_method
- payment_status atau paymentStatus
- va_number atau virtual_account
- gopay_deeplink atau gopay_url atau deeplink
- paid_at
- items
- created_at
```

Walaupun frontend cukup fleksibel, backend sebaiknya mengirim snake_case konsisten:

```text
id
total_amount
payment_method
payment_status
va_number
gopay_deeplink
created_at
```

## Error Response Standard

Gunakan format error seperti ini:

```json
{
  "message": "Pesan error"
}
```

Status code yang disarankan:

```text
400 Bad Request        -> input invalid
401 Unauthorized       -> token tidak ada/tidak valid
403 Forbidden          -> akses bukan milik user
404 Not Found          -> item/order/product tidak ditemukan
409 Conflict           -> stock tidak cukup
500 Internal Server    -> error server/database
```

## Acceptance Tests

Pastikan test berikut berhasil lewat Postman atau curl:

```text
GET /v1/products dengan Bearer token -> 200

POST /v1/cart dengan Bearer token:
body {"product_id":1,"quantity":2}
-> bukan 404, idealnya 200/201

GET /v1/cart dengan Bearer token
-> data.items berisi produk yang ditambahkan

PUT /v1/cart/:id
body {"quantity":3}
-> quantity berubah

DELETE /v1/cart/:id
-> item terhapus

POST /v1/orders/checkout
body {
  "shipping_address":"Alamat",
  "notes":"",
  "payment_method":"virtual_account"
}
-> order dibuat, cart kosong, order_items terisi

GET /v1/orders
-> daftar order user muncul

GET /v1/orders/:id
-> detail order dan items muncul
```

## Important Notes

- Jangan hanya membuat tabel database. Route Gin harus benar-benar didaftarkan, karena masalah saat ini adalah `404 Not Found`.
- Endpoint harus memakai auth middleware yang sama dengan `/v1/products`.
- Ambil `user_id` dari token/backend user context, bukan dari request body.
- Checkout harus memakai database transaction.
- Nama endpoint yang diharapkan frontend adalah `/v1/cart` dan `/v1/orders/...`.
- Jika backend ingin memakai nama lain seperti `/v1/cart_items`, frontend harus ikut diubah. Rekomendasi: tetap gunakan `/v1/cart` agar sesuai materi Flutter.
