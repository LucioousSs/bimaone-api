package brands

import (
	"context"
	"bimaone-api/internal/database"
)

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

// CreateBrand: Simpan brand baru, otomatis menolak jika nama sudah ada
func CreateBrand(b *Brand) error {
	query := `INSERT INTO brands (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	
	err := database.DB.QueryRow(context.Background(), query, b.Name, b.Description).Scan(
		&b.ID, &b.CreatedAt, &b.UpdatedAt,
	)
	return err
}

// UpdateBrand: Ubah brand, otomatis menolak jika nama baru sudah dipakai brand lain
func UpdateBrand(b *Brand) error {
	query := `UPDATE brands SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := database.DB.Exec(context.Background(), query, b.Name, b.Description, b.ID)
	return err
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