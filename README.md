# BIMAONE API

Backend aplikasi BIMAONE. Go + Gin untuk HTTP, pgx dengan SQL tulis tangan (tanpa ORM),
PostgreSQL 16, dan golang-migrate untuk skema database.

Entitas yang tersedia: **brands** dan **warehouses**.

## Prasyarat

- **Go 1.27** atau lebih baru (lihat `go.mod`). Brief mensyaratkan minimal 1.23.
- **Docker Desktop**, untuk menjalankan PostgreSQL.
- **golang-migrate CLI**. Cara memasang:

  ```
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

  Hasilnya ada di `%USERPROFILE%\go\bin` (Windows) atau `$HOME/go/bin` (macOS/Linux).
  Pastikan folder itu sudah masuk `PATH`, lalu cek dengan `migrate -version`.

## Menjalankan dari nol

Ikuti urutannya. Semua perintah dijalankan dari folder root proyek.

### 1. Ambil kode

```
git clone https://github.com/LucioousSs/bimaone-api.git
cd bimaone-api
```

### 2. Buat file `.env`

```
copy .env.example .env
```

macOS/Linux: `cp .env.example .env`

Nilai di dalamnya sudah cocok untuk pemakaian lokal, jadi belum perlu diubah.
File `.env` sengaja tidak ikut masuk git, jadi setiap orang punya salinannya sendiri.

### 3. Nyalakan database

```
docker compose up -d
```

Container bernama `bimaone_postgres` akan jalan di port 5432. Cek dengan:

```
docker ps
```

### 4. Jalankan migrasi

DSN di bawah memakai nilai contoh dari `.env.example`. Kalau kamu mengubah `.env`,
samakan `user`, `password`, dan nama database-nya.

```
migrate -path db/migrations -database "postgres://postgres:password123@localhost:5432/bimaone_postgres?sslmode=disable" up
```

Kalau berhasil, perintah ini membuat tabel `brands`, `warehouses`, dan `schema_migrations`.
Menjalankannya lagi akan menjawab `no change`, artinya aman diulang.

### 5. Isi data contoh

```
Get-Content db\seeds\brands.sql -Raw | docker exec -i bimaone_postgres psql -U postgres -d bimaone_postgres -f -
```

macOS/Linux:

```
docker exec -i bimaone_postgres psql -U postgres -d bimaone_postgres -f - < db/seeds/brands.sql
```

Perintah ini mengisi 5 brand contoh (4 aktif, 1 diarsipkan). Isinya memakai
`WHERE NOT EXISTS`, jadi aman dijalankan berulang tanpa membuat data dobel.

### 6. Jalankan API

```
go run ./cmd/api
```

Server jalan di `http://localhost:8080`.

### 7. Pastikan semuanya hidup

```
curl http://localhost:8080/health
curl "http://localhost:8080/api/v1/brands?status=all"
```

Yang kedua harus mengembalikan 5 brand contoh.

## Environment variables

| Variabel | Isi | Contoh |
|---|---|---|
| `PORT` | port HTTP server | `8080` |
| `DB_HOST` | alamat database | `localhost` |
| `DB_PORT` | port database | `5432` |
| `DB_USER` | user database | `postgres` |
| `DB_PASSWORD` | password database | `password123` |
| `DB_NAME` | nama database | `bimaone_postgres` |
| `CORS_ALLOWED_ORIGINS` | asal frontend yang boleh memanggil API | `http://localhost:3000` |
| `DATABASE_URL` | opsional, satu DSN penuh yang menimpa `DB_*` | `postgres://...` |

`PORT` dan `DATABASE_URL` tidak ada di `.env.example` karena keduanya opsional:
kalau kosong, port memakai `8080` dan koneksi disusun dari variabel `DB_*`.

Jangan pernah menulis alamat, password, atau kredensial langsung di kode. Semuanya
dibaca dari environment. `.env` sudah masuk `.gitignore`, jangan dipaksa di-commit.

`docker compose` membaca file `.env` secara otomatis, jadi langkah 2 di atas harus
selesai sebelum langkah 3. Kalau `.env` belum ada, `POSTGRES_PASSWORD` jadi kosong
dan container gagal menyala.

