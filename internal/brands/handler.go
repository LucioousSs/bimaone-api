package brands

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Kode error yang dipakai di kontrak API.
const (
	codeBadRequest = "BAD_REQUEST"
	codeNotFound   = "NOT_FOUND"
	codeDuplicate  = "DUPLICATE"
	codeValidation = "VALIDATION_ERROR"
	codeInternal   = "INTERNAL_ERROR"
)

const (
	msgInvalidID     = "ID brand harus berupa angka"
	msgNotFound      = "Brand tidak ditemukan"
	msgInvalidJSON   = "Format JSON tidak valid"
	msgInvalidData   = "Data tidak valid"
	msgDuplicateName = "Nama brand sudah dipakai"
)

// Batas panjang harus sama dengan CHECK constraint di migration.
const (
	nameMinLength        = 2
	nameMaxLength        = 100
	descriptionMaxLength = 500
)

// Kode unique_violation dari PostgreSQL.
const uniqueViolationCode = "23505"

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func respondError(c *gin.Context, status int, code, message string, fields map[string]string) {
	// Map nil akan jadi null di JSON, sedangkan kontrak API meminta {}.
	if fields == nil {
		fields = map[string]string{}
	}

	c.JSON(status, ErrorResponse{
		Error: APIError{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
	})
}

// Detail teknis hanya masuk log server, frontend cukup menerima pesan umum.
func respondInternalError(c *gin.Context, message string, err error) {
	log.Printf("brands: %s: %v", message, err)
	respondError(c, http.StatusInternalServerError, codeInternal, message, nil)
}

func duplicateNameFields() map[string]string {
	return map[string]string{"name": msgDuplicateName}
}

// isUniqueViolation mendeteksi nama brand kembar. Constraint di database yang
// jadi penentu, jadi tetap benar walau dua request masuk bersamaan.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode
}

// parseIDParam mengambil :id dari URL. Kalau tidak valid, respons 400 sudah
// dikirim dan fungsi ini mengembalikan ok = false.
func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		respondError(c, http.StatusBadRequest, codeBadRequest, msgInvalidID, nil)
		return 0, false
	}

	return id, true
}

// parseStatus menolak status yang tidak dikenal, supaya salah ketik di frontend
// tidak diam-diam berubah menjadi "tampilkan semua".
func parseStatus(c *gin.Context) (string, bool) {
	status := c.DefaultQuery("status", "active")

	switch status {
	case "", "active", "archived", "all":
		return status, true
	default:
		respondError(c, http.StatusBadRequest, codeBadRequest,
			"Parameter status harus salah satu dari: active, archived, all", nil)
		return "", false
	}
}

func parseQueryInt(c *gin.Context, key string, defaultValue int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(value)
}

// List menangani GET /api/v1/brands
func (h *Handler) List(c *gin.Context) {
	page, err := parseQueryInt(c, "page", 1)
	if err != nil || page < 1 {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Parameter page harus berupa bilangan bulat positif", nil)
		return
	}

	perPage, err := parseQueryInt(c, "per_page", 10)
	if err != nil || perPage < 1 || perPage > 100 {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Parameter per_page harus antara 1 dan 100", nil)
		return
	}

	status, ok := parseStatus(c)
	if !ok {
		return
	}

	result, err := h.repo.List(c.Request.Context(), ListParams{
		Search:  c.Query("search"),
		Status:  status,
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		respondInternalError(c, "Gagal mengambil daftar brand", err)
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

// GetByID menangani GET /api/v1/brands/:id
func (h *Handler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	brand, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, codeNotFound, msgNotFound, nil)
			return
		}

		respondInternalError(c, "Gagal mengambil brand", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": brand,
	})
}

// Create menangani POST /api/v1/brands
func (h *Handler) Create(c *gin.Context) {
	var input BrandInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, msgInvalidJSON, nil)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = normalizeOptional(input.Description)

	if fields := validateBrandInput(input); len(fields) > 0 {
		respondError(c, http.StatusUnprocessableEntity, codeValidation, msgInvalidData, fields)
		return
	}

	brand, err := h.repo.Create(c.Request.Context(), input)
	if err != nil {
		if isUniqueViolation(err) {
			respondError(c, http.StatusConflict, codeDuplicate, msgDuplicateName, duplicateNameFields())
			return
		}

		respondInternalError(c, "Gagal menyimpan brand", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Brand berhasil ditambahkan",
		"data":    brand,
	})
}

// Update menangani PUT /api/v1/brands/:id
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var input BrandInput
	if err := c.ShouldBindJSON(&input); err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, msgInvalidJSON, nil)
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Description = normalizeOptional(input.Description)

	if fields := validateBrandInput(input); len(fields) > 0 {
		respondError(c, http.StatusUnprocessableEntity, codeValidation, msgInvalidData, fields)
		return
	}

	brand, err := h.repo.Update(c.Request.Context(), id, input)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, codeNotFound, msgNotFound, nil)
			return
		}
		if isUniqueViolation(err) {
			respondError(c, http.StatusConflict, codeDuplicate, msgDuplicateName, duplicateNameFields())
			return
		}

		respondInternalError(c, "Gagal memperbarui brand", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil diperbarui",
		"data":    brand,
	})
}

// Archive menangani PATCH /api/v1/brands/:id/archive
func (h *Handler) Archive(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	brand, err := h.repo.Archive(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, codeNotFound, "Brand tidak ditemukan atau sudah diarsipkan", nil)
			return
		}

		respondInternalError(c, "Gagal mengarsipkan brand", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil diarsipkan",
		"data":    brand,
	})
}

// Restore menangani PATCH /api/v1/brands/:id/restore
func (h *Handler) Restore(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	brand, err := h.repo.Restore(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, codeNotFound, "Brand tidak ditemukan atau belum diarsipkan", nil)
			return
		}

		respondInternalError(c, "Gagal memulihkan brand", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil dipulihkan",
		"data":    brand,
	})
}

// validateBrandInput mengembalikan pesan per field. Map kosong berarti datanya
// valid, jadi pemanggil cukup memeriksa len(fields) > 0.
func validateBrandInput(input BrandInput) map[string]string {
	fields := make(map[string]string)

	// RuneCountInString menghitung karakter, bukan byte, sehingga huruf beraksen
	// dihitung sama seperti VARCHAR di PostgreSQL.
	nameLength := utf8.RuneCountInString(input.Name)
	if input.Name == "" {
		fields["name"] = "Nama brand wajib diisi"
	} else if nameLength < nameMinLength || nameLength > nameMaxLength {
		fields["name"] = fmt.Sprintf("Nama brand harus terdiri dari %d sampai %d karakter", nameMinLength, nameMaxLength)
	}

	if input.Description != nil && utf8.RuneCountInString(*input.Description) > descriptionMaxLength {
		fields["description"] = fmt.Sprintf("Deskripsi maksimal %d karakter", descriptionMaxLength)
	}

	return fields
}

// Deskripsi kosong disimpan sebagai NULL, bukan string kosong, supaya tidak ada dua bentuk "kosong".
func normalizeOptional(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
