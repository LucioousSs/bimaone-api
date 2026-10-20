# CLAUDE.md

Panduan untuk AI assistant (Claude/agent) yang bekerja di repo **bimaone-api**.
Ikuti aturan di file ini. Kalau saran AI bertentangan dengan file ini, ikuti file ini.

## Tentang proyek

Backend BIMAONE (Tahap 1). Latihan membuat CRUD end-to-end untuk entitas **Brand**
dan **Gudang (Warehouse)**: list, cari, tambah, ubah, arsip, dan pulihkan.
Semua data diakses lewat API Go; frontend (`bimaone-web`) memanggil API ini.

## Stack yang dikunci (jangan diganti tanpa izin mentor)

- Bahasa: **Go 1.23+** (repo saat ini memakai `go 1.27`)
- Router: **Gin**
- Akses DB: **pgx v5 + SQL tulis tangan** (tanpa ORM)
- Database: **PostgreSQL 16** via **Docker Compose**
- Migrasi: **golang-migrate**; setiap versi wajib punya `.up.sql` dan `.down.sql`
- Frontend (repo terpisah `bimaone-web`): Next.js App Router + TypeScript +
  Tailwind CSS + shadcn/ui, memanggil API lewat satu `lib/api-client.ts`
- **Jangan menambah library baru tanpa izin.**

## Struktur folder

```
cmd/api/main.go              entry point: env, koneksi DB, routing, CORS
internal/database/db.go      connection pool PostgreSQL
internal/brands/             model.go, repository.go, handler.go
internal/warehouses/         model.go, repository.go, handler.go
db/migrations/               skema; tiap versi punya .up.sql dan .down.sql
db/seeds/                    data contoh (dijalankan manual)
postman/                     koleksi Postman + environment
docker-compose.yml           PostgreSQL 16
```

Setiap paket entitas dibagi tiga:
- `model.go` — bentuk data, input request, format response/error
- `repository.go` — seluruh SQL
- `handler.go` — validasi, pemetaan error, response HTTP

## Perintah penting

Semua dijalankan dari root repo. Nilai DB mengikuti `.env`
(salin dari `.env.example` lebih dulu).

```bash
# 1. Database
docker compose up -d

# 2. Migrasi
migrate -path db/migrations -database "postgres://postgres:password123@localhost:5432/bimaone_postgres?sslmode=disable" up

# 3. Seed contoh (PowerShell)
Get-Content db\seeds\brands.sql -Raw | docker exec -i bimaone_postgres psql -U postgres -d bimaone_postgres -f -

# 4. Jalankan API
go run ./cmd/api

# 5. Cek hidup
curl http://localhost:8080/health

# 6. Build, vet, test
go build ./...
go vet ./...
go test ./...
```

Catatan PowerShell: laporan `migrate` ditulis ke `stderr` sehingga tampak merah,
tetapi exit code `0` berarti berhasil.

## Kontrak API

Semua endpoint data berawalan `/api/v1`.

| Method | Path | Sukses |
|--------|------|--------|
| GET | `/brands` | 200 |
| POST | `/brands` | 201 |
| GET | `/brands/{id}` | 200 |
| PUT | `/brands/{id}` | 200 |
| PATCH | `/brands/{id}/archive` | 200 |
| PATCH | `/brands/{id}/restore` | 200 |
| GET | `/warehouses` | 200 |
| POST | `/warehouses` | 201 |
| GET | `/warehouses/{id}` | 200 |
| PUT | `/warehouses/{id}` | 200 |
| PATCH | `/warehouses/{id}/archive` | 200 |
| PATCH | `/warehouses/{id}/restore` | 200 |
| GET | `/health` | 200 |

Query list: `search`, `status` (`active` default | `archived` | `all`),
`page` (mulai 1), `per_page` (default 10, maks 100). Urutan default: nama A–Z.

Response sukses:
- List: `{ "data": [...], "meta": { "page", "per_page", "total", "total_pages" } }`
- Detail/tambah/ubah/arsip/pulihkan: `{ "data": { ... } }` tanpa `meta`

Response error:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Data tidak valid",
    "fields": { "name": "Nama wajib diisi" }
  }
}
```

`fields` **selalu ada**; berisi `{}` bila error bukan soal isi field.

Kode status:

| Status | Code | Kapan |
|--------|------|-------|
| 400 | `BAD_REQUEST` | JSON rusak, `id`/parameter query bukan angka |
| 404 | `NOT_FOUND` | data dengan id itu tidak ada |
| 409 | `DUPLICATE` | `name` (brand) / `code` (gudang) sudah dipakai |
| 422 | `VALIDATION_ERROR` | field wajib kosong, terlalu panjang, format salah |
| 500 | `INTERNAL_ERROR` | error tak terduga; detail hanya di log server |

## Aturan data & kode

- **Arsip, bukan hapus.** Tidak ada `DELETE`. Nonaktifkan lewat `archived_at`
  (NULL = aktif). Pulihkan dengan mengosongkannya kembali.
- **Tanggal selalu ISO 8601 UTC.** `scanBrand` / `scanWarehouse` mengonversi
  waktu dengan `.UTC()` saat membaca dari database.
- **JSON snake_case**, sama dengan nama kolom database.
- **Validasi di dua sisi.** BE tetap memvalidasi ulang meski FE sudah memvalidasi.
  Constraint database menjadi pengaman terakhir.
- **Pesan error berbahasa Indonesia** agar bisa ditampilkan langsung ke user.
- **Jangan menulis URL/password/kredensial di kode.** Baca dari environment.
  `.env` tidak ikut git; sediakan `.env.example`.
- Brand: `name` unik tanpa peduli huruf besar/kecil (index `lower(trim(name))`),
  2–100 karakter; `description` maks 500 karakter.
- Gudang: `code` wajib, huruf besar, unik, maks 20 karakter; `name` 2–100;
  `phone` hanya angka, `+`, spasi, dan `-`; `pic_name` maks 100.

## Git

- Kerja di branch sendiri per pekerjaan; jangan langsung ke `dev`/`main`.
- PR ditujukan ke `dev`, bukan `main`.
- Buat commit kecil dan rapi; satu commit satu tujuan.
- Jalankan `gofmt` pada file yang diubah dan `go test ./...` sebelum commit.
- Jangan push/merge/release tanpa persetujuan pemilik repo.

## Aturan pakai AI (dari brief mentor)

1. Kalau tidak bisa menjelaskan kode, kode tidak boleh di-merge.
2. Minta penjelasan langkah demi langkah, bukan sekadar jawaban.
3. Satu langkah kecil per prompt; jangan minta satu modul sekaligus.
4. Wajib mengikuti aturan di file ini.
5. Jangan menambah library tanpa izin.
6. Jalankan dan tes sendiri sebelum PR; kode yang belum pernah dijalankan
   tidak boleh masuk PR.
7. Catat pembelajaran harian (3 baris) di `NOTES.md` pribadi.
8. Jangan pernah menempelkan rahasia (password, token, isi `.env`) ke AI.
