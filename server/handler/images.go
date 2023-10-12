package handler

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/hostinger/fireactions/api"
	"github.com/hostinger/fireactions/server/httperr"
	"github.com/hostinger/fireactions/server/store"
	"github.com/hostinger/fireactions/server/structs"
	"github.com/rs/zerolog"
)

// RegisterImagesV1 registers all HTTP handlers for the Images v1 API.
func RegisterImagesV1(r gin.IRouter, log *zerolog.Logger, store store.Store) {
	r.GET("/images", GetImagesHandlerFuncV1(log, store))
	r.GET("/images/:id", GetImageHandlerFuncV1(log, store))
	r.DELETE("/images/:id", DeleteImageHandlerFuncV1(log, store))
}

// GetImagesHandlerFuncV1 returns a HTTP handler function that returns all Images. The Images are returned in the v1
// format.
func GetImagesHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		images, err := store.ListImages(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, gin.H{"images": convertImagesToImagesV1(images...)})
	}

	return f
}

// GetImageHandlerFuncV1 returns a HTTP handler function that returns a single Image by id. The Image is returned in
// the v1 format.
func GetImageHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		image, err := store.GetImage(ctx, ctx.Param("id"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, convertImageToImageV1(image))
	}

	return f
}

// DeleteImageHandlerFuncV1 returns a HTTP handler function that deletes a single Image by ID.
func DeleteImageHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		image, err := store.GetImageByID(ctx, ctx.Param("id"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		err = store.DeleteImage(ctx, image.ID)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.Status(204)
	}

	return f
}

func convertImageToImageV1(image *structs.Image) *v1.Image {
	i := &v1.Image{
		ID:   image.ID,
		Name: image.Name,
		URL:  image.URL,
	}

	return i
}

func convertImagesToImagesV1(images ...*structs.Image) []*v1.Image {
	var is []*v1.Image
	for _, image := range images {
		is = append(is, convertImageToImageV1(image))
	}

	return is
}