## Struktur folder

```
cmd/api/main.go              entry point: koneksi database, routing, CORS
internal/database/db.go      connection pool ke PostgreSQL
internal/brands/             model.go, repository.go, handler.go
internal/warehouses/         model.go, repository.go, handler.go
db/migrations/               skema. Setiap versi punya file .up.sql dan .down.sql
db/seeds/                    data contoh, dijalankan manual setelah migrate
docker-compose.yml           PostgreSQL 16
```

Setiap paket entitas dibagi tiga:

- `model.go` berisi bentuk data dan struct request.
- `repository.go` berisi seluruh SQL.
- `handler.go` berisi validasi, pemetaan error, dan bentuk response HTTP.

## Endpoint

Semua endpoint data diawali `/api/v1`, kecuali `GET /health` yang dipakai untuk
mengecek server hidup.

| Method | Path | Keterangan | Sukses |
|---|---|---|---|
| `GET` | `/brands` | daftar brand | 200 |
| `POST` | `/brands` | tambah brand | 201 |
| `GET` | `/brands/:id` | detail brand | 200 |
| `PUT` | `/brands/:id` | ubah brand | 200 |
| `PATCH` | `/brands/:id/archive` | arsipkan brand | 200 |
| `PATCH` | `/brands/:id/restore` | pulihkan brand | 200 |
| `GET` | `/warehouses` | daftar gudang | 200 |
| `POST` | `/warehouses` | tambah gudang | 201 |
| `GET` | `/warehouses/:id` | detail gudang | 200 |
| `PUT` | `/warehouses/:id` | ubah gudang | 200 |
| `PATCH` | `/warehouses/:id/archive` | arsipkan gudang | 200 |
| `PATCH` | `/warehouses/:id/restore` | pulihkan gudang | 200 |
| `GET` | `/health` | cek server hidup | 200 |

Parameter query untuk endpoint daftar:

| Parameter | Nilai | Default |
|---|---|---|
| `status` | `active`, `archived`, `all` | `active` |
| `search` | kata kunci nama | kosong |
| `page` | bilangan bulat mulai 1 | `1` |
| `per_page` | 1 sampai 100 | `10` |

## Format response

Daftar, memakai `meta`:

```json
{
  "data": [ { "id": 1, "name": "Grundfos" } ],
  "meta": { "page": 1, "per_page": 10, "total": 1, "total_pages": 1 }
}
```

Detail, tambah, ubah, arsip, dan pulihkan. `data` berisi satu objek, tanpa `meta`:

```json
{
  "message": "Brand berhasil ditambahkan",
  "data": { "id": 1, "name": "Grundfos" }
}
```

