package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHealthzHandlerFunc returns a handler function that returns a 200 OK
func GetHealthzHandlerFunc() gin.HandlerFunc {
	f := func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	}

	return f
}
