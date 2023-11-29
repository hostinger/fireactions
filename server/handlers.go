package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/version"
)

func (s *Server) handleGetHealthz(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

func (s *Server) handleGetVersion(ctx *gin.Context) {
	ctx.String(http.StatusOK, version.String())
}
