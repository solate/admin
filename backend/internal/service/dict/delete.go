package dict

import (
	"admin/internal/dal/model"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/convert"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// DeleteSystemDict 删除系统字典（超管专用）
func (s *Service) DeleteSystemDict(ctx context.Context, typeCode string) (err error) {
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
func (s *Service) BatchDeleteSystemDicts(ctx context.Context, typeIDs []string) (err error) {
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

// DeleteSystemDictItem 删除系统字典项（超管专用）
func (s *Service) DeleteSystemDictItem(ctx context.Context, typeCode, value string) error {
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
