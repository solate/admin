package constants

// 租户相关常量
const (
	// DefaultTenantCode 默认租户编码（超管租户，用于 Casbin domain）
	DefaultTenantCode = "default"
)

// 角色相关常量
const (
	// SuperAdmin 超级管理员角色code
	SuperAdmin = "super_admin"
	// Admin 租户管理员角色code
	Admin = "admin"
	// Auditor 审计员/监管员角色code
	Auditor = "auditor"
	// User 普通用户角色code
	User = "user"
)

// 权限类型常量
const (
	TypeMenu   = "MENU"   // 菜单权限
	TypeButton = "BUTTON" // 按钮权限
	TypeAPI    = "API"    // 接口权限
	TypeData   = "DATA"   // 数据权限
)

// 菜单状态常量
const (
	MenuStatusShow   = 1 // 显示
	MenuStatusHidden = 2 // 隐藏
)

// PermissionTypeText 权限类型中文描述映射
var PermissionTypeText = map[string]string{
	TypeMenu:   "菜单",
	TypeButton: "按钮",
	TypeAPI:    "接口",
	TypeData:   "数据",
}

// MenuStatusText 菜单状态中文描述映射
var MenuStatusText = map[int16]string{
	MenuStatusShow:   "显示",
	MenuStatusHidden: "隐藏",
}

// RSAKey RSA私钥（用于密码解密）
// 对应公钥应提供给前端用于加密
const RSAKey = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQC+1GvoPXuNkVyR
yPmC+Qw3AmlYjf+ol6cyB6jM++sB9U7vlY6owOxzLeK/R2JAiY61nhcO90vJWgmz
CeE9uq6704xrf2ONCCll5DNd3zz3WhpknTHuOXS2/3e2u26eiFR94ZHMctnD34jo
LA6VL+qHi3mTqZ6RlDeES26IvbKZB0NhsbZ2cmaN7oxBVoJzHipHyf0/2YWJQQJS
IukIJbl9RNc4IDhdCPeyfK7SJGUbZIjsiraBs9fn1Sz89jpVjeMHXnHQDGMver7c
XI8StrmdSbd5zVoxc8ZG5r1TpyQPEeyb/uE8zxwsCKusGMMcB37J5fFvUWMSJ4HA
kSBoQO5/AgMBAAECggEABMxyqjRhlv3Axim3nIOGuxtkasWnWCX4Hlny9LShBDuW
8I9iNvwi9gKBYS36WoUbAZYoHkg5r6aD9+yXrWW0XyTCszFQ34sE/3rtj769Wbr6
Tu1lBAiN1sw1xnKQJYxoE4JImEuLDlHgr3XsJ/Q8gYwQUpZBVofTnZAIB4g9pXtu
JKmU4OkyLQaQqZwvIjtWtVjNSeC+VGqh3p5FP13jv2kBlzFksrtYYY/yDB6Uj4c8
Grj8RwOEmz02txsUU3TPbcYSiLjjUwcK9pgHic3uPMSv4wy6Mvm6ZA75XWgBOsrd
Lg6l1tsz8Fz4utnfRbVIdsrQ9z3LHLTTnnOpKf3PAQKBgQDpAu6wtVEgnxzGyFup
Weciwk9zP3qXTzD68rNy4EuUo0z6crbbo++aHJ10l6GKtTjXnvBPfMsoi9zqOCEQ
8dpNMe7e+X2mMCGuJz8e83OOGedroGi8BWoNRyysRZO3BdXzPja48Dx2wRgqQ5Km
pBrDuORctPJv9c98rzxXwv/plwKBgQDRqB/8qw9lRu/0qy2vF5aSmXAwHAQuC+69
2697flrnIozVSK0bdfxMqvKM2/eQITtInFixlRiWsq1k5N1T3o3bBko+hT/Rg/N5
doOhi8WmH4MgZRp7bAyeAMHsVfDXea+NG1K8kgDQW+Kqcg+5Hb3KlrXQDxGJF2Hp
N/ejuoovWQKBgQCi4X3g4J5ZY2BGRIBunX3I+nN3aIRViPIAOe/e+ZNbz9tbpxzT
5ID1BdO7UNOHlq6pa10o819AdKR0xc+3fJjRJXqJO3Xt2e9xQdYJ2LyKNOlkfrk3
1cEQjxRXSDu90MKCSpcOKEDb8pbl1F6LRmO/NVvMwmBGi1oDGqvf3Vvu+QKBgQCL
MozKPOij3U1DrMNQFOErxCPwTSmZSOLhuxHvdBz2iMHoebA1I0i3vmf7ja/4SZgK
xYM9pDgHFep5qloobQLSAIMar22HtYvZgQ40G5DGkvWEdJv4hex6mxYly4l0Bp6/
mPx9ppJTxC3h7Ijz5wMzloxv7xE9bADdzwLj+d31QQKBgQDkTwtclVg90MLMOyx+
9XNbucBnnKMQvMd753n1y/2whhgqdgffMGgRALIhkHLsGkM4irACJytFU6T5mq7n
xqto3Ux5up3yALGP0jfXfcV5IjfbXkkgu+w0xMaWvXx5hgiftY/SBeZz/ujUJ2XG
aSyRaef1MpYPbhjZmFIY0+dI9w==
-----END PRIVATE KEY-----`
