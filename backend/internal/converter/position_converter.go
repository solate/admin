package converter

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// ModelToPositionInfo 将数据库模型转换为岗位信息 DTO
func ModelToPositionInfo(position *model.Position) *dto.PositionInfo {
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

// ModelListToPositionInfoList 批量将数据库模型转换为岗位信息 DTO
func ModelListToPositionInfoList(positions []*model.Position) []*dto.PositionInfo {
	if len(positions) == 0 {
		return nil
	}

	result := make([]*dto.PositionInfo, len(positions))
	for i, position := range positions {
		result[i] = ModelToPositionInfo(position)
	}
	return result
}
