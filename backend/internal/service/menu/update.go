package menu

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateMenu 更新菜单
func (s *Service) UpdateMenu(ctx context.Context, menuID string, req *dto.UpdateMenuRequest) (resp *dto.MenuInfo, err error) {
	var oldMenu, newMenu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleMenu),
				audit.WithError(err),
			)
		} else if newMenu != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleMenu),
				audit.WithResource(constants.ResourceTypeMenu, newMenu.MenuID, newMenu.Name),
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
	if req.Status != constants.StatusZero {
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
	return ModelToMenuInfo(newMenu), nil
}

// UpdateMenuStatus 更新菜单状态
func (s *Service) UpdateMenuStatus(ctx context.Context, menuID string, status int) (err error) {
	var oldMenu, newMenu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleMenu),
				audit.WithError(err),
			)
		} else if newMenu != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleMenu),
				audit.WithResource(constants.ResourceTypeMenu, newMenu.MenuID, newMenu.Name),
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
