package warehouses

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_CreateValidation(t *testing.T) {
	h := NewHandler(nil)

	tests := []struct {
		name           string
		body           map[string]any
		expectedStatus int
		expectedField  string
	}{
		{
			name: "Kode gudang kosong",
			body: map[string]any{
				"code": "",
				"name": "Gudang Utama",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedField:  "code",
		},
		{
			name: "Nama gudang kurang dari 2 karakter",
			body: map[string]any{
				"code": "GDG-01",
				"name": "A",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedField:  "name",
		},
		{
			name: "Format nomor telepon tidak valid",
			body: map[string]any{
				"code":  "GDG-01",
				"name":  "Gudang Utama",
				"phone": "08123abc",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedField:  "phone",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			payload, _ := json.Marshal(tc.body)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/warehouses", bytes.NewBuffer(payload))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Create(c)

			if w.Code != tc.expectedStatus {
				t.Fatalf("status code salah, diharapkan %d tetapi didapat %d", tc.expectedStatus, w.Code)
			}

			var resp ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("gagal deserialize respons: %v", err)
			}

			if resp.Error.Code != "VALIDATION_ERROR" {
				t.Fatalf("error code salah, diharapkan VALIDATION_ERROR tetapi didapat %s", resp.Error.Code)
			}

			if _, ok := resp.Error.Fields[tc.expectedField]; !ok {
				t.Fatalf("field %s tidak ada dalam respons fields: %v", tc.expectedField, resp.Error.Fields)
			}
		})
	}
}

func TestHandler_InvalidID(t *testing.T) {
	h := NewHandler(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/warehouses/abc", nil)

	h.GetByID(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status code salah, diharapkan 400 tetapi didapat %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("gagal unmarshal respons: %v", err)
	}

	if resp.Error.Code != "BAD_REQUEST" {
		t.Fatalf("error code salah, diharapkan BAD_REQUEST tetapi didapat %s", resp.Error.Code)
	}
}
