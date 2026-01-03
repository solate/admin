package casbin

// DefaultModel 默认的RBAC模型
//
// Domain 使用规范：
// 1. default domain：用于存储角色模板的权限（超管在 default 租户创建的角色）
// 2. 租户 domain（如 tenant-a, tenant-b）：用于存储租户特定的用户-角色绑定和权限
//
// 策略存储规则：
// - p, role_code, domain, resource, action：权限策略
//   - 模板角色的权限存储在 default domain
//   - 租户角色继承模板时，不再物理复制权限，而是通过应用层实时查询
// - g, username, role_code, tenant_code：用户-角色绑定（必须带租户）
// - g2, child_role, parent_role：角色继承（不需要 domain，用于模板继承）
//
// 重要：由于 matcher 要求 r.dom == p.dom，继承角色的跨 domain 权限通过应用层处理
// 见 UserMenuService.getMenuPermissionsForRoles 中的实现
func DefaultModel() string {
	return `
[request_definition]
r = sub, dom, obj, act # 用户, 租户, 资源, 操作

[policy_definition]
p = sub, dom, obj, act # 角色, 租户, 资源, 操作

[role_definition]
g = _, _, _    # 用户-角色-租户关系 (user, role, domain)
g2 = _, _      # 角色继承关系 (父角色, 子角色)

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)  # 支持操作通配符

`
}
