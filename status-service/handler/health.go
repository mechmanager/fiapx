package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health responde ao healthcheck do serviço.
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "status-service"})
}
