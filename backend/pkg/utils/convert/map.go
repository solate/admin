package convert

// ToMap 将切片转换为 map（通用版本）
//
// 示例：
//
//	deviceMap := convert.ToMap(devices, func(d *model.Device) string { return d.DeviceID })
//	tenantMap := convert.ToMap(tenants, func(t *model.Tenant) string { return t.TenantID })
func ToMap[T any, K comparable](slice []T, keyFunc func(T) K) map[K]T {
	result := make(map[K]T, len(slice))
	for _, item := range slice {
		result[keyFunc(item)] = item
	}
	return result
}
