package server

import (
	"net/http"
	"sort"

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

		// Sort pools by name
		sort.Slice(pools, func(i, j int) bool {
			return pools[i].config.Name < pools[j].config.Name
		})

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

		var req struct {
			Replicas *int `json:"replicas" binding:"required,min=0"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := p.ScalePool(ctx, id, *req.Replicas); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Pool replicas updated successfully"})
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
		// Pool ID is optional - if not provided, list all microvms
		poolID := ctx.Param("id")

		microVMs, err := m.ListMicroVMs(ctx, poolID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if microVMs == nil {
			microVMs = []*MicroVM{}
		}

		// Sort microVMs by Pool name first, then by VMID
		sort.Slice(microVMs, func(i, j int) bool {
			if microVMs[i].Pool != microVMs[j].Pool {
				return microVMs[i].Pool < microVMs[j].Pool
			}
			return microVMs[i].VMID < microVMs[j].VMID
		})

		ctx.JSON(http.StatusOK, gin.H{"micro_vms": convertMicroVMs(microVMs)})
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
