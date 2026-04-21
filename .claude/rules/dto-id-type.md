# DTO ID 字段类型规范

## 规则
所有 ID 字段必须使用 `string` 类型，禁止使用 `int64`

## 适用范围
- 所有 DTO 文件（`internal/dto/*.go`）
- ID 字段类型：`UserID`, `RoleID`, `MenuID`, `TenantID`, `APIResourceID`, `DepartmentID` 等

## 代码示例

### ❌ 错误
```go
type AssignPermissionsRequest struct {
    RoleID int64 `json:"role_id"`
}
```

### ✅ 正确
```go
type AssignPermissionsRequest struct {
    RoleID string `json:"role_id"`
}
```

## 检查命令
```bash
grep -n "ID.*int64\|int64.*ID" internal/dto/*.go
```

## 修复记录
- `api_resource_dto.go`: 4 处 RoleID 字段 (int64 → string)
- `menu_dto.go`: 4 处 RoleID 字段 (int64 → string)
