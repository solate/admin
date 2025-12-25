package xerr

// 预定义错误
var (
	// 状态码 200
	ErrSuccess = New(200, "成功")

	// 通用错误 1000-1999
	ErrInternal        = New(1000, "内部服务器错误")
	ErrInvalidParams   = New(1001, "请求参数错误")
	ErrNotFound        = New(1002, "资源不存在")
	ErrUnauthorized    = New(1003, "未授权")
	ErrForbidden       = New(1004, "无权限访问")
	ErrConflict        = New(1005, "资源冲突")
	ErrTooManyRequests = New(1006, "请求过于频繁")
	ErrInvalidateParam = New(1007, "参数校验错误")

	// 数据库错误 2000-2099
	ErrRecordNotFound  = New(2000, "未找到该记录")
	ErrDbRecordExist   = New(2001, "数据库记录已存在")
	ErrUnmarshalFailed = New(2002, "数据序列化错误")
	ErrQueryError      = New(2003, "查询错误")
	ErrUpdateError     = New(2004, "更新错误")
	ErrCreateError     = New(2005, "保存错误")

	// 认证错误 2100-2199
	ErrInvalidCredentials     = New(2100, "用户名或密码错误")
	ErrTokenExpired           = New(2101, "Token已过期")
	ErrTokenInvalid           = New(2102, "Token无效")
	ErrUserNotFound           = New(2103, "用户不存在")
	ErrUserExists             = New(2104, "用户已存在")
	ErrCaptchaInvalid         = New(2105, "验证码错误")
	ErrUserDisabled           = New(2106, "用户已禁用")
	ErrUserHasMultipleTenants = New(2107, "用户关联多个租户")
	ErrUserNoTenants          = New(2108, "用户未关联任何租户")
	ErrUserTenantAccessDenied = New(2109, "用户无该租户访问权限")
	ErrUserNoRoles            = New(2110, "用户在租户中无任何角色")

	// 角色错误 2200-2299
	ErrRoleNotFound   = New(2200, "角色不存在")
	ErrRoleExists     = New(2201, "角色已存在")
	ErrRoleCodeExists = New(2202, "角色编码已存在")
	ErrRoleInUse      = New(2203, "角色正在使用中")

	// 菜单错误 2300-2399
	ErrMenuNotFound        = New(2300, "菜单不存在")
	ErrMenuExists          = New(2301, "菜单已存在")
	ErrMenuCodeExists      = New(2302, "菜单编码已存在")
	ErrMenuInUse           = New(2303, "菜单正在使用中")
	ErrMenuHasChildren     = New(2304, "菜单下有子菜单，无法删除")
	ErrMenuInvalidParent   = New(2305, "父菜单无效")
	ErrMenuCannotMoveToSelf = New(2306, "不能将菜单移动到自己或其子菜单下")
)
