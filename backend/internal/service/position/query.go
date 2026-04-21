package position

import (
	"admin/internal/dto"
	"admin/pkg/utils/pagination"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// GetPositionByID 获取岗位详情
func (s *Service) GetPositionByID(ctx context.Context, positionID string) (*dto.PositionInfo, error) {
	position, err := s.positionRepo.GetByID(ctx, positionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("position_id", positionID).Msg("岗位不存在")
			return nil, xerr.ErrPositionNotFound
		}
		log.Error().Err(err).Str("position_id", positionID).Msg("查询岗位失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位失败", err)
	}

	return modelToPositionInfo(position), nil
}

// ListPositions 获取岗位列表
func (s *Service) ListPositions(ctx context.Context, req *dto.ListPositionsRequest) (*dto.ListPositionsResponse, error) {
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
	positionInfos := modelListToPositionInfoList(positions)

	return &dto.ListPositionsResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     positionInfos,
	}, nil
}

// ListAllPositions 获取所有岗位（不分页，用于下拉选择）
func (s *Service) ListAllPositions(ctx context.Context) ([]*dto.PositionInfo, error) {
	positions, err := s.positionRepo.ListAll(ctx)
	if err != nil {
		log.Error().Err(err).Msg("查询所有岗位列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询岗位列表失败", err)
	}

	responses := modelListToPositionInfoList(positions)

	return responses, nil
}
