package warehouses

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}
type Warehouse struct {
	ID         int64   `json:"id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Address    *string `json:"address"`
	PICName    *string `json:"pic_name"`
	Phone      *string `json:"phone"`
	ArchivedAt *string `json:"archived_at"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type ListParams struct {
	Search  string
	Status  string
	Page    int
	PerPage int
}

type ListResult struct {
	Data       []Warehouse `json:"data"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	Total      int64       `json:"total"`
	TotalPages int64       `json:"total_pages"`
}

func (repo *Repository) List(
	ctx context.Context,
	params ListParams,
) (ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	search := strings.TrimSpace(params.Search)

	where := []string{}
	args := []any{}

	if search != "" {
		args = append(args, "%"+search+"%")
		where = append(where, fmt.Sprintf(
			"(code ILIKE $%d OR name ILIKE $%d)",
			len(args), len(args),
		))
	}

	switch params.Status {
	case "", "active":
		where = append(where, "archived_at IS NULL")
	case "archived":
		where = append(where, "archived_at IS NOT NULL")
	case "all":
		// Tidak menambahkan filter status.
	default:
		return ListResult{}, fmt.Errorf("status tidak valid")
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM warehouses" + whereSQL
	if err := repo.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return ListResult{}, fmt.Errorf("gagal menghitung gudang: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, params.PerPage, offset)

	query := `
		SELECT id, code, name, address, pic_name, phone,
		       archived_at, created_at, updated_at
		FROM warehouses` + whereSQL + fmt.Sprintf(`
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d`,
		len(listArgs)-1, len(listArgs),
	)

	rows, err := repo.db.Query(ctx, query, listArgs...)
	if err != nil {
		return ListResult{}, fmt.Errorf("gagal mengambil daftar gudang: %w", err)
	}
	defer rows.Close()

	data := make([]Warehouse, 0)
	for rows.Next() {
		var item Warehouse
		var archivedAt, createdAt, updatedAt *string

		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.Address,
			&item.PICName,
			&item.Phone,
			&archivedAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return ListResult{}, fmt.Errorf("gagal membaca data gudang: %w", err)
		}

		item.ArchivedAt = archivedAt
		if createdAt != nil {
			item.CreatedAt = *createdAt
		}
		if updatedAt != nil {
			item.UpdatedAt = *updatedAt
		}

		data = append(data, item)
	}

	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("gagal membaca hasil query: %w", err)
	}

	totalPages := total / int64(params.PerPage)
	if total%int64(params.PerPage) != 0 {
		totalPages++
	}

	return ListResult{
		Data:       data,
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
