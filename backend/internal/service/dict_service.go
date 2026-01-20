package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/cache"
	"admin/pkg/constants"
	"admin/pkg/convert"
	"admin/pkg/idgen"
	"admin/pkg/pagination"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DictService 字典服务
type DictService struct {
	dictTypeRepo *repository.DictTypeRepo
	dictItemRepo *repository.DictItemRepo
	tenantCache  *cache.TenantCache
	recorder     *audit.Recorder
}

// NewDictService 创建字典服务
func NewDictService(dictTypeRepo *repository.DictTypeRepo, dictItemRepo *repository.DictItemRepo, tenantCache *cache.TenantCache, recorder *audit.Recorder) *DictService {
	return &DictService{
		dictTypeRepo: dictTypeRepo,
		dictItemRepo: dictItemRepo,
		tenantCache:  tenantCache,
		recorder:     recorder,
	}
}

// CreateSystemDict 创建系统字典（超管专用）
func (s *DictService) CreateSystemDict(ctx context.Context, req *dto.CreateSystemDictRequest) (err error) {
	var dictType *model.DictType

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleDict),
				audit.WithError(err),
			)
		} else if dictType != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleDict),
				audit.WithResource(constants.ResourceTypeDict, dictType.TypeID, dictType.TypeName),
				audit.WithValue(nil, dictType),
			)
		}
	}()

	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 检查字典编码是否已存在（默认租户内唯一）
	var exists bool
	exists, err = s.dictTypeRepo.CheckExists(ctx, defaultTenantID, req.TypeCode)
	if err != nil {
		log.Error().Err(err).Str("type_code", req.TypeCode).Str("tenant_id", defaultTenantID).Msg("检查字典编码是否存在失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "检查字典编码是否存在失败", err)
	}
	if exists {
		log.Warn().Str("type_code", req.TypeCode).Str("tenant_id", defaultTenantID).Msg("字典编码已存在")
		return xerr.New(xerr.ErrInvalidParams.Code, "字典编码已存在")
	}

	// 生成字典类型ID
	var typeID string
	typeID, err = idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成字典类型ID失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "生成字典类型ID失败", err)
	}

	// 构建字典类型模型
	dictType = &model.DictType{
		TypeID:      typeID,
		TenantID:    defaultTenantID,
		TypeCode:    req.TypeCode,
		TypeName:    req.TypeName,
		Description: req.Description,
	}

	// 创建字典类型
	if err := s.dictTypeRepo.Create(ctx, dictType); err != nil {
		log.Error().Err(err).Str("type_id", typeID).Str("type_code", req.TypeCode).Msg("创建字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "创建字典类型失败", err)
	}

	// 创建字典项
	for _, itemReq := range req.Items {
		itemID, err := idgen.GenerateUUID()
		if err != nil {
			log.Error().Err(err).Msg("生成字典项ID失败")
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
			log.Error().Err(err).Str("item_id", itemID).Str("type_id", typeID).Msg("创建字典项失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
		}
	}

	log.Info().Str("type_id", typeID).Str("type_code", req.TypeCode).Int("item_count", len(req.Items)).Msg("创建系统字典成功")
	return nil
}

// UpdateSystemDict 更新系统字典（超管专用）
func (s *DictService) UpdateSystemDict(ctx context.Context, typeCode string, req *dto.UpdateSystemDictRequest) (err error) {
	var oldDictType, newDictType *model.DictType

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleDict),
				audit.WithError(err),
			)
		} else if newDictType != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleDict),
				audit.WithResource(constants.ResourceTypeDict, newDictType.TypeID, newDictType.TypeName),
				audit.WithValue(oldDictType, newDictType),
			)
		}
	}()

	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取旧的系统字典类型
	oldDictType, err = s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("字典类型不存在")
			return xerr.ErrNotFound
		}
		log.Error().Err(err).Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("查询字典类型失败")
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
	if err := s.dictTypeRepo.Update(ctx, oldDictType.TypeID, updates); err != nil {
		log.Error().Err(err).Str("type_id", oldDictType.TypeID).Str("type_code", typeCode).Interface("updates", updates).Msg("更新字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "更新字典类型失败", err)
	}

	// 如果提供了字典项列表，则更新字典项
	if len(req.Items) > 0 {
		// 删除所有旧的系统字典项
		if err := s.dictItemRepo.DeleteByTypeID(ctx, oldDictType.TypeID); err != nil {
			log.Error().Err(err).Str("type_id", oldDictType.TypeID).Msg("删除旧字典项失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "删除旧字典项失败", err)
		}

		// 创建新的字典项
		for _, itemReq := range req.Items {
			itemID, err := idgen.GenerateUUID()
			if err != nil {
				log.Error().Err(err).Msg("生成字典项ID失败")
				return xerr.Wrap(xerr.ErrInternal.Code, "生成字典项ID失败", err)
			}

			dictItem := &model.DictItem{
				ItemID:   itemID,
				TypeID:   oldDictType.TypeID,
				TenantID: defaultTenantID,
				Label:    itemReq.Label,
				Value:    itemReq.Value,
				Sort:     int32(itemReq.Sort),
			}

			if err := s.dictItemRepo.Create(ctx, dictItem); err != nil {
				log.Error().Err(err).Str("item_id", itemID).Str("type_id", oldDictType.TypeID).Msg("创建字典项失败")
				return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
			}
		}
	}

	// 获取更新后的字典类型
	newDictType, err = s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		log.Error().Err(err).Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("获取更新后字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "获取更新后字典类型失败", err)
	}

	log.Info().Str("type_id", oldDictType.TypeID).Str("type_code", typeCode).Int("item_count", len(req.Items)).Msg("更新系统字典成功")
	return nil
}

// DeleteSystemDict 删除系统字典（超管专用）
func (s *DictService) DeleteSystemDict(ctx context.Context, typeCode string) (err error) {
	var dictType *model.DictType

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleDict),
				audit.WithError(err),
			)
		} else if dictType != nil {
			s.recorder.Log(ctx,
				audit.WithDelete(constants.ModuleDict),
				audit.WithResource(constants.ResourceTypeDict, dictType.TypeID, dictType.TypeName),
				audit.WithValue(dictType, nil),
			)
			log.Info().Str("type_id", dictType.TypeID).Str("type_code", typeCode).Msg("删除系统字典成功")
		}
	}()

	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err = s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("字典类型不存在")
			return xerr.ErrNotFound
		}
		log.Error().Err(err).Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("查询字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 删除字典项（先删除子记录）
	if err := s.dictItemRepo.DeleteByTypeID(ctx, dictType.TypeID); err != nil {
		log.Error().Err(err).Str("type_id", dictType.TypeID).Msg("删除字典项失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除字典项失败", err)
	}

	// 删除字典类型
	if err := s.dictTypeRepo.Delete(ctx, dictType.TypeID); err != nil {
		log.Error().Err(err).Str("type_id", dictType.TypeID).Msg("删除字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除字典类型失败", err)
	}

	return nil
}

