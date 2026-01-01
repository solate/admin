package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/auditlog"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xerr"
	"context"
	"time"

	"gorm.io/gorm"
)

// MenuService 菜单服务
type MenuService struct {
	menuRepo *repository.MenuRepo
}

// NewMenuService 创建菜单服务
func NewMenuService(menuRepo *repository.MenuRepo) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
	}
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(ctx context.Context, req *dto.CreateMenuRequest) (*dto.MenuInfo, error) {
	// 生成菜单ID
	menuID, err := idgen.GenerateUUID()
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成菜单ID失败", err)
	}

	// 如果有父菜单，检查父菜单是否存在
	if req.ParentID != "" {
		_, err := s.menuRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, xerr.ErrMenuInvalidParent
			}
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父菜单失败", err)
		}
	}

	// 构建菜单模型
	status := int16(req.Status)
	if req.Status == 0 {
		status = constants.MenuStatusShow
	}

	menu := &model.Menu{
		MenuID:      menuID,
		ParentID:    req.ParentID,
		Name:        req.Name,
		Path:        req.Path,
		Component:   req.Component,
		Redirect:    req.Redirect,
		Icon:        req.Icon,
		Sort:        int32(*req.Sort),
		Status:      status,
		Description: "",
	}

	// 创建菜单
	if err := s.menuRepo.Create(ctx, menu); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建菜单失败", err)
	}

	// 记录操作日志
	ctx = auditlog.RecordCreate(ctx, constants.ModuleMenu, constants.ResourceTypeMenu, menu.MenuID, menu.Name, menu)

	return s.toMenuInfo(menu), nil
}

// GetMenuByID 获取菜单详情
func (s *MenuService) GetMenuByID(ctx context.Context, menuID string) (*dto.MenuInfo, error) {
	menu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrMenuNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	return s.toMenuInfo(menu), nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(ctx context.Context, menuID string, req *dto.UpdateMenuRequest) (*dto.MenuInfo, error) {
	// 检查菜单是否存在，获取旧值用于日志
	oldMenu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrMenuNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 如果更新父菜单，需要检查
	if req.ParentID != "" {
		// 不能将菜单移动到自己或其子菜单下
		if req.ParentID == menuID {
			return nil, xerr.ErrMenuCannotMoveToSelf
		}

		// 检查父菜单是否存在
		_, err := s.menuRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, xerr.ErrMenuInvalidParent
			}
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父菜单失败", err)
		}
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ParentID != "" {
		updates["parent_id"] = req.ParentID
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Component != "" {
		updates["component"] = req.Component
	}
	if req.Redirect != "" {
		updates["redirect"] = req.Redirect
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Sort != nil {
		updates["sort"] = int32(*req.Sort)
	}
	if req.Status != 0 {
		updates["status"] = int16(req.Status)
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新菜单
	if err := s.menuRepo.Update(ctx, menuID, updates); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新菜单失败", err)
	}

	// 获取更新后的菜单信息
	updatedMenu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后菜单信息失败", err)
	}

	// 记录操作日志
	ctx = auditlog.RecordUpdate(ctx, constants.ModuleMenu, constants.ResourceTypeMenu, updatedMenu.MenuID, updatedMenu.Name, oldMenu, updatedMenu)

	return s.toMenuInfo(updatedMenu), nil
}

// DeleteMenu 删除菜单
func (s *MenuService) DeleteMenu(ctx context.Context, menuID string) error {
	// 检查菜单是否存在，获取菜单信息用于日志
	menu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrMenuNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 检查是否有子菜单
	childrenCount, err := s.menuRepo.GetChildrenCount(ctx, menuID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "检查子菜单失败", err)
	}
	if childrenCount > 0 {
		return xerr.ErrMenuHasChildren
	}

	// 删除菜单
	if err := s.menuRepo.Delete(ctx, menuID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除菜单失败", err)
	}

	// 记录操作日志
	auditlog.RecordDelete(ctx, constants.ModuleMenu, constants.ResourceTypeMenu, menu.MenuID, menu.Name, menu)

	return nil
}

// ListMenus 获取菜单列表
func (s *MenuService) ListMenus(ctx context.Context, req *dto.ListMenusRequest) (*dto.ListMenusResponse, error) {
	// 获取菜单列表和总数，支持筛选条件
	var statusFilter *int16
	if req.Status != nil {
		statusFilter = req.Status
	}
	menus, total, err := s.menuRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Name, statusFilter)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单列表失败", err)
	}

	// 转换为响应格式
	menuInfos := make([]*dto.MenuInfo, len(menus))
	for i, menu := range menus {
		menuInfos[i] = s.toMenuInfo(menu)
	}

	return &dto.ListMenusResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     menuInfos,
	}, nil
}

// GetAllMenus 获取所有菜单（平铺）
func (s *MenuService) GetAllMenus(ctx context.Context) (*dto.AllMenusResponse, error) {
	// 获取所有菜单
	menus, err := s.menuRepo.List(ctx)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单列表失败", err)
	}

	// 转换为响应格式
	menuInfos := make([]*dto.MenuInfo, len(menus))
	for i, menu := range menus {
		menuInfos[i] = s.toMenuInfo(menu)
	}

	return &dto.AllMenusResponse{
		List: menuInfos,
	}, nil
}

// GetMenuTree 获取菜单树
func (s *MenuService) GetMenuTree(ctx context.Context) (*dto.MenuTreeResponse, error) {
	// 获取所有菜单
	menus, err := s.menuRepo.List(ctx)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单列表失败", err)
	}

	// 构建菜单树
	tree := s.buildMenuTree(menus)

	return &dto.MenuTreeResponse{
		List: tree,
	}, nil
}

// UpdateMenuStatus 更新菜单状态
func (s *MenuService) UpdateMenuStatus(ctx context.Context, menuID string, status int) error {
	// 检查菜单是否存在，获取旧值用于日志
	oldMenu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrMenuNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 更新菜单状态
	if err := s.menuRepo.UpdateStatus(ctx, menuID, int16(status)); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新菜单状态失败", err)
	}

	// 获取更新后的菜单信息
	updatedMenu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后菜单信息失败", err)
	}

	// 记录操作日志
	auditlog.RecordUpdate(ctx, constants.ModuleMenu, constants.ResourceTypeMenu, updatedMenu.MenuID, updatedMenu.Name, oldMenu, updatedMenu)

	return nil
}

// buildMenuTree 构建菜单树
func (s *MenuService) buildMenuTree(menus []*model.Menu) []*dto.MenuTreeNode {
	// 创建节点映射
	nodeMap := make(map[string]*dto.MenuTreeNode)
	for _, menu := range menus {
		nodeMap[menu.MenuID] = &dto.MenuTreeNode{
			MenuInfo: s.toMenuInfo(menu),
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

// toMenuInfo 转换为菜单信息格式
func (s *MenuService) toMenuInfo(menu *model.Menu) *dto.MenuInfo {
	sort := int16(menu.Sort)
	return &dto.MenuInfo{
		PermissionID: menu.MenuID,
		Name:         menu.Name,
		Type:         constants.TypeMenu,
		ParentID:     &menu.ParentID,
		Path:         &menu.Path,
		Component:    &menu.Component,
		Redirect:     &menu.Redirect,
		Icon:         &menu.Icon,
		Sort:         &sort,
		Status:       menu.Status,
		Description:  &menu.Description,
		CreatedAt:    menu.CreatedAt,
		UpdatedAt:    menu.UpdatedAt,
	}
}
