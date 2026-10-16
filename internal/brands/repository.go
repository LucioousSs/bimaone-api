package brands

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
	Data       []Brand
	Page       int
	PerPage    int
	Total      int64
	TotalPages int64
}

// brandColumns ditulis sekali supaya daftar kolom dan urutan Scan tidak pernah
// berbeda antar query.
const brandColumns = `id, name, description, archived_at, created_at, updated_at`

// rowScanner dipenuhi oleh pgx.Row (satu baris) maupun pgx.Rows (banyak baris),
// jadi cara membaca kolomnya cukup ditulis sekali.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanBrand(row rowScanner, brand *Brand) error {
	if err := row.Scan(
		&brand.ID,
		&brand.Name,
		&brand.Description,
		&brand.ArchivedAt,
		&brand.CreatedAt,
		&brand.UpdatedAt,
	); err != nil {
		return err
	}

	// pgx mengembalikan waktu dalam zona lokal mesin, sedangkan kontrak API
	// meminta UTC. Semua baris dibaca lewat fungsi ini, jadi cukup di sini.
	brand.CreatedAt = brand.CreatedAt.UTC()
	brand.UpdatedAt = brand.UpdatedAt.UTC()
	if brand.ArchivedAt != nil {
		archived := brand.ArchivedAt.UTC()
		brand.ArchivedAt = &archived
	}

	return nil
}

func (repo *Repository) List(ctx context.Context, params ListParams) (ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	}
	if params.PerPage > 100 {
		params.PerPage = 100
	}

	// where menampung syarat yang dipakai bersama oleh query hitung dan query data,
	// supaya total selalu cocok dengan isi halamannya.
	where := []string{}
	args := []any{}

	switch params.Status {
	case "archived":
		where = append(where, "archived_at IS NOT NULL")
	case "all":
		// Tidak memfilter status arsip.
	default:
		// "active" dan string kosong diperlakukan sama.
		where = append(where, "archived_at IS NULL")
	}

	if search := strings.TrimSpace(params.Search); search != "" {
		args = append(args, "%"+escapeLikePattern(search)+"%")
		where = append(where, fmt.Sprintf("name ILIKE $%d", len(args)))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	countQuery := `SELECT COUNT(*) FROM brands` + whereSQL
	if err := repo.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return ListResult{}, fmt.Errorf("gagal menghitung brand: %w", err)
	}

	offset := (params.Page - 1) * params.PerPage
	listArgs := append([]any{}, args...)
	listArgs = append(listArgs, params.PerPage, offset)

	// Urut A-Z tanpa peduli huruf besar/kecil, sama seperti yang dilihat user di tabel.
	query := `SELECT ` + brandColumns + ` FROM brands` + whereSQL +
		fmt.Sprintf(` ORDER BY LOWER(name) ASC LIMIT $%d OFFSET $%d`, len(listArgs)-1, len(listArgs))

	rows, err := repo.db.Query(ctx, query, listArgs...)
	if err != nil {
		return ListResult{}, fmt.Errorf("gagal mengambil daftar brand: %w", err)
	}
	defer rows.Close()

	data := make([]Brand, 0)

	for rows.Next() {
		var item Brand
		if err := scanBrand(rows, &item); err != nil {
			return ListResult{}, fmt.Errorf("gagal membaca data brand: %w", err)
		}

		data = append(data, item)
	}

	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("gagal membaca hasil query: %w", err)
	}

	// Minimal 1 supaya frontend tidak menampilkan "halaman 1 dari 0" saat datanya kosong.
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

func (repo *Repository) GetByID(ctx context.Context, id int64) (Brand, error) {
	query := `SELECT ` + brandColumns + ` FROM brands WHERE id = $1`

	var item Brand
	if err := scanBrand(repo.db.QueryRow(ctx, query, id), &item); err != nil {
		// %w membuat pgx.ErrNoRows tetap bisa dikenali handler lewat errors.Is.
		return Brand{}, fmt.Errorf("gagal mengambil brand: %w", err)
	}

	return item, nil
}

func (repo *Repository) Create(ctx context.Context, input BrandInput) (Brand, error) {
	query := `INSERT INTO brands (name, description)
	          VALUES ($1, $2)
	          RETURNING ` + brandColumns

	var item Brand
	if err := scanBrand(repo.db.QueryRow(ctx, query, input.Name, input.Description), &item); err != nil {
		return Brand{}, fmt.Errorf("gagal menyimpan brand: %w", err)
	}

	return item, nil
}

func (repo *Repository) Update(ctx context.Context, id int64, input BrandInput) (Brand, error) {
	query := `UPDATE brands
	          SET name = $1, description = $2, updated_at = NOW()
	          WHERE id = $3
	          RETURNING ` + brandColumns

	var item Brand
	if err := scanBrand(repo.db.QueryRow(ctx, query, input.Name, input.Description, id), &item); err != nil {
		return Brand{}, fmt.Errorf("gagal memperbarui brand: %w", err)
	}

	return item, nil
}

func (repo *Repository) Archive(ctx context.Context, id int64) (Brand, error) {
	// archived_at IS NULL membuat brand yang sudah diarsipkan tidak ikut ter-update,
	// sehingga pgx.ErrNoRows muncul dan handler bisa membalas 404.
	query := `UPDATE brands
	          SET archived_at = NOW(), updated_at = NOW()
	          WHERE id = $1 AND archived_at IS NULL
	          RETURNING ` + brandColumns

	var item Brand
	if err := scanBrand(repo.db.QueryRow(ctx, query, id), &item); err != nil {
		return Brand{}, fmt.Errorf("gagal mengarsipkan brand: %w", err)
	}

	return item, nil
}

func (repo *Repository) Restore(ctx context.Context, id int64) (Brand, error) {
	// Kebalikan dari Archive: hanya brand yang sedang diarsipkan yang bisa dipulihkan.
	query := `UPDATE brands
	          SET archived_at = NULL, updated_at = NOW()
	          WHERE id = $1 AND archived_at IS NOT NULL
	          RETURNING ` + brandColumns

	var item Brand
	if err := scanBrand(repo.db.QueryRow(ctx, query, id), &item); err != nil {
		return Brand{}, fmt.Errorf("gagal memulihkan brand: %w", err)
	}

	return item, nil
}

// Karakter % dan _ punya arti khusus di LIKE. Fungsi ini membuat kata yang
// dicari user diperlakukan sebagai teks biasa.
func escapeLikePattern(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
