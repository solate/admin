package main

import (
	"admin/pkg/config"
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func main() {

	// 1. 加载配置 (必须最先加载，其他组件依赖配置)
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 配置加载失败: %v\n", err)
		os.Exit(1)
	}

	dsn := cfg.Database.GetDSN() + " TimeZone=Asia/Shanghai"

	// 2. 添加更健壮的 GORM 配置
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true, // 启用预编译
		SkipDefaultTransaction: true, // 禁用默认事务
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 数据库连接失败: %v\n", err)
		os.Exit(1)
	}

	// 3. 创建生成器（保持你的配置，但修正路径）
	g := gen.NewGenerator(gen.Config{
		OutPath: "./internal/dal/query",
		OutFile: "gen.go",
		// ModelPkgPath:      "./internal/dal/model",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})

	g.UseDB(db)

	// 生成单个表的model和query
	g.ApplyBasic(
	// Generate structs from all tables of current database
	//	g.GenerateAllTable()...,

	// 生成users表的model和query
	// g.GenerateModel("users"),
	)
	// Generate the code
	g.Execute()

}
