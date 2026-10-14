package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // Endpoint health check untuk tes jaringan sesuai brief
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
            "message": "BIMAONE API Backend is running!",
        })
    })

    r.Run(":8080")
}