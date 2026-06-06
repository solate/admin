package role

import (
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/pkg/audit"
	"admin/pkg/constants"
	"admin/pkg/utils/idgen"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CreateRole 创建角色（支持继承父角色）
func (s *Service) CreateRole(ctx context.Context, req *dto.CreateRoleRequest) (resp *dto.RoleInfo, err error) {
	var role *model.Role

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleRole),
				audit.WithError(err),
			)
		} else if role != nil {
			s.recorder.Log(ctx,
				audit.WithCreate(constants.ModuleRole),
				audit.WithResource(constants.ResourceTypeRole, role.RoleID, role.Name),
				audit.WithValue(nil, role),
			)
		}
	}()

	tenantID := xcontext.GetTenantID(ctx)
	if tenantID == "" {
		return nil, xerr.ErrUnauthorized
	}

	// 检查角色编码是否已存在（租户内唯一）
	var exists bool
	exists, err = s.roleRepo.CheckExists(ctx, tenantID, req.RoleCode)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("检查角色编码是否存在失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "检查角色编码是否存在失败", err)
	}
	if exists {
		log.Warn().Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("角色编码已存在")
		return nil, xerr.ErrRoleCodeExists
	}

	// 如果有父角色，验证父角色属于 default 租户的角色模板
	var parentRoleCode *string
	if req.ParentRoleCode != nil {
		// 父角色必须在 default 租户中查找
		defaultTenantID, err := s.getDefaultTenantID(ctx)
		if err != nil {
			return nil, err
		}
		parentRole, err := s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, *req.ParentRoleCode)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("parent_role_code", *req.ParentRoleCode).Msg("父角色不存在或只能继承 default 租户的角色模板")
				return nil, xerr.New(xerr.ErrInvalidParams.Code, "父角色不存在或只能继承 default 租户的角色模板")
			}
			log.Error().Err(err).Str("parent_role_code", *req.ParentRoleCode).Msg("查询父角色失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询父角色失败", err)
		}
		parentRoleCode = &parentRole.RoleCode
	}

	// 生成角色ID
	var roleID string
	roleID, err = idgen.GenerateUUID()
	if err != nil {
		log.Error().Err(err).Msg("生成角色ID失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "生成角色ID失败", err)
	}

	// 构建角色模型，设置 ParentRoleID
	role = &model.Role{
		RoleID:      roleID,
		TenantID:    tenantID,
		RoleCode:    req.RoleCode,
		Name:        req.Name,
		Description: req.Description,
		Status:      int16(req.Status),
	}

	// 设置默认状态
	if role.Status == int16(constants.StatusZero) {
		role.Status = int16(constants.StatusEnabled) // 默认启用状态
	}

	// 如果有父角色，设置 parent_role_id
	if parentRoleCode != nil {
		// 通过 parent_role_code 查找父角色的 RoleID
		defaultTenantID, _ := s.getDefaultTenantID(ctx)
		parentRole, err := s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, *parentRoleCode)
		if err == nil && parentRole != nil {
			role.ParentRoleID = parentRole.RoleID
		}
	}

	// 创建角色
	if err := s.roleRepo.Create(ctx, role); err != nil {
		log.Error().Err(err).Str("role_id", roleID).Str("tenant_id", tenantID).Str("role_code", req.RoleCode).Msg("创建角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "创建角色失败", err)
	}

	return ModelToRoleInfoWithParent(role, parentRoleCode), nil
}
