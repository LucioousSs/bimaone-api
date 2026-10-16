package brands

import (
	"net/http"
	"strconv"
	"strings"
	"github.com/gin-gonic/gin"
)

// GetBrandByIDHandler: Mengambil satu brand berdasarkan ID
func GetBrandByIDHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "ID brand harus berupa angka",
				"fields":  gin.H{},
			},
		})
		return
	}

	brand, err := GetBrandByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Brand tidak ditemukan",
				"fields":  gin.H{},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": brand,
	})
}

// CreateBrandHandler: Menambahkan brand baru (POST)
func CreateBrandHandler(c *gin.Context) {
	var newBrand Brand
	if err := c.ShouldBindJSON(&newBrand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Format JSON rusak atau tidak valid",
				"fields": gin.H{
					"name": "Format salah",
				},
			},
		})
		return
	}

	newBrand.Name = strings.TrimSpace(newBrand.Name)

	if newBrand.Name == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Validasi gagal",
				"fields": gin.H{
					"name": "Nama brand tidak boleh kosong",
				},
			},
		})
		return
	}

	err := CreateBrand(&newBrand)
	if err != nil {
		// Tangkap error duplikat nama
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "DUPLICATE",
				"message": "Nama brand sudah dipakai",
				"fields": gin.H{
					"name": "Nama brand sudah dipakai",
				},
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Brand berhasil ditambahkan!",
		"data":    newBrand,
	})
}

// UpdateBrandHandler: Mengubah nama & deskripsi brand (PUT)
func UpdateBrandHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "ID tidak valid",
				"fields":  gin.H{},
			},
		})
		return
	}

	var req Brand
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Format data tidak valid",
				"fields":  gin.H{},
			},
		})
		return
	}

	req.ID = id
	req.Name = strings.TrimSpace(req.Name)

	err = UpdateBrand(&req)
	if err != nil {
		if err.Error() == "not_found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "NOT_FOUND",
					"message": "Brand tidak ditemukan",
					"fields":  gin.H{},
				},
			})
			return
		}
		// Jika nama kembar saat update -> 409 DUPLICATE dengan fields.name
		c.JSON(http.StatusConflict, gin.H{
			"error": gin.H{
				"code":    "DUPLICATE",
				"message": "Nama brand sudah dipakai",
				"fields": gin.H{
					"name": "Nama brand sudah dipakai",
				},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil diperbarui",
		"data":    req,
	})
}

// GetBrandsHandler: Mengambil daftar brand dengan filter, search, & pagination
func GetBrandsHandler(c *gin.Context) {
	status := c.DefaultQuery("status", "active")
	search := c.Query("search")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 10
	}

	brandsList, total, err := GetFilteredBrands(status, search, page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "Gagal mengambil data brands",
				"fields":  gin.H{},
			},
		})
		return
	}

	totalPages := (total + perPage - 1) / perPage
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"data": brandsList,
		"meta": gin.H{
			"page":        page,
			"per_page":    perPage,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// ArchiveBrandHandler: Mengarsipkan brand (PATCH)
func ArchiveBrandHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	brand, err := ArchiveBrand(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Brand tidak ditemukan",
				"fields":  gin.H{},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil diarsipkan",
		"data":    brand,
	})
}

// RestoreBrandHandler: Memulihkan brand dari arsip (PATCH)
func RestoreBrandHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	brand, err := RestoreBrand(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Brand tidak ditemukan",
				"fields":  gin.H{},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil dipulihkan",
		"data":    brand,
	})
}