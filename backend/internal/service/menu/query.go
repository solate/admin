package menu

import (
	"admin/internal/dto"
	"admin/pkg/utils/pagination"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetMenuByID 获取菜单详情
func (s *Service) GetMenuByID(ctx context.Context, menuID string) (*dto.MenuInfo, error) {
	menu, err := s.menuRepo.GetByID(ctx, menuID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("menu_id", menuID).Msg("菜单不存在")
			return nil, xerr.ErrMenuNotFound
		}
		log.Error().Err(err).Str("menu_id", menuID).Msg("查询菜单失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单失败", err)
	}

	return ModelToMenuInfo(menu), nil
}

// ListMenus 获取菜单列表
func (s *Service) ListMenus(ctx context.Context, req *dto.ListMenusRequest) (*dto.ListMenusResponse, error) {
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
	menuInfos := ModelListToMenuInfoList(menus)

	return &dto.ListMenusResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     menuInfos,
	}, nil
}

// GetAllMenus 获取所有菜单（平铺）
func (s *Service) GetAllMenus(ctx context.Context) (*dto.AllMenusResponse, error) {
	// 获取所有菜单
	menus, err := s.menuRepo.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("查询菜单列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询菜单列表失败", err)
	}

	// 转换为响应格式
	menuInfos := ModelListToMenuInfoList(menus)

	return &dto.AllMenusResponse{
		List: menuInfos,
	}, nil
}