// BatchDeleteSystemDicts 批量删除系统字典
func (s *DictService) BatchDeleteSystemDicts(ctx context.Context, typeIDs []string) (err error) {
	var dictTypeMap map[string]*model.DictType

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleDict),
				audit.WithError(err),
			)
		} else if len(dictTypeMap) > 0 {
			// 收集资源信息用于批量审计日志
			ids := make([]string, 0, len(dictTypeMap))
			names := make([]string, 0, len(dictTypeMap))
			for _, dictType := range dictTypeMap {
				ids = append(ids, dictType.TypeID)
				names = append(names, dictType.TypeName)
			}
			// 记录批量删除审计日志（单条日志记录所有资源）
			s.recorder.Log(ctx,
				audit.WithBatchDelete(constants.ModuleDict),
				audit.WithBatchResource(constants.ResourceTypeDict, ids, names),
				audit.WithValue(dictTypeMap, nil),
			)
			log.Info().Strs("type_ids", typeIDs).Int("count", len(typeIDs)).Msg("批量删除系统字典成功")
		}
	}()

	// 获取所有字典类型信息
	dictTypes, err := s.dictTypeRepo.GetByIDs(ctx, typeIDs)
	if err != nil {
		log.Error().Err(err).Strs("type_ids", typeIDs).Msg("查询字典类型信息失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型信息失败", err)
	}
	dictTypeMap = convert.ToMap(dictTypes, func(dt *model.DictType) string { return dt.TypeID })

	// 验证所有字典类型都存在
	if len(dictTypeMap) != len(typeIDs) {
		var missingIDs []string
		for _, id := range typeIDs {
			if _, exists := dictTypeMap[id]; !exists {
				missingIDs = append(missingIDs, id)
			}
		}
		log.Warn().Strs("missing_ids", missingIDs).Msg("部分字典类型不存在")
		return xerr.New(xerr.ErrNotFound.Code, "部分字典类型不存在")
	}

	// 批量删除字典项（先删除子记录）
	for _, typeID := range typeIDs {
		if err := s.dictItemRepo.DeleteByTypeID(ctx, typeID); err != nil {
			log.Error().Err(err).Str("type_id", typeID).Msg("删除字典项失败")
			return xerr.Wrap(xerr.ErrInternal.Code, "删除字典项失败", err)
		}
	}

	// 批量删除字典类型
	if err := s.dictTypeRepo.BatchDelete(ctx, typeIDs); err != nil {
		log.Error().Err(err).Strs("type_ids", typeIDs).Msg("批量删除字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "批量删除字典类型失败", err)
	}

	return nil
}

