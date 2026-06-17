package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health trata GET /health e indica que o serviço está no ar.
func Health(c *gin.Context) {
	// Resposta simples usada por healthchecks e monitoramento.
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
