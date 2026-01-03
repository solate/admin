package seeds

import (
	"admin/internal/dal/model"
	"fmt"

	"gorm.io/gorm"
)

// DictTypeDefinition 字典类型定义
type DictTypeDefinition struct {
	TypeCode    string
	TypeName    string
	Description string
	Items       []DictItemDefinition
}

// DictItemDefinition 字典项定义
type DictItemDefinition struct {
	Label string
	Value string
	Sort  int
}

// DefaultDictTypeDefinitions 返回默认字典类型定义
func DefaultDictTypeDefinitions() []DictTypeDefinition {
	return []DictTypeDefinition{
		{
			TypeCode:    "common_status",
			TypeName:    "通用状态",
			Description: "系统通用状态字典",
			Items: []DictItemDefinition{
				{Label: "启用", Value: "1", Sort: 1},
				{Label: "禁用", Value: "0", Sort: 2},
			},
		},
		{
			TypeCode:    "common_gender",
			TypeName:    "性别",
			Description: "用户性别字典",
			Items: []DictItemDefinition{
				{Label: "男", Value: "1", Sort: 1},
				{Label: "女", Value: "2", Sort: 2},
				{Label: "保密", Value: "0", Sort: 3},
			},
		},
		{
			TypeCode:    "common_yes_no",
			TypeName:    "是否",
			Description: "是否选项字典",
			Items: []DictItemDefinition{
				{Label: "是", Value: "1", Sort: 1},
				{Label: "否", Value: "0", Sort: 2},
			},
		},
		{
			TypeCode:    "user_status",
			TypeName:    "用户状态",
			Description: "用户账号状态",
			Items: []DictItemDefinition{
				{Label: "正常", Value: "1", Sort: 1},
				{Label: "禁用", Value: "0", Sort: 2},
				{Label: "锁定", Value: "2", Sort: 3},
				{Label: "过期", Value: "3", Sort: 4},
			},
		},
		{
			TypeCode:    "tenant_status",
			TypeName:    "租户状态",
			Description: "租户状态字典",
			Items: []DictItemDefinition{
				{Label: "正常", Value: "1", Sort: 1},
				{Label: "禁用", Value: "0", Sort: 2},
				{Label: "过期", Value: "2", Sort: 3},
			},
		},
		{
			TypeCode:    "role_status",
			TypeName:    "角色状态",
			Description: "角色状态字典",
			Items: []DictItemDefinition{
				{Label: "正常", Value: "1", Sort: 1},
				{Label: "禁用", Value: "0", Sort: 2},
			},
		},
		{
			TypeCode:    "menu_type",
			TypeName:    "菜单类型",
			Description: "菜单类型字典",
			Items: []DictItemDefinition{
				{Label: "目录", Value: "0", Sort: 1},
				{Label: "菜单", Value: "1", Sort: 2},
				{Label: "按钮", Value: "2", Sort: 3},
			},
		},
		{
			TypeCode:    "menu_status",
			TypeName:    "菜单状态",
			Description: "菜单状态字典",
			Items: []DictItemDefinition{
				{Label: "启用", Value: "1", Sort: 1},
				{Label: "禁用", Value: "0", Sort: 2},
			},
		},
		{
			TypeCode:    "dept_status",
			TypeName:    "部门状态",
			Description: "部门状态字典",
			Items: []DictItemDefinition{
				{Label: "正常", Value: "1", Sort: 1},
				{Label: "禁用", Value: "0", Sort: 2},
			},
		},
		{
			TypeCode:    "position_status",
			TypeName:    "岗位状态",
			Description: "岗位状态字典",
			Items: []DictItemDefinition{
				{Label: "正常", Value: "1", Sort: 1},
				{Label: "禁用", Value: "0", Sort: 2},
			},
		},
		{
			TypeCode:    "log_level",
			TypeName:    "日志级别",
			Description: "系统日志级别",
			Items: []DictItemDefinition{
				{Label: "调试", Value: "DEBUG", Sort: 1},
				{Label: "信息", Value: "INFO", Sort: 2},
				{Label: "警告", Value: "WARN", Sort: 3},
				{Label: "错误", Value: "ERROR", Sort: 4},
				{Label: "致命", Value: "FATAL", Sort: 5},
			},
		},
		{
			TypeCode:    "login_status",
			TypeName:    "登录状态",
			Description: "用户登录状态",
			Items: []DictItemDefinition{
				{Label: "成功", Value: "1", Sort: 1},
				{Label: "失败", Value: "0", Sort: 2},
			},
		},
		{
			TypeCode:    "operation_type",
			TypeName:    "操作类型",
			Description: "操作日志类型",
			Items: []DictItemDefinition{
				{Label: "创建", Value: "CREATE", Sort: 1},
				{Label: "更新", Value: "UPDATE", Sort: 2},
				{Label: "删除", Value: "DELETE", Sort: 3},
				{Label: "查询", Value: "READ", Sort: 4},
				{Label: "导出", Value: "EXPORT", Sort: 5},
				{Label: "导入", Value: "IMPORT", Sort: 6},
				{Label: "其他", Value: "OTHER", Sort: 99},
			},
		},
	}
}

// SeedDicts 初始化字典数据
func SeedDicts(db *gorm.DB, dictDefs []DictTypeDefinition, tenantID string, ids []string) ([]*model.DictType, error) {
	fmt.Println("📚 开始初始化系统字典")

	var dictTypes []*model.DictType
	idIndex := 0

	for _, dictDef := range dictDefs {
		// 检查字典类型是否已存在
		var existingType model.DictType
		err := db.Where("type_code = ? AND tenant_id = ?", dictDef.TypeCode, tenantID).First(&existingType).Error
		if err == nil {
			fmt.Printf("   ⏭️  字典类型 %s 已存在，跳过\n", dictDef.TypeCode)
			dictTypes = append(dictTypes, &existingType)
			// 跳过已存在字典的项ID
			idIndex += len(dictDef.Items)
			continue
		} else if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("检查字典类型失败: %w", err)
		}

		// 检查ID数量是否足够
		requiredIDs := 1 + len(dictDef.Items) // 1个类型ID + N个项ID
		if idIndex+requiredIDs > len(ids) {
			return nil, fmt.Errorf("ID数量不足，需要 %d 个，剩余 %d 个", requiredIDs, len(ids)-idIndex)
		}

		// 创建字典类型
		dictType := &model.DictType{
			TypeID:      ids[idIndex],
			TenantID:    tenantID,
			TypeCode:    dictDef.TypeCode,
			TypeName:    dictDef.TypeName,
			Description: dictDef.Description,
		}
		idIndex++

		if err := db.Create(dictType).Error; err != nil {
			return nil, fmt.Errorf("创建字典类型失败: %w", err)
		}

		// 创建字典项
		for _, itemDef := range dictDef.Items {
			dictItem := &model.DictItem{
				ItemID:   ids[idIndex],
				TypeID:   dictType.TypeID,
				TenantID: tenantID,
				Label:    itemDef.Label,
				Value:    itemDef.Value,
				Sort:     int32(itemDef.Sort),
			}
			idIndex++

			if err := db.Create(dictItem).Error; err != nil {
				return nil, fmt.Errorf("创建字典项失败: %w", err)
			}
		}

		dictTypes = append(dictTypes, dictType)
		fmt.Printf("   ✓ 创建字典类型: %s (%d个选项)\n", dictDef.TypeName, len(dictDef.Items))
	}

	fmt.Printf("   📊 共初始化 %d 个字典类型\n", len(dictTypes))
	return dictTypes, nil
}
