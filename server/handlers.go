package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions"
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

func reloadHandler(p PoolManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		if err := p.Reload(ctx); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Pools reloaded successfully"})
	}

	return f
}

func listMicroVMsHandler(m MicroVMManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		poolID := ctx.Param("id")

		microVMs, err := m.ListMicroVMs(ctx, poolID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if microVMs == nil {
			microVMs = []*MicroVM{}
		}

		ctx.JSON(http.StatusOK, gin.H{"micro_vms": microVMs})
	}

	return f
}

func getMicroVMHandler(m MicroVMManager) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		vmID := ctx.Param("id")
		vm, err := m.GetMicroVM(ctx, vmID)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"micro_vm": convertMicroVM(vm)})
	}

	return f
}
