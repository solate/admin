package menu

import (
	"admin/internal/dal/model"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeleteMenu 删除菜单
// 说明：删除菜单时会自动清理所有租户中该菜单的权限策略
func (s *Service) DeleteMenu(ctx context.Context, menuID string) (err error) {
	var menu *model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleMenu),
				audit.WithError(err),
			)
		} else if menu != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleMenu),
				audit.WithResource(constants.ResourceTypeMenu, menu.MenuID, menu.Name),
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

	// 通知权限缓存刷新
	s.cache.NotifyRefresh()

	return nil
}

// BatchDeleteMenus 批量删除菜单
func (s *Service) BatchDeleteMenus(ctx context.Context, menuIDs []string) (err error) {
	var menuMap map[string]*model.Menu

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleMenu),
				audit.WithError(err),
			)
		} else if len(menuMap) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(menuMap))
			names := make([]string, 0, len(menuMap))
			for _, menu := range menuMap {
				ids = append(ids, menu.MenuID)
				names = append(names, menu.Name)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleMenu),
				audit.WithBatchResource(constants.ResourceTypeMenu, ids, names),
				audit.WithValue(menuMap, nil),
			)
			log.Info().Strs("menu_ids", menuIDs).Int("count", len(menuIDs)).Msg("批量删除菜单成功")
		}
	}()

	// 获取所有菜单信息
	menuMap, err = s.menuRepo.GetByIDsAsMap(ctx, menuIDs)
	if err != nil {
		log.Error().Err(err).Strs("menu_ids", menuIDs).Msg("查询菜单信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询菜单信息失败", err)
	}

	// 验证所有菜单都存在
	if len(menuMap) != len(menuIDs) {
		var missingIDs []string
		for _, id := range menuIDs {
			if _, exists := menuMap[id]; !exists {
				missingIDs = append(missingIDs, id)
			}
		}
		log.Warn().Strs("missing_ids", missingIDs).Msg("部分菜单不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分菜单不存在")
	}

	// 检查是否有菜单存在子菜单
	for _, menuID := range menuIDs {
		childrenCount, err := s.menuRepo.GetChildrenCount(ctx, menuID)
		if err != nil {
			log.Error().Err(err).Str("menu_id", menuID).Msg("检查子菜单失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "检查子菜单失败", err)
		}
		if childrenCount > 0 {
			log.Warn().Str("menu_id", menuID).Int64("children_count", childrenCount).Msg("菜单存在子菜单，无法删除")
			return xerr.ErrMenuHasChildren
		}
	}

	// 批量删除菜单
	if err := s.menuRepo.DeleteBatch(ctx, menuIDs); err != nil {
		log.Error().Err(err).Strs("menu_ids", menuIDs).Msg("批量删除菜单失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除菜单失败", err)
	}

	return nil
}