// ListDictTypes 获取字典类型列表（超管专用）
func (s *DictService) ListDictTypes(ctx context.Context, req *dto.ListDictTypesRequest) (*dto.ListDictTypesResponse, error) {
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
		List:     converter.ModelListToDictTypeInfoList(dictTypes),
	}, nil
}

// GetDictByCode 获取字典（合并系统+覆盖）
func (s *DictService) GetDictByCode(ctx context.Context, typeCode string) (*dto.DictInfo, error) {
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

// BatchUpdateDictItems 批量更新字典项
func (s *DictService) BatchUpdateDictItems(ctx context.Context, req *dto.BatchUpdateDictItemsRequest) error {
	currentTenantID := xcontext.GetTenantID(ctx)
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, req.TypeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("type_code", req.TypeCode).Str("tenant_id", defaultTenantID).Msg("字典类型不存在")
			return xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		log.Error().Err(err).Str("type_code", req.TypeCode).Str("tenant_id", defaultTenantID).Msg("查询字典类型失败")
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
		log.Error().Err(err).Str("type_id", dictType.TypeID).Msg("查询现有字典项失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询现有字典项失败", err)
	}
	if existingItems == nil {
		existingItems = []*model.DictItem{}
	}

	// 创建 value -> item 的映射（使用 value 作为唯一键）
	existingMap := make(map[string]*model.DictItem)
	for _, item := range existingItems {
		existingMap[item.Value] = item
	}

	// 记录已处理的 value
	processedValues := make(map[string]bool)

	// 处理请求中的字典项
	for _, itemReq := range req.Items {
		existingItem := existingMap[itemReq.Value]

		if existingItem != nil {
			// 更新现有项
			processedValues[itemReq.Value] = true
			updates := map[string]interface{}{
				"label":      itemReq.Label,
				"sort":       int32(itemReq.Sort),
				"updated_at": time.Now().UnixMilli(),
			}
			if err := s.dictItemRepo.Update(ctx, existingItem.ItemID, updates); err != nil {
				log.Error().Err(err).Str("item_id", existingItem.ItemID).Str("value", itemReq.Value).Msg("更新字典项失败")
				return xerr.Wrap(xerr.ErrInternal.Code, "更新字典项失败", err)
			}
		} else {
			// 创建新项（超管可以添加新项）
			itemID, err := idgen.GenerateUUID()
			if err != nil {
				log.Error().Err(err).Msg("生成字典项ID失败")
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
				log.Error().Err(err).Str("item_id", itemID).Str("type_id", dictType.TypeID).Msg("创建字典项失败")
				return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
			}
		}
	}

	// 删除没有在请求中的项（超管可以删除）
	for _, item := range existingItems {
		if !processedValues[item.Value] {
			if err := s.dictItemRepo.Delete(ctx, item.ItemID); err != nil {
				log.Error().Err(err).Str("item_id", item.ItemID).Str("value", item.Value).Msg("删除字典项失败")
				return xerr.Wrap(xerr.ErrInternal.Code, "删除字典项失败", err)
			}
		}
	}

	log.Info().Str("type_id", dictType.TypeID).Int("item_count", len(req.Items)).Msg("批量更新系统字典项成功")
	return nil
}

