package casbin

// DefaultModel 默认的RBAC模型
func DefaultModel() string {
	return `
[request_definition]
r = sub, dom, obj, act # 用户, 租户, 资源, 操作

[policy_definition]
p = sub, dom, obj, act # 角色, 租户, 资源, 操作

[role_definition]
g = _, _, _   # 用户, 角色, 租户

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)  # 支持操作通配符

`
}
