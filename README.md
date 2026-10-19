# BIMAONE API

Backend REST API untuk BIMAONE, dibangun dengan **Go** dan **PostgreSQL**.

> **Status:** masih tahap awal. Yang sudah jalan baru endpoint health check dan koneksi database. Fitur brand dan warehouse masih berupa kerangka folder yang belum diisi.

## Tech Stack

| Bagian | Teknologi |
| --- | --- |
| Bahasa | Go 1.27 |
| Web framework | Gin |
| Database | PostgreSQL (pgxpool) |
| Container | Docker Compose |

## Struktur Project

```
cmd/api/main.go              Entry point dan setup router
internal/database/db.go      Koneksi PostgreSQL pakai pgxpool
internal/brands/             Model, handler, repository untuk brand
internal/warehouses/         Model, handler, repository untuk warehouse
db/migrations/               Migrasi SQL
docker-compose.yml           PostgreSQL untuk development
```

## Menjalankan

1. Salin file environment.

   ```bash
   cp .env.example .env
   ```

2. Nyalakan PostgreSQL.

   ```bash
   docker compose up -d
   ```

3. Unduh dependency dan jalankan.

   ```bash
   go mod download
   go run cmd/api/main.go
   ```

Server berjalan di `http://localhost:8080`.

## Endpoint

| Method | Path | Keterangan |
| --- | --- | --- |
| GET | `/health` | Cek apakah service berjalan |

```json
{
  "status": "ok",
  "message": "BIMAONE API Backend is running!"
}
```

## Environment

| Variabel | Dipakai | Keterangan |
| --- | --- | --- |
| `DATABASE_URL` | ya | DSN PostgreSQL lengkap |
| `DB_HOST` | ya | Host database |
| `DB_PORT` | ya | Port database |
| `DB_USER` | ya | User database |
| `DB_PASSWORD` | ya | Password database |
| `DB_NAME` | ya | Nama database |
| `PORT` | belum | Masih ada di `.env.example`, server masih di-hardcode ke 8080 |
| `CORS_ALLOWED_ORIGINS` | belum | Belum dipakai di kode |

Kalau `DATABASE_URL` kosong, koneksi akan dibangun otomatis dari `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, dan `DB_NAME`.