Error:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Data tidak valid",
    "fields": { "name": "Nama brand wajib diisi" }
  }
}
```

`fields` selalu ada. Isinya objek kosong `{}` kalau errornya bukan soal isi field.

Catatan: perilaku itu sudah diterapkan di `brands`. Di `warehouses` saat ini `fields`
ikut hilang kalau kosong, karena memakai `json:"fields,omitempty"`. Jadi kalau ragu,
lihat paket `brands` sebagai acuan.

| Status | Code | Kapan |
|---|---|---|
| 400 | `BAD_REQUEST` | JSON rusak, `id` bukan angka, parameter query tidak valid |
| 404 | `NOT_FOUND` | data dengan `id` itu tidak ada |
| 409 | `DUPLICATE` | nama (brand) atau kode (gudang) sudah dipakai, `fields` menunjuk kolomnya |
| 422 | `VALIDATION_ERROR` | field wajib kosong, terlalu panjang, atau salah format |
| 500 | `INTERNAL_ERROR` | error tak terduga. Detail teknis hanya masuk log server |

## Aturan data

- **Arsip, bukan hapus.** Tidak ada `DELETE` ke database. Master data dinonaktifkan
  dengan mengisi kolom `archived_at`. Daftar memfilter kolom ini sesuai `status`.
- **Nama brand unik tanpa peduli huruf besar/kecil.** Penjagaannya ada di database
  lewat index unik `lower(trim(name))`, jadi tetap benar walau ada dua request bersamaan.
- **Nama 2 sampai 100 karakter**, dihitung per karakter, bukan per byte.
- **Deskripsi maksimal 500 karakter.** Deskripsi kosong disimpan sebagai `NULL`,
  bukan string kosong, supaya "kosong" hanya punya satu bentuk.
- **Tanggal selalu ISO 8601 UTC**, contoh `2026-10-16T07:14:44.427171Z`.
- **Nama field JSON memakai snake_case**, sama dengan nama kolom database.

## Menjalankan test

```
go test ./...
```

Test handler dan repository baru ada di paket `warehouses`. Untuk `brands` belum ada.

## Catatan yang sering bikin bingung

**1. Password database hanya dipakai saat volume dibuat pertama kali.**
Kalau folder `pgdata` sudah ada, mengubah `DB_PASSWORD` di `.env` tidak akan mengubah
password di dalam database. Gejalanya `password authentication failed` saat menjalankan
`migrate`. Pilihanmu: samakan `DB_PASSWORD` dengan password volume yang sudah ada, atau
hapus volumenya dengan `docker compose down -v` lalu `docker compose up -d`.
Ingat, `-v` menghapus semua data.

**2. `migrate down` itu berbahaya.**
`db/migrations/000001_create_brands_table.down.sql` berisi `DROP TABLE brands`. Menjalankannya
menghapus tabel beserta seluruh datanya, dan tidak bisa dibatalkan.

**3. Di PowerShell, pesan sukses `migrate` muncul seperti error.**
`migrate` menulis laporannya ke `stderr`, dan PowerShell menampilkan apa pun dari `stderr`
sebagai blok merah `NativeCommandError`. Padahal perintahnya berhasil, exit code-nya `0`.
Jadi jangan panik dan jangan menganggapnya gagal. Yang perlu dibaca isinya:
`no change` berarti skema sudah paling baru, sedangkan pesan berawalan `error:` berarti gagal
sungguhan.

**4. Semua file di repo ini memakai akhir baris CRLF.**
`gofmt -l .` akan menandai hampir semua file karena hal ini, padahal isi kodenya sudah
sesuai. Untuk menormalkannya, tambahkan file `.gitattributes` berisi `* text=auto eol=lf`.

## Kalau database kamu sudah punya tabel lama

Tabel di proyek ini awalnya dibuat manual, sebelum golang-migrate dipakai. Artinya
database lamamu belum punya tabel `schema_migrations`, sehingga `migrate up` akan gagal
di versi 1 karena tabel `brands` sudah ada.

Periksa dulu keadaannya:

```sql
SELECT to_regclass('public.schema_migrations') AS ada_tabel_migration;

SELECT column_name, data_type, character_maximum_length
FROM information_schema.columns
WHERE table_name = 'brands' ORDER BY ordinal_position;
```

Kalau `ada_tabel_migration` kosong (`NULL`), berarti golang-migrate belum pernah dipakai
di database itu. Pilih salah satu:

**Jalur A, kalau data lama tidak perlu disimpan.** Paling bersih dan paling sedikit salah:

```
docker compose down -v
docker compose up -d
migrate -path db/migrations -database "postgres://..." up
```

**Jalur B, kalau data harus dipertahankan.** Tandai migrasi sebagai sudah sampai versi 2,
karena tabelnya memang sudah ada:

```
migrate -path db/migrations -database "postgres://..." force 2
```

Setelah itu sesuaikan bentuk tabel `brands` dengan `db/migrations/000001_create_brands_table.up.sql`
secara manual. Langkah ini tidak tersimpan di repo, jadi kalau databasenya di-reset nanti
kamu harus mengulanginya.

Verifikasi hasilnya:

```sql
SELECT version, dirty FROM schema_migrations;
SELECT indexname FROM pg_indexes WHERE tablename = 'brands';
```

Yang pertama harus berisi `2` dan `f` (false), yang kedua harus memuat `brands_name_unique_idx`.

## Alur kontribusi

- Buat branch sendiri per pekerjaan, jangan langsung ke `dev`.
- PR ditujukan ke `dev`, bukan `main`.
- Jalankan `gofmt` pada file yang kamu ubah sebelum commit.
- Setiap migration baru wajib punya pasangan `.up.sql` dan `.down.sql`.
