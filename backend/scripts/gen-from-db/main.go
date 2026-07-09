// gen-from-db 是 GORM Gen 的代码生成入口(database-first)。
//
// 用法:make gen-db(等价 go run ./scripts/gen-from-db/)。
// 流程:读 internal/config → 连活库 → 反射所有表 → 生成 internal/dal/model + internal/dal/query。
//
// 约定(见 .claude/rules/migration-data-convention.md):
//   - migrations/ 是表结构真相源,改表后重跑本脚本刷新生成代码,勿手改 *.gen.go。
//   - 时间戳统一 bigint 毫秒:created_at→autoCreateTime:milli、updated_at→autoUpdateTime:milli。
//   - 软删 deleted_at→soft_delete.DeletedAt(milli)。
//   - 排除 golang-migrate 的记账表 schema_migrations。
//
// 脚本布局约定:每个一次性可执行工具独占一个子目录(各自 package main),按目录运行
// (go run ./scripts/xxx/)。这样 scripts/ 可容纳多个 main 而不冲突。
package main

import (
	"fmt"
	"log"
	"strings"

	"admin/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// codegen 专用连接:与运行时连接分离,开启预编译、跳过默认事务(仅反射建表用不到事务)。
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

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

	// 全局字段配置:所有表统一套用,新增表自动继承,无需逐表重复传 opts。
	// 时间戳统一 bigint 毫秒,软删 deleted_at→soft_delete.DeletedAt(milli)。
	genCfg.WithOpts(
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
	)

	g := gen.NewGenerator(genCfg)
	g.UseDB(db)

	tables, err := db.Migrator().GetTables()
	if err != nil {
		log.Fatalf("读取数据库表失败: %v", err)
	}

	// schema_migrations 是 golang-migrate 的版本记账表,不生成 model。
	excludeTables := map[string]bool{
		"schema_migrations": true,
	}

	var models []any
	for _, table := range tables {
		if excludeTables[strings.ToLower(table)] {
			continue
		}
		// 字段配置已提到 genCfg.WithOpts 全局套用,此处逐表无需再传 opts。
		models = append(models, g.GenerateModel(table))
	}

	if len(models) == 0 {
		log.Println("⚠️  未发现任何业务表(除 schema_migrations),仅生成 query 基础代码")
	}

	g.ApplyBasic(models...)
	g.Execute()

	fmt.Printf("✅ 代码生成完成:%d 张表 → internal/dal/model + internal/dal/query\n", len(models))
}
