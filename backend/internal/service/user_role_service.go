package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// UserRoleService 用户角色服务
type UserRoleService struct {
	userRepo     *repository.UserRepo
	userRoleRepo *repository.UserRoleRepo
	roleRepo     *repository.RoleRepo
	tenantRepo   *repository.TenantRepo
	enforcer     *casbin.Enforcer
	recorder     *audit.Recorder
}

// NewUserRoleService 创建用户角色服务
func NewUserRoleService(userRepo *repository.UserRepo, userRoleRepo *repository.UserRoleRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, enforcer *casbin.Enforcer, recorder *audit.Recorder) *UserRoleService {
	return &UserRoleService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		tenantRepo:   tenantRepo,
		enforcer:     enforcer,
		recorder:     recorder,
	}
}

// AssignRoles 为用户分配角色（覆盖式）
func (s *UserRoleService) AssignRoles(ctx context.Context, userID string, req *dto.AssignRolesRequest) (err error) {
	var user *model.User

	defer func() {
		if err != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleRole),
				audit.WithError(err),
			)
		} else if user != nil {
			s.recorder.Log(ctx,
				audit.WithUpdate(constants.ModuleRole),
				audit.WithResource(constants.ResourceTypeUser, user.UserID, user.UserName),
				audit.WithValue(nil, map[string]interface{}{
					"role_codes": req.RoleCodes,
				}),
			)
		}
	}()

	// 获取用户信息
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 验证所有角色是否存在（从 default 租户查询）
	// 因为所有角色都存储在 default 租户中
	defaultTenants, err := s.tenantRepo.GetByCodes(ctx, []string{constants.DefaultTenantCode})
	if err != nil || len(defaultTenants) == 0 {
		log.Error().Err(err).Str("default_tenant_code", constants.DefaultTenantCode).Msg("查询 default 租户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询 default 租户失败", err)
	}
	defaultTenant := defaultTenants[0]

	roles, err := s.roleRepo.ListByCodesWithTenant(ctx, defaultTenant.TenantID, req.RoleCodes)
	if err != nil {
		log.Error().Err(err).Strs("role_codes", req.RoleCodes).Msg("查询角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	if len(roles) != len(req.RoleCodes) {
		log.Warn().Strs("role_codes", req.RoleCodes).Msg("部分角色不存在")
		return xerr.Wrap(xerr.ErrInvalidParams.Code, "部分角色不存在", nil)
	}

	// 分配角色（覆盖式）
	// 关键修改：统一使用 default 域
	if err := s.userRoleRepo.AssignRoles(ctx, user.UserName, req.RoleCodes, constants.DefaultTenantCode); err != nil {
		log.Error().Err(err).Str("user_id", userID).Strs("role_codes", req.RoleCodes).Msg("分配角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "分配角色失败", err)
	}

	log.Info().
		Str("user_id", userID).
		Str("username", user.UserName).
		Strs("role_codes", req.RoleCodes).
		Msg("分配用户角色成功")

	return nil
}

// GetUserRoles 获取用户的角色列表
func (s *UserRoleService) GetUserRoles(ctx context.Context, userID string) (*dto.UserRolesResponse, error) {
	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	tenantCode := xcontext.GetTenantCode(ctx)

	// 使用辅助方法获取用户角色
	roles, err := s.getUserRoles(ctx, user.UserName, tenantCode)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户角色失败", err)
	}

	return &dto.UserRolesResponse{
		UserID:   user.UserID,
		UserName: user.UserName,
		Roles:    converter.ModelListToRoleInfoList(roles),
	}, nil
}

// getUserRoles 获取用户的角色详情列表（辅助方法）
// 返回用户的完整角色信息，如果查询失败则返回空切片和错误
//
// 注意：由于角色分配时使用 default 域（g, username, role_code, default）
//
//	角色也存储在 default 租户中，因此查询角色详情时需要使用 default 租户ID
func (s *UserRoleService) getUserRoles(ctx context.Context, userName, tenantID string) ([]*model.Role, error) {
	// 获取用户角色编码列表
	// 注意：Casbin 策略存储格式：g, username, role_code, default
	// 所以这里使用 default 域查询角色编码
	roleCodes, err := s.userRoleRepo.GetUserRoles(ctx, userName, constants.DefaultTenantCode)
	if err != nil {
		log.Error().Err(err).Str("username", userName).Str("tenant_code", constants.DefaultTenantCode).Msg("查询用户角色编码失败")
		return nil, err
	}

	// 如果没有角色，直接返回空切片
	if len(roleCodes) == 0 {
		return []*model.Role{}, nil
	}

	// 获取 default 租户的租户ID（用于查询角色详情）
	defaultTenants, err := s.tenantRepo.GetByCodes(ctx, []string{constants.DefaultTenantCode})
	if err != nil || len(defaultTenants) == 0 {
		log.Error().Err(err).Str("default_tenant_code", constants.DefaultTenantCode).Msg("查询 default 租户失败")
		return nil, err
	}
	defaultTenant := defaultTenants[0]

	// 关键修改：使用 default 租户ID 查询角色详情
	// 因为所有角色都存储在 default 租户中
	roles, err := s.roleRepo.ListByCodesWithTenant(ctx, defaultTenant.TenantID, roleCodes)
	if err != nil {
		log.Error().Err(err).Strs("role_codes", roleCodes).Str("tenant_id", defaultTenant.TenantID).Msg("查询角色详情失败")
		return nil, err
	}

	return roles, nil
}

