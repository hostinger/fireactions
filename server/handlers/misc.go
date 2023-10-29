package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthzHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /healthz
func HealthzHandlerFunc() gin.HandlerFunc {
	f := func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	}

	return f
}
