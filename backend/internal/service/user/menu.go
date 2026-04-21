package user

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/internal/rbac"
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"strings"

	menuconv "admin/internal/service/menu"

	"gorm.io/gorm"
)

// MenuService 用户菜单服务
type MenuService struct {
	menuRepo       *repository.MenuRepo
	permissionRepo *repository.PermissionRepo
	cache          *rbac.PermissionCache
}

// NewMenuService 创建用户菜单服务
func NewMenuService(db *gorm.DB, cache *rbac.PermissionCache) *MenuService {
	return &MenuService{
		menuRepo:       repository.NewMenuRepo(db),
		permissionRepo: repository.NewPermissionRepo(db),
		cache:          cache,
	}
}

// GetUserMenu 获取用户菜单树
func (s *MenuService) GetUserMenu(ctx context.Context, userName, tenantCode string) (*dto.UserMenuResponse, error) {
	// 超管特殊处理：返回所有菜单
	if xcontext.HasRole(ctx, constants.SuperAdmin) {
		return s.getAllMenus(ctx)
	}

	roleIDs := xcontext.GetRoleIDs(ctx)
	if len(roleIDs) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 从 PermissionCache 获取菜单ID列表
	menuIDs := s.cache.GetMenuIDs(roleIDs)
	if len(menuIDs) == 0 {
		return &dto.UserMenuResponse{List: []*dto.MenuTreeNode{}}, nil
	}

	// 查询菜单详情
	menus, err := s.menuRepo.GetByIDs(ctx, menuIDs)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单详情失败", err)
	}

	visibleMenus := s.filterVisibleMenus(menus)
	tree := s.buildMenuTree(visibleMenus)

	return &dto.UserMenuResponse{List: tree}, nil
}

// GetUserButtons 获取指定菜单的按钮权限
func (s *MenuService) GetUserButtons(ctx context.Context, menuID string) (*dto.UserButtonsResponse, error) {
	// 超管特殊处理：返回所有按钮
	if xcontext.HasRole(ctx, constants.SuperAdmin) {
		return s.getAllButtonsForMenu(ctx, menuID)
	}

	roleIDs := xcontext.GetRoleIDs(ctx)
	if len(roleIDs) == 0 {
		return &dto.UserButtonsResponse{Buttons: []*dto.ButtonInfo{}}, nil
	}

	// 从缓存获取用户角色拥有的按钮权限 ID
	userButtonPermIDs := s.cache.GetButtonPermissionIDs(roleIDs)
	if len(userButtonPermIDs) == 0 {
		return &dto.UserButtonsResponse{Buttons: []*dto.ButtonInfo{}}, nil
	}
	userPermSet := make(map[string]bool, len(userButtonPermIDs))
	for _, id := range userButtonPermIDs {
		userPermSet[id] = true
	}

	// 查询该菜单的所有按钮，只返回用户有权限的
	buttons, err := s.permissionRepo.ListByType(ctx, "BUTTON")
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询按钮失败", err)
	}

	var menuButtons []*dto.ButtonInfo
	for _, btn := range buttons {
		if strings.HasPrefix(btn.Resource, "btn:"+menuID+":") && userPermSet[btn.PermissionID] {
			menuButtons = append(menuButtons, &dto.ButtonInfo{
				PermissionID: btn.PermissionID,
				Name:         btn.Name,
				Action:       &btn.Action,
				Resource:     &btn.Resource,
			})
		}
	}

	return &dto.UserButtonsResponse{Buttons: menuButtons}, nil
}

// getAllMenus 获取所有菜单（超管使用）
func (s *MenuService) getAllMenus(ctx context.Context) (*dto.UserMenuResponse, error) {
	menus, err := s.menuRepo.List(ctx)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	visibleMenus := s.filterVisibleMenus(menus)
	tree := s.buildMenuTree(visibleMenus)

	return &dto.UserMenuResponse{List: tree}, nil
}

// getAllButtonsForMenu 获取菜单的所有按钮（超管使用）
func (s *MenuService) getAllButtonsForMenu(ctx context.Context, menuID string) (*dto.UserButtonsResponse, error) {
	buttons, err := s.permissionRepo.ListByType(ctx, "BUTTON")
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询按钮失败", err)
	}

	var buttonInfos []*dto.ButtonInfo
	for _, btn := range buttons {
		if strings.HasPrefix(btn.Resource, "btn:"+menuID+":") {
			buttonInfos = append(buttonInfos, &dto.ButtonInfo{
				PermissionID: btn.PermissionID,
				Name:         btn.Name,
				Action:       &btn.Action,
				Resource:     &btn.Resource,
			})
		}
	}

	return &dto.UserButtonsResponse{Buttons: buttonInfos}, nil
}

// filterVisibleMenus 过滤启用状态的菜单
func (s *MenuService) filterVisibleMenus(menus []*model.Menu) []*model.Menu {
	var result []*model.Menu
	for _, menu := range menus {
		if menu.Status == constants.StatusEnabled {
			result = append(result, menu)
		}
	}
	return result
}

// buildMenuTree 构建菜单树
func (s *MenuService) buildMenuTree(menus []*model.Menu) []*dto.MenuTreeNode {
	menuMap := make(map[string]*dto.MenuTreeNode)
	for _, menu := range menus {
		menuMap[menu.MenuID] = &dto.MenuTreeNode{
			MenuInfo: menuconv.ModelToMenuInfo(menu),
			Children: []*dto.MenuTreeNode{},
		}
	}

	var roots []*dto.MenuTreeNode
	for _, menu := range menus {
		node := menuMap[menu.MenuID]
		if menu.ParentID == "" || menu.ParentID == "0" {
			roots = append(roots, node)
		} else if parent, exists := menuMap[menu.ParentID]; exists {
			parent.Children = append(parent.Children, node)
		}
	}

	return roots
}
