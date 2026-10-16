package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"bimaone-api/internal/brands"
	"bimaone-api/internal/database"
	"bimaone-api/internal/warehouses"

	"github.com/joho/godotenv"
)

func main() {
	// Memuat konfigurasi lokal jika file .env tersedia.
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan; menggunakan environment variables yang tersedia.")
	}

	// Tanpa koneksi database tidak ada gunanya melanjutkan, jadi langsung berhenti.
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	r := gin.Default()

	// Origin dikonfigurasi melalui environment variable agar frontend
	// dapat diizinkan tanpa membuka akses ke semua origin.
	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3000"
	}

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set(
			"Access-Control-Allow-Headers",
			"Content-Type, Authorization, Accept, Origin, X-Requested-With",
		)
		c.Writer.Header().Set(
			"Access-Control-Allow-Methods",
			"GET, POST, PUT, PATCH, DELETE, OPTIONS",
		)

		// Preflight harus dijawab sebelum request diteruskan ke handler.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// Health check untuk memastikan server merespons.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "API is running",
		})
	})

	brandRepo := brands.NewRepository(database.DB)
	brandHandler := brands.NewHandler(brandRepo)

	warehouseRepo := warehouses.NewRepository(database.DB)
	warehouseHandler := warehouses.NewHandler(warehouseRepo)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/brands", brandHandler.List)
		v1.POST("/brands", brandHandler.Create)
		v1.GET("/brands/:id", brandHandler.GetByID)
		v1.PUT("/brands/:id", brandHandler.Update)
		v1.PATCH("/brands/:id/archive", brandHandler.Archive)
		v1.PATCH("/brands/:id/restore", brandHandler.Restore)

		v1.GET("/warehouses", warehouseHandler.List)
		v1.POST("/warehouses", warehouseHandler.Create)
		v1.GET("/warehouses/:id", warehouseHandler.GetByID)
		v1.PUT("/warehouses/:id", warehouseHandler.Update)
		v1.PATCH("/warehouses/:id/archive", warehouseHandler.Archive)
		v1.PATCH("/warehouses/:id/restore", warehouseHandler.Restore)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server berjalan di port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
