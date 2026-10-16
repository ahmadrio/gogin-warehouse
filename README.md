# Gogin Warehouse

REST API manajemen gudang multi-lokasi, ditulis dengan Go, Gin, GORM, dan PostgreSQL. API ini mencatat barang masuk, barang keluar, transfer antar-rak, dan penyesuaian stok. Setiap transaksi melewati persetujuan supervisor sebelum mengubah saldo stok.

## Fitur

- **Autentikasi JWT** dengan hak akses berbasis role dan permission (`resource:operation`).
- **Pembatasan per gudang**: user hanya melihat dan mengubah data gudang yang ditugaskan kepadanya.
- **Master data**: kategori produk, produk, gudang, dan lokasi (rak) di dalam gudang.
- **Transaksi stok** dengan alur pengajuan → persetujuan:
  - In-bound (barang masuk, lalu putaway ke rak)
  - Out-bound (barang keluar)
  - Stock transfer (pindah rak dalam satu gudang)
  - Stock adjustment (stock opname, barang rusak, barang hilang, koreksi)
- **Ledger stok**: setiap perubahan saldo tercatat di `stock_movements` beserta saldo akhirnya.
- **Audit trail** untuk setiap create, update, delete, approve, reject, dan putaway.
- **Nomor dokumen otomatis** per bulan, misalnya `INB-202610-0001`.

## Tech stack

