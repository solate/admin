package middleware

import (
	"admin/internal/repository"
	"admin/pkg/database"
	"admin/pkg/response"
	"admin/pkg/xcontext"
	"admin/pkg/xerr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SkipTenantCheck 跳过租户检查的中间件
//
// 说明：
// - 某些接口（如登录、注册）在处理时需要跨租户查询，此时需要跳过 GORM 的租户拦截器
// - 该中间件会在 context 中设置跳过标记，后续的数据库查询将不再自动添加租户条件
// - 使用场景：登录、注册、获取验证码等无需预先知道租户信息的接口
//
// 注意：
// - 该中间件应谨慎使用，仅在明确需要跨租户查询的接口上使用
// - 正常的业务接口都应该经过租户检查，确保数据隔离
func SkipTenantCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := database.SkipTenantCheck(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func TenantFromCode(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantCode := c.Param("tenant_code")
		if tenantCode == "" {
			response.Error(c, xerr.New(xerr.ErrInvalidParams.Code, "租户编码不能为空"))
			c.Abort()
			return
		}

		tenant, err := repository.NewTenantRepo(db).GetByCodeManual(c.Request.Context(), tenantCode)
		if err != nil {
			response.Error(c, xerr.New(xerr.ErrNotFound.Code, "租户不存在"))
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = database.WithTenantID(ctx, tenant.TenantID)
		ctx = xcontext.SetTenantID(ctx, tenant.TenantID)
		ctx = xcontext.SetTenantCode(ctx, tenant.TenantCode)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
