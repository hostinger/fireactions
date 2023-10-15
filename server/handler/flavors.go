package handler

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/server/httperr"
	"github.com/hostinger/fireactions/server/models"
	"github.com/hostinger/fireactions/server/store"
	"github.com/rs/zerolog"
)

// RegisterFlavorsV1 registers all HTTP handlers for the Flavors v1 API.
func RegisterFlavorsV1(r gin.IRouter, log *zerolog.Logger, store store.Store) {
	r.GET("/flavors",
		GetFlavorsHandlerFuncV1(log, store))
	r.GET("/flavors/:name",
		GetFlavorHandlerFuncV1(log, store))
	r.PATCH("/flavors/:name/disable",
		DisableFlavorHandlerFuncV1(log, store))
	r.PATCH("/flavors/:name/enable",
		EnableFlavorHandlerFuncV1(log, store))
	r.DELETE("/flavors/:name",
		DeleteFlavorHandlerFuncV1(log, store))
}

// GetFlavorsHandlerFuncV1 returns a HTTP handler function that returns all Flavors. The Flavors are returned in the v1
// format.
func GetFlavorsHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		flavors, err := store.ListFlavors(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, gin.H{"flavors": convertFlavorsToFlavorsV1(flavors...)})
	}

	return f
}

// GetFlavorHandlerFuncV1 returns a HTTP handler function that returns a single Flavor by name. The Flavor is returned in
// the v1 format.
func GetFlavorHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		flavor, err := store.GetFlavor(ctx, ctx.Param("name"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, convertFlavorToFlavorV1(flavor))
	}

	return f
}

// DisableFlavorHandlerFuncV1 returns a HTTP handler function that disables a Flavor by name.
func DisableFlavorHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		flavor, err := store.GetFlavor(ctx, ctx.Param("name"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		flavor.Disable()
		err = store.SaveFlavor(ctx, flavor)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.Status(204)
	}

	return f
}

// EnableFlavorHandlerFuncV1 returns a HTTP handler function that enables a Flavor by name.
func EnableFlavorHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		flavor, err := store.GetFlavor(ctx, ctx.Param("name"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		flavor.Enable()
		err = store.SaveFlavor(ctx, flavor)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.Status(204)
	}

	return f
}

// DeleteFlavorHandlerFuncV1 returns a HTTP handler function that deletes a Flavor by name.
func DeleteFlavorHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		flavor, err := store.GetFlavor(ctx, ctx.Param("name"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		err = store.DeleteFlavor(ctx, flavor.Name)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.Status(204)
	}

	return f
}

func convertFlavorToFlavorV1(flavor *models.Flavor) v1.Flavor {
	f := v1.Flavor{
		Name:         flavor.Name,
		DiskSizeGB:   flavor.DiskSizeGB,
		MemorySizeMB: flavor.MemorySizeMB,
		VCPUs:        flavor.VCPUs,
		Image:        flavor.Image,
		Enabled:      flavor.Enabled,
	}

	return f
}

func convertFlavorsToFlavorsV1(flavors ...*models.Flavor) []v1.Flavor {
	flavorsV1 := make([]v1.Flavor, 0, len(flavors))
	for _, f := range flavors {
		flavorsV1 = append(flavorsV1, convertFlavorToFlavorV1(f))
	}

	return flavorsV1
}
