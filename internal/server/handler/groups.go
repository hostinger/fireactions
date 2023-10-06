package handler

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/server/store"
	"github.com/hostinger/fireactions/internal/server/structs"
	"github.com/rs/zerolog"
)

// RegisterGroupsV1 registers all HTTP handlers for the Groups v1 API.
func RegisterGroupsV1(r gin.IRouter, log *zerolog.Logger, store store.Store) {
	r.GET("/groups",
		GetGroupsHandlerFuncV1(log, store))
	r.GET("/groups/:name",
		GetGroupHandlerFuncV1(log, store))
	r.PATCH("/groups/:name/disable",
		DisableGroupHandlerFuncV1(log, store))
	r.PATCH("/groups/:name/enable",
		EnableGroupHandlerFuncV1(log, store))
}

// GetGroupsHandlerFuncV1 returns a HTTP handler function that returns all Groups. The Groups are returned in the v1
// format.
func GetGroupsHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		groups, err := store.ListGroups(ctx)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, gin.H{"groups": convertGroupsToGroupsV1(groups...)})
	}

	return f
}

// GetGroupHandlerFuncV1 returns a HTTP handler function that returns a single Group by name. The Group is returned in
// the v1 format.
func GetGroupHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		name := ctx.Param("name")

		group, err := store.GetGroup(ctx, name)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.JSON(200, convertGroupToGroupV1(group))
	}

	return f
}

// DisableGroupHandlerFuncV1 returns a HTTP handler function that disables a Group by name.
func DisableGroupHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		group, err := store.GetGroup(ctx, ctx.Param("name"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		group.Disable()
		err = store.SaveGroup(ctx, group)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.Status(204)
	}

	return f
}

// EnableGroupHandlerFuncV1 returns a HTTP handler function that enables a Group by name.
func EnableGroupHandlerFuncV1(log *zerolog.Logger, store store.Store) gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		group, err := store.GetGroup(ctx, ctx.Param("name"))
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		group.Enable()
		err = store.SaveGroup(ctx, group)
		if err != nil {
			httperr.E(ctx, err)
			return
		}

		ctx.Status(204)
	}

	return f
}

func convertGroupToGroupV1(group *structs.Group) *v1.Group {
	g := &v1.Group{
		Name:    group.Name,
		Enabled: group.Enabled,
	}

	return g
}

func convertGroupsToGroupsV1(groups ...*structs.Group) []*v1.Group {
	groupsV1 := make([]*v1.Group, 0, len(groups))
	for _, g := range groups {
		groupsV1 = append(groupsV1, convertGroupToGroupV1(g))
	}

	return groupsV1
}
