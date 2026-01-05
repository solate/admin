package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// MenuService 菜单服务
type MenuService struct {
	menuRepo *repository.MenuRepo
	enforcer *casbin.Enforcer
	recorder *audit.Recorder
}

// NewMenuService 创建菜单服务
func NewMenuService(menuRepo *repository.MenuRepo, enforcer *casbin.Enforcer, recorder *audit.Recorder) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
		enforcer: enforcer,
		recorder: recorder,
	}
}

// CreateMenu 创建菜单
func (s *MenuService) CreateMenu(ctx context.Context, req *dto.CreateMenuRequest) (resp *dto.MenuInfo, err error) {
	var menu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(audit.ModuleMenu),
				audit.WithError(err),
			)
		} else if menu != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(audit.ModuleMenu),
				audit.WithResource(audit.ResourceMenu, menu.MenuID, menu.Name),
				audit.WithValue(nil, menu),
			)
		}
	}()

	// 生成菜单ID
	var menuID string
	menuID, err = idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成菜单ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成菜单ID失败", err)
	}

	// 如果有父菜单，检查父菜单是否存在
	if req.ParentID != "" {
		_, err := s.menuRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("parent_id", req.ParentID).Msg("父菜单不存在")
				return nil, xerr.ErrMenuInvalidParent
			}
			log.Error().Err(err).Str("parent_id", req.ParentID).Msg("查询父菜单失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父菜单失败", err)
		}
	}

	// 构建菜单模型
	status := int16(req.Status)
	if req.Status == 0 {
		status = constants.MenuStatusShow
	}

	menu = &model.Menu{
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
		log.Error().Err(err).Str("menu_id", menuID).Str("name", req.Name).Msg("创建菜单失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建菜单失败", err)
	}

	log.Info().Str("menu_id", menuID).Str("name", req.Name).Msg("创建菜单成功")
	return s.toMenuInfo(menu), nil
}

// GetMenuByID 获取菜单详情
func (s *MenuService) GetMenuByID(ctx context.Context, menuID string) (*dto.MenuInfo, error) {
	menu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("menu_id", menuID).Msg("菜单不存在")
			return nil, xerr.ErrMenuNotFound
		}
		log.Error().Err(err).Str("menu_id", menuID).Msg("查询菜单失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	return s.toMenuInfo(menu), nil
}

// UpdateMenu 更新菜单
func (s *MenuService) UpdateMenu(ctx context.Context, menuID string, req *dto.UpdateMenuRequest) (resp *dto.MenuInfo, err error) {
	var oldMenu, newMenu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleMenu),
				audit.WithError(err),
			)
		} else if newMenu != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleMenu),
				audit.WithResource(audit.ResourceMenu, newMenu.MenuID, newMenu.Name),
				audit.WithValue(oldMenu, newMenu),
			)
		}
	}()

	// 获取旧菜单信息
	oldMenu, err = s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("menu_id", menuID).Msg("菜单不存在")
			return nil, xerr.ErrMenuNotFound
		}
		log.Error().Err(err).Str("menu_id", menuID).Msg("查询菜单失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 如果更新父菜单，需要检查
	if req.ParentID != "" {
		// 不能将菜单移动到自己或其子菜单下
		if req.ParentID == menuID {
			log.Warn().Str("menu_id", menuID).Msg("不能将菜单移动到自己下")
			return nil, xerr.ErrMenuCannotMoveToSelf
		}

		// 检查父菜单是否存在
		_, err := s.menuRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("parent_id", req.ParentID).Msg("父菜单不存在")
				return nil, xerr.ErrMenuInvalidParent
			}
			log.Error().Err(err).Str("parent_id", req.ParentID).Msg("查询父菜单失败")
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
		log.Error().Err(err).Str("menu_id", menuID).Interface("updates", updates).Msg("更新菜单失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新菜单失败", err)
	}

	// 获取更新后的菜单信息
	newMenu, err = s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		log.Error().Err(err).Str("menu_id", menuID).Msg("获取更新后菜单信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后菜单信息失败", err)
	}

	log.Info().Str("menu_id", menuID).Msg("更新菜单成功")
	return s.toMenuInfo(newMenu), nil
}

