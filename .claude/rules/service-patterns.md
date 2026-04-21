# Service 层开发规范

## 规则 1：禁止循环中数据库操作

用 `IN` 查询或批量方法替代。

```go
// ❌ for _, id := range ids { s.repo.Update(ctx, id, updates) }
// ✅ s.repo.BatchUpdateStatus(ctx, ids, status)
```

## 规则 2：避免 N+1 查询

收集外键 ID → 批量查询 → converter 处理。

## 规则 3：Converter 接收 list，不接收 map

Converter 内部自行构建 map，调用方传 `[]*model.XXX`。

## 规则 4：禁止忽略错误

`_` 忽略 error 只允许在明确不影响的场景，且必须加注释说明。

## 规则 5：避免冗余查询

方法已返回的数据不要再查一遍。List 返回的 count 直接使用。

## 规则 6：详情页不需要冗余 count 字段

如果返回完整子列表，前端用 `list.length` 计算。

## 规则 7：批量操作前验证租户

批量操作需先获取记录、过滤出当前租户的 ID、再执行。

## 规则 8：事务中用 xerr.Wrap 包装错误

```go
// ❌ return err                          // 前端收到 pq: foreign_key_violation...
// ✅ return xerr.Wrap(xerr.ErrInternal.Code, "删除失败", err)
```

## 规则 9：Repository 使用 GORM Gen

```go
// ❌ tx.Table("resources").Create(resources)
// ✅ r.q.Resource.WithContext(ctx).Create(resources...)
```

## 规则 10：事务由外部 db 决定

Repo 方法直接用 `r.q`，不内部管理事务。Service 需要事务时用 `s.db.Transaction()`。

## 规则 11：事务边界在 Service 层

跨多 Repo 操作在 Service 层包事务。事务内用 `tx` 创建临时 Repo 或直接 `tx.Table()`。
