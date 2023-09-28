package server

import (
	"github.com/gin-gonic/gin"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/structs"
)

func (s *Server) handleGetFlavors(ctx *gin.Context) {
	flavors, err := s.fm.ListFlavors()
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"flavors": convertFlavorsToFlavorsV1(flavors...)})
}

func (s *Server) handleGetFlavor(ctx *gin.Context) {
	flavor, err := s.fm.GetFlavor(ctx.Param("name"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, convertFlavorToFlavorV1(flavor))
}

func convertFlavorToFlavorV1(flavor *structs.Flavor) api.Flavor {
	f := api.Flavor{
		Name:         flavor.Name,
		DiskSizeGB:   flavor.DiskSizeGB,
		MemorySizeMB: flavor.MemorySizeMB,
		VCPUs:        flavor.VCPUs,
		ImageName:    flavor.ImageName,
	}

	return f
}

func convertFlavorsToFlavorsV1(flavors ...*structs.Flavor) []api.Flavor {
	flavorsV1 := make([]api.Flavor, 0, len(flavors))
	for _, f := range flavors {
		flavorsV1 = append(flavorsV1, convertFlavorToFlavorV1(f))
	}

	return flavorsV1
}
