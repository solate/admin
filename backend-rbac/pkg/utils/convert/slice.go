package convert

import "golang.org/x/exp/constraints"

// SliceToMap 将切片转换为 map
// K: map 的键类型
// V: map 的值类型(切片元素类型)
// keyFunc: 用于从切片元素中提取键的函数
//
// 示例:
//
//	userMap := SliceToMap(users, func(u *model.User) string { return u.UserID })
//	tenantMap := SliceToMap(tenants, func(t *model.Tenant) string { return t.TenantID })
func SliceToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	if len(slice) == 0 {
		return make(map[K]T)
	}

	result := make(map[K]T, len(slice))
	for _, item := range slice {
		key := keyFunc(item)
		result[key] = item
	}
	return result
}

// SliceToMapPtr 将切片转换为指针 map
// K: map 的键类型
// V: map 的值类型(切片元素类型的指针)
// keyFunc: 用于从切片元素中提取键的函数
//
// 示例:
//
//	userMap := SliceToMapPtr(users, func(u *model.User) string { return u.UserID })
//	tenantMap := SliceToMapPtr(tenants, func(t *model.Tenant) string { return t.TenantID })
func SliceToMapPtr[T any, K comparable](slice []T, keyFunc func(T) K) map[K]*T {
	if len(slice) == 0 {
		return make(map[K]*T)
	}

	result := make(map[K]*T, len(slice))
	for i := range slice {
		key := keyFunc(slice[i])
		result[key] = &slice[i]
	}
	return result
}

// MapToSlice 将 map 转换为切片
// K: map 的键类型
// V: map 的值类型
//
// 示例:
//
//	users := MapToSlice(userMap)
func MapToSlice[K comparable, V any](m map[K]V) []V {
	if len(m) == 0 {
		return make([]V, 0)
	}

	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

// MapKeys 提取 map 的所有 key 为切片
// 示例:
//
//	userIDs := MapKeys(userMap)
func MapKeys[K comparable, V any](m map[K]V) []K {
	if len(m) == 0 {
		return make([]K, 0)
	}

	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// MapValues 提取 map 的所有 value 为切片
// 示例:
//
//	users := MapValues(userMap)
func MapValues[K comparable, V any](m map[K]V) []V {
	if len(m) == 0 {
		return make([]V, 0)
	}

	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

// SliceToUnique 将切片去重
// 示例:
//
//	uniqueIDs := SliceToUnique(userIDs)
func SliceToUnique[T comparable](slice []T) []T {
	if len(slice) == 0 {
		return make([]T, 0)
	}

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

// SliceFilter 过滤切片
// 示例:
//
//	activeUsers := SliceFilter(users, func(u *model.User) bool { return u.Status == 1 })
func SliceFilter[T any](slice []T, predicate func(T) bool) []T {
	if len(slice) == 0 {
		return make([]T, 0)
	}

	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// SliceMap 转换切片中的每个元素
// 示例:
//
//	userIDs := SliceMap(users, func(u *model.User) string { return u.UserID })
func SliceMap[T any, R any](slice []T, mapper func(T) R) []R {
	if len(slice) == 0 {
		return make([]R, 0)
	}

	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = mapper(item)
	}
	return result
}

// SliceReduce 将切片归约为单个值
// initial: 初始值
// reducer: 归约函数,接收累加器和当前元素,返回新的累加值
//
// 示例:
//
//	total := SliceReduce(numbers, 0, func(acc, curr int) int { return acc + curr })
func SliceReduce[T any, R any](slice []T, initial R, reducer func(R, T) R) R {
	result := initial
	for _, item := range slice {
		result = reducer(result, item)
	}
	return result
}

// SliceContains 检查切片是否包含某个元素
// 示例:
//
//	hasAdmin := SliceContains(userIDs, "admin")
func SliceContains[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// SliceGroupBy 按指定键对切片进行分组
// 示例:
//
//	usersByTenant := SliceGroupBy(users, func(u *model.User) string { return u.TenantID })
func SliceGroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	if len(slice) == 0 {
		return make(map[K][]T)
	}

	result := make(map[K][]T, len(slice))
	for _, item := range slice {
		key := keyFunc(item)
		result[key] = append(result[key], item)
	}
	return result
}

// SliceChunk 将切片分块
// 示例:
//
//	chunks := SliceChunk(users, 100) // 每块最多100个元素
func SliceChunk[T any](slice []T, size int) [][]T {
	if len(slice) == 0 || size <= 0 {
		return [][]T{}
	}

	var result [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}
	return result
}

// SliceReverse 反转切片
// 示例:
//
//	reversed := SliceReverse(users)
func SliceReverse[T any](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	result := make([]T, len(slice))
	for i, item := range slice {
		result[len(slice)-1-i] = item
	}
	return result
}

// SliceMax 获取切片中的最大值
// 示例:
//
//	maxAge := SliceMax(ages)
func SliceMax[T constraints.Ordered](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}

	max := slice[0]
	for _, item := range slice[1:] {
		if item > max {
			max = item
		}
	}
	return max
}

// SliceMin 获取切片中的最小值
// 示例:
//
//	minAge := SliceMin(ages)
func SliceMin[T constraints.Ordered](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}

	min := slice[0]
	for _, item := range slice[1:] {
		if item < min {
			min = item
		}
	}
	return min
}
