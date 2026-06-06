# Step 14: Swagger 文档 + 集成测试

## 目标

为所有 API 生成 Swagger 文档，编写关键路径的集成测试。

## 前置条件

- Step 01-13 全部完成

## 文件清单

```
backend-rbac/
├── docs/
│   ├── swagger.yaml           # 生成的 Swagger 文档
│   └── swagger.json
├── Makefile                   # 追加 swagger 和 test 命令
└── tests/
    └── integration/
        ├── auth_test.go       # 登录/刷新/登出测试
        ├── user_test.go       # 用户 CRUD 测试
        ├── role_test.go       # 角色 + 权限测试
        └── tenant_test.go     # 租户管理测试
```

## 实现细节

### 1. Swagger 注解规范

每个 handler 方法上的 Swagger 注解遵循以下规则：

```go
// @Summary     创建用户
// @Description 管理员创建新用户
// @Tags        用户管理
// @Accept      json
// @Produce     json
// @Param       body body dto.CreateUserRequest true "用户信息"
// @Success     200 {object} response.Response{data=dto.UserInfo}
// @Failure     400 {object} response.Response
// @Failure     401 {object} response.Response
// @Router      /api/v1/users [post]
// @Security    BearerAuth
```

**注解规则**（来自 swagger-rules.md）：
- ID 字段用 `string`，不用 `int64`
- `@Success` 必须指定 `data` 类型：`{object} response.Response{data=dto.UserInfo}`
- DTO 字段必须包含 `example`（必填字段）
- Enums 逗号后有空格：`Enums(A, B, C)`
- 时间戳参数注明单位：`(毫秒级时间戳)`

### 2. Makefile 追加

```makefile
swagger:
	swag init -g cmd/server/main.go -o docs/ --parseDependency --parseInternal

test:
	go test ./... -v -count=1

test-integration:
	go test ./tests/integration/... -v -count=1 -tags=integration
```

### 3. 集成测试策略

```go
// tests/integration/auth_test.go
func TestLogin(t *testing.T) {
    // 1. 准备测试数据（用事务，测试完回滚）
    // 2. 调用 POST /api/v1/auth/login
    // 3. 断言响应码、Token 结构
}

func TestLogin_WrongPassword(t *testing.T) {
    // 断言 401
}

func TestFullAuthFlow(t *testing.T) {
    // 登录 → 用 token 调接口 → 刷新 → 登出 → token 失效
}
```

**测试基础设施**：

```go
// tests/integration/helper.go
func SetupTestServer(t *testing.T) *TestServer {
    // 1. 加载测试配置
    // 2. 连接测试数据库
    // 3. 运行迁移
    // 4. 插入种子数据
    // 5. 启动 Gin（httptest.NewServer）
    return &TestServer{...}
}

func (s *TestServer) TearDown() {
    // 清理测试数据
    // 关闭服务器
}
```

### 4. 关键测试用例

| 测试场景 | 验证点 |
|----------|--------|
| 登录成功 | 返回 Token Pair，包含正确 Claims |
| 登录失败 | 错误密码返回 401 |
| Token 有效 | 用 access_token 调接口返回 200 |
| Token 过期 | 返回 401 |
| 租户隔离 | 用户 A 看不到租户 B 的数据 |
| 超管绕行 | super_admin 访问所有接口都通过 |
| RBAC 拦截 | 无权限的角色访问被拒绝 |
| 角色继承 | 子角色自动拥有父角色的权限 |
| 数据权限 | data_scope=5 只能看到自己的数据 |
| 批量删除 | 多个 ID 批量软删除成功 |

## 验收标准

- [ ] `make swagger` 生成完整的 swagger.yaml
- [ ] Swagger UI 能正确显示所有 API
- [ ] 所有 DTO 的 example 字段完整
- [ ] `make test` 所有单元测试通过
- [ ] `make test-integration` 关键路径测试通过
- [ ] 测试覆盖：登录/鉴权/租户隔离/RBAC/数据权限
- [ ] 测试不依赖外部环境（用测试数据库 + 自动迁移）

## AI 协作提示

```
请按 step-14-swagger-test.md 完成 Swagger 文档和集成测试。
1. 先给所有 handler 方法添加 Swagger 注解
2. 运行 make swagger 生成文档
3. 编写 auth_test.go 作为第一个集成测试
4. 确保测试用 httptest，不需要真实启动服务器
```
