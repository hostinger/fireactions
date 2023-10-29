package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/server/scheduler"
	"github.com/hostinger/fireactions/server/store"
	"github.com/rs/zerolog"
)

type GitHubTokenGetter interface {
	GetRegistrationToken(ctx context.Context, organisation string) (string, error)
	GetRemoveToken(ctx context.Context, organisation string) (string, error)
}

// RegisterRunnersHandlers registers HTTP handlers for /api/v1/runners/* endpoints
// to the provided router.
func RegisterRunnersHandlers(
	logger *zerolog.Logger, router *gin.RouterGroup, scheduler *scheduler.Scheduler, store store.Store,
	tokenGetter GitHubTokenGetter,
) {
	runners := router.Group("/runners")
	{
		runners.GET("", RunnersHandlerFunc(logger, store))
		runners.GET("/:id", RunnerHandlerFunc(logger, store))
		runners.GET("/:id/registration-token", RunnerRegistrationTokenHandlerFunc(logger, store, tokenGetter))
		runners.GET("/:id/remove-token", RunnerRemoveTokenHandlerFunc(logger, store, tokenGetter))
		runners.PATCH("/:id/status", RunnerSetStatusHandlerFunc(logger, store))
		runners.DELETE("/:id", RunnerDeleteHandlerFunc(logger, store))
	}
}

// RunnersHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /api/v1/runners
func RunnersHandlerFunc(logger *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		type Query struct {
			Organisation *string `form:"organisation"`
		}

		var query Query
		err := ctx.ShouldBindQuery(&query)
		if err != nil {
			ctx.Error(err)
			return
		}

		runners, err := store.GetRunners(ctx, func(r *fireactions.Runner) bool {
			if r.DeletedAt != nil {
				return false
			}

			if query.Organisation != nil && *query.Organisation != r.Organisation {
				return false
			}

			return true
		})
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(200, gin.H{"runners": runners})
	}

	return f
}

// RunnerHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /api/v1/runners/:id
func RunnerHandlerFunc(logger *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		runner, err := store.GetRunner(ctx, ctx.Param("id"))
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(200, runner)
	}

	return f
}

// RunnerRegistrationTokenHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /api/v1/runners/:id/registration-token
func RunnerRegistrationTokenHandlerFunc(logger *zerolog.Logger, store store.Store, tokenGetter GitHubTokenGetter) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		runner, err := store.GetRunner(ctx, ctx.Param("id"))
		if err != nil {
			ctx.Error(err)
			return
		}

		token, err := tokenGetter.GetRegistrationToken(ctx, runner.Organisation)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(200, fireactions.RunnerRegistrationToken{Token: token})
	}

	return f
}

// RunnerRemoveTokenHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint GET /api/v1/runners/:id/remove-token
func RunnerRemoveTokenHandlerFunc(logger *zerolog.Logger, store store.Store, tokenGetter GitHubTokenGetter) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		runner, err := store.GetRunner(ctx, ctx.Param("id"))
		if err != nil {
			ctx.Error(err)
			return
		}

		token, err := tokenGetter.GetRemoveToken(ctx, runner.Organisation)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.JSON(200, fireactions.RunnerRemoveToken{Token: token})
	}

	return f
}

// RunnerSetStatusHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint PATCH /api/v1/runners/:id/status
func RunnerSetStatusHandlerFunc(logger *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		var request fireactions.RunnerSetStatusRequest
		err := ctx.ShouldBindJSON(&request)
		if err != nil {
			ctx.Error(err)
			return
		}

		_, err = store.SetRunnerStatus(ctx, ctx.Param("id"), fireactions.RunnerStatus(request))
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.Status(204)
	}

	return f
}

// RunnerDeleteHandlerFunc returns a HandlerFunc that handles HTTP requests to
// endpoint DELETE /api/v1/runners/:id
func RunnerDeleteHandlerFunc(logger *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		err := store.DeallocateRunner(ctx, ctx.Param("id"))
		if err != nil {
			ctx.Error(err)
			return
		}

		err = store.SoftDeleteRunner(ctx, ctx.Param("id"))
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.Status(204)
	}

	return f
}
