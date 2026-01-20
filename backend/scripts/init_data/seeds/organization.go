package seeds

import (
	"admin/internal/dal/model"
	"fmt"

	"gorm.io/gorm"
)

// DepartmentDefinition 部门定义
type DepartmentDefinition struct {
	DepartmentID   string
	ParentID       string
	DepartmentName string
	Description    string
	Sort           int32
}

// PositionDefinition 岗位定义
type PositionDefinition struct {
	PositionID   string
	PositionCode string
	PositionName string
	Level        int32
	Description  string
	Sort         int32
}

// DefaultDepartmentDefinitions 返回默认部门定义
// 参考：典型企业组织架构 - 总经办、技术部、产品部、运营部、市场部、销售部、人事部、财务部
func DefaultDepartmentDefinitions(departmentIDs []string) []DepartmentDefinition {
	return []DepartmentDefinition{
		// 根部门（一级）
		{departmentIDs[0], "", "总公司", "企业总部", 0},

		// 总经办（二级）
		{departmentIDs[1], departmentIDs[0], "总经办", "负责公司整体战略决策和管理", 1},

		// 技术中心（二级）
		{departmentIDs[2], departmentIDs[0], "技术中心", "负责产品研发和技术支持", 2},
		// 技术中心下属部门（三级）
		{departmentIDs[3], departmentIDs[2], "研发部", "负责核心产品研发", 1},
		{departmentIDs[4], departmentIDs[2], "测试部", "负责产品质量保证", 2},
		{departmentIDs[5], departmentIDs[2], "运维部", "负责系统运维和基础设施", 3},

		// 产品中心（二级）
		{departmentIDs[6], departmentIDs[0], "产品中心", "负责产品规划和设计", 3},
		// 产品中心下属部门（三级）
		{departmentIDs[7], departmentIDs[6], "产品部", "负责产品设计和需求管理", 1},
		{departmentIDs[8], departmentIDs[6], "设计部", "负责UI/UX设计", 2},

		// 运营中心（二级）
		{departmentIDs[9], departmentIDs[0], "运营中心", "负责用户运营和内容运营", 4},
		// 运营中心下属部门（三级）
		{departmentIDs[10], departmentIDs[9], "用户运营部", "负责用户增长和留存", 1},
		{departmentIDs[11], departmentIDs[9], "内容运营部", "负责内容策划和管理", 2},

		// 市场中心（二级）
		{departmentIDs[12], departmentIDs[0], "市场中心", "负责市场推广和品牌建设", 5},

		// 销售中心（二级）
		{departmentIDs[13], departmentIDs[0], "销售中心", "负责销售业务和客户管理", 6},
		// 销售中心下属部门（三级）
		{departmentIDs[14], departmentIDs[13], "直销部", "负责直接销售业务", 1},
		{departmentIDs[15], departmentIDs[13], "渠道部", "负责渠道销售业务", 2},

		// 职能部门（二级）
		{departmentIDs[16], departmentIDs[0], "人力资源部", "负责人力资源管理", 7},
		{departmentIDs[17], departmentIDs[0], "财务部", "负责财务管理", 8},
		{departmentIDs[18], departmentIDs[0], "行政部", "负责行政管理", 9},
	}
}

