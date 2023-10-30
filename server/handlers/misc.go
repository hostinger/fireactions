package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/build"
)

// RegisterMiscHandlers registers HTTP handlers for /healthz, /readyz, /livez
// and /version endpoints to the provided router.
func RegisterMiscHandlers(r gin.IRouter) {
	r.GET("/healthz", HealthzHandlerFunc())
	r.GET("/readyz", ReadyzHandlerFunc())
	r.GET("/version", VersionHandlerFunc())
	r.GET("/livez", LivezHandlerFunc())
}

// HealthzHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /healthz
func HealthzHandlerFunc() gin.HandlerFunc {
	f := func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	}

	return f
}

// ReadyHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /readyz
func ReadyzHandlerFunc() gin.HandlerFunc {
	f := func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	}

	return f
}

// LivezHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /livez
func LivezHandlerFunc() gin.HandlerFunc {
	f := func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	}

	return f
}

// VersionHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /version
func VersionHandlerFunc() gin.HandlerFunc {
	f := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": build.GitTag, "commit": build.GitCommit, "date": build.BuildDate})
	}

	return f
}
