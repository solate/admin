package menu

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
)

// GetMenuTree 获取菜单树
func (s *Service) GetMenuTree(ctx context.Context) (*dto.MenuTreeResponse, error) {
	// 获取所有菜单
	menus, err := s.menuRepo.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("查询菜单列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单列表失败", err)
	}

	// 构建菜单树
	tree := s.buildMenuTree(menus)

	return &dto.MenuTreeResponse{
		List: tree,
	}, nil
}

// buildMenuTree 构建菜单树
func (s *Service) buildMenuTree(menus []*model.Menu) []*dto.MenuTreeNode {
	// 创建节点映射
	nodeMap := make(map[string]*dto.MenuTreeNode)
	for _, menu := range menus {
		nodeMap[menu.MenuID] = &dto.MenuTreeNode{
			MenuInfo: ModelToMenuInfo(menu),
			Children: []*dto.MenuTreeNode{},
		}
	}

	// 构建树结构
	var roots []*dto.MenuTreeNode
	for _, menu := range menus {
		node := nodeMap[menu.MenuID]
		if menu.ParentID == "" {
			// 顶级菜单
			roots = append(roots, node)
		} else if parent, exists := nodeMap[menu.ParentID]; exists {
			// 添加到父菜单的子节点
			parent.Children = append(parent.Children, node)
		}
	}

	return roots
}
