package position

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

// UpdatePosition 更新岗位
func (s *Service) UpdatePosition(ctx context.Context, positionID string, req *dto.UpdatePositionRequest) (resp *dto.PositionInfo, err error) {
	var oldPosition, newPosition *model.Position

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModulePosition),
				audit.WithError(err),
			)
		} else if newPosition != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModulePosition),
				audit.WithResource(constants.ResourceTypePosition, newPosition.PositionID, newPosition.PositionName),
				audit.WithValue(oldPosition, newPosition),
			)
		}
	}()

	// 获取旧岗位信息
	oldPosition, err = s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("position_id", positionID).Msg("岗位不存在")
			return nil, xerr.ErrPositionNotFound
		}
		log.Error().Err(err).Str("position_id", positionID).Msg("查询岗位失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	// 如果要修改岗位编码，检查编码是否已存在
	if req.PositionCode != "" && req.PositionCode != oldPosition.PositionCode {
		var exists bool
		exists, err = s.positionRepo.CheckExistsByID(ctx, oldPosition.TenantID, req.PositionCode, positionID)
		if err != nil {
			log.Error().Err(err).Str("position_id", positionID).Str("tenant_id", oldPosition.TenantID).Str("position_code", req.PositionCode).Msg("检查岗位编码是否存在失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查岗位编码是否存在失败", err)
		}
		if exists {
			log.Warn().Str("position_id", positionID).Str("tenant_id", oldPosition.TenantID).Str("position_code", req.PositionCode).Msg("岗位编码已存在")
			return nil, xerr.ErrPositionCodeExists
		}
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.PositionCode != "" {
		updates["position_code"] = req.PositionCode
	}
	if req.PositionName != "" {
		updates["position_name"] = req.PositionName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Level != 0 {
		updates["level"] = req.Level
	}
	if req.Sort != 0 {
		updates["sort"] = req.Sort
	}
	if req.Status != constants.StatusZero {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新岗位
	if err := s.positionRepo.Update(ctx, positionID, updates); err != nil {
		log.Error().Err(err).Str("position_id", positionID).Interface("updates", updates).Msg("更新岗位失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新岗位失败", err)
	}

	// 获取更新后的岗位信息
	newPosition, err = s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		log.Error().Err(err).Str("position_id", positionID).Msg("获取更新后岗位信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后岗位信息失败", err)
	}

	return modelToPositionInfo(newPosition), nil
}

// UpdatePositionStatus 更新岗位状态
func (s *Service) UpdatePositionStatus(ctx context.Context, positionID string, status int) (err error) {
	var oldPosition, newPosition *model.Position

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModulePosition),
				audit.WithError(err),
			)
		} else if newPosition != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModulePosition),
				audit.WithResource(constants.ResourceTypePosition, newPosition.PositionID, newPosition.PositionName),
				audit.WithValue(oldPosition, newPosition),
			)
			log.Info().Str("position_id", positionID).Int("status", status).Str("tenant_id", newPosition.TenantID).Msg("更新岗位状态成功")
		}
	}()

	// 获取旧岗位信息
	oldPosition, err = s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("position_id", positionID).Msg("岗位不存在")
			return xerr.ErrPositionNotFound
		}
		log.Error().Err(err).Str("position_id", positionID).Msg("查询岗位失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	// 更新岗位状态
	if err := s.positionRepo.UpdateStatus(ctx, positionID, status); err != nil {
		log.Error().Err(err).Str("position_id", positionID).Int("status", status).Msg("更新岗位状态失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新岗位状态失败", err)
	}

	// 获取更新后的岗位信息
	newPosition, err = s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		log.Error().Err(err).Str("position_id", positionID).Msg("获取更新后岗位信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后岗位信息失败", err)
	}

	return nil
}
