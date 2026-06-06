package position

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
)

// CreatePosition 创建岗位
func (s *Service) CreatePosition(ctx context.Context, req *dto.CreatePositionRequest) (resp *dto.PositionInfo, err error) {
	var position *model.Position

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModulePosition),
				audit.WithError(err),
			)
		} else if position != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModulePosition),
				audit.WithResource(constants.ResourceTypePosition, position.PositionID, position.PositionName),
				audit.WithValue(nil, position),
			)
		}
	}()

	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 检查岗位编码是否已存在（租户内唯一）
	var exists bool
	exists, err = s.positionRepo.CheckExists(ctx, tenantID, req.PositionCode)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("position_code", req.PositionCode).Msg("检查岗位编码失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查岗位编码是否存在失败", err)
	}
	if exists {
		log.Warn().Str("tenant_id", tenantID).Str("position_code", req.PositionCode).Msg("岗位编码已存在")
		return nil, xerr.ErrPositionCodeExists
	}

	// 生成岗位ID
	var positionID string
	positionID, err = idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成岗位ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成岗位ID失败", err)
	}

	// 设置默认值
	sort := req.Sort
	if sort == 0 {
		sort = 0 // 默认排序
	}
	status := req.Status
	if status == 0 {
		status = 1 // 默认启用
	}

	// 构建岗位模型
	position = &model.Position{
		PositionID:   positionID,
		TenantID:     tenantID,
		PositionCode: req.PositionCode,
		PositionName: req.PositionName,
		Level:        int32(req.Level),
		Description:  req.Description,
		Sort:         int32(sort),
		Status:       int16(status),
	}

	// 创建岗位
	if err := s.positionRepo.Create(ctx, position); err != nil {
		log.Error().Err(err).Str("position_id", positionID).Str("tenant_id", tenantID).Str("position_code", req.PositionCode).Msg("创建岗位失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建岗位失败", err)
	}

	return modelToPositionInfo(position), nil
}
