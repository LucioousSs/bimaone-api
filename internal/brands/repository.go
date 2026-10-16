package brands

import (
	"context"
	"errors"
	"bimaone-api/internal/database"
	"strconv"
	"strings"
	
)

// Struktur untuk menampung data dan meta pagination
type PaginationResult struct {
	Brands      []Brand
	Total       int
	TotalPages  int
}

// 1. AMBIL SEMUA DATA (Menampilkan daftar brand di layar)
func GetAllBrands() ([]Brand, error) {
	query := `SELECT id, name, description, archived_at, created_at, updated_at FROM brands ORDER BY id DESC`
	
	rows, err := database.DB.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []Brand

	for rows.Next() {
		var b Brand
		// Mencocokkan kolom database ke dalam kotak variabel (struct) Go
		err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.ArchivedAt, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}

	return brands, nil
}

// Mengambil satu brand berdasarkan ID
func GetBrandByID(id int) (*Brand, error) {
	query := `SELECT id, name, description, archived_at, created_at, updated_at FROM brands WHERE id = $1`
	var b Brand
	err := database.DB.QueryRow(context.Background(), query, id).Scan(
		&b.ID, &b.Name, &b.Description, &b.ArchivedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err // Jika tidak ketemu, error akan ditangkap di handler
	}
	return &b, nil
}

// CreateBrand: Menambahkan brand baru dengan pengecekan duplikat case-insensitive
func CreateBrand(b *Brand) error {
	// Cek dulu apakah nama brand sudah ada (mengabaikan huruf besar/kecil)
	var existingID int
	checkQuery := `SELECT id FROM brands WHERE LOWER(name) = LOWER($1)`
	err := database.DB.QueryRow(context.Background(), checkQuery, b.Name).Scan(&existingID)
	
	// Jika tidak error (berarti datanya KETEMU / sudah ada yang punya nama sama)
	if err == nil {
		return errors.New("duplicate_name")
	}

	// Kalau belum ada, lanjut insert ke database
	query := `INSERT INTO brands (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at, archived_at`
	err = database.DB.QueryRow(context.Background(), query, b.Name, b.Description).Scan(
		&b.ID, &b.CreatedAt, &b.UpdatedAt, &b.ArchivedAt,
	)
	return err
}

// UpdateBrand: Ubah brand, otomatis menolak jika nama baru sudah dipakai brand lain
// Contoh di UpdateBrand
func UpdateBrand(b *Brand) error {
	query := `UPDATE brands SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	result, err := database.DB.Exec(context.Background(), query, b.Name, b.Description, b.ID)
	if err != nil {
		return err // Bisa karena duplikat nama (unique constraint)
	}
	if result.RowsAffected() == 0 {
		return errors.New("not_found")
	}
	return nil
}

// ArchiveBrand berfungsi untuk mencatat waktu arsip dan mengembalikan data brand
func ArchiveBrand(id int) (*Brand, error) {
	query := `UPDATE brands SET archived_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, name, description, archived_at, created_at, updated_at`
	
	var b Brand
	err := database.DB.QueryRow(context.Background(), query, id).Scan(
		&b.ID, &b.Name, &b.Description, &b.ArchivedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// RestoreBrand berfungsi untuk memulihkan brand dan mengembalikan data brand
func RestoreBrand(id int) (*Brand, error) {
	query := `UPDATE brands SET archived_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, name, description, archived_at, created_at, updated_at`
	
	var b Brand
	err := database.DB.QueryRow(context.Background(), query, id).Scan(
		&b.ID, &b.Name, &b.Description, &b.ArchivedAt, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}


func GetFilteredBrands(status string, search string, page int, perPage int) ([]Brand, int, error) {
	// Hitung offset untuk pagination (misal halaman 2 dengan 10 data per halaman, mulai dari data ke-10)
	offset := (page - 1) * perPage

	// Base query dasar
	query := `SELECT id, name, description, archived_at, created_at, updated_at FROM brands WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM brands WHERE 1=1`
	
	var args []interface{}
	argId := 1

	// 1. Filter Status
	if status == "active" {
		query += ` AND archived_at IS NULL`
		countQuery += ` AND archived_at IS NULL`
	} else if status == "archived" {
		query += ` AND archived_at IS NOT NULL`
		countQuery += ` AND archived_at IS NOT NULL`
	}
	// Kalau "all", tidak perlu filter archived_at

	// 2. Filter Search (Pencarian nama tanpa peduli huruf besar/kecil)
	if search != "" {
		query += ` AND LOWER(name) LIKE $` + strconv.Itoa(argId)
		countQuery += ` AND LOWER(name) LIKE $` + strconv.Itoa(argId)
		args = append(args, "%"+strings.ToLower(search)+"%")
		argId++
	}

	// Hitung total data keseluruhan dulu untuk meta
	var total int
	err := database.DB.QueryRow(context.Background(), countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 3. Urutan A-Z dan Pagination (LIMIT & OFFSET)
	// Di dalam fungsi GetFilteredBrands (bagian ORDER BY):
	query += ` ORDER BY LOWER(name) ASC LIMIT $` + strconv.Itoa(argId) + ` OFFSET $` + strconv.Itoa(argId+1)
	args = append(args, perPage, offset)

	rows, err := database.DB.Query(context.Background(), query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var brands []Brand = []Brand{} // Inisialisasi slice kosong agar tidak nil
	for rows.Next() {
		var b Brand
		err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.ArchivedAt, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		brands = append(brands, b)
	}

	return brands, total, nil
}