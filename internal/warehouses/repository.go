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

// rowScanner dipenuhi oleh pgx.Row (satu baris) maupun pgx.Rows (banyak baris),
// jadi cara membaca kolomnya cukup ditulis sekali.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanWarehouse(row rowScanner, item *Warehouse) error {
	if err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Address,
		&item.PICName,
		&item.Phone,
		&item.ArchivedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return err
	}

	// pgx mengembalikan waktu dalam zona lokal mesin, sedangkan kontrak API
	// meminta UTC. Semua baris dibaca lewat fungsi ini, jadi cukup di sini.
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	if item.ArchivedAt != nil {
		archived := item.ArchivedAt.UTC()
		item.ArchivedAt = &archived
	}

	return nil
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
		args = append(args, "%"+escapeLikePattern(search)+"%")
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
		// Tidak memfilter berdasarkan status arsip.
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
		ORDER BY LOWER(name) ASC
		LIMIT $%d OFFSET $%d`,
		len(listArgs)-1, len(listArgs),
	)

	rows, err := repo.db.Query(ctx, query, listArgs...)
	if err != nil {
		return ListResult{}, fmt.Errorf(
			"gagal mengambil daftar gudang: %w", err,
		)
	}
	defer rows.Close()

	data := make([]Warehouse, 0)

	for rows.Next() {
		var item Warehouse

		if err := scanWarehouse(rows, &item); err != nil {
			return ListResult{}, fmt.Errorf(
				"gagal membaca data gudang: %w", err,
			)
		}

		data = append(data, item)
	}

	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf(
			"gagal membaca hasil query: %w", err,
		)
	}

	totalPages := total / int64(params.PerPage)
	if total%int64(params.PerPage) != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return ListResult{
		Data:       data,
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func (repo *Repository) Create(
	ctx context.Context,
	input CreateWarehouseInput,
) (Warehouse, error) {
	const query = `
		INSERT INTO warehouses (code, name, address, pic_name, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, code, name, address, pic_name, phone,
		          archived_at, created_at, updated_at
	`

	var item Warehouse

	err := scanWarehouse(repo.db.QueryRow(
		ctx,
		query,
		input.Code,
		input.Name,
		input.Address,
		input.PICName,
		input.Phone,
	), &item)
	if err != nil {
		return Warehouse{}, fmt.Errorf("gagal membuat gudang: %w", err)
	}

	return item, nil
}

func (repo *Repository) GetByID(
	ctx context.Context,
	id int64,
) (Warehouse, error) {
	const query = `
		SELECT id, code, name, address, pic_name, phone,
		       archived_at, created_at, updated_at
		FROM warehouses
		WHERE id = $1
	`

	var item Warehouse

	err := scanWarehouse(repo.db.QueryRow(ctx, query, id), &item)
	if err != nil {
		return Warehouse{}, fmt.Errorf("gagal mengambil gudang: %w", err)
	}

	return item, nil
}

func (repo *Repository) Update(
	ctx context.Context,
	id int64,
	input UpdateWarehouseInput,
) (Warehouse, error) {
	const query = `
		UPDATE warehouses
		SET code = $1,
		    name = $2,
		    address = $3,
		    pic_name = $4,
		    phone = $5,
		    updated_at = NOW()
		WHERE id = $6
		RETURNING id, code, name, address, pic_name, phone,
		          archived_at, created_at, updated_at
	`

	var item Warehouse

	err := scanWarehouse(repo.db.QueryRow(
		ctx,
		query,
		input.Code,
		input.Name,
		input.Address,
		input.PICName,
		input.Phone,
		id,
	), &item)
	if err != nil {
		return Warehouse{}, fmt.Errorf("gagal memperbarui gudang: %w", err)
	}

	return item, nil
}

func (repo *Repository) Archive(
	ctx context.Context,
	id int64,
) (Warehouse, error) {
	const query = `
		UPDATE warehouses
		SET archived_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1 AND archived_at IS NULL
		RETURNING id, code, name, address, pic_name, phone,
		          archived_at, created_at, updated_at
	`

	var item Warehouse

	err := scanWarehouse(repo.db.QueryRow(ctx, query, id), &item)
	if err != nil {
		return Warehouse{}, fmt.Errorf("gagal mengarsipkan gudang: %w", err)
	}

	return item, nil
}

func (repo *Repository) Restore(
	ctx context.Context,
	id int64,
) (Warehouse, error) {
	const query = `
		UPDATE warehouses
		SET archived_at = NULL,
		    updated_at = NOW()
		WHERE id = $1 AND archived_at IS NOT NULL
		RETURNING id, code, name, address, pic_name, phone,
		          archived_at, created_at, updated_at
	`

	var item Warehouse

	err := scanWarehouse(repo.db.QueryRow(ctx, query, id), &item)
	if err != nil {
		return Warehouse{}, fmt.Errorf("gagal memulihkan gudang: %w", err)
	}

	return item, nil
}

// Karakter % dan _ punya arti khusus di LIKE. Fungsi ini membuat kata yang
// dicari user diperlakukan sebagai teks biasa.
func escapeLikePattern(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
