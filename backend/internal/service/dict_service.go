package service

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/cache"
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

// DictService 字典服务
type DictService struct {
	dictTypeRepo *repository.DictTypeRepo
	dictItemRepo *repository.DictItemRepo
	tenantCache  *cache.TenantCache
}

// NewDictService 创建字典服务
func NewDictService(dictTypeRepo *repository.DictTypeRepo, dictItemRepo *repository.DictItemRepo, tenantCache *cache.TenantCache) *DictService {
	return &DictService{
		dictTypeRepo: dictTypeRepo,
		dictItemRepo: dictItemRepo,
		tenantCache:  tenantCache,
	}
}

// CreateSystemDict 创建系统字典（超管专用）
func (s *DictService) CreateSystemDict(ctx context.Context, req *dto.CreateSystemDictRequest) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 检查字典编码是否已存在（默认租户内唯一）
	exists, err := s.dictTypeRepo.CheckExists(ctx, defaultTenantID, req.TypeCode)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "检查字典编码是否存在失败", err)
	}
	if exists {
		return xerr.New(xerr.ErrInvalidParams.Code, "字典编码已存在")
	}

	// 生成字典类型ID
	typeID, err := idgen.GenerateUUID()
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "生成字典类型ID失败", err)
	}

	// 构建字典类型模型
	dictType := &model.DictType{
		TypeID:      typeID,
		TenantID:    defaultTenantID,
		TypeCode:    req.TypeCode,
		TypeName:    req.TypeName,
		Description: req.Description,
	}

	// 创建字典类型
	if err := s.dictTypeRepo.Create(ctx, dictType); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "创建字典类型失败", err)
	}

	// 创建字典项
	for _, itemReq := range req.Items {
		itemID, err := idgen.GenerateUUID()
		if err != nil {
			return xerr.Wrap(xerr.ErrInternal.Code, "生成字典项ID失败", err)
		}

		dictItem := &model.DictItem{
			ItemID:   itemID,
			TypeID:   typeID,
			TenantID: defaultTenantID,
			Label:    itemReq.Label,
			Value:    itemReq.Value,
			Sort:     int32(itemReq.Sort),
		}

		if err := s.dictItemRepo.Create(ctx, dictItem); err != nil {
			return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
		}
	}

	// 记录操作日志
	ctx = auditlog.RecordCreate(ctx, constants.ModuleDict, constants.ResourceTypeDict, dictType.TypeID, dictType.TypeName, dictType)

	return nil
}

// UpdateSystemDict 更新系统字典（超管专用）
func (s *DictService) UpdateSystemDict(ctx context.Context, typeCode string, req *dto.UpdateSystemDictRequest) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 准备更新数据
	updates := make(map[string]interface{})
	if req.TypeName != "" {
		updates["type_name"] = req.TypeName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["updated_at"] = time.Now().UnixMilli()

	// 更新字典类型
	if err := s.dictTypeRepo.Update(ctx, dictType.TypeID, updates); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "更新字典类型失败", err)
	}

	// 记录操作日志
	ctx = auditlog.RecordUpdate(ctx, constants.ModuleDict, constants.ResourceTypeDict, dictType.TypeID, dictType.TypeName, dictType, updates)

	return nil
}

// DeleteSystemDict 删除系统字典（超管专用）
func (s *DictService) DeleteSystemDict(ctx context.Context, typeCode string) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.ErrNotFound
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 删除字典项（先删除子记录）
	if err := s.dictItemRepo.DeleteByTypeID(ctx, dictType.TypeID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除字典项失败", err)
	}

	// 删除字典类型
	if err := s.dictTypeRepo.Delete(ctx, dictType.TypeID); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除字典类型失败", err)
	}

	// 记录操作日志
	auditlog.RecordDelete(ctx, constants.ModuleDict, constants.ResourceTypeDict, dictType.TypeID, dictType.TypeName, dictType)

	return nil
}

