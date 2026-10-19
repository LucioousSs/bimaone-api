package brands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestContext(method, target string, body io.Reader) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, body)
	c.Request.Header.Set("Content-Type", "application/json")

	return w, c
}

func decodeError(t *testing.T, body []byte) ErrorResponse {
	t.Helper()

	var resp ErrorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("gagal membaca respons error: %v", err)
	}

	return resp
}

func stringPtr(value string) *string {
	return &value
}

// Semua kasus di sini sengaja tidak valid. Input yang lolos validasi akan
// memanggil repository, sedangkan handler di tes ini tidak punya database.
func TestHandler_CreateValidation(t *testing.T) {
	h := NewHandler(nil)

	tests := []struct {
		name          string
		body          map[string]any
		expectedField string
	}{
		{
			name:          "Nama kosong",
			body:          map[string]any{"name": ""},
			expectedField: "name",
		},
		{
			name:          "Nama hanya spasi",
			body:          map[string]any{"name": "   "},
			expectedField: "name",
		},
		{
			name:          "Nama 1 karakter",
			body:          map[string]any{"name": "A"},
			expectedField: "name",
		},
		{
			name:          "Spasi di ujung dipangkas sebelum dihitung",
			body:          map[string]any{"name": " A "},
			expectedField: "name",
		},
		{
			name:          "Nama 101 karakter",
			body:          map[string]any{"name": strings.Repeat("a", 101)},
			expectedField: "name",
		},
		{
			name:          "Nama 1 huruf non-latin",
			body:          map[string]any{"name": "あ"},
			expectedField: "name",
		},
		{
			name:          "Deskripsi 501 karakter",
			body:          map[string]any{"name": "Grundfos", "description": strings.Repeat("a", 501)},
			expectedField: "description",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatalf("gagal menyiapkan body: %v", err)
			}

			w, c := newTestContext(http.MethodPost, "/api/v1/brands", bytes.NewReader(payload))

			h.Create(c)

			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status diharapkan 422 tetapi didapat %d", w.Code)
			}

			resp := decodeError(t, w.Body.Bytes())
			if resp.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("code diharapkan VALIDATION_ERROR tetapi didapat %s", resp.Error.Code)
			}
			if _, ok := resp.Error.Fields[tc.expectedField]; !ok {
				t.Fatalf("fields tidak memuat %q: %v", tc.expectedField, resp.Error.Fields)
			}
		})
	}
}

func TestHandler_CreateInvalidJSON(t *testing.T) {
	h := NewHandler(nil)

	w, c := newTestContext(http.MethodPost, "/api/v1/brands", strings.NewReader(`{"name":`))

	h.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status diharapkan 400 tetapi didapat %d", w.Code)
	}

	resp := decodeError(t, w.Body.Bytes())
	if resp.Error.Code != "BAD_REQUEST" {
		t.Fatalf("code diharapkan BAD_REQUEST tetapi didapat %s", resp.Error.Code)
	}
}

func TestHandler_InvalidID(t *testing.T) {
	h := NewHandler(nil)

	tests := []struct {
		name     string
		id       string
		jalankan func(*gin.Context)
	}{
		{name: "GetByID id bukan angka", id: "abc", jalankan: h.GetByID},
		{name: "GetByID id nol", id: "0", jalankan: h.GetByID},
		{name: "GetByID id negatif", id: "-1", jalankan: h.GetByID},
		{name: "Update id bukan angka", id: "abc", jalankan: h.Update},
		{name: "Archive id bukan angka", id: "abc", jalankan: h.Archive},
		{name: "Restore id bukan angka", id: "abc", jalankan: h.Restore},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, c := newTestContext(http.MethodGet, "/api/v1/brands/"+tc.id, nil)
			c.Params = gin.Params{{Key: "id", Value: tc.id}}

			tc.jalankan(c)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status diharapkan 400 tetapi didapat %d", w.Code)
			}

			resp := decodeError(t, w.Body.Bytes())
			if resp.Error.Code != "BAD_REQUEST" {
				t.Fatalf("code diharapkan BAD_REQUEST tetapi didapat %s", resp.Error.Code)
			}
		})
	}
}

// Parameter yang salah ketik harus ditolak, bukan diam-diam dianggap default.
func TestHandler_InvalidListQuery(t *testing.T) {
	h := NewHandler(nil)

	tests := []struct {
		name  string
		query string
	}{
		{name: "page bukan angka", query: "page=abc"},
		{name: "page nol", query: "page=0"},
		{name: "page negatif", query: "page=-1"},
		{name: "per_page bukan angka", query: "per_page=abc"},
		{name: "per_page nol", query: "per_page=0"},
		{name: "per_page lebih dari 100", query: "per_page=101"},
		{name: "status tidak dikenal", query: "status=bogus"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w, c := newTestContext(http.MethodGet, "/api/v1/brands?"+tc.query, nil)

			h.List(c)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("status diharapkan 400 tetapi didapat %d", w.Code)
			}

			resp := decodeError(t, w.Body.Bytes())
			if resp.Error.Code != "BAD_REQUEST" {
				t.Fatalf("code diharapkan BAD_REQUEST tetapi didapat %s", resp.Error.Code)
			}
		})
	}
}

