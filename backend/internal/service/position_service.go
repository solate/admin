package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/constants"
	"admin/pkg/idgen"
	"admin/pkg/auditlog"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"gorm.io/gorm"
)

// PositionService 岗位服务
type PositionService struct {
	positionRepo *repository.PositionRepo
}

// NewPositionService 创建岗位服务
func NewPositionService(positionRepo *repository.PositionRepo) *PositionService {
	return &PositionService{
		positionRepo: positionRepo,
	}
}

// CreatePosition 创建岗位
func (s *PositionService) CreatePosition(ctx context.Context, req *dto.CreatePositionRequest) (*dto.PositionResponse, error) {
	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 检查岗位编码是否已存在（租户内唯一）
	exists, err := s.positionRepo.CheckExists(ctx, tenantID, req.PositionCode)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查岗位编码是否存在失败", err)
	}
	if exists {
		return nil, xerr.ErrPositionCodeExists
	}

	// 生成岗位ID
	positionID, err := idgen.GenerateUUID()
	if err != nil {
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
	position := &model.Position{
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
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建岗位失败", err)
	}

	// 记录操作日志
	ctx = auditlog.RecordCreate(ctx, constants.ModulePosition, constants.ResourceTypePosition, position.PositionID, position.PositionName, position)

	return s.toPositionResponse(position), nil
}

// GetPositionByID 获取岗位详情
func (s *PositionService) GetPositionByID(ctx context.Context, positionID string) (*dto.PositionResponse, error) {
	position, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrPositionNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	return s.toPositionResponse(position), nil
}

// UpdatePosition 更新岗位
func (s *PositionService) UpdatePosition(ctx context.Context, positionID string, req *dto.UpdatePositionRequest) (*dto.PositionResponse, error) {
	// 检查岗位是否存在，获取旧值用于日志
	oldPosition, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.ErrPositionNotFound
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	// 如果要修改岗位编码，检查编码是否已存在
	if req.PositionCode != "" && req.PositionCode != oldPosition.PositionCode {
		exists, err := s.positionRepo.CheckExistsByID(ctx, oldPosition.TenantID, req.PositionCode, positionID)
		if err != nil {
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查岗位编码是否存在失败", err)
		}
		if exists {
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
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新岗位
	if err := s.positionRepo.Update(ctx, positionID, updates); err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "更新岗位失败", err)
	}

	// 获取更新后的岗位信息
	updatedPosition, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "获取更新后岗位信息失败", err)
	}

	// 记录操作日志
	ctx = auditlog.RecordUpdate(ctx, constants.ModulePosition, constants.ResourceTypePosition, updatedPosition.PositionID, updatedPosition.PositionName, oldPosition, updatedPosition)

	return s.toPositionResponse(updatedPosition), nil
}

// DeletePosition 删除岗位
func (s *PositionService) DeletePosition(ctx context.Context, positionID string) error {
	// 检查岗位是否存在，获取岗位信息用于日志
	position, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrPositionNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	// 删除岗位
	if err := s.positionRepo.Delete(ctx, positionID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除岗位失败", err)
	}

	// 记录操作日志
	auditlog.RecordDelete(ctx, constants.ModulePosition, constants.ResourceTypePosition, position.PositionID, position.PositionName, position)

	return nil
}

// ListPositions 获取岗位列表
func (s *PositionService) ListPositions(ctx context.Context, req *dto.ListPositionsRequest) (*dto.ListPositionsResponse, error) {
	positions, total, err := s.positionRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword, req.Status)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位列表失败", err)
	}

	// 转换为响应格式
	positionResponses := make([]*dto.PositionResponse, len(positions))
	for i, position := range positions {
		positionResponses[i] = s.toPositionResponse(position)
	}

	return &dto.ListPositionsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     positionResponses,
	}, nil
}

// ListAllPositions 获取所有岗位（不分页，用于下拉选择）
func (s *PositionService) ListAllPositions(ctx context.Context) ([]*dto.PositionResponse, error) {
	positions, err := s.positionRepo.ListAll(ctx)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位列表失败", err)
	}

	responses := make([]*dto.PositionResponse, len(positions))
	for i, position := range positions {
		responses[i] = s.toPositionResponse(position)
	}

	return responses, nil
}

// UpdatePositionStatus 更新岗位状态
func (s *PositionService) UpdatePositionStatus(ctx context.Context, positionID string, status int) error {
	// 检查岗位是否存在，获取旧值用于日志
	oldPosition, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrPositionNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	// 更新岗位状态
	if err := s.positionRepo.UpdateStatus(ctx, positionID, status); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新岗位状态失败", err)
	}

	// 获取更新后的岗位信息
	updatedPosition, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后岗位信息失败", err)
	}

	// 记录操作日志
	auditlog.RecordUpdate(ctx, constants.ModulePosition, constants.ResourceTypePosition, updatedPosition.PositionID, updatedPosition.PositionName, oldPosition, updatedPosition)

	return nil
}

// toPositionResponse 转换为岗位响应格式
func (s *PositionService) toPositionResponse(position *model.Position) *dto.PositionResponse {
	return &dto.PositionResponse{
		PositionID:   position.PositionID,
		PositionCode: position.PositionCode,
		PositionName: position.PositionName,
		Level:        int(position.Level),
		Description:  position.Description,
		Sort:         int(position.Sort),
		Status:       int(position.Status),
		CreatedAt:    position.CreatedAt,
		UpdatedAt:    position.UpdatedAt,
	}
}