// DefaultPositionDefinitions 返回默认岗位定义
// 参考：典型企业岗位体系，岗位编码与 Casbin 角色对应
func DefaultPositionDefinitions(positionIDs []string) []PositionDefinition {
	return []PositionDefinition{
		// 管理岗位
		{positionIDs[0], "CEO", "首席执行官", 100, "公司最高负责人", 1},
		{positionIDs[1], "CTO", "首席技术官", 90, "技术负责人", 2},
		{positionIDs[2], "CPO", "首席产品官", 85, "产品负责人", 3},
		{positionIDs[3], "COO", "首席运营官", 88, "运营负责人", 4},
		{positionIDs[4], "CFO", "首席财务官", 87, "财务负责人", 5},

		// 部门管理岗位
		{positionIDs[5], "DEPT_MANAGER", "部门经理", 60, "部门负责人", 10},
		{positionIDs[6], "DEPT_LEADER", "部门主管", 50, "部门主管", 11},
		{positionIDs[7], "TEAM_LEADER", "团队组长", 40, "团队负责人", 12},

		// 技术岗位
		{positionIDs[8], "ARCHITECT", "架构师", 55, "技术架构师", 20},
		{positionIDs[9], "SENIOR_ENGINEER", "高级工程师", 45, "高级开发工程师", 21},
		{positionIDs[10], "ENGINEER", "工程师", 35, "开发工程师", 22},
		{positionIDs[11], "JUNIOR_ENGINEER", "初级工程师", 25, "初级开发工程师", 23},

		{positionIDs[12], "QA_LEAD", "测试主管", 48, "测试部门负责人", 30},
		{positionIDs[13], "QA_ENGINEER", "测试工程师", 38, "软件测试工程师", 31},
		{positionIDs[14], "OPS_ENGINEER", "运维工程师", 38, "系统运维工程师", 32},

		// 产品岗位
		{positionIDs[15], "PRODUCT_MANAGER", "产品经理", 50, "产品规划和设计", 40},
		{positionIDs[16], "PRODUCT_OWNER", "产品负责人", 45, "产品线负责人", 41},
		{positionIDs[17], "UI_DESIGNER", "UI设计师", 38, "用户界面设计", 42},
		{positionIDs[18], "UX_DESIGNER", "UX设计师", 38, "用户体验设计", 43},

		// 运营岗位
		{positionIDs[19], "OPERATION_MANAGER", "运营经理", 52, "运营部门负责人", 50},
		{positionIDs[20], "CONTENT_OPERATOR", "内容运营", 35, "内容策划和运营", 51},
		{positionIDs[21], "USER_OPERATOR", "用户运营", 35, "用户增长和运营", 52},

		// 市场岗位
		{positionIDs[22], "MARKETING_MANAGER", "市场经理", 52, "市场部门负责人", 60},
		{positionIDs[23], "BRAND_MANAGER", "品牌经理", 45, "品牌建设和管理", 61},

		// 销售岗位
		{positionIDs[24], "SALES_MANAGER", "销售经理", 52, "销售部门负责人", 70},
		{positionIDs[25], "SALES_DIRECTOR", "销售总监", 55, "销售业务负责人", 71},
		{positionIDs[26], "SALES_REP", "销售代表", 30, "销售业务员", 72},

		// 职能岗位
		{positionIDs[27], "HR_MANAGER", "人事经理", 50, "人力资源负责人", 80},
		{positionIDs[28], "HR_SPECIALIST", "人事专员", 35, "人事专员", 81},
		{positionIDs[29], "HR_RECRUITER", "招聘专员", 35, "负责招聘工作", 82},

		{positionIDs[30], "FINANCE_MANAGER", "财务经理", 50, "财务部门负责人", 90},
		{positionIDs[31], "ACCOUNTANT", "会计", 38, "会计核算", 91},
		{positionIDs[32], "CASHIER", "出纳", 30, "现金管理", 92},

		{positionIDs[33], "ADMIN_MANAGER", "行政经理", 48, "行政部门负责人", 100},
		{positionIDs[34], "ADMIN_STAFF", "行政专员", 32, "行政事务处理", 101},

		// 通用岗位
		{positionIDs[35], "EMPLOYEE", "员工", 10, "普通员工", 200},
		{positionIDs[36], "INTERN", "实习生", 5, "实习岗位", 201},
	}
}

// SeedDepartments 初始化默认部门
func SeedDepartments(db *gorm.DB, deptDefs []DepartmentDefinition, tenantID string) ([]model.Department, error) {
	departments := make([]model.Department, 0, len(deptDefs))

	for _, def := range deptDefs {
		var dept model.Department
		if err := db.Where("department_id = ? AND tenant_id = ?", def.DepartmentID, tenantID).First(&dept).Error; err != nil {
			// 部门不存在，创建新部门
			dept = model.Department{
				DepartmentID:   def.DepartmentID,
				TenantID:       tenantID,
				ParentID:       def.ParentID,
				DepartmentName: def.DepartmentName,
				Description:    def.Description,
				Sort:           def.Sort,
				Status:         1,
			}
			if err := db.Create(&dept).Error; err != nil {
				return nil, fmt.Errorf("创建部门 %s 失败: %w", def.DepartmentName, err)
			}
			fmt.Printf("✅ 部门创建成功 dept_id=%s name=%s parent_id=%s\n", dept.DepartmentID, dept.DepartmentName, dept.ParentID)
		} else {
			fmt.Printf("ℹ️  部门已存在 dept_id=%s name=%s\n", dept.DepartmentID, dept.DepartmentName)
		}
		departments = append(departments, dept)
	}

	return departments, nil
}

// SeedPositions 初始化默认岗位
func SeedPositions(db *gorm.DB, posDefs []PositionDefinition, tenantID string) ([]model.Position, error) {
	positions := make([]model.Position, 0, len(posDefs))

	for _, def := range posDefs {
		var pos model.Position
		if err := db.Where("position_code = ? AND tenant_id = ?", def.PositionCode, tenantID).First(&pos).Error; err != nil {
			// 岗位不存在，创建新岗位
			pos = model.Position{
				PositionID:   def.PositionID,
				TenantID:     tenantID,
				PositionCode: def.PositionCode,
				PositionName: def.PositionName,
				Level:        def.Level,
				Description:  def.Description,
				Sort:         def.Sort,
				Status:       1,
			}
			if err := db.Create(&pos).Error; err != nil {
				return nil, fmt.Errorf("创建岗位 %s 失败: %w", def.PositionName, err)
			}
			fmt.Printf("✅ 岗位创建成功 position_id=%s code=%s name=%s level=%d\n", pos.PositionID, pos.PositionCode, pos.PositionName, pos.Level)
		} else {
			fmt.Printf("ℹ️  岗位已存在 position_id=%s code=%s name=%s\n", pos.PositionID, pos.PositionCode, pos.PositionName)
		}
		positions = append(positions, pos)
	}

	return positions, nil
}
