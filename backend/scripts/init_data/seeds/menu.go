package seeds

import (
	"admin/internal/dal/model"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

// APIPath API 路径定义
type APIPath struct {
	Path    string   `json:"path"`    // API 路径
	Methods []string `json:"methods"` // HTTP 方法列表
}

// MenuDefinition 菜单定义
type MenuDefinition struct {
	MenuID     string
	ParentID   string
	Name       string
	Path       string
	Component  string
	Icon       string
	Redirect   string
	Sort       int
	Status     int
	Description string
	APIPaths   []APIPath // API 路径列表
}

// ToModel 转换为 Menu 模型
func (m *MenuDefinition) ToModel() *model.Menu {
	apiPathsJSON := ""
	if len(m.APIPaths) > 0 {
		data, _ := json.Marshal(m.APIPaths)
		apiPathsJSON = string(data)
	}

	return &model.Menu{
		MenuID:      m.MenuID,
		ParentID:    m.ParentID,
		Name:        m.Name,
		Path:        m.Path,
		Component:   m.Component,
		Icon:        m.Icon,
		Redirect:    m.Redirect,
		Sort:        int32(m.Sort),
		Status:      int16(m.Status),
		Description: m.Description,
		APIPaths:    apiPathsJSON,
	}
}

// DefaultMenuDefinitions 返回默认菜单定义
// 根据前端 Layout.vue 中的菜单结构生成
// 新增：为菜单配置 API 路径，实现菜单权限自动关联 API 权限
func DefaultMenuDefinitions(menuIDs []string) []MenuDefinition {
	return []MenuDefinition{
		// ==================== 工作台 ====================
		{
			MenuID: menuIDs[0], ParentID: "", Name: "工作台", Path: "/",
			Component: "views/Dashboard.vue", Icon: "DataBoard", Redirect: "", Sort: 1, Status: 1,
			Description: "系统首页工作台",
			APIPaths: []APIPath{}, // 工作台无需 API 权限
		},

		// ==================== 租户管理 ====================
		{
			MenuID: menuIDs[1], ParentID: "", Name: "租户管理", Path: "/tenant",
			Component: "", Icon: "OfficeBuilding", Redirect: "", Sort: 2, Status: 1,
			Description: "租户相关管理",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[2], ParentID: menuIDs[1], Name: "租户列表", Path: "/tenant/list",
			Component: "views/tenant/TenantList.vue", Icon: "", Redirect: "", Sort: 1, Status: 1,
			Description: "管理所有租户信息",
			APIPaths: []APIPath{
				{Path: "/api/v1/tenants", Methods: []string{"GET", "POST"}},
				{Path: "/api/v1/tenants/:tenant_id", Methods: []string{"GET", "PUT", "DELETE"}},
				{Path: "/api/v1/tenants/:tenant_id/status/:status", Methods: []string{"PUT"}},
			},
		},
		{
			MenuID: menuIDs[3], ParentID: menuIDs[1], Name: "套餐管理", Path: "/tenant/packages",
			Component: "views/tenant/TenantPackages.vue", Icon: "", Redirect: "", Sort: 2, Status: 1,
			Description: "管理租户套餐配置",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[4], ParentID: menuIDs[1], Name: "订阅管理", Path: "/tenant/subscription",
			Component: "views/tenant/Subscription.vue", Icon: "", Redirect: "", Sort: 3, Status: 1,
			Description: "管理租户订阅信息",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[5], ParentID: menuIDs[1], Name: "账单管理", Path: "/tenant/billing",
			Component: "views/tenant/Billing.vue", Icon: "", Redirect: "", Sort: 4, Status: 1,
			Description: "管理租户账单",
			APIPaths: []APIPath{},
		},

		// ==================== 组织架构 ====================
		{
			MenuID: menuIDs[6], ParentID: "", Name: "组织架构", Path: "/organization",
			Component: "", Icon: "Share", Redirect: "", Sort: 3, Status: 1,
			Description: "组织架构管理",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[7], ParentID: menuIDs[6], Name: "部门管理", Path: "/organization/departments",
			Component: "views/organization/Departments.vue", Icon: "", Redirect: "", Sort: 1, Status: 1,
			Description: "管理公司组织架构部门",
			APIPaths: []APIPath{
				{Path: "/api/v1/departments", Methods: []string{"GET", "POST"}},
				{Path: "/api/v1/departments/:department_id", Methods: []string{"GET", "PUT", "DELETE"}},
				{Path: "/api/v1/departments/:department_id/status/:status", Methods: []string{"PUT"}},
				{Path: "/api/v1/departments/:department_id/children", Methods: []string{"GET"}},
				{Path: "/api/v1/departments/tree", Methods: []string{"GET"}},
			},
		},
		{
			MenuID: menuIDs[8], ParentID: menuIDs[6], Name: "岗位管理", Path: "/organization/positions",
			Component: "views/organization/Positions.vue", Icon: "", Redirect: "", Sort: 2, Status: 1,
			Description: "管理公司岗位信息",
			APIPaths: []APIPath{
				{Path: "/api/v1/positions", Methods: []string{"GET", "POST"}},
				{Path: "/api/v1/positions/all", Methods: []string{"GET"}},
				{Path: "/api/v1/positions/:position_id", Methods: []string{"GET", "PUT", "DELETE"}},
				{Path: "/api/v1/positions/:position_id/status/:status", Methods: []string{"PUT"}},
			},
		},

		// ==================== 用户与权限 ====================
		{
			MenuID: menuIDs[9], ParentID: "", Name: "用户与权限", Path: "/access",
			Component: "", Icon: "Lock", Redirect: "", Sort: 4, Status: 1,
			Description: "用户与权限管理",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[10], ParentID: menuIDs[9], Name: "用户管理", Path: "/access/users",
			Component: "views/access/Users.vue", Icon: "", Redirect: "", Sort: 1, Status: 1,
			Description: "管理系统用户",
			APIPaths: []APIPath{
				{Path: "/api/v1/users", Methods: []string{"GET", "POST"}},
				{Path: "/api/v1/users/:user_id", Methods: []string{"GET", "PUT", "DELETE"}},
				{Path: "/api/v1/users/:user_id/status/:status", Methods: []string{"PUT"}},
			},
		},
		{
			MenuID: menuIDs[11], ParentID: menuIDs[9], Name: "角色管理", Path: "/access/roles",
			Component: "views/access/Roles.vue", Icon: "", Redirect: "", Sort: 2, Status: 1,
			Description: "管理系统角色和权限",
			APIPaths: []APIPath{
				{Path: "/api/v1/roles", Methods: []string{"GET", "POST"}},
				{Path: "/api/v1/roles/:role_id", Methods: []string{"GET", "PUT", "DELETE"}},
				{Path: "/api/v1/roles/:role_id/status/:status", Methods: []string{"PUT"}},
				{Path: "/api/v1/roles/:role_id/permissions", Methods: []string{"GET", "PUT"}},
			},
		},
		{
			MenuID: menuIDs[12], ParentID: menuIDs[9], Name: "菜单权限", Path: "/access/menus",
			Component: "views/access/Menus.vue", Icon: "", Redirect: "", Sort: 3, Status: 1,
			Description: "管理菜单和权限配置",
			APIPaths: []APIPath{
				{Path: "/api/v1/menus", Methods: []string{"GET", "POST"}},
				{Path: "/api/v1/menus/all", Methods: []string{"GET"}},
				{Path: "/api/v1/menus/tree", Methods: []string{"GET"}},
				{Path: "/api/v1/menus/:menu_id", Methods: []string{"GET", "PUT", "DELETE"}},
				{Path: "/api/v1/menus/:menu_id/status/:status", Methods: []string{"PUT"}},
			},
		},
		{
			MenuID: menuIDs[13], ParentID: menuIDs[9], Name: "数据权限", Path: "/access/data-permissions",
			Component: "views/access/DataPermissions.vue", Icon: "", Redirect: "", Sort: 4, Status: 1,
			Description: "管理数据权限范围",
			APIPaths: []APIPath{},
		},

		// ==================== 业务管理 ====================
		{
			MenuID: menuIDs[14], ParentID: "", Name: "业务管理", Path: "/business",
			Component: "", Icon: "Briefcase", Redirect: "", Sort: 5, Status: 1,
			Description: "业务数据管理",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[15], ParentID: menuIDs[14], Name: "工厂管理", Path: "/business/factories",
			Component: "views/business/Factories.vue", Icon: "", Redirect: "", Sort: 1, Status: 1,
			Description: "管理工厂信息",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[16], ParentID: menuIDs[14], Name: "商品管理", Path: "/business/products",
			Component: "views/business/Products.vue", Icon: "", Redirect: "", Sort: 2, Status: 1,
			Description: "管理商品信息",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[17], ParentID: menuIDs[14], Name: "订单管理", Path: "/business/orders",
			Component: "views/business/Orders.vue", Icon: "", Redirect: "", Sort: 3, Status: 1,
			Description: "管理订单信息",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[18], ParentID: menuIDs[14], Name: "数据统计", Path: "/business/statistics",
			Component: "views/business/Statistics.vue", Icon: "", Redirect: "", Sort: 4, Status: 1,
			Description: "业务数据统计分析",
			APIPaths: []APIPath{},
		},

		// ==================== 审计日志 ====================
		{
			MenuID: menuIDs[19], ParentID: "", Name: "审计日志", Path: "/audit",
			Component: "", Icon: "Document", Redirect: "", Sort: 6, Status: 1,
			Description: "系统审计日志",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[20], ParentID: menuIDs[19], Name: "登录日志", Path: "/audit/login",
			Component: "views/audit/LoginLog.vue", Icon: "", Redirect: "", Sort: 1, Status: 1,
			Description: "用户登录日志",
			APIPaths: []APIPath{
				{Path: "/api/v1/login-logs", Methods: []string{"GET"}},
				{Path: "/api/v1/login-logs/:log_id", Methods: []string{"GET"}},
			},
		},
		{
			MenuID: menuIDs[21], ParentID: menuIDs[19], Name: "操作日志", Path: "/audit/operation",
			Component: "views/audit/OperationLog.vue", Icon: "", Redirect: "", Sort: 2, Status: 1,
			Description: "用户操作日志",
			APIPaths: []APIPath{
				{Path: "/api/v1/operation-logs", Methods: []string{"GET"}},
				{Path: "/api/v1/operation-logs/:log_id", Methods: []string{"GET"}},
			},
		},
		{
			MenuID: menuIDs[22], ParentID: menuIDs[19], Name: "数据变更", Path: "/audit/data",
			Component: "views/audit/DataChange.vue", Icon: "", Redirect: "", Sort: 3, Status: 1,
			Description: "数据变更记录",
			APIPaths: []APIPath{},
		},

		// ==================== 系统设置 ====================
		{
			MenuID: menuIDs[23], ParentID: "", Name: "系统设置", Path: "/settings",
			Component: "", Icon: "Setting", Redirect: "", Sort: 7, Status: 1,
			Description: "系统配置管理",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[24], ParentID: menuIDs[23], Name: "字典管理", Path: "/settings/dictionary",
			Component: "views/settings/Dictionary.vue", Icon: "", Redirect: "", Sort: 1, Status: 1,
			Description: "管理数据字典",
			APIPaths: []APIPath{
				{Path: "/api/v1/dict/:type_code", Methods: []string{"GET", "PUT"}},
				{Path: "/api/v1/dict-types", Methods: []string{"GET"}},
			},
		},
		{
			MenuID: menuIDs[25], ParentID: menuIDs[23], Name: "系统参数", Path: "/settings/parameters",
			Component: "views/settings/Parameters.vue", Icon: "", Redirect: "", Sort: 2, Status: 1,
			Description: "系统参数配置",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[26], ParentID: menuIDs[23], Name: "通知配置", Path: "/settings/notifications",
			Component: "views/settings/Notifications.vue", Icon: "", Redirect: "", Sort: 3, Status: 1,
			Description: "系统通知配置",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[27], ParentID: menuIDs[23], Name: "存储配置", Path: "/settings/storage",
			Component: "views/settings/Storage.vue", Icon: "", Redirect: "", Sort: 4, Status: 1,
			Description: "存储服务配置",
			APIPaths: []APIPath{},
		},
		{
			MenuID: menuIDs[28], ParentID: menuIDs[23], Name: "系统监控", Path: "/settings/monitor",
			Component: "views/settings/Monitor.vue", Icon: "", Redirect: "", Sort: 5, Status: 1,
			Description: "系统运行监控",
			APIPaths: []APIPath{},
		},
	}
}

// SeedSystemMenus 初始化系统菜单
func SeedSystemMenus(db *gorm.DB, menuDefs []MenuDefinition) error {
	// 检查是否已有菜单
	var count int64
	if err := db.Model(&model.Menu{}).Where("deleted_at = 0").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		fmt.Println("ℹ️  系统菜单已存在，跳过初始化")
		return nil
	}

	fmt.Println("\n📋 开始初始化系统菜单")

	for _, def := range menuDefs {
		// 使用 ToModel 方法转换，自动处理 APIPaths JSON 序列化
		menu := def.ToModel()
		if err := db.Create(menu).Error; err != nil {
			return fmt.Errorf("创建菜单 %s 失败: %w", def.Name, err)
		}
		fmt.Printf("✅ 菜单创建成功: %s (%s)\n", def.Name, def.Path)
	}

	fmt.Printf("📋 系统菜单初始化完成: 共 %d 个菜单\n", len(menuDefs))
	return nil
}
