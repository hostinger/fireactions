package server

import (
	"context"
	"net/http"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v63/github"
	"github.com/hostinger/fireactions"
	"github.com/samber/lo"
)

func getHealthzHandler() gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "OK"})
	}

	return f
}

func getVersionHandler() gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"version": fireactions.String()})
	}

	return f
}

func listPoolsHandler(p PoolManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		pools, err := p.ListPools(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"pools": convertPools(pools)})
	}

	return f
}

func getPoolHandler(p PoolManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")
		pool, err := p.GetPool(ctx, id)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"pool": convertPool(pool)})
	}

	return f
}

func scalePoolHandler(p PoolManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")
		if err := p.ScalePool(ctx, id, 1); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Pool scaled successfully"})
	}

	return f
}

func pausePoolHandler(p PoolManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")
		if err := p.PausePool(ctx, id); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Pool paused successfully"})
	}

	return f
}

func resumePoolHandler(p PoolManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		id := ctx.Param("id")
		if err := p.ResumePool(ctx, id); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Pool resumed successfully"})
	}

	return f
}

func restartHandler(p PoolManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		if err := p.Restart(ctx); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Pools restarted successfully"})
	}

	return f
}

func webhookGitHubHandler(p PoolManager, secret string) gin.HandlerFunc {
	ge := githubevents.New(secret)
	ge.OnWorkflowJobEventQueued(func(deliveryID, eventName string, event *github.WorkflowJobEvent) error {
		pools, err := p.ListPools(context.Background())
		if err != nil {
			return err
		}

		for _, pool := range pools {
			if !lo.Every(pool.config.Runner.Labels, event.WorkflowJob.Labels) {
				continue
			}

			return pool.Scale(context.Background(), 1)
		}

		return nil
	})

	f := func(ctx *gin.Context) {
		err := ge.HandleEventRequest(ctx.Request)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.Status(http.StatusOK)
	}

	return f
}
