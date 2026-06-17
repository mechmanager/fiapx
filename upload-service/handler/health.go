// Package handler contains HTTP handlers for the upload-service.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health handles GET /health and returns a simple liveness response.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
