package warehouses

import (
	"context"
	"os"
	"testing"
	"time"

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

type fakeWarehouseRow struct {
	at time.Time
}

func (f fakeWarehouseRow) Scan(dest ...any) error {
	*dest[0].(*int64) = 1
	*dest[1].(*string) = "GDG-01"
	*dest[2].(*string) = "Gudang Uji"
	*dest[3].(**string) = nil
	*dest[4].(**string) = nil
	*dest[5].(**string) = nil
	*dest[6].(**time.Time) = nil
	*dest[7].(*time.Time) = f.at
	*dest[8].(*time.Time) = f.at
	return nil
}

func TestScanWarehouseConvertsToUTC(t *testing.T) {
	wita := time.FixedZone("WITA", 8*60*60)
	local := time.Date(2026, 10, 16, 11, 50, 4, 0, wita)

	var item Warehouse
	if err := scanWarehouse(fakeWarehouseRow{at: local}, &item); err != nil {
		t.Fatalf("scanWarehouse gagal: %v", err)
	}

	if item.CreatedAt.Location() != time.UTC {
		t.Fatalf("created_at harus bertimezone UTC, didapat %v", item.CreatedAt.Location())
	}
	if item.UpdatedAt.Location() != time.UTC {
		t.Fatalf("updated_at harus bertimezone UTC, didapat %v", item.UpdatedAt.Location())
	}
	if !item.CreatedAt.Equal(local) {
		t.Fatalf("momen waktu berubah: %v vs %v", item.CreatedAt, local)
	}
}
