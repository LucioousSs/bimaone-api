package warehouses

import (
	"context"
	"os"
	"testing"

	"bimaone-api/internal/database"
)

func TestListIntegration(t *testing.T) {
	if os.Getenv("WAREHOUSES_INTEGRATION") != "1" {
		t.Skip("set WAREHOUSES_INTEGRATION=1 untuk menjalankan tes database")
	}

	if err := database.ConnectDB(); err != nil {
		t.Fatalf("koneksi database gagal: %v", err)
	}
	defer database.CloseDB()

	repo := NewRepository(database.DB)

	result, err := repo.List(context.Background(), ListParams{
		Search:  "WH-TEST-001",
		Status:  "active",
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("List gagal: %v", err)
	}

	if result.Total < 1 {
		t.Fatalf("data gudang uji tidak ditemukan; total=%d", result.Total)
	}
	if len(result.Data) < 1 {
		t.Fatal("hasil List tidak memuat data gudang uji")
	}
	if result.Data[0].Code != "WH-TEST-001" {
		t.Fatalf("kode gudang tidak sesuai: %q", result.Data[0].Code)
	}
	if result.Data[0].CreatedAt.IsZero() {
		t.Fatal("created_at kosong")
	}
}
