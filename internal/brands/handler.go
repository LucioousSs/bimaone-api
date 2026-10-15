package brands

import (
	"net/http"
	"strconv" // Untuk mengubah teks ID dari URL (misal: "10") menjadi angka (10)
	"github.com/gin-gonic/gin"
)

// GetBrandsHandler: Melihat semua daftar brand
func GetBrandsHandler(c *gin.Context) {
	brandsList, err := GetAllBrands()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data brands: " + err.Error(),
		})
		return
	}

	if brandsList == nil {
		brandsList = []Brand{}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data brands",
		"data":    brandsList,
	})
}

// CreateBrandHandler: Menambah brand baru
func CreateBrandHandler(c *gin.Context) {
	var newBrand Brand

	if err := c.ShouldBindJSON(&newBrand); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Format data tidak valid: " + err.Error(),
		})
		return
	}

	err := CreateBrand(&newBrand)
	if err != nil {
		// Jika error karena nama kembar (duplikat), berikan pesan yang jelas
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gagal menyimpan brand: Nama brand sudah terdaftar!",
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
	id, _ := strconv.Atoi(c.Param("id"))

	var req Brand
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data tidak valid"})
		return
	}

	req.ID = id
	err := UpdateBrand(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gagal mengupdate brand: Nama brand mungkin sudah dipakai!",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil diperbarui",
		"data":    req,
	})
}

// ArchiveBrandHandler: Mengarsipkan brand (PATCH) - Mengembalikan data di "data"
func ArchiveBrandHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	brand, err := ArchiveBrand(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengarsipkan brand"})
		return
	}

	// Mengirim respons sukses lengkap dengan objek "data" sesuai brief frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil diarsipkan",
		"data":    brand,
	})
}

// RestoreBrandHandler: Memulihkan brand dari arsip (PATCH) - Mengembalikan data di "data"
func RestoreBrandHandler(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	brand, err := RestoreBrand(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memulihkan brand"})
		return
	}

	// Mengirim respons sukses lengkap dengan objek "data" sesuai brief frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Brand berhasil dipulihkan",
		"data":    brand,
	})
}