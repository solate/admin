package dict

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/utils/idgen"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// BatchUpdateDictItems 批量更新字典项
func (s *Service) BatchUpdateDictItems(ctx context.Context, req *dto.BatchUpdateDictItemsRequest) error {
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

// ResetDictItem 恢复字典项系统默认值
func (s *Service) ResetDictItem(ctx context.Context, typeCode, value string) error {
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

// batchUpdateSystemDictItems 超管批量更新系统字典项（可以增删改）
func (s *Service) batchUpdateSystemDictItems(ctx context.Context, dictType *model.DictType, req *dto.BatchUpdateDictItemsRequest) error {
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
func (s *Service) batchUpdateTenantDictItems(ctx context.Context, dictType *model.DictType, currentTenantID string, req *dto.BatchUpdateDictItemsRequest) error {
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
