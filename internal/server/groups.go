package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	api "github.com/hostinger/fireactions/apiv1"
	"github.com/hostinger/fireactions/internal/structs"
)

func (s *Server) GetGroups() []*structs.Group {
	return s.cfg.Groups
}

func (s *Server) GetGroupByName(name string) (*structs.Group, error) {
	for _, g := range s.cfg.Groups {
		if g.Name != name {
			continue
		}

		return g, nil
	}

	return nil, fmt.Errorf("group not found: %s", name)
}

func (s *Server) handleGetGroups(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"groups": convertGroupsToGroupsV1(s.cfg.Groups...)})
}

func (s *Server) handleGetGroup(ctx *gin.Context) {
	name := ctx.Param("name")

	group, err := s.GetGroupByName(name)
	if err != nil {
		ctx.JSON(404, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"group": convertGroupToGroupV1(group)})
}

func convertGroupToGroupV1(group *structs.Group) *api.Group {
	g := &api.Group{
		Name: group.Name,
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
