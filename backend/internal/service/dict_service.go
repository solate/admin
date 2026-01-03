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

	// 如果提供了字典项列表，则更新字典项
	if len(req.Items) > 0 {
		// 删除所有旧的系统字典项
		if err := s.dictItemRepo.DeleteByTypeID(ctx, dictType.TypeID); err != nil {
			return xerr.Wrap(xerr.ErrInternal.Code, "删除旧字典项失败", err)
		}

		// 创建新的字典项
		for _, itemReq := range req.Items {
			itemID, err := idgen.GenerateUUID()
			if err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "生成字典项ID失败", err)
			}

			dictItem := &model.DictItem{
				ItemID:   itemID,
				TypeID:   dictType.TypeID,
				TenantID: defaultTenantID,
				Label:    itemReq.Label,
				Value:    itemReq.Value,
				Sort:     int32(itemReq.Sort),
			}

			if err := s.dictItemRepo.Create(ctx, dictItem); err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
			}
		}
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

// BatchUpdateDictItems 批量更新字典项
func (s *DictService) BatchUpdateDictItems(ctx context.Context, req *dto.BatchUpdateDictItemsRequest) error {
	currentTenantID := xcontext.GetTenantID(ctx)
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, req.TypeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 如果是超管（默认租户），直接更新系统字典项
	if currentTenantID == defaultTenantID {
		return s.batchUpdateSystemDictItems(ctx, dictType, req)
	}

	// 否则是租户覆盖逻辑
	return s.batchUpdateTenantDictItems(ctx, dictType, currentTenantID, req)
}

// batchUpdateSystemDictItems 超管批量更新系统字典项（可以增删改）
func (s *DictService) batchUpdateSystemDictItems(ctx context.Context, dictType *model.DictType, req *dto.BatchUpdateDictItemsRequest) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取现有的系统字典项
	existingItems, err := s.dictItemRepo.GetByTypeAndTenant(ctx, dictType.TypeID, defaultTenantID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return xerr.Wrap(xerr.ErrInternal.Code, "查询现有字典项失败", err)
	}
	if existingItems == nil {
		existingItems = []*model.DictItem{}
	}

	// 创建 item_id -> item 的映射
	existingMap := make(map[string]*model.DictItem)
	for _, item := range existingItems {
		existingMap[item.ItemID] = item
	}

	// 记录已处理的 item_id
	processedIDs := make(map[string]bool)

	// 处理请求中的字典项
	for _, itemReq := range req.Items {
		// 尝试根据 label 找到现有项（前端传的是 label 和 sort，没有 value）
		var existingItem *model.DictItem
		for _, item := range existingItems {
			if item.Label == itemReq.Label || item.Sort == int32(itemReq.Sort) {
				existingItem = item
				break
			}
		}

		if existingItem != nil {
			// 更新现有项
			processedIDs[existingItem.ItemID] = true
			updates := map[string]interface{}{
				"label":     itemReq.Label,
				"sort":      int32(itemReq.Sort),
				"updated_at": time.Now().UnixMilli(),
			}
			if err := s.dictItemRepo.Update(ctx, existingItem.ItemID, updates); err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "更新字典项失败", err)
			}
		} else {
			// 创建新项（超管可以添加新项，需要生成 value）
			itemID, err := idgen.GenerateUUID()
			if err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "生成字典项ID失败", err)
			}

			// 使用 label 作为 value（超管可以自定义）
			dictItem := &model.DictItem{
				ItemID:   itemID,
				TypeID:   dictType.TypeID,
				TenantID: defaultTenantID,
				Label:    itemReq.Label,
				Value:    itemReq.Label, // 超管添加新项时，使用 label 作为 value
				Sort:     int32(itemReq.Sort),
			}

			if err := s.dictItemRepo.Create(ctx, dictItem); err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
			}
		}
	}

	// 删除没有在请求中的项（超管可以删除）
	for _, item := range existingItems {
		if !processedIDs[item.ItemID] {
			if err := s.dictItemRepo.Delete(ctx, item.ItemID); err != nil {
				return xerr.Wrap(xerr.ErrInternal.Code, "删除字典项失败", err)
			}
		}
	}

	return nil
}

// batchUpdateTenantDictItems 租户批量更新字典项（只能覆盖，不能增删）
func (s *DictService) batchUpdateTenantDictItems(ctx context.Context, dictType *model.DictType, currentTenantID string, req *dto.BatchUpdateDictItemsRequest) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统默认的字典项（已按 sort 排序）
	systemItems, err := s.dictItemRepo.GetByTypeAndTenant(ctx, dictType.TypeID, defaultTenantID)
	if err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "查询系统字典项失败", err)
	}

	// 获取该租户现有的所有覆盖记录
	existingItems, err := s.dictItemRepo.GetByTypeAndTenant(ctx, dictType.TypeID, currentTenantID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return xerr.Wrap(xerr.ErrInternal.Code, "查询现有字典项失败", err)
	}
	if existingItems == nil {
		existingItems = []*model.DictItem{}
	}

	// 创建 value -> item 的映射
	existingMap := make(map[string]*model.DictItem)
	for _, item := range existingItems {
		existingMap[item.Value] = item
	}

	// 批量更新：按索引顺序，从系统字典项中获取 value
	for i, itemReq := range req.Items {
		if i >= len(systemItems) {
			return xerr.New(xerr.ErrInvalidParams.Code, "字典项数量超过系统默认项数量")
		}

		systemItem := systemItems[i]
		existing := existingMap[systemItem.Value]

		if existing != nil {
			// 更新现有记录
			updates := map[string]interface{}{
				"label":      itemReq.Label,
				"sort":       int32(itemReq.Sort),
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
				Value:    systemItem.Value,
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

// DeleteSystemDictItem 删除系统字典项（超管专用）
func (s *DictService) DeleteSystemDictItem(ctx context.Context, typeCode, value string) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 删除系统字典项（真正的删除）
	if err := s.dictItemRepo.DeleteByTypeAndValue(ctx, dictType.TypeID, defaultTenantID, value); err != nil {
		return xerr.Wrap(xerr.ErrInternal.Code, "删除字典项失败", err)
	}

	// 记录操作日志
	auditlog.RecordDelete(ctx, constants.ModuleDict, constants.ResourceTypeDictItem, dictType.TypeID, value, value)

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
