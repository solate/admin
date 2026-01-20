package converter

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// ModelToDepartmentInfo 将数据库模型转换为部门信息 DTO
func ModelToDepartmentInfo(dept *model.Department) *dto.DepartmentInfo {
	if dept == nil {
		return nil
	}

	return &dto.DepartmentInfo{
		DepartmentID:   dept.DepartmentID,
		ParentID:       dept.ParentID,
		DepartmentName: dept.DepartmentName,
		Description:    dept.Description,
		Sort:           int(dept.Sort),
		Status:         int(dept.Status),
		CreatedAt:      dept.CreatedAt,
		UpdatedAt:      dept.UpdatedAt,
	}
}

// ModelListToDepartmentInfoList 批量将数据库模型转换为部门信息 DTO
func ModelListToDepartmentInfoList(depts []*model.Department) []*dto.DepartmentInfo {
	if len(depts) == 0 {
		return nil
	}

	result := make([]*dto.DepartmentInfo, len(depts))
	for i, dept := range depts {
		result[i] = ModelToDepartmentInfo(dept)
	}
	return result
}