// Kontrak API meminta "fields" selalu ada dan berbentuk objek, walau kosong.
// Kalau jadi null atau hilang, frontend gagal membaca error.fields.name.
func TestHandler_ErrorFieldsAlwaysObject(t *testing.T) {
	h := NewHandler(nil)

	assertFieldsNotNull := func(t *testing.T, body []byte) {
		t.Helper()

		if decodeError(t, body).Error.Fields == nil {
			t.Fatal(`"fields" harus ada dan berbentuk {} bukan null`)
		}
	}

	t.Run("400 id bukan angka", func(t *testing.T) {
		w, c := newTestContext(http.MethodGet, "/api/v1/brands/abc", nil)
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		h.GetByID(c)

		assertFieldsNotNull(t, w.Body.Bytes())
	})

	t.Run("400 status tidak dikenal", func(t *testing.T) {
		w, c := newTestContext(http.MethodGet, "/api/v1/brands?status=bogus", nil)

		h.List(c)

		assertFieldsNotNull(t, w.Body.Bytes())
	})

	t.Run("422 validasi", func(t *testing.T) {
		w, c := newTestContext(http.MethodPost, "/api/v1/brands", strings.NewReader(`{"name":""}`))

		h.Create(c)

		assertFieldsNotNull(t, w.Body.Bytes())
	})
}

// Panjang dihitung per karakter, bukan per byte. Kalau memakai len(), huruf
// non-latin yang pendek akan ikut ditolak.
func TestValidateBrandInput(t *testing.T) {
	namaValid := "Grundfos"

	tests := []struct {
		name      string
		input     BrandInput
		wantValid bool
		wantField string
	}{
		{
			name:      "Nama kosong",
			input:     BrandInput{Name: ""},
			wantValid: false,
			wantField: "name",
		},
		{
			name:      "Nama 1 karakter",
			input:     BrandInput{Name: "A"},
			wantValid: false,
			wantField: "name",
		},
		{
			name:      "Nama 2 karakter",
			input:     BrandInput{Name: "AB"},
			wantValid: true,
		},
		{
			name:      "Nama 100 karakter",
			input:     BrandInput{Name: strings.Repeat("a", 100)},
			wantValid: true,
		},
		{
			name:      "Nama 101 karakter",
			input:     BrandInput{Name: strings.Repeat("a", 101)},
			wantValid: false,
			wantField: "name",
		},
		{
			name:      "Nama 1 huruf non-latin",
			input:     BrandInput{Name: "あ"},
			wantValid: false,
			wantField: "name",
		},
		{
			name:      "Nama 2 huruf non-latin",
			input:     BrandInput{Name: "日本"},
			wantValid: true,
		},
		{
			name:      "Nama 100 huruf non-latin",
			input:     BrandInput{Name: strings.Repeat("あ", 100)},
			wantValid: true,
		},
		{
			name:      "Nama 101 huruf non-latin",
			input:     BrandInput{Name: strings.Repeat("あ", 101)},
			wantValid: false,
			wantField: "name",
		},
		{
			name:      "Deskripsi kosong",
			input:     BrandInput{Name: namaValid, Description: stringPtr("")},
			wantValid: true,
		},
		{
			name:      "Deskripsi 500 karakter",
			input:     BrandInput{Name: namaValid, Description: stringPtr(strings.Repeat("a", 500))},
			wantValid: true,
		},
		{
			name:      "Deskripsi 501 karakter",
			input:     BrandInput{Name: namaValid, Description: stringPtr(strings.Repeat("a", 501))},
			wantValid: false,
			wantField: "description",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fields := validateBrandInput(tc.input)

			if tc.wantValid {
				if len(fields) > 0 {
					t.Fatalf("input dianggap tidak valid: %v", fields)
				}
				return
			}

			if _, ok := fields[tc.wantField]; !ok {
				t.Fatalf("fields tidak memuat %q: %v", tc.wantField, fields)
			}
		})
	}
}

// Deskripsi kosong disimpan sebagai NULL, jadi hanya ada satu bentuk "kosong".
func TestNormalizeOptional(t *testing.T) {
	tests := []struct {
		name  string
		input *string
		want  *string
	}{
		{name: "nil tetap nil", input: nil, want: nil},
		{name: "string kosong jadi nil", input: stringPtr(""), want: nil},
		{name: "hanya spasi jadi nil", input: stringPtr("   "), want: nil},
		{name: "spasi di ujung dipangkas", input: stringPtr("  Pompa  "), want: stringPtr("Pompa")},
		{name: "teks biasa dibiarkan", input: stringPtr("Pompa"), want: stringPtr("Pompa")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeOptional(tc.input)

			if tc.want == nil {
				if got != nil {
					t.Fatalf("diharapkan nil tetapi didapat %q", *got)
				}
				return
			}

			if got == nil {
				t.Fatalf("diharapkan %q tetapi didapat nil", *tc.want)
			}
			if *got != *tc.want {
				t.Fatalf("diharapkan %q tetapi didapat %q", *tc.want, *got)
			}
		})
	}
}

// Repository membungkus error dengan %w, jadi deteksinya harus tetap bekerja
// lewat errors.As.
func TestIsUniqueViolation(t *testing.T) {
	t.Run("kode 23505 yang dibungkus", func(t *testing.T) {
		err := fmt.Errorf("gagal menyimpan brand: %w", &pgconn.PgError{Code: "23505"})

		if !isUniqueViolation(err) {
			t.Fatal("error 23505 seharusnya terdeteksi")
		}
	})

	t.Run("kode 23505 tanpa pembungkus", func(t *testing.T) {
		if !isUniqueViolation(&pgconn.PgError{Code: "23505"}) {
			t.Fatal("error 23505 seharusnya terdeteksi")
		}
	})

	t.Run("kode lain", func(t *testing.T) {
		err := fmt.Errorf("gagal menyimpan brand: %w", &pgconn.PgError{Code: "23503"})

		if isUniqueViolation(err) {
			t.Fatal("kode selain 23505 tidak boleh dianggap duplikat")
		}
	})

	t.Run("error biasa", func(t *testing.T) {
		if isUniqueViolation(errors.New("error biasa")) {
			t.Fatal("error biasa tidak boleh dianggap duplikat")
		}
	})
}
