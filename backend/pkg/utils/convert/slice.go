// Package convert 提供切片与 map 的泛型转换工具。
//
// 只保留 Go 标准库（slices / maps / cmp）没有、而业务里高频使用的函数。
// 标准库已覆盖的场景请直接使用标准库：
//
//	是否包含   slices.Contains
//	反转       slices.Reverse
//	最大/最小  slices.Max / slices.Min
//	相邻去重   slices.Compact
//	分块       slices.Chunk
//	map 的键   slices.Collect(maps.Keys(m))
//	map 的值   slices.Collect(maps.Values(m))
package convert

// ToMap 将切片转换为 map，keyFunc 从元素提取键。
// 键重复时后者覆盖前者。
//
// 示例:
//
//	userMap := convert.ToMap(users, func(u *model.User) string { return u.UserID })
func ToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T, len(slice))
	for _, item := range slice {
		result[keyFunc(item)] = item
	}
	return result
}

// ToMapPtr 将切片转换为指针 map，值指向原切片元素，可避免值拷贝。
// 注意：修改指针指向的值会影响原切片。
//
// 示例:
//
//	userMap := convert.ToMapPtr(users, func(u model.User) string { return u.UserID })
func ToMapPtr[T any, K comparable](slice []T, keyFunc func(T) K) map[K]*T {
	result := make(map[K]*T, len(slice))
	for i := range slice {
		result[keyFunc(slice[i])] = &slice[i]
	}
	return result
}

// Map 转换切片中的每个元素为新类型。
// 常用于提取字段、model → dto 转换。
//
// 示例:
//
//	userIDs := convert.Map(users, func(u *model.User) string { return u.UserID })
func Map[T any, R any](slice []T, mapper func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// GroupBy 按 keyFunc 提取的键对切片分组（一对多）。
// 与 ToMap（一对一）互补，常用于批量关联数据归组。
//
// 示例:
//
//	permsByRole := convert.GroupBy(perms, func(p *model.RolePermission) string { return p.RoleID })
func GroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T, len(slice))
	for _, item := range slice {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}

// Unique 对切片去重并保持原顺序。
// 与 slices.Compact 不同，Unique 不要求预先排序、能去除非相邻的重复元素。
//
// 示例:
//
//	uniqueIDs := convert.Unique(userIDs)
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// Filter 返回满足 predicate 的元素组成的新切片（不修改原切片）。
//
// 示例:
//
//	activeUsers := convert.Filter(users, func(u *model.User) bool { return u.Status == 1 })
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Reduce 将切片归约为单个值。
// initial 为初始累加值，reducer 接收累加器和当前元素，返回新的累加值。
//
// 示例:
//
//	total := convert.Reduce(numbers, 0, func(acc, curr int) int { return acc + curr })
func Reduce[T any, R any](slice []T, initial R, reducer func(R, T) R) R {
	result := initial
	for _, item := range slice {
		result = reducer(result, item)
	}
	return result
}
