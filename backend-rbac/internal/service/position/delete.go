package position

import (
	"admin/internal/dal/model"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/convert"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeletePosition 删除岗位
func (s *Service) DeletePosition(ctx context.Context, positionID string) (err error) {
	var position *model.Position

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModulePosition),
				audit.WithError(err),
			)
		} else if position != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModulePosition),
				audit.WithResource(constants.ResourceTypePosition, position.PositionID, position.PositionName),
				audit.WithValue(position, nil),
			)
			log.Info().Str("position_id", positionID).Str("tenant_id", position.TenantID).Msg("删除岗位成功")
		}
	}()

	// 检查岗位是否存在
	position, err = s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("position_id", positionID).Msg("岗位不存在")
			return xerr.ErrPositionNotFound
		}
		log.Error().Err(err).Str("position_id", positionID).Msg("查询岗位失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	// 删除岗位
	if err := s.positionRepo.Delete(ctx, positionID); err != nil {
		log.Error().Err(err).Str("position_id", positionID).Str("tenant_id", position.TenantID).Msg("删除岗位失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除岗位失败", err)
	}

	return nil
}

// BatchDeletePositions 批量删除岗位
func (s *Service) BatchDeletePositions(ctx context.Context, positionIDs []string) (err error) {
	var positionMap map[string]*model.Position

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModulePosition),
				audit.WithError(err),
			)
		} else if len(positionMap) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(positionMap))
			names := make([]string, 0, len(positionMap))
			for _, position := range positionMap {
				ids = append(ids, position.PositionID)
				names = append(names, position.PositionName)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModulePosition),
				audit.WithBatchResource(constants.ResourceTypePosition, ids, names),
				audit.WithValue(positionMap, nil),
			)
			log.Info().Strs("position_ids", positionIDs).Int("count", len(positionIDs)).Msg("批量删除岗位成功")
		}
	}()

	// 获取所有岗位信息
	positions, err := s.positionRepo.GetByIDs(ctx, positionIDs)
	if err != nil {
		log.Error().Err(err).Strs("position_ids", positionIDs).Msg("查询岗位信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询岗位信息失败", err)
	}
	positionMap = convert.ToMap(positions, func(p *model.Position) string { return p.PositionID })

	// 验证所有岗位都存在
	if len(positionMap) != len(positionIDs) {
		var missingIDs []string
		for _, id := range positionIDs {
			if _, exists := positionMap[id]; !exists {
				missingIDs = append(missingIDs, id)
			}
		}
		log.Warn().Strs("missing_ids", missingIDs).Msg("部分岗位不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分岗位不存在")
	}

	// 批量删除岗位
	if err := s.positionRepo.BatchDelete(ctx, positionIDs); err != nil {
		log.Error().Err(err).Strs("position_ids", positionIDs).Msg("批量删除岗位失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除岗位失败", err)
	}

	return nil
}
