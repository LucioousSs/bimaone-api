package warehouses

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func respondError(c *gin.Context, status int, code, message string, fields map[string]string) {
	c.JSON(status, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
	})
}

// GET /api/v1/warehouses
func (h *Handler) List(c *gin.Context) {
	page, err := parseQueryInt(c, "page", 1)
	if err != nil || page < 1 {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "Parameter page harus berupa bilangan bulat positif", nil)
		return
	}

	perPage, err := parseQueryInt(c, "per_page", 10)
	if err != nil || perPage < 1 || perPage > 100 {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "Parameter per_page harus antara 1 dan 100", nil)
		return
	}

	result, err := h.repo.List(c.Request.Context(), ListParams{
		Search:  c.Query("search"),
		Status:  c.DefaultQuery("status", "active"),
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "status tidak valid") {
			respondError(c, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil)
			return
		}

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil daftar gudang", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result.Data,
		"meta": ListMeta{
			Page:       result.Page,
			PerPage:    result.PerPage,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	})
}

// POST /api/v1/warehouses
func (h *Handler) Create(c *gin.Context) {
	var input CreateWarehouseInput

	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "Format JSON tidak valid", nil)
		return
	}

	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	input.Name = strings.TrimSpace(input.Name)

	fields := make(map[string]string)

	if input.Code == "" {
		fields["code"] = "Kode gudang wajib diisi"
	} else if len(input.Code) > 20 {
		fields["code"] = "Kode gudang maksimal 20 karakter"
	}

	if input.Name == "" {
		fields["name"] = "Nama gudang wajib diisi"
	} else {
		nameLen := utf8.RuneCountInString(input.Name)
		if nameLen < 2 || nameLen > 100 {
			fields["name"] = "Nama gudang harus terdiri dari 2 sampai 100 karakter"
		}
	}

	input.Address = normalizeOptional(input.Address)
	input.PICName = normalizeOptional(input.PICName)
	input.Phone = normalizeOptional(input.Phone)

	if input.PICName != nil && utf8.RuneCountInString(*input.PICName) > 100 {
		fields["pic_name"] = "Nama PIC maksimal 100 karakter"
	}

	if input.Phone != nil {
		if len(*input.Phone) > 20 {
			fields["phone"] = "Nomor telepon maksimal 20 karakter"
		} else if !isValidPhone(*input.Phone) {
			fields["phone"] = "Format nomor telepon tidak valid"
		}
	}

	if len(fields) > 0 {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Data tidak valid", fields)
		return
	}

	warehouse, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			respondError(c, http.StatusConflict, "DUPLICATE", "Kode gudang sudah digunakan", map[string]string{
				"code": "Kode gudang sudah digunakan",
			})
			return
		}

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal menyimpan gudang", nil)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": warehouse,
	})
}

// GET /api/v1/warehouses/:id
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "ID gudang tidak valid", nil)
		return
	}

	warehouse, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Data dengan id itu tidak ada", nil)
			return
		}

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengambil detail gudang", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warehouse,
	})
}

// PUT /api/v1/warehouses/:id
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "ID gudang tidak valid", nil)
		return
	}

	var input UpdateWarehouseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "Format JSON tidak valid", nil)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	fields := make(map[string]string)

	if input.Name == "" {
		fields["name"] = "Nama gudang wajib diisi"
	} else {
		nameLen := utf8.RuneCountInString(input.Name)
		if nameLen < 2 || nameLen > 100 {
			fields["name"] = "Nama gudang harus terdiri dari 2 sampai 100 karakter"
		}
	}

	input.Address = normalizeOptional(input.Address)
	input.PICName = normalizeOptional(input.PICName)
	input.Phone = normalizeOptional(input.Phone)

	if input.PICName != nil && utf8.RuneCountInString(*input.PICName) > 100 {
		fields["pic_name"] = "Nama PIC maksimal 100 karakter"
	}

	if input.Phone != nil {
		if len(*input.Phone) > 20 {
			fields["phone"] = "Nomor telepon maksimal 20 karakter"
		} else if !isValidPhone(*input.Phone) {
			fields["phone"] = "Format nomor telepon tidak valid"
		}
	}

	if len(fields) > 0 {
		respondError(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Data tidak valid", fields)
		return
	}

	warehouse, err := h.repo.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Data dengan id itu tidak ada", nil)
			return
		}

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memperbarui gudang", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warehouse,
	})
}

// PATCH /api/v1/warehouses/:id/archive
func (h *Handler) Archive(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "ID gudang tidak valid", nil)
		return
	}

	warehouse, err := h.repo.Archive(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Data dengan id itu tidak ada atau sudah diarsipkan", nil)
			return
		}

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal mengarsipkan gudang", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warehouse,
	})
}

// PATCH /api/v1/warehouses/:id/restore
func (h *Handler) Restore(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "ID gudang tidak valid", nil)
		return
	}

	warehouse, err := h.repo.Restore(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "NOT_FOUND", "Data dengan id itu tidak ada atau belum diarsipkan", nil)
			return
		}

		respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Gagal memulihkan gudang", nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": warehouse,
	})
}

func isValidPhone(phone string) bool {
	for _, ch := range phone {
		if !((ch >= '0' && ch <= '9') || ch == '+' || ch == ' ' || ch == '-') {
			return false
		}
	}
	return true
}

func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}

	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}

	return &normalized
}

func parseQueryInt(
	c *gin.Context,
	key string,
	defaultValue int,
) (int, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(value)
}
