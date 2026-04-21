package service

import (
	"admin/internal/converter"
	"admin/internal/dal/model"
	"admin/internal/dto"
	"admin/internal/repository"
	"admin/pkg/audit"
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
	recorder     *audit.Recorder
}

// NewUserRoleService 创建用户角色服务
func NewUserRoleService(userRepo *repository.UserRepo, userRoleRepo *repository.UserRoleRepo, roleRepo *repository.RoleRepo, tenantRepo *repository.TenantRepo, recorder *audit.Recorder) *UserRoleService {
	return &UserRoleService{
		userRepo:     userRepo,
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
		tenantRepo:   tenantRepo,
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

	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 验证所有角色是否存在
	tenantID := xcontext.GetTenantID(ctx)
	roles, err := s.roleRepo.ListByIDs(ctx, req.RoleCodes)
	if err != nil {
		log.Error().Err(err).Strs("role_ids", req.RoleCodes).Msg("查询角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
	}

	if len(roles) != len(req.RoleCodes) {
		log.Warn().Strs("role_ids", req.RoleCodes).Msg("部分角色不存在")
		return xerr.Wrap(xerr.ErrInvalidParams.Code, "部分角色不存在", nil)
	}

	// 使用角色ID分配
	roleIDs := make([]string, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.RoleID
	}

	if err := s.userRoleRepo.AssignRoles(ctx, user.UserID, roleIDs, tenantID); err != nil {
		log.Error().Err(err).Str("user_id", userID).Strs("role_ids", roleIDs).Msg("分配角色失败")
		return xerr.Wrap(xerr.ErrInternal.Code, "分配角色失败", err)
	}

	log.Info().
		Str("user_id", userID).
		Str("username", user.UserName).
		Strs("role_ids", req.RoleCodes).
		Msg("分配用户角色成功")

	return nil
}

// GetUserRoles 获取用户的角色列表
func (s *UserRoleService) GetUserRoles(ctx context.Context, userID string) (*dto.UserRolesResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Warn().Str("user_id", userID).Msg("用户不存在")
			return nil, xerr.ErrUserNotFound
		}
		log.Error().Err(err).Str("user_id", userID).Msg("查询用户失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	tenantID := xcontext.GetTenantID(ctx)
	roles, err := s.getUserRoles(ctx, user.UserID, tenantID)
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

// getUserRoles 获取用户的角色详情列表
func (s *UserRoleService) getUserRoles(ctx context.Context, userID, tenantID string) ([]*model.Role, error) {
	roleIDs, err := s.userRoleRepo.GetUserRoleIDs(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) == 0 {
		return []*model.Role{}, nil
	}

	// 使用跨租户查询获取角色详情
	return s.roleRepo.GetByIDs(ctx, roleIDs)
}

// ValidateAndAssignRoles 验证角色代码并分配
func (s *UserRoleService) ValidateAndAssignRoles(ctx context.Context, userID string, roleCodes []string, tenantID string) ([]*model.Role, error) {
	if len(roleCodes) == 0 {
		return []*model.Role{}, nil
	}

	if tenantID == "" {
		log.Error().Msg("tenantID 不能为空")
		return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "tenantID 不能为空", nil)
	}

	// 查找用户
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询用户失败", err)
	}

	// 验证角色
	var allRoles []*model.Role
	roleIDs := make([]string, 0, len(roleCodes))

	for _, roleCode := range roleCodes {
		role, err := s.roleRepo.GetByCode(ctx, roleCode)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Warn().Str("role_code", roleCode).Msg("角色不存在")
				return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "角色不存在: "+roleCode, nil)
			}
			return nil, xerr.Wrap(xerr.ErrInternal.Code, "查询角色失败", err)
		}
		allRoles = append(allRoles, role)
		roleIDs = append(roleIDs, role.RoleID)
	}

	// 使用角色ID分配
	if err := s.userRoleRepo.AssignRoles(ctx, user.UserID, roleIDs, tenantID); err != nil {
		log.Error().Err(err).Str("user_id", userID).Strs("role_codes", roleCodes).Msg("分配角色失败")
		return nil, xerr.Wrap(xerr.ErrInternal.Code, "分配角色失败", err)
	}

	return allRoles, nil
}

// GetUserRoleDetails 获取用户角色详情
func (s *UserRoleService) GetUserRoleDetails(ctx context.Context, userID, tenantID string) ([]*model.Role, error) {
	if tenantID == "" {
		log.Error().Msg("tenantID 不能为空")
		return nil, xerr.Wrap(xerr.ErrInvalidParams.Code, "tenantID 不能为空", nil)
	}

	return s.getUserRoles(ctx, userID, tenantID)
}
