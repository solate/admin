package middleware

import (
	"admin/pkg/database"

	"github.com/gin-gonic/gin"
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
