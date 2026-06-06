package position

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// modelToPositionInfo 将数据库模型转换为岗位信息 DTO
func modelToPositionInfo(position *model.Position) *dto.PositionInfo {
	if position == nil {
		return nil
	}

	return &dto.PositionInfo{
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

// modelListToPositionInfoList 批量将数据库模型转换为岗位信息 DTO
func modelListToPositionInfoList(positions []*model.Position) []*dto.PositionInfo {
	if len(positions) == 0 {
		return nil
	}

	result := make([]*dto.PositionInfo, len(positions))
	for i, position := range positions {
		result[i] = modelToPositionInfo(position)
	}
	return result
}
