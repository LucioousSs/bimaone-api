package main

import (
	"log"
	"net/http"

	"bimaone-api/internal/database"
	"bimaone-api/internal/warehouses"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("gagal menghubungkan database: %v", err)
	}
	defer database.CloseDB()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "BIMAONE API Backend is running!",
		})
	})

	warehouseRepo := warehouses.NewRepository(database.DB)
	warehouseHandler := warehouses.NewHandler(warehouseRepo)

	r.GET("/warehouses", warehouseHandler.List)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("gagal menjalankan server: %v", err)
	}
}
