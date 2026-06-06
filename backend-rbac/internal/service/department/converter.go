package department

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// modelToDepartmentInfo 将数据库模型转换为部门信息 DTO
func modelToDepartmentInfo(dept *model.Department) *dto.DepartmentInfo {
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

// modelListToDepartmentInfoList 批量将数据库模型转换为部门信息 DTO
func modelListToDepartmentInfoList(depts []*model.Department) []*dto.DepartmentInfo {
	if len(depts) == 0 {
		return nil
	}

	result := make([]*dto.DepartmentInfo, len(depts))
	for i, dept := range depts {
		result[i] = modelToDepartmentInfo(dept)
	}
	return result
}
