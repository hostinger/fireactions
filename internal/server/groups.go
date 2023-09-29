package server

import (
	"github.com/gin-gonic/gin"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/server/httperr"
	"github.com/hostinger/fireactions/internal/structs"
)

func (s *Server) handleGetGroups(ctx *gin.Context) {
	groups, err := s.gm.ListGroups()
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, gin.H{"groups": convertGroupsToGroupsV1(groups...)})
}

func (s *Server) handleGetGroup(ctx *gin.Context) {
	name := ctx.Param("name")

	group, err := s.gm.GetGroup(name)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.JSON(200, convertGroupToGroupV1(group))
}

func (s *Server) handleDisableGroup(ctx *gin.Context) {
	group, err := s.gm.GetGroup(ctx.Param("name"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	err = s.gm.DisableGroup(group.Name)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.Status(204)
}

func (s *Server) handleEnableGroup(ctx *gin.Context) {
	group, err := s.gm.GetGroup(ctx.Param("name"))
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	err = s.gm.EnableGroup(group.Name)
	if err != nil {
		httperr.E(ctx, err)
		return
	}

	ctx.Status(204)
}

func convertGroupToGroupV1(group *structs.Group) *api.Group {
	g := &api.Group{
		Name:    group.Name,
		Enabled: group.Enabled,
	}

	return g
}

func convertGroupsToGroupsV1(groups ...*structs.Group) []*api.Group {
	groupsV1 := make([]*api.Group, 0, len(groups))
	for _, g := range groups {
		groupsV1 = append(groupsV1, convertGroupToGroupV1(g))
	}

	return groupsV1
}
