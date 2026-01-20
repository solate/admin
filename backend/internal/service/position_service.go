package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/convert"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// PositionService 岗位服务
type PositionService struct {
	positionRepo *repository.PositionRepo
	recorder     *audit.Recorder
}

// NewPositionService 创建岗位服务
func NewPositionService(positionRepo *repository.PositionRepo, recorder *audit.Recorder) *PositionService {
	return &PositionService{
		positionRepo: positionRepo,
		recorder:     recorder,
	}
}

// CreatePosition 创建岗位
func (s *PositionService) CreatePosition(ctx context.Context, req *dto.CreatePositionRequest) (resp *dto.PositionInfo, err error) {
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

	return converter.ModelToPositionInfo(position), nil
}

// GetPositionByID 获取岗位详情
func (s *PositionService) GetPositionByID(ctx context.Context, positionID string) (*dto.PositionInfo, error) {
	position, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("position_id", positionID).Msg("岗位不存在")
			return nil, xerr.ErrPositionNotFound
		}
		log.Error().Err(err).Str("position_id", positionID).Msg("查询岗位失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	return converter.ModelToPositionInfo(position), nil
}

// UpdatePosition 更新岗位
func (s *PositionService) UpdatePosition(ctx context.Context, positionID string, req *dto.UpdatePositionRequest) (resp *dto.PositionInfo, err error) {
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

	return converter.ModelToPositionInfo(newPosition), nil
}

// DeletePosition 删除岗位
func (s *PositionService) DeletePosition(ctx context.Context, positionID string) (err error) {
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
func (s *PositionService) BatchDeletePositions(ctx context.Context, positionIDs []string) (err error) {
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

// ListPositions 获取岗位列表
func (s *PositionService) ListPositions(ctx context.Context, req *dto.ListPositionsRequest) (*dto.ListPositionsResponse, error) {
	positions, total, err := s.positionRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.PositionName, req.PositionCode, req.Status)
	if err != nil {
		log.Error().Err(err).
			Str("position_name", req.PositionName).
			Str("position_code", req.PositionCode).
			Int("status", req.Status).
			Int("offset", req.GetOffset()).
			Int("limit", req.GetLimit()).
			Msg("查询岗位列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位列表失败", err)
	}

	// 转换为响应格式
	positionInfos := converter.ModelListToPositionInfoList(positions)

	return &dto.ListPositionsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     positionInfos,
	}, nil
}

// ListAllPositions 获取所有岗位（不分页，用于下拉选择）
func (s *PositionService) ListAllPositions(ctx context.Context) ([]*dto.PositionInfo, error) {
	positions, err := s.positionRepo.ListAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("查询所有岗位列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位列表失败", err)
	}

	responses := converter.ModelListToPositionInfoList(positions)

	return responses, nil
}

// UpdatePositionStatus 更新岗位状态
func (s *PositionService) UpdatePositionStatus(ctx context.Context, positionID string, status int) (err error) {
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
