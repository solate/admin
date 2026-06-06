package dict

import (
	"admin/internal/dto"
	"admin/pkg/utils/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ListDictTypes 获取字典类型列表（超管专用）
func (s *Service) ListDictTypes(ctx context.Context, req *dto.ListDictTypesRequest) (*dto.ListDictTypesResponse, error) {
	dictTypes, total, err := s.dictTypeRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.TypeName, req.TypeCode)
	if err != nil {
		log.Error().Err(err).
			Str("type_name", req.TypeName).
			Str("type_code", req.TypeCode).
			Msg("查询字典类型列表失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型列表失败", err)
	}

	return &dto.ListDictTypesResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     modelListToDictTypeInfoList(dictTypes),
	}, nil
}

// GetDictByCode 获取字典（合并系统+覆盖）
func (s *Service) GetDictByCode(ctx context.Context, typeCode string) (*dto.DictInfo, error) {
	currentTenantID := xcontext.GetTenantID(ctx)
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取字典类型及合并后的字典项
	dictType, items, err := s.dictItemRepo.GetDictTypeWithItems(ctx, typeCode, defaultTenantID, currentTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("type_code", typeCode).Str("tenant_id", currentTenantID).Msg("字典不存在")
			return nil, xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		log.Error().Err(err).Str("type_code", typeCode).Str("tenant_id", currentTenantID).Msg("查询字典失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询字典失败", err)
	}

	// 转换为 Info 格式
	itemInfos := make([]*dto.DictItemInfo, len(items))
	for i, item := range items {
		source := "system"
		if item.TenantID == currentTenantID {
			source = "custom"
		}
		itemInfos[i] = &dto.DictItemInfo{
			ItemID: item.ItemID,
			Label:  item.Label,
			Value:  item.Value,
			Sort:   int(item.Sort),
			Source: source,
		}
	}

	return &dto.DictInfo{
		TypeID:   dictType.TypeID,
		TypeCode: dictType.TypeCode,
		TypeName: dictType.TypeName,
		Items:    itemInfos,
	}, nil
}

// ListSystemDictTypes 获取系统字典类型列表（超管专用，与 ListDictTypes 逻辑一致）
func (s *Service) ListSystemDictTypes(ctx context.Context, req *dto.ListDictTypesRequest) (*dto.ListDictTypesResponse, error) {
	return s.ListDictTypes(ctx, req)
}
