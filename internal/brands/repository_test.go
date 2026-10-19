package brands

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"bimaone-api/internal/database"

	"github.com/jackc/pgx/v5"
)

// Tanpa escape, user yang mencari "100%" akan dianggap memakai wildcard dan
// hasilnya bukan yang dia maksud.
func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "teks biasa tidak berubah", input: "Grundfos", want: "Grundfos"},
		{name: "persen di-escape", input: "100%", want: `100\%`},
		{name: "garis bawah di-escape", input: "a_b", want: `a\_b`},
		{name: "backslash di-escape", input: `c:\data`, want: `c:\\data`},
		{name: "gabungan", input: `%_\`, want: `\%\_\\`},
		{name: "backslash dan persen berurutan", input: `\%`, want: `\\\%`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapeLikePattern(tc.input); got != tc.want {
				t.Fatalf("escapeLikePattern(%q) = %q, diharapkan %q", tc.input, got, tc.want)
			}
		})
	}
}

// Tes di bawah butuh PostgreSQL dan sengaja dilewati secara default, supaya
// "go test ./..." tetap hijau di mesin yang tidak menyalakan database.
func siapkanRepoIntegrasi(t *testing.T) *Repository {
	t.Helper()

	if os.Getenv("BRANDS_INTEGRATION") != "1" {
		t.Skip("set BRANDS_INTEGRATION=1 untuk menjalankan tes database")
	}

	if err := database.ConnectDB(); err != nil {
		t.Fatalf("koneksi database gagal: %v", err)
	}
	t.Cleanup(database.CloseDB)

	return NewRepository(database.DB)
}

// Nama sengaja aneh supaya tidak bentrok dengan brand asli.
const (
	namaBrandUji      = "ZZ Brand Uji Integrasi"
	namaBrandUjiKedua = "ZZ Brand Uji Kedua"
)

// siapkanBrandUji mengembalikan brand uji yang aktif dan mengarsipkannya lagi
// saat tes selesai. Sisa dari eksekusi sebelumnya dipakai ulang, bukan dibuat
// baru, karena index unik nama berlaku juga untuk brand yang diarsipkan.
func siapkanBrandUji(t *testing.T, repo *Repository, nama string) Brand {
	t.Helper()

	ctx := context.Background()

	hasil, err := repo.List(ctx, ListParams{Search: nama, Status: "all", Page: 1, PerPage: 10})
	if err != nil {
		t.Fatalf("gagal mencari brand uji: %v", err)
	}

	var brandUji Brand
	for _, item := range hasil.Data {
		if item.Name == nama {
			brandUji = item
			break
		}
	}

	switch {
	case brandUji.ID == 0:
		brandUji, err = repo.Create(ctx, BrandInput{Name: nama})
		if err != nil {
			t.Fatalf("gagal membuat brand uji: %v", err)
		}
	case brandUji.ArchivedAt != nil:
		brandUji, err = repo.Restore(ctx, brandUji.ID)
		if err != nil {
			t.Fatalf("gagal memulihkan brand uji: %v", err)
		}
	}

	t.Cleanup(func() {
		if _, err := repo.Archive(ctx, brandUji.ID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("gagal mengarsipkan brand uji %q: %v", nama, err)
		}
	})

	return brandUji
}

// cekTimestampUTC memeriksa nilai yang benar-benar diterima frontend, bukan
// hanya nilai di memori.
func cekTimestampUTC(t *testing.T, brand Brand, kolom string) {
	t.Helper()

	data, err := json.Marshal(brand)
	if err != nil {
		t.Fatalf("gagal encode brand: %v", err)
	}

	var terkirim map[string]any
	if err := json.Unmarshal(data, &terkirim); err != nil {
		t.Fatalf("gagal decode brand: %v", err)
	}

	nilai, ok := terkirim[kolom].(string)
	if !ok {
		t.Fatalf("%s bukan teks: %v", kolom, terkirim[kolom])
	}
	if !strings.HasSuffix(nilai, "Z") {
		t.Fatalf("%s dikirim sebagai %q, seharusnya berakhiran Z", kolom, nilai)
	}
}

func TestListIntegration(t *testing.T) {
	repo := siapkanRepoIntegrasi(t)
	ctx := context.Background()

	aktif, err := repo.List(ctx, ListParams{Status: "active", Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List status active gagal: %v", err)
	}

	diarsipkan, err := repo.List(ctx, ListParams{Status: "archived", Page: 1, PerPage: 100})
	if err != nil {
		t.Fatalf("List status archived gagal: %v", err)
	}

	t.Run("status active tidak memuat brand diarsipkan", func(t *testing.T) {
		for _, item := range aktif.Data {
			if item.ArchivedAt != nil {
				t.Fatalf("brand %q sedang diarsipkan tetapi ikut muncul", item.Name)
			}
		}
	})

	t.Run("status archived hanya berisi brand diarsipkan", func(t *testing.T) {
		for _, item := range diarsipkan.Data {
			if item.ArchivedAt == nil {
				t.Fatalf("brand %q masih aktif tetapi ikut muncul di daftar arsip", item.Name)
			}
		}
	})

	t.Run("urutan A-Z tanpa peduli huruf besar kecil", func(t *testing.T) {
		semua, err := repo.List(ctx, ListParams{Status: "all", Page: 1, PerPage: 100})
		if err != nil {
			t.Fatalf("List status all gagal: %v", err)
		}

		for i := 1; i < len(semua.Data); i++ {
			sebelum := strings.ToLower(semua.Data[i-1].Name)
			sesudah := strings.ToLower(semua.Data[i].Name)
			if sebelum > sesudah {
				t.Fatalf("urutan salah: %q muncul sebelum %q", semua.Data[i-1].Name, semua.Data[i].Name)
			}
		}
	})

	t.Run("pencarian tidak peduli huruf besar kecil", func(t *testing.T) {
		if len(aktif.Data) == 0 {
			t.Skip("tidak ada brand aktif yang bisa dijadikan contoh")
		}

		contoh := aktif.Data[0].Name
		kataKunci := strings.ToLower(contoh)

		hasil, err := repo.List(ctx, ListParams{Search: kataKunci, Status: "active", Page: 1, PerPage: 100})
		if err != nil {
			t.Fatalf("List dengan search gagal: %v", err)
		}

		for _, item := range hasil.Data {
			if item.Name == contoh {
				return
			}
		}

		t.Fatalf("mencari %q tidak menemukan %q", kataKunci, contoh)
	})

	// Sekaligus membuktikan total_pages tidak menjadi 0 saat datanya kosong.
	t.Run("wildcard diperlakukan sebagai teks biasa", func(t *testing.T) {
		hasil, err := repo.List(ctx, ListParams{Search: "%", Status: "all", Page: 1, PerPage: 100})
		if err != nil {
			t.Fatalf("List dengan search gagal: %v", err)
		}

		if hasil.Total != 0 {
			t.Fatalf("karakter %% seharusnya jadi teks biasa, tetapi menemukan %d brand", hasil.Total)
		}
		if len(hasil.Data) != 0 {
			t.Fatalf("data seharusnya kosong, tetapi berisi %d baris", len(hasil.Data))
		}
		if hasil.TotalPages != 1 {
			t.Fatalf("total_pages diharapkan 1 saat data kosong, tetapi %d", hasil.TotalPages)
		}
	})
}

func TestBrandCRUDIntegration(t *testing.T) {
	repo := siapkanRepoIntegrasi(t)
	ctx := context.Background()

	brand := siapkanBrandUji(t, repo, namaBrandUji)
	kedua := siapkanBrandUji(t, repo, namaBrandUjiKedua)

	t.Run("waktu dikirim ke frontend sebagai UTC", func(t *testing.T) {
		for _, kolom := range []string{"created_at", "updated_at"} {
			cekTimestampUTC(t, brand, kolom)
		}
	})

	t.Run("nama dobel ditolak walau beda huruf besar kecil", func(t *testing.T) {
		_, err := repo.Create(ctx, BrandInput{Name: strings.ToUpper(namaBrandUji)})
		if !isUniqueViolation(err) {
			t.Fatalf("duplikat seharusnya terdeteksi, tetapi err = %v", err)
		}
	})

	t.Run("nama dobel saat update ditolak", func(t *testing.T) {
		_, err := repo.Update(ctx, kedua.ID, BrandInput{Name: namaBrandUji})
		if !isUniqueViolation(err) {
			t.Fatalf("duplikat saat update seharusnya terdeteksi, tetapi err = %v", err)
		}
	})

	t.Run("brand yang tidak ada", func(t *testing.T) {
		const idTidakAda = 999999999

		if _, err := repo.GetByID(ctx, idTidakAda); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("GetByID seharusnya ErrNoRows, tetapi err = %v", err)
		}
		if _, err := repo.Update(ctx, idTidakAda, BrandInput{Name: "Brand Tidak Ada"}); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("Update seharusnya ErrNoRows, tetapi err = %v", err)
		}
	})

	t.Run("arsip lalu pulihkan", func(t *testing.T) {
		diarsipkan, err := repo.Archive(ctx, brand.ID)
		if err != nil {
			t.Fatalf("Archive gagal: %v", err)
		}
		if diarsipkan.ArchivedAt == nil {
			t.Fatal("archived_at masih kosong setelah diarsipkan")
		}
		cekTimestampUTC(t, diarsipkan, "archived_at")

		if _, err := repo.Archive(ctx, brand.ID); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("mengarsipkan dua kali seharusnya ErrNoRows, tetapi err = %v", err)
		}

		daftarAktif, err := repo.List(ctx, ListParams{Search: namaBrandUji, Status: "active", Page: 1, PerPage: 10})
		if err != nil {
			t.Fatalf("List status active gagal: %v", err)
		}
		for _, item := range daftarAktif.Data {
			if item.ID == brand.ID {
				t.Fatal("brand yang diarsipkan masih muncul di daftar aktif")
			}
		}

		daftarArsip, err := repo.List(ctx, ListParams{Search: namaBrandUji, Status: "archived", Page: 1, PerPage: 10})
		if err != nil {
			t.Fatalf("List status archived gagal: %v", err)
		}
		var adaDiArsip bool
		for _, item := range daftarArsip.Data {
			if item.ID == brand.ID {
				adaDiArsip = true
				break
			}
		}
		if !adaDiArsip {
			t.Fatal("brand yang diarsipkan tidak muncul di daftar arsip")
		}

		dipulihkan, err := repo.Restore(ctx, brand.ID)
		if err != nil {
			t.Fatalf("Restore gagal: %v", err)
		}
		if dipulihkan.ArchivedAt != nil {
			t.Fatal("archived_at masih terisi setelah dipulihkan")
		}

		if _, err := repo.Restore(ctx, brand.ID); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("memulihkan dua kali seharusnya ErrNoRows, tetapi err = %v", err)
		}
	})
}