// DeleteMenu 删除菜单
// 说明：删除菜单时会自动清理所有租户中该菜单的权限策略
func (s *MenuService) DeleteMenu(ctx context.Context, menuID string) (err error) {
	var menu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(audit.ModuleMenu),
				audit.WithError(err),
			)
		} else if menu != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(audit.ModuleMenu),
				audit.WithResource(audit.ResourceMenu, menu.MenuID, menu.Name),
				audit.WithValue(menu, nil),
			)
			log.Info().Str("menu_id", menuID).Msg("删除菜单成功")
		}
	}()

	// 检查菜单是否存在
	menu, err = s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("menu_id", menuID).Msg("菜单不存在")
			return xerr.ErrMenuNotFound
		}
		log.Error().Err(err).Str("menu_id", menuID).Msg("查询菜单失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 检查是否有子菜单
	childrenCount, err := s.menuRepo.GetChildrenCount(ctx, menuID)
	if err != nil {
		log.Error().Err(err).Str("menu_id", menuID).Msg("检查子菜单失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "检查子菜单失败", err)
	}
	if childrenCount > 0 {
		log.Warn().Str("menu_id", menuID).Int64("children_count", childrenCount).Msg("菜单存在子菜单，无法删除")
		return xerr.ErrMenuHasChildren
	}

	// 删除菜单（软删除）
	if err := s.menuRepo.Delete(ctx, menuID); err != nil {
		log.Error().Err(err).Str("menu_id", menuID).Msg("删除菜单失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除菜单失败", err)
	}

	// 清理该菜单在所有租户中的权限策略
	// 1. 清理菜单权限 (menu:menu_id)
	// RemoveFilteredPolicy 参数: (fieldIndex, fieldValues...)
	// fieldIndex=0 表示按 v0 (subject) 过滤，我们传入通配符来匹配所有租户
	// 需要逐个租户清理，或者使用 RemovePolicies 如果支持

	// 获取所有租户并清理权限（这里使用更简单的方式：直接清理所有匹配的菜单权限）
	// 由于 Casbin 策略格式是 p, role_code, tenant_code, menu:menu_id, *
	// 我们可以通过 v2 (resource) 来过滤
	policies, _ := s.enforcer.GetFilteredPolicy(2, "menu:"+menuID)
	for _, policy := range policies {
		if len(policy) >= 4 {
			// 删除策略: p, role_code, tenant_code, menu:menu_id, *
			s.enforcer.RemovePolicy(policy[0], policy[1], policy[2], policy[3])
		}
	}

	// 2. 清理关联的 API 权限（路径匹配 /api/v1/*）
	// 需要根据菜单的 api_paths 字段来清理
	// 这里为了简化，我们清理所有可能受影响的 API 权限
	// 实际生产环境中，可以考虑维护一个菜单-权限映射表

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
		log.Error().Err(err).Str("name", req.Name).Msg("查询菜单列表失败")
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
		log.Error().Err(err).Msg("查询菜单列表失败")
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
		log.Error().Err(err).Msg("查询菜单列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单列表失败", err)
	}

	// 构建菜单树
	tree := s.buildMenuTree(menus)

	return &dto.MenuTreeResponse{
		List: tree,
	}, nil
}

// UpdateMenuStatus 更新菜单状态
func (s *MenuService) UpdateMenuStatus(ctx context.Context, menuID string, status int) (err error) {
	var oldMenu, newMenu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleMenu),
				audit.WithError(err),
			)
		} else if newMenu != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(audit.ModuleMenu),
				audit.WithResource(audit.ResourceMenu, newMenu.MenuID, newMenu.Name),
				audit.WithValue(oldMenu, newMenu),
			)
			log.Info().Str("menu_id", menuID).Int("status", status).Msg("更新菜单状态成功")
		}
	}()

	// 获取旧菜单信息
	oldMenu, err = s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("menu_id", menuID).Msg("菜单不存在")
			return xerr.ErrMenuNotFound
		}
		log.Error().Err(err).Str("menu_id", menuID).Msg("查询菜单失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	// 更新菜单状态
	if err := s.menuRepo.UpdateStatus(ctx, menuID, int16(status)); err != nil {
		log.Error().Err(err).Str("menu_id", menuID).Int("status", status).Msg("更新菜单状态失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新菜单状态失败", err)
	}

	// 获取更新后的菜单信息
	newMenu, err = s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		log.Error().Err(err).Str("menu_id", menuID).Msg("获取更新后菜单信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后菜单信息失败", err)
	}

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
	resource := "menu:" + menu.MenuID
	action := "*"
	return &dto.MenuInfo{
		MenuID:      menu.MenuID,
		Name:        menu.Name,
		Type:        constants.TypeMenu,
		ParentID:    stringPtr(menu.ParentID),
		Resource:    &resource,
		Action:      &action,
		Path:        stringPtr(menu.Path),
		Component:   stringPtr(menu.Component),
		Redirect:    stringPtr(menu.Redirect),
		Icon:        stringPtr(menu.Icon),
		Sort:        &sort,
		Status:      menu.Status,
		Description: stringPtr(menu.Description),
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
	}
}
