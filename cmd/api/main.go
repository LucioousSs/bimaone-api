package main

import (
	"log"
	"net/http"

	"bimaone-api/internal/brands"
	"bimaone-api/internal/database"
	"bimaone-api/internal/warehouses"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: File .env tidak ditemukan")
	}

	err := database.ConnectDB()
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	defer database.CloseDB()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Izinkan semua origin untuk development
    c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

    if c.Request.Method == http.MethodOptions {
        c.AbortWithStatus(http.StatusNoContent)
        return
    }
    c.Next()
})

	// Endpoint health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "BIMAONE API Backend is running!",
		})
	})

	// --- INISIALISASI REPOSITORY & HANDLER WAREHOUSES ---
	warehouseRepo := warehouses.NewRepository(database.DB)
	warehouseHandler := warehouses.NewHandler(warehouseRepo)
	// ---------------------------------------------------

	// --- GROUP ROUTE /api/v1 SESUAI PERMINTAAN FRONTEND ---
	v1 := r.Group("/api/v1")
	{
		// Brands routes
		v1.GET("/brands", brands.GetBrandsHandler)
		v1.POST("/brands", brands.CreateBrandHandler)

		v1.GET("/brands/:id", brands.GetBrandByIDHandler)

        v1.PUT("/brands/:id", brands.UpdateBrandHandler)       // <-- Tambahan untuk Edit
        v1.PATCH("/brands/:id/archive", brands.ArchiveBrandHandler)
        v1.PATCH("/brands/:id/restore", brands.RestoreBrandHandler)
		

		// Warehouses routes (disesuaikan ke group /api/v1 agar seragam)
		v1.GET("/warehouses", warehouseHandler.List)
	}
	// -----------------------------------------------------

	log.Println("Server berjalan di port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("gagal menjalankan server: %v", err)
	}
}