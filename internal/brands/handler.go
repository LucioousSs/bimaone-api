package brands

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// GetBrandsHandler untuk menangani permintaan melihat semua daftar brand
func GetBrandsHandler(c *gin.Context) {
    // Panggil fungsi dari repository.go
    brandsList, err := GetAllBrands()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Gagal mengambil data brands: " + err.Error(),
        })
        return
    }

    // Kirim hasil data dalam bentuk JSON
    c.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mengambil data brands",
        "data":    brandsList,
    })
}

// CreateBrandHandler untuk menangani permintaan menambah brand baru
func CreateBrandHandler(c *gin.Context) {
    var newBrand Brand

    // Baca data JSON yang dikirim dari luar
    if err := c.ShouldBindJSON(&newBrand); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Format data tidak valid: " + err.Error(),
        })
        return
    }

    // Simpan ke database lewat repository.go
    err := CreateBrand(&newBrand)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Gagal menyimpan brand: " + err.Error(),
        })
        return
    }

    // Kirim respon sukses
    c.JSON(http.StatusCreated, gin.H{
        "message": "Brand berhasil ditambahkan!",
        "data":    newBrand,
    })
}