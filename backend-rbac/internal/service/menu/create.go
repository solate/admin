package menu

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateMenu 创建菜单
func (s *Service) CreateMenu(ctx context.Context, req *dto.CreateMenuRequest) (resp *dto.MenuInfo, err error) {
	var menu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleMenu),
				audit.WithError(err),
			)
		} else if menu != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleMenu),
				audit.WithResource(constants.ResourceTypeMenu, menu.MenuID, menu.Name),
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
	if req.Status == constants.StatusZero {
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
	return ModelToMenuInfo(menu), nil
}
