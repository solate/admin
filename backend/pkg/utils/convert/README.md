# convert 包

切片与 map 的泛型转换工具。

> **核心原则**：Go 标准库（`slices` / `maps` / `cmp`，go 1.21+）能做的，一律用标准库，不在本包重复封装。
> 本包只保留标准库**没有**、而业务里高频使用的 7 个函数。
> 新增函数前，先查下面「标准库已覆盖」表，确认官方真的没有再封装。

---

## 本包保留的函数（标准库无等价）

| 函数 | 作用 | 示例 |
|---|---|---|
| `ToMap` | 切片 → map（一对一，自选 key） | `convert.ToMap(users, func(u *model.User) string { return u.UserID })` |
| `ToMapPtr` | 切片 → 指针 map（避免值拷贝） | `convert.ToMapPtr(users, func(u model.User) string { return u.UserID })` |
| `Map` | 逐元素转换（提字段、model→dto） | `ids := convert.Map(users, func(u *model.User) string { return u.UserID })` |
| `GroupBy` | 切片 → map（一对多分组） | `convert.GroupBy(perms, func(p *model.RolePermission) string { return p.RoleID })` |
| `Unique` | 去重且**保持原顺序** | `convert.Unique(userIDs)` |
| `Filter` | 过滤，返回新切片（不改原切片） | `convert.Filter(users, func(u *model.User) bool { return u.Status == 1 })` |
| `Reduce` | 归约为单个值 | `convert.Reduce(nums, 0, func(acc, n int) int { return acc + n })` |

---

## 标准库已覆盖——不要再封装，直接用官方

需要 `import "slices"` / `import "maps"`。

### 切片操作（`slices`，go 1.21+）

| 需求 | 官方写法 | 说明 |
|---|---|---|
| 是否包含元素 | `slices.Contains(s, x)` | |
| 查元素下标 | `slices.Index(s, x)` | 找不到返回 -1 |
| 反转 | `slices.Reverse(s)` | **原地**修改，无返回值 |
| 最大值 | `slices.Max(s)` | 空切片会 panic |
| 最小值 | `slices.Min(s)` | 空切片会 panic |
| 相邻去重 | `slices.Compact(s)` | 只去**相邻**重复；要全量去重用本包 `Unique` |
| 分块 | `slices.Chunk(s, n)` | 返回迭代器，见下方「迭代器」 |
| 排序 | `slices.Sort(s)` | 原地；自定义比较用 `slices.SortFunc` |
| 是否相等 | `slices.Equal(a, b)` | |
| 克隆 | `slices.Clone(s)` | 浅拷贝 |

### map 操作（`maps`，go 1.21+）

| 需求 | 官方写法 | 说明 |
|---|---|---|
| 克隆 | `maps.Clone(m)` | 浅拷贝 |
| 合并 | `maps.Copy(dst, src)` | src 覆盖 dst |
| 是否相等 | `maps.Equal(a, b)` | |
| 删除满足条件的键 | `maps.DeleteFunc(m, fn)` | |

---

## 迭代器：map → 切片（go 1.23+ 的坑）

`maps.Keys(m)` / `maps.Values(m)` 返回的**不是切片**，是迭代器（`iter.Seq`）。
要拿到真正的切片，必须用 `slices.Collect` 收集一层：

```go
import (
    "maps"
    "slices"
)

// map 的所有 value 成切片
users := slices.Collect(maps.Values(userMap))   // []*model.User

// map 的所有 key 成切片
ids := slices.Collect(maps.Keys(userMap))        // []string

// slices.Chunk 同理，是迭代器，用 range 遍历
for chunk := range slices.Chunk(ids, 100) {
    repo.BatchQuery(ctx, chunk)                  // 每批最多 100 个
}
```

**记法**：`maps.Xxx` 只产出序列，`slices.Collect` 落成切片，组合使用。

---

## 方向对照（最易混）

| 方向 | 用什么 |
|---|---|
| 切片 → map（建索引） | `convert.ToMap` / `convert.GroupBy`（本包） |
| map → 切片（取值/取键） | `slices.Collect(maps.Values(m))` / `slices.Collect(maps.Keys(m))`（官方） |

---

**历史**：本包曾有 16 个函数，2026-07 精简为 7 个——删掉了标准库已有的
（`Contains`/`Reverse`/`Max`/`Min`/`Chunk`/`MapKeys`/`MapValues`）和重复实现（`MapToSlice`），
并去掉 `SliceXxx` 冗余前缀（包名 `convert` 已提供上下文）。
