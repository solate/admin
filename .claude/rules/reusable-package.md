# 独立可复用封装包规范

> 适用于设计为"可整目录 copy 到其他项目"的**通用库封装包**(如 xredis / xkafka 这类对第三方库的通用封装)。
> 区别于只为本项目服务的 pkg 业务包(audit / cache / response 等)。
> 数据库连接封装同 xredis 一样自包含(无 internal 依赖),已提升为 `pkg/xgorm`。
>
> **不适用**:项目专属、结构静态的部分,如**配置结构体本身** —— 每个项目字段不同,是"复制后改字段"的项目模板,放在项目的 `internal/config`,不套用本规范。
>
> 配置加载的封装划分:通用加载机制(读 base + overlay 合并 + 环境变量覆盖)抽成 `pkg/xviper`(泛型 `Load[T]`,套用本规范);项目专属的 `Config` 结构体与校验规则放 `internal/config`(不套用本规范,复制后改字段)。详见 `docs/blog/从零设计Go配置加载-viper封装与约定式设计.md`。

## 规则 1：x 前缀命名,避免 import 别名

独立封装包用 `x` 前缀命名(xredis / xkafka / xes 等),避免与项目内部常见包名
(config / cache 等)冲突,import 时**无需别名**。

```go
// ✅ x 前缀,无冲突
import "admin/pkg/utils/xredis"
client := xredis.New(xredis.Options{...})

// ❌ 普通名,容易与项目内部 config 冲突,被迫用别名
import cfgload "admin/pkg/utils/config"
cfgload.New(cfgload.Options{...})
```

## 规则 2：自包含 + 完整功能

- 核心功能在包内**完整实现**(加载、覆盖、校验…),不把关键逻辑留给调用方补
- 只依赖标准库 + 第三方库,**不依赖宿主项目的 internal 包**
- 行为正确性靠实现保证,不靠"调用方记得补"(避免静默失效)

## 规则 3：必须带完整单元测试

- 包内 `xxx_test.go` 覆盖**全部公开行为**(正常 / 异常 / 边界)
- 测试自包含:`t.TempDir()` 写临时配置、`t.Setenv()` 设环境变量,不依赖外部服务/状态
- 跑 `go test ./<pkg>/` 全绿才算完成

## 规则 4：可移植

- 整个目录(代码 + 测试)**可直接复制**到其他 Go module 使用
- 配置通过 Options / 参数传入,**不写死**项目相关路径或字段
- 不持有全局状态(无全局单例),实例由调用方持有

---

**最后更新**:2026-07-03(配置加载重构为 `pkg/xviper` 泛型封装 `Load[T]`——通用加载机制套用本规范、项目专属 `Config` 与校验放 `internal/config`;见 `docs/blog/从零设计Go配置加载-viper封装与约定式设计.md`。本规范同样适用于 xredis 等跨项目共享的第三方库封装)
