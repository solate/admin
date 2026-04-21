# Swagger 配置规则

## 规则 1：ID 字段用 string

```go
// ❌ @Param role_id query int64 true "角色ID"
// ✅ @Param role_id query string true "角色ID"
```

## 规则 2：@Success 必须指定 data 类型

```go
// ❌ @Success 200 {object} response.Response
// ✅ @Success 200 {object} response.Response{data=dto.RoleInfo}
```

## 规则 3：DTO 字段必须包含 example

必填 example：ID、Name、Status、时间戳、枚举字段。不需要：`[]*XxxInfo` 数组、可选对象。

## 规则 4：Handler 参数类型与 DTO 一致

DTO 用 `string` 则 Swagger 注解也用 `string`。

## 规则 5：Enums 格式

```go
// ❌ Enums(A,B,C)
// ✅ Enums(A, B, C)  — 逗号后有空格
```

## 规则 6：时间戳参数注明单位

```go
// @Param start_date query int false "开始时间(毫秒级时间戳)"
```
