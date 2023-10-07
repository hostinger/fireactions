package handler

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/server/httperr"
	"github.com/rs/zerolog"
)

// GitHubTokenGetter is an interface that returns a GitHub registration token for an organisation.
type GitHubTokenGetter interface {
	GetRegistrationToken(ctx context.Context, org string) (string, error)
	GetRemoveToken(ctx context.Context, org string) (string, error)
}

// GetGitHubRegistrationTokenHandlerFuncV1 returns a HTTP handler function that returns a GitHub registration token for
// an organisation.
func GetGitHubRegistrationTokenHandlerFuncV1(log *zerolog.Logger, tg GitHubTokenGetter) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		org := ctx.Param("organisation")

		token, err := tg.GetRegistrationToken(ctx, org)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, gin.H{"token": token})
	}

	return f
}

// GetGitHubRemoveTokenHandlerFuncV1 returns a HTTP handler function that returns a GitHub remove token for an
// organisation.
func GetGitHubRemoveTokenHandlerFuncV1(log *zerolog.Logger, tg GitHubTokenGetter) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		org := ctx.Param("organisation")

		token, err := tg.GetRemoveToken(ctx, org)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, gin.H{"token": token})
	}

	return f
}

// RegisterGitHubV1 registers all HTTP handlers for the GitHub v1 API.
func RegisterGitHubV1(r gin.IRouter, log *zerolog.Logger, tg GitHubTokenGetter) {
	r.POST("/github/:organisation/registration-token",
		GetGitHubRegistrationTokenHandlerFuncV1(log, tg))
	r.POST("/github/:organisation/remove-token",
		GetGitHubRemoveTokenHandlerFuncV1(log, tg))
}
