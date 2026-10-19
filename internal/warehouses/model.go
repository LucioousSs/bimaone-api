package warehouses

import "time"

type Warehouse struct {
	ID         int64      `json:"id"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	Address    *string    `json:"address"`
	PICName    *string    `json:"pic_name"`
	Phone      *string    `json:"phone"`
	ArchivedAt *time.Time `json:"archived_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type CreateWarehouseInput struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Address *string `json:"address"`
	PICName *string `json:"pic_name"`
	Phone   *string `json:"phone"`
}

type UpdateWarehouseInput struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Address *string `json:"address"`
	PICName *string `json:"pic_name"`
	Phone   *string `json:"phone"`
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
	Fields  map[string]string `json:"fields,omitempty"`
}

type ErrorResponse struct {
	Error APIError `json:"error"`
}
