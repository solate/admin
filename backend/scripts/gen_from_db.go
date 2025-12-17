package main

import (
	"fmt"
	"os"
	"strings"

	"admin/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func main() {
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
	genCfg := gen.Config{
		OutPath:           "./internal/dal/query",
		OutFile:           "gen.go",
		ModelPkgPath:      "./internal/dal/model",
		Mode:              gen.WithDefaultQuery | gen.WithQueryInterface,
		FieldNullable:     true,
		FieldCoverable:    false,
		FieldSignable:     false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	}
	genCfg.WithImportPkgPath("gorm.io/plugin/soft_delete")
	g := gen.NewGenerator(genCfg)

	g.UseDB(db)

	// 4. 正确过滤不需要的表
	tables, err := db.Migrator().GetTables()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 获取表列表失败: %v\n", err)
		os.Exit(1)
	}

	// 定义需要排除的表
	excludeTables := map[string]bool{
		"schema_migrations": true, // 精确匹配表名
		// 可添加其他需要排除的表
	}

	// 5. 只生成需要的表
	var models []any
	for _, table := range tables {
		if excludeTables[strings.ToLower(table)] {
			fmt.Printf("⏭️  跳过表: %s\n", table)
			continue
		}
		fmt.Printf("✅ 生成表: %s\n", table)

		opts := []gen.ModelOpt{
			gen.FieldGORMTag("created_at", func(tag field.GormTag) field.GormTag {
				tag.Set("autoCreateTime", "milli")
				return tag
			}),
			gen.FieldGORMTag("updated_at", func(tag field.GormTag) field.GormTag {
				tag.Set("autoUpdateTime", "milli")
				return tag
			}),
			gen.FieldType("deleted_at", "soft_delete.DeletedAt"),
			gen.FieldGORMTag("deleted_at", func(tag field.GormTag) field.GormTag {
				tag.Set("softDelete", "milli")
				return tag
			}),
		}
		models = append(models, g.GenerateModel(table, opts...))
	}

	if len(models) == 0 {
		fmt.Fprintln(os.Stderr, "⚠️  没有找到需要生成的表!")
		os.Exit(1)
	}

	g.ApplyBasic(models...)
	g.Execute()

	fmt.Println("✨ 代码生成成功!")

	if len(models) == 0 {
		fmt.Fprintln(os.Stderr, "⚠️  没有找到需要生成的表!")
		os.Exit(1)
	}

	g.ApplyBasic(models...)
	g.Execute()

	fmt.Println("✨ 代码生成成功!")
}
