package middleware

import (
	"admin/pkg/casbin"
	"admin/pkg/constants"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// CasbinMiddleware Casbin权限中间件
// 职责：
//  1. 超管（super_admin 角色）：跳过权限检查（租户检查由 TenantSkipMiddleware 处理）
//  2. 普通用户：通过 Casbin 验证 (username, default, path, method)
//
// 执行顺序：CasbinMiddleware（权限校验） → TenantSkipMiddleware（租户检查跳过）
//
// 注意：所有租户的用户统一使用 default 域进行权限验证
//
//	因为角色分配时使用的是 default 域：g, username, role_code, default
func CasbinMiddleware(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 超管跳过权限检查
		if xcontext.HasRole(ctx, constants.SuperAdmin) {
			c.Next()
			return
		}

		userName := xcontext.GetUserName(ctx)
		tenantCode := xcontext.GetTenantCode(ctx)

		if userName == "" || tenantCode == "" {
			response.Error(c, xerr.ErrUnauthorized)
			c.Abort()
			return
		}

		// Object: Request Path
		obj := c.Request.URL.Path
		// Action: Request Method
		act := c.Request.Method

		// 关键修改：统一使用 default 域进行权限验证
		// 因为角色分配时使用的是：g, username, role_code, default
		// 权限策略也存储在 default 域：p, role_code, default, resource, action
		authDomain := constants.DefaultTenantCode

		log.Debug().
			Str("sub", userName).
			Str("user_tenant", tenantCode).
			Str("auth_domain", authDomain).
			Str("obj", obj).
			Str("act", act).
			Msg("[CasbinMiddleware] 开始权限检查")

		ok, err := enforcer.Enforce(userName, authDomain, obj, act)
		if err != nil {
			log.Error().Err(err).
				Str("sub", userName).
				Str("auth_domain", authDomain).
				Str("obj", obj).
				Str("act", act).
				Msg("[CasbinMiddleware] 权限检查失败")
			response.Error(c, xerr.ErrInternal)
			c.Abort()
			return
		}

		if !ok {
			log.Warn().
				Str("sub", userName).
				Str("user_tenant", tenantCode).
				Str("auth_domain", authDomain).
				Str("obj", obj).
				Str("act", act).
				Msg("[CasbinMiddleware] 权限不足")
			response.Error(c, xerr.ErrForbidden)
			c.Abort()
			return
		}

		log.Debug().
			Str("sub", userName).
			Str("obj", obj).
			Msg("[CasbinMiddleware] 权限检查通过")

		c.Next()
	}
}
