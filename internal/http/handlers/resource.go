package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Resource responds for the protected API endpoint.
func Resource(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "protected resource accessed",
	})
}
