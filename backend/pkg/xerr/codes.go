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
	ErrEmailOrPhoneExists     = New(2105, "邮箱或手机号已存在")
	ErrCaptchaInvalid         = New(2106, "验证码错误")
	ErrUserDisabled           = New(2107, "用户已禁用")
	ErrUserHasMultipleTenants = New(2108, "用户关联多个租户")
	ErrUserNoTenants          = New(2109, "用户未关联任何租户")
	ErrUserTenantAccessDenied = New(2110, "用户无该租户访问权限")
	ErrUserNoRoles            = New(2111, "用户在租户中无任何角色")

	// 租户错误 2200-2299
	ErrTenantCodeRequired = New(2200, "租户编码不能为空")
	ErrTenantNotFound     = New(2201, "租户不存在")
	ErrTenantExists       = New(2202, "租户已存在")
	ErrTenantCodeExists   = New(2203, "租户编码已存在")
	ErrTenantDisabled     = New(2204, "租户已禁用")

	// 角色错误 2300-2399
	ErrRoleNotFound   = New(2300, "角色不存在")
	ErrRoleExists     = New(2301, "角色已存在")
	ErrRoleCodeExists = New(2302, "角色编码已存在")
	ErrRoleInUse      = New(2303, "角色正在使用中")

	// 菜单错误 2400-2499
	ErrMenuNotFound         = New(2400, "菜单不存在")
	ErrMenuExists           = New(2401, "菜单已存在")
	ErrMenuCodeExists       = New(2402, "菜单编码已存在")
	ErrMenuInUse            = New(2403, "菜单正在使用中")
	ErrMenuHasChildren      = New(2404, "菜单下有子菜单，无法删除")
	ErrMenuInvalidParent    = New(2405, "父菜单无效")
	ErrMenuCannotMoveToSelf = New(2406, "不能将菜单移动到自己或其子菜单下")

	// 部门错误 2500-2599
	ErrDeptNotFound       = New(2500, "部门不存在")
	ErrDeptExists         = New(2501, "部门已存在")
	ErrDeptHasChildren    = New(2502, "部门下有子部门，无法删除")
	ErrDeptHasUsers       = New(2503, "部门下有用户，无法删除")
	ErrParentDeptNotFound = New(2504, "父部门不存在")
	ErrInvalidParentDept  = New(2505, "父部门无效")

	// 岗位错误 2600-2699
	ErrPositionNotFound   = New(2600, "岗位不存在")
	ErrPositionExists     = New(2601, "岗位已存在")
	ErrPositionCodeExists = New(2602, "岗位编码已存在")
	ErrPositionInUse      = New(2603, "岗位正在使用中")
)
