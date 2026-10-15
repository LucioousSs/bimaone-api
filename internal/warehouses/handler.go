package warehouses

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(c *gin.Context) {
	page, err := parseQueryInt(c, "page", 1)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameter page harus berupa bilangan bulat positif",
		})
		return
	}

	perPage, err := parseQueryInt(c, "per_page", 10)
	if err != nil || perPage < 1 || perPage > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "parameter per_page harus antara 1 dan 100",
		})
		return
	}

	result, err := h.repo.List(c.Request.Context(), ListParams{
		Search:  c.Query("search"),
		Status:  c.DefaultQuery("status", "active"),
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		if err.Error() == "status tidak valid" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal mengambil daftar gudang",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseQueryInt(c *gin.Context, key string, defaultValue int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(value)
}