// ListDictTypes 获取字典类型列表（超管专用）
func (s *DictService) ListDictTypes(ctx context.Context, req *dto.ListDictTypesRequest) (*dto.ListDictTypesResponse, error) {
	dictTypes, total, err := s.dictTypeRepo.ListWithFilters(ctx, req.GetOffset(), req.GetLimit(), req.Keyword)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型列表失败", err)
	}

	// 转换为响应格式
	dictTypeResponses := make([]*dto.DictTypeResponse, len(dictTypes))
	for i, dictType := range dictTypes {
		dictTypeResponses[i] = s.toDictTypeResponse(dictType)
	}

	return &dto.ListDictTypesResponse{
		Response: pagination.NewResponse(req.Request, total),
		List:     dictTypeResponses,
	}, nil
}

// GetDictByCode 获取字典（合并系统+覆盖）
func (s *DictService) GetDictByCode(ctx context.Context, typeCode string) (*dto.DictResponse, error) {
	currentTenantID := xcontext.GetTenantID(ctx)
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取字典类型及合并后的字典项
	dictType, items, err := s.dictItemRepo.GetDictTypeWithItems(ctx, typeCode, defaultTenantID, currentTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询字典失败", err)
	}

	// 转换为 Response 格式
	itemResponses := make([]*dto.DictItemResponse, len(items))
	for i, item := range items {
		source := "system"
		if item.TenantID == currentTenantID {
			source = "custom"
		}
		itemResponses[i] = &dto.DictItemResponse{
			ItemID: item.ItemID,
			Label:  item.Label,
			Value:  item.Value,
			Sort:   int(item.Sort),
			Source: source,
		}
	}

	return &dto.DictResponse{
		TypeID:   dictType.TypeID,
		TypeCode: dictType.TypeCode,
		TypeName: dictType.TypeName,
		Items:    itemResponses,
	}, nil
}

// BatchUpdateDictItems 批量更新字典项（租户覆盖）
func (s *DictService) BatchUpdateDictItems(ctx context.Context, req *dto.BatchUpdateDictItemsRequest) error {
	currentTenantID := xcontext.GetTenantID(ctx)
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, req.Items[0].TypeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 批量更新字典项
	for _, itemReq := range req.Items {
		if itemReq.TypeCode != dictType.TypeCode {
			return xerr.New(xerr.ErrInvalidParams.Code, "所有字典项必须属于同一字典类型")
		}

		// 检查租户是否已有覆盖记录
		existing, err := s.dictItemRepo.GetByTypeAndValue(ctx, dictType.TypeID, currentTenantID, itemReq.Value)
		if err != nil && err != gorm.ErrRecordNotFound {
			return xerr.Wrap(xerr.ErrInternal.Code, "查询字典项失败", err)
		}

		if existing != nil {
			// 更新现有记录
			updates := map[string]interface{}{
				"label":     itemReq.Label,
				"sort":      int32(itemReq.Sort),
				"updated_at": time.Now().UnixMilli(),
			}
			if err := s.dictItemRepo.Update(ctx, existing.ItemID, updates); err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "更新字典项失败", err)
			}
		} else {
			// 创建新的覆盖记录
			itemID, err := idgen.GenerateUUID()
			if err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "生成字典项ID失败", err)
			}

			dictItem := &model.DictItem{
				ItemID:   itemID,
				TypeID:   dictType.TypeID,
				TenantID: currentTenantID,
				Label:    itemReq.Label,
				Value:    itemReq.Value,
				Sort:     int32(itemReq.Sort),
			}

			if err := s.dictItemRepo.Create(ctx, dictItem); err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
			}
		}
	}

	return nil
}

// ResetDictItem 恢复字典项系统默认值
func (s *DictService) ResetDictItem(ctx context.Context, typeCode, value string) error {
	currentTenantID := xcontext.GetTenantID(ctx)
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 删除租户的覆盖记录（恢复系统默认）
	if err := s.dictItemRepo.DeleteByTypeAndValue(ctx, dictType.TypeID, currentTenantID, value); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "恢复系统默认值失败", err)
	}

	return nil
}

// toDictTypeResponse 转换为字典类型响应格式
func (s *DictService) toDictTypeResponse(dictType *model.DictType) *dto.DictTypeResponse {
	return &dto.DictTypeResponse{
		TypeID:      dictType.TypeID,
		TenantID:    dictType.TenantID,
		TypeCode:    dictType.TypeCode,
		TypeName:    dictType.TypeName,
		Description: dictType.Description,
		CreatedAt:   dictType.CreatedAt,
		UpdatedAt:   dictType.UpdatedAt,
	}
}
