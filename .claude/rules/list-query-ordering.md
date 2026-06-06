# 列表查询排序规范

## 规则
所有分页列表查询必须使用**双排序字段**：`created_at DESC` + 主键 `DESC`，确保排序确定性。

## 原因
- 批量导入的数据可能共享相同的 `created_at` 时间戳
- PostgreSQL 对相同排序值的记录不保证顺序
- UPDATE 后 MVCC 机制改变行物理位置，导致相同时间戳的记录相对顺序跳变

## 代码示例

### ❌ 错误 — 单排序字段
```go
words, err := query.Order(r.q.SystemSensitiveWord.CreatedAt.Desc()).Offset(offset).Limit(limit).Find()
```

### ✅ 正确 — 主排序 + 主键兜底
```go
words, err := query.Order(r.q.SystemSensitiveWord.CreatedAt.Desc()).Order(r.q.SystemSensitiveWord.WordID.Desc()).Offset(offset).Limit(limit).Find()
```

## 写法
使用链式 `.Order().Order()` 风格（非多参数），和项目现有 role_repo 保持一致。

## 不分页查询同样适用
`GetAll`、`ListAll` 等不分页方法如果有 Order，也需要双排序字段。

## 例外
如果排序字段是唯一的（如单条查询 `.First()`），不需要第二排序字段。

## 检查命令
```bash
# 检查是否存在单排序字段的列表查询
grep -rn "\.Order(.*\.Desc())\.Offset(" internal/repository/ | grep -v "\.Order(.*\.Desc())\.Order("
```

---

**最后更新**：2026-04-22
