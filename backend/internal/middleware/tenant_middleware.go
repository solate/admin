package middleware

import (
	"admin/pkg/constants"
	"admin/pkg/xcontext"

	"github.com/gin-gonic/gin"
)

// TenantSkipMiddleware 跳过租户检查的中间件
//
// 职责：
//  1. 超管（super_admin 角色）：跳过租户检查，可以查看所有租户数据
//  2. 审核员（auditor 角色）：跳过租户检查，可以查看所有租户数据
//  3. 普通用户：不做任何处理，保持原有的租户检查
//
// 执行顺序：CasbinMiddleware（权限校验） → SkipTenantCheck（租户检查跳过） → AuditMiddleware
//
// 工作原理：
//   - 能执行到这里说明已经通过了 CasbinMiddleware 的权限验证
//   - 超管和审核员在 Casbin 中配置了可访问的接口
//   - 这些接口允许跨租户查看数据，所以跳过租户检查
//   - 后续的数据库查询将不再自动添加租户条件
//
// 使用场景：
//   - 登录、注册、获取验证码等无需预先知道租户信息的接口（对所有用户）
//   - 需要跨租户查看数据的接口（仅超管和审核员，通过 Casbin 验证）
//
// 注意：
//   - 正常的业务接口都应该经过租户检查，确保数据隔离
//   - 该中间件在认证路由组上使用，配合 Casbin 实现权限控制
func TenantSkipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// 超管或审核员跳过租户检查
		if xcontext.HasRole(ctx, constants.SuperAdmin) || xcontext.HasRole(ctx, constants.Auditor) {
			ctx = xcontext.SkipTenantCheck(ctx)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
	}
}
