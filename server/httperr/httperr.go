package httperr

import (
	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/server/store"
)

func E(ctx *gin.Context, err error) {
	var status int
	switch err.(type) {
	case nil:
		return
	case store.ErrNotFound:
		status = 404
	default:
		status = 500
	}

	ctx.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
}
