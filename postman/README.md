# Koleksi Postman

Isi folder ini:

- `BIMAONE.postman_collection.json` — semua endpoint `brands`, `warehouses`, dan
  `health`, ditambah folder **Contoh Error** (400, 404, 409, 422).
- `BIMAONE.postman_environment.json` — environment `BIMAONE Local` dengan
  variabel `baseUrl`.

## Cara pakai

1. Jalankan API dulu (`go run ./cmd/api`) di `http://localhost:8080`.
2. Buka Postman → **Import** → pilih kedua file JSON di folder ini.
3. Pilih environment **BIMAONE Local** di kanan atas.
4. Jalankan request. Kalau API di alamat lain, ubah nilai `baseUrl` di environment.

Catatan: request `409 - Kode duplikat` baru menghasilkan 409 setelah `code`
tersebut sudah ada (misalnya setelah menjalankan `Tambah Gudang`).
