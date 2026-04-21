package dict

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
)

// CreateSystemDict 创建系统字典（超管专用）
func (s *Service) CreateSystemDict(ctx context.Context, req *dto.CreateSystemDictRequest) (err error) {
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
