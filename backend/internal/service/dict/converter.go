package dict

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
)

// modelToDictTypeInfo 将数据库模型转换为字典类型信息 DTO
func modelToDictTypeInfo(dictType *model.DictType) *dto.DictTypeInfo {
	if dictType == nil {
		return nil
	}

	return &dto.DictTypeInfo{
		TypeID:      dictType.TypeID,
		TenantID:    dictType.TenantID,
		TypeCode:    dictType.TypeCode,
		TypeName:    dictType.TypeName,
		Description: dictType.Description,
		CreatedAt:   dictType.CreatedAt,
		UpdatedAt:   dictType.UpdatedAt,
	}
}

// modelListToDictTypeInfoList 批量将数据库模型转换为字典类型信息 DTO
func modelListToDictTypeInfoList(dictTypes []*model.DictType) []*dto.DictTypeInfo {
	if len(dictTypes) == 0 {
		return nil
	}

	result := make([]*dto.DictTypeInfo, len(dictTypes))
	for i, dictType := range dictTypes {
		result[i] = modelToDictTypeInfo(dictType)
	}
	return result
}