| Komponen | Pilihan |
|---|---|
| Bahasa | Go 1.27 |
| HTTP framework | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) + driver PostgreSQL |
| Database | PostgreSQL 16 (via Docker Compose) |
| Migrasi | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Auth | [golang-jwt/jwt v5](https://github.com/golang-jwt/jwt), bcrypt |

## Menjalankan secara lokal

### Prasyarat

- Go 1.27 atau lebih baru
- Docker dan Docker Compose
- CLI golang-migrate dengan driver PostgreSQL:

  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

  Binary terpasang di `$(go env GOPATH)/bin`. Pastikan folder itu ada di `PATH`.

### Langkah

```bash
# 1. Clone dan masuk ke folder project
git clone https://github.com/ahmadrio/gogin-warehouse.git
cd gogin-warehouse

# 2. Siapkan konfigurasi, lalu ganti JWT_SECRET dengan string acak yang panjang
cp .env.example .env

# 3. Jalankan PostgreSQL
docker compose up -d

# 4. Buat tabel
migrate -path migrations \
  -database "postgres://postgres:postgres@localhost:5432/warehouse?sslmode=disable" up

# 5. Isi data contoh (role, permission, gudang, produk, user)
go run ./cmd/seed

# 6. Jalankan API
go run ./cmd/api
```

Cek server dengan `curl http://localhost:8080/health`. Respons `{"status":"ok"}` berarti API dan database siap.

### Akun hasil seeder

Semua akun memakai password `password123`.

| Email | Role | Akses gudang |
|---|---|---|
| `superadmin@warehouse.test` | superadmin | Semua gudang |
| `supervisor@warehouse.test` | supervisor | WH-MALANG-01, WH-SURABAYA-01 |
| `staff.malang@warehouse.test` | staff | WH-MALANG-01 |
| `staff.surabaya@warehouse.test` | staff | WH-SURABAYA-01 |

Seeder juga membuat tiga gudang (masing-masing dengan tiga rak), dua kategori, dan tiga produk. Seeder aman dijalankan berulang kali karena memakai `FirstOrCreate`.

## Konfigurasi

Aplikasi membaca variabel berikut dari environment atau file `.env`.

| Variabel | Default | Keterangan |
|---|---|---|
| `APP_ENV` | `development` | Nama environment |
| `APP_PORT` | `8080` | Port HTTP |
| `DB_HOST` | `localhost` | Host PostgreSQL |
| `DB_PORT` | `5432` | Port PostgreSQL |
| `DB_USER` | `postgres` | User database |
| `DB_PASSWORD` | _(kosong)_ | Password database |
| `DB_NAME` | `warehouse` | Nama database |
| `DB_SSLMODE` | `disable` | Mode SSL koneksi |
| `DB_TIMEZONE` | `Asia/Jakarta` | Zona waktu koneksi |
| `JWT_SECRET` | _(wajib)_ | Kunci penanda tangan JWT; aplikasi menolak start tanpa nilai ini |
| `JWT_TTL_HOURS` | `24` | Masa berlaku token (jam) |

## API

Semua endpoint berada di bawah `/api/v1`. Kecuali login, setiap request wajib membawa header `Authorization: Bearer <access_token>`.

### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{ "email": "superadmin@warehouse.test", "password": "password123" }
```

Respons berisi `access_token`, `expires_at`, dan profil user (role, permission, dan daftar gudang). `GET /api/v1/auth/me` mengembalikan profil yang sama untuk token aktif.

### Daftar endpoint

| Method | Path | Permission |
|---|---|---|
| `GET` | `/health` | publik |
| `POST` | `/auth/login` | publik |
| `GET` | `/auth/me` | login |
| `GET, POST` | `/products` | `products:read`, `products:create` |
| `GET, PUT, DELETE` | `/products/:id` | `products:read`, `:update`, `:delete` |
| `GET, POST` | `/product-categories` | `product_categories:read`, `:create` |
| `GET, PUT, DELETE` | `/product-categories/:id` | `product_categories:read`, `:update`, `:delete` |
| `GET, POST` | `/warehouses` | `warehouses:read`, `:create` |
| `GET` | `/warehouses/:id` | `warehouses:read` |
| `GET, POST` | `/warehouses/:id/locations` | `warehouse_locations:read`, `:create` |
| `GET, PUT, DELETE` | `/warehouses/:id/locations/:location_id` | `warehouse_locations:read`, `:update`, `:delete` |
| `GET` | `/stocks` | `stocks:read` |
| `GET, POST` | `/stock-adjustments` | `stock_adjustments:read`, `:create` |
| `GET, PUT` | `/stock-adjustments/:id` | `stock_adjustments:read`, `:update` |
| `POST` | `/stock-adjustments/:id/approve` | `stock_adjustments:approve` |
| `POST` | `/stock-adjustments/:id/reject` | `stock_adjustments:reject` |
| `GET, POST` | `/outbounds` | `outbounds:read`, `:create` |
| `GET, PUT` | `/outbounds/:id` | `outbounds:read`, `:update` |
| `POST` | `/outbounds/:id/approve` | `outbounds:approve` |
| `POST` | `/outbounds/:id/reject` | `outbounds:reject` |
| `GET, POST` | `/stock-transfers` | `stock_transfers:read`, `:create` |
| `GET, PUT` | `/stock-transfers/:id` | `stock_transfers:read`, `:update` |
| `POST` | `/stock-transfers/:id/approve` | `stock_transfers:approve` |
| `POST` | `/stock-transfers/:id/reject` | `stock_transfers:reject` |
| `GET, POST` | `/inbounds` | `inbounds:read`, `:create` |
| `GET, PUT` | `/inbounds/:id` | `inbounds:read`, `:update` |
| `POST` | `/inbounds/:id/approve` | `inbounds:approve` |
| `POST` | `/inbounds/:id/reject` | `inbounds:reject` |
| `POST` | `/inbounds/:id/putaway` | `inbounds:update` |

### Pagination dan filter

Endpoint list menerima `?page=` (default 1), `?per_page=` (default 15, maksimal 100), dan `?search=`. Endpoint transaksi juga menerima `?warehouse_id=` dan `?status=`. `GET /stocks` menerima `?warehouse_id=`, `?location_id=`, dan `?product_id=`.

### Format respons

```jsonc
// Sukses
{ "data": { ... } }

// List
{ "data": [ ... ], "meta": { "page": 1, "per_page": 15, "total": 42, "total_pages": 3 } }

// Error
{ "message": "data tidak valid", "errors": { "items[0].quantity": "..." } }
```

| Status | Arti |
|---|---|
| `400` | Format JSON atau parameter UUID tidak valid |
| `401` | Token tidak ada, tidak valid, atau kedaluwarsa |
| `403` | User tidak punya permission yang dibutuhkan |
| `404` | Data tidak ditemukan, atau berada di gudang di luar akses user |
| `409` | Konflik data atau status transaksi tidak mengizinkan aksi tersebut |
| `422` | Validasi gagal, atau stok tidak cukup |

## Alur transaksi

Out-bound, stock transfer, dan stock adjustment memakai alur yang sama:

```text
create ──► pending ──approve──► approved   (stok berubah di sini)
              │
              └──reject──► rejected ──revisi (PUT)──► pending
```

In-bound menambah satu langkah. Stok baru bertambah saat putaway, ketika setiap item diberi rak tujuan:

```text
create ──► pending ──approve──► approved ──putaway──► completed   (stok bertambah)
```

Aturan yang berlaku untuk semua transaksi:

- Stok hanya berubah saat approve (atau putaway untuk in-bound). Pengajuan yang ditolak tidak menyentuh stok.
- Transaksi yang sudah `approved` tidak bisa direvisi.
- Saldo stok tidak boleh negatif. Jika satu item gagal, seluruh transaksi di-rollback.
- `StockService` adalah satu-satunya kode yang mengubah saldo. Service ini mengunci baris saldo dalam urutan tetap untuk mencegah deadlock, lalu menulis ledger `stock_movements`.
- Stock transfer memindahkan barang antar-rak tanpa mengubah total stok produk di gudang.

## Hak akses

| Role | Wewenang |
|---|---|
| superadmin | Semua permission di semua gudang |
| supervisor | Melihat master data dan stok; menyetujui atau menolak transaksi |
| staff | Melihat master data dan stok; membuat dan merevisi transaksi |

Permission menentukan fitur yang boleh dipakai. Daftar gudang user menentukan data mana yang terlihat. Supervisor hasil seeder, misalnya, mendapat `404` saat membuka transaksi di WH-SURABAYA-02 karena gudang itu di luar aksesnya.

## Struktur project

```text
cmd/
  api/            entrypoint server HTTP
  seed/           seeder data contoh
internal/
  auth/           JWT, definisi role & permission, user di context
  config/         pembacaan environment
  database/       koneksi PostgreSQL
  dto/            bentuk request & response
  handler/        lapisan HTTP (binding, validasi, respons)
  middleware/     autentikasi & cek permission
  model/          struct tabel GORM
  repository/     query database
  router/         wiring dependency & daftar route
  service/        logika bisnis & transaksi
pkg/
  apperror/       error aplikasi beserta status HTTP
  pagination/     parameter & meta pagination
  response/       helper respons JSON
migrations/       file SQL golang-migrate (up & down)
.http             koleksi request untuk REST Client
```

## Mencoba API dengan file `.http`

File `.http` berisi request siap pakai untuk setiap endpoint, termasuk skenario gagal (401, 403, 404, 409, 422). Buka file ini di VS Code dengan ekstensi [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) atau di JetBrains HTTP Client. Port dibaca dari `APP_PORT` di `.env`, dan token hasil login dipakai ulang oleh request berikutnya.