// batchUpdateTenantDictItems 租户批量更新字典项（只能覆盖，不能增删）
func (s *DictService) batchUpdateTenantDictItems(ctx context.Context, dictType *model.DictType, currentTenantID string, req *dto.BatchUpdateDictItemsRequest) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统默认的字典项（已按 sort 排序）
	systemItems, err := s.dictItemRepo.GetByTypeAndTenant(ctx, dictType.TypeID, defaultTenantID)
	if err != nil {
		log.Error().Err(err).Str("type_id", dictType.TypeID).Msg("查询系统字典项失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询系统字典项失败", err)
	}

	// 获取该租户现有的所有覆盖记录
	existingItems, err := s.dictItemRepo.GetByTypeAndTenant(ctx, dictType.TypeID, currentTenantID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error().Err(err).Str("type_id", dictType.TypeID).Str("tenant_id", currentTenantID).Msg("查询现有字典项失败")
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
			log.Warn().Int("request_count", len(req.Items)).Int("system_count", len(systemItems)).Msg("字典项数量超过系统默认项数量")
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
				log.Error().Err(err).Str("item_id", existing.ItemID).Str("value", systemItem.Value).Msg("更新字典项失败")
				return xerr.Wrap(xerr.ErrInternal.Code, "更新字典项失败", err)
			}
		} else {
			// 创建新的覆盖记录
			itemID, err := idgen.GenerateUUID()
			if err != nil {
				log.Error().Err(err).Msg("生成字典项ID失败")
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
				log.Error().Err(err).Str("item_id", itemID).Str("type_id", dictType.TypeID).Msg("创建字典项失败")
				return xerr.Wrap(xerr.ErrInternal.Code, "创建字典项失败", err)
			}
		}
	}

	log.Info().Str("type_id", dictType.TypeID).Str("tenant_id", currentTenantID).Int("item_count", len(req.Items)).Msg("批量更新租户字典项成功")
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
			log.Warn().Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("字典类型不存在")
			return xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		log.Error().Err(err).Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("查询字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 删除租户的覆盖记录（恢复系统默认）
	if err := s.dictItemRepo.DeleteByTypeAndValue(ctx, dictType.TypeID, currentTenantID, value); err != nil {
		log.Error().Err(err).Str("type_id", dictType.TypeID).Str("tenant_id", currentTenantID).Str("value", value).Msg("恢复系统默认值失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "恢复系统默认值失败", err)
	}

	log.Info().Str("type_id", dictType.TypeID).Str("type_code", typeCode).Str("value", value).Msg("恢复字典项系统默认值成功")
	return nil
}

// DeleteSystemDictItem 删除系统字典项（超管专用）
func (s *DictService) DeleteSystemDictItem(ctx context.Context, typeCode, value string) error {
	defaultTenantID := s.tenantCache.GetDefaultTenantID()

	// 获取系统字典类型
	dictType, err := s.dictTypeRepo.GetByCodeAndTenant(ctx, typeCode, defaultTenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("字典类型不存在")
			return xerr.New(xerr.ErrNotFound.Code, "字典不存在")
		}
		log.Error().Err(err).Str("type_code", typeCode).Str("tenant_id", defaultTenantID).Msg("查询字典类型失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询字典类型失败", err)
	}

	// 删除系统字典项（真正的删除）
	if err := s.dictItemRepo.DeleteByTypeAndValue(ctx, dictType.TypeID, defaultTenantID, value); err != nil {
		log.Error().Err(err).Str("type_id", dictType.TypeID).Str("tenant_id", defaultTenantID).Str("value", value).Msg("删除字典项失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "删除字典项失败", err)
	}

	log.Info().Str("type_id", dictType.TypeID).Str("type_code", typeCode).Str("value", value).Msg("删除系统字典项成功")
	return nil
}