// ValidateAndAssignRoles 验证角色代码并分配（供其他服务复用）
// tenantID: 租户ID，用于查询租户代码（仅供日志记录使用）
// 返回角色详情列表和可能的错误
//
// 核心逻辑（简化版）：
// 1. 验证角色在 default 租户中是否存在（作为角色模板）
// 2. 为用户分配角色，统一使用 default 域（不创建租户专属角色）
//
// 注意：所有租户共享 default 租户的角色模板，权限验证时统一使用 default 域
func (s *UserRoleService) ValidateAndAssignRoles(ctx context.Context, userName string, roleCodes []string, tenantID string) ([]*model.Role, error) {
	// 如果没有角色代码，直接返回空
	if len(roleCodes) == 0 {
		return []*model.Role{}, nil
	}

	if tenantID == "" {
		log.Error().Msg("tenantID 不能为空")
		return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "tenantID 不能为空", nil)
	}

	// 查询租户信息（仅供日志记录）
	tenant, err := s.tenantRepo.GetByIDManual(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("查询租户信息失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询租户信息失败", err)
	}

	tenantCode := tenant.TenantCode
	isDefaultTenant := tenantCode == constants.DefaultTenantCode

	// 获取 default 租户ID（用于角色模板查询）
	defaultTenantID := ""
	if !isDefaultTenant {
		defaultTenants, err := s.tenantRepo.GetByCodes(ctx, []string{constants.DefaultTenantCode})
		if err != nil || len(defaultTenants) == 0 {
			log.Error().Err(err).Str("default_tenant_code", constants.DefaultTenantCode).Msg("查询 default 租户失败")
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询 default 租户失败", err)
		}
		defaultTenantID = defaultTenants[0].TenantID
	}

	var allRoles []*model.Role

	// 处理每个角色代码
	for _, roleCode := range roleCodes {
		var role *model.Role

		// 统一从 default 租户获取角色模板
		if isDefaultTenant {
			// default 租户：直接查询角色
			role, err = s.roleRepo.GetByCode(ctx, roleCode)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Warn().Str("role_code", roleCode).Msg("角色不存在")
					return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "角色不存在: "+roleCode, nil)
				}
				log.Error().Err(err).Str("role_code", roleCode).Msg("查询角色失败")
				return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
			}
		} else {
			// 非默认租户：从 default 租户获取角色模板（不创建租户专属角色）
			role, err = s.roleRepo.GetByCodeWithTenant(ctx, defaultTenantID, roleCode)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Warn().Str("role_code", roleCode).Msg("default 租户中不存在该角色模板")
					return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "角色模板不存在: "+roleCode, nil)
				}
				log.Error().Err(err).Str("role_code", roleCode).Msg("查询 default 租户角色模板失败")
				return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色模板失败", err)
			}

			log.Info().
				Str("role_code", roleCode).
				Str("tenant_code", tenantCode).
				Msg("使用 default 租户角色模板")
		}

		allRoles = append(allRoles, role)
	}

	// 分配角色到用户（g 策略：g, username, role_code, default）
	// 关键修改：统一使用 default 域，而不是用户的租户代码
	if err := s.userRoleRepo.AssignRoles(ctx, userName, roleCodes, constants.DefaultTenantCode); err != nil {
		log.Error().Err(err).Str("username", userName).Strs("role_codes", roleCodes).Msg("分配角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "分配角色失败", err)
	}

	log.Info().
		Str("username", userName).
		Strs("role_codes", roleCodes).
		Str("user_tenant", tenantCode).
		Str("auth_domain", constants.DefaultTenantCode).
		Msg("角色分配成功（使用 default 域）")

	return allRoles, nil
}

// GetUserRoleDetails 获取用户角色详情（供其他服务复用）
func (s *UserRoleService) GetUserRoleDetails(ctx context.Context, userName, tenantID string) ([]*model.Role, error) {
	if tenantID == "" {
		log.Error().Msg("tenantID 不能为空")
		return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "tenantID 不能为空", nil)
	}

	return s.getUserRoles(ctx, userName, tenantID)
}
