package xerr

// 预定义错误
var (
	// 状态码 200
	ErrSuccess = New(200, "成功")

	// 通用错误 1000-1999
	ErrInternal        = New(1000, "内部服务器错误")
	ErrBadRequest      = New(1001, "请求参数错误")
	ErrNotFound        = New(1002, "资源不存在")
	ErrUnauthorized    = New(1003, "未授权")
	ErrForbidden       = New(1004, "无权限访问")
	ErrConflict        = New(1005, "资源冲突")
	ErrTooManyRequests = New(1006, "请求过于频繁")

	// 数据库错误 2000-2099
	ErrRecordNotFound  = New(2000, "未找到该记录")
	ErrDbRecordExist   = New(2001, "数据库记录已存在")
	ErrUnmarshalFailed = New(2002, "数据序列化错误")
	ErrQueryError      = New(2003, "查询错误")
	ErrUpdateError     = New(2004, "更新错误")
	ErrCreateError     = New(2005, "保存错误")

	// 认证错误 2100-2199
	ErrInvalidCredentials = New(2100, "用户名或密码错误")
	ErrTokenExpired       = New(2101, "Token已过期")
	ErrTokenInvalid       = New(2102, "Token无效")
	ErrUserNotFound       = New(2103, "用户不存在")
	ErrUserExists         = New(2104, "用户已存在")
)
