package brands

import "time"

// Brand adalah bentuk data yang dikirim ke frontend, sama dengan kolom tabel brands.
type Brand struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	ArchivedAt  *time.Time `json:"archived_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// BrandInput adalah body request untuk create dan update. Keduanya menerima
// field yang sama, jadi cukup satu struct.
type BrandInput struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type ListMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}
