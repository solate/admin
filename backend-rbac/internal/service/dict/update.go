package dict

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/xerr"
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UpdateSystemDict 更新系统字典（超管专用）
func (s *Service) UpdateSystemDict(ctx context.Context, typeCode string, req *dto.UpdateSystemDictRequest) (err error) {
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
