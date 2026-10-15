package brands

import (
    "context"
    "bimaone-api/internal/database"
)

// GetAllBrands berfungsi untuk mengambil semua data dari tabel brands menggunakan SQL tulis tangan
func GetAllBrands() ([]Brand, error) {
    // Perintah SQL murni
    query := `SELECT id, name, description, created_at, updated_at FROM brands ORDER BY id DESC`
    
    // Menjalankan query ke database menggunakan pool yang sudah ada
    rows, err := database.DB.Query(context.Background(), query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var brands []Brand

    // Membaca data baris per baris dari hasil query
    for rows.Next() {
        var b Brand
        err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt)
        if err != nil {
            return nil, err
        }
        brands = append(brands, b)
    }

    return brands, nil
}

// CreateBrand berfungsi untuk menyimpan data brand baru ke database
func CreateBrand(b *Brand) error {
    // Perintah SQL INSERT dan mengembalikan data baru
    query := `INSERT INTO brands (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`
    
    err := database.DB.QueryRow(context.Background(), query, b.Name, b.Description).Scan(
        &b.ID, &b.CreatedAt, &b.UpdatedAt,
    )
    if err != nil {
        return err
    }

    return nil
}