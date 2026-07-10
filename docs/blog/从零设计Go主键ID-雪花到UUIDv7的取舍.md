# 从零设计 Go 主键 ID：雪花到 UUIDv7 的取舍（2026）

> 读完你能想清楚一件事：**一个 2026 年新起的后台项目，主键该用雪花还是 UUIDv7，为什么，以及怎么落。**
> 不是「哪个更潮」，是把「为什么不用原来的、换了会遇到什么担忧、每个担忧怎么算、最后怎么落地」这条链走完。
> 场景：多租户 SaaS 后台、PostgreSQL 18、前端 Vue3 + TypeScript + axios。
> 选型全过程与逐条论证见 research：
> [01 要不要换](../saas-backend/research/idgen/01-ID生成方案选型-雪花vs-UUID.md) ·
> [02 担忧拆解](../saas-backend/research/idgen/02-UUIDv7落地细节-前端影响与生成方式.md) ·
> [03 存储机制与落地](../saas-backend/research/idgen/03-UUIDv7存储机制与落地-PG18字段设计与文本二进制转换.md)

---

## 0. 开篇：主键这事，其实一直在「伺候」ID

主键选型看着是小事，真写起来你会发现，雪花算法让你操的心不少。把「用雪花」这三个字摊开，背后是一串
**「你得伺候它」**的待办：

1. **机器 ID 得协调**——多机部署要保证每台机器 ID 不撞，还得有一套兜底逻辑去凑这个 ID。
2. **启动可能 panic**——机器 ID 解析失败、实例创建失败，`init()` 里直接挂，生成 ID 这件事多了一个启动失败面。
3. **时钟和 epoch 得管**——定一个纪元起点、上线后永不能改，还要担心服务器时钟回拨。
4. **64 位数字必须转成字符串**——否则前端 JavaScript 静默丢精度，ID 尾数悄悄错掉。
5. **换个方案会不会更麻烦**——UUID 不是又长又占地方吗？前端要不要跟着改？以后换低版本数据库怎么办？

这篇就是把这五条逐个拆开，算清楚每一笔账。**先说立场：本项目已于 2026-07 在 `backend/` 落地了 UUIDv7**——
主键改用 PostgreSQL 18 原生 `uuid` 列 + `DEFAULT uuidv7()`（库端自动生成），`idgen` 包保留一个
`uuid.NewV7()` 的应用层备用封装。结论很简单：**UUIDv7 更省心，且换得越早越便宜**。为什么，往下看。

## 1. 反面镜鉴：老项目的雪花为什么让人「伺候」

不是雪花不好，是它带着一串**「你得伺候它」**的机制成本。拿老项目里那套 sony/sonyflake 实现当镜子照一照：

- **机器 ID 要协调，还带碰撞隐患**。雪花靠「机器 ID + 时间 + 序列」拼唯一，多机部署就得保证每台机器
  ID 不撞。老项目的实现用「环境变量 → `/etc/machine-id` → MAC 哈希」**三级兜底**去凑这个 ID，注释里
  自己写着**「机器到 300 台时碰撞概率超 50%」**——这是一颗需要时时惦记的雷。
- **启动可能 panic**。机器 ID 解析失败、sonyflake 实例创建失败，`init()` 里直接 panic。为了「宁停不错」
  这设计没毛病，但它意味着**「生成 ID」这件事有一个启动失败面**：配置一错，服务起不来。
- **时钟和 epoch 要管**。雪花要定一个纪元起点（epoch），上线后**永远不能改**，否则 ID 重叠；还要担心
  服务器时钟回拨导致 ID 倒退或重复，以及纪元决定的**约 174 年可用上限**这种边界。

这些成本换来的是「短、有序」两个好处。但下一节会看到，「短」这个好处，在有前端的项目里**已经被迫缩水**了。

## 2. 最咬人的一条：JavaScript 的 2^53

这是选型里最实际的一条，也是「前端说 ID 太长、后端改成字符串」这条约定的技术根因。

**JavaScript 的 `Number` 是 IEEE 754 双精度浮点**，能精确表示的最大整数是
`Number.MAX_SAFE_INTEGER = 2^53 − 1 = 9007199254740991`（16 位十进制）。而 sonyflake ID 是 64 位、
**18–19 位十进制数，远超 2^53**。后果是 JS 解析时**静默丢精度、不报错**：

```js
JSON.parse('{"id": 186429408913933113}').id
// → 186429408913933110   ← 末位 3 被改成 0，ID 悄悄错掉
```

最坑的地方在于**无异常**：ID 看着还在、只是尾数错了，前端拿着这个错 ID 去查详情或删除 →
**404，甚至删错行**。

### 2.1 为什么前端不自己扛

有人会说，前端上个 `json-bigint` 不就行了？本项目前端（`frontend/` 与 UI 原型 `frontend-uiux/`，都是
Vue3 + TypeScript + Vite + axios）探查下来：**全程把 ID 当 `string` 处理**，所有 `*_id` 字段、按 ID
操作的函数入参都是 `string`，搜不到 `id: number`；也**没有任何** `json-bigint` / `bignumber.js` /
自定义 `transformResponse`——axios 走默认 JSON 解析。

也就是说，前端是靠**「后端把 ID 输出成 string」这条隐性约定**在安全工作。标准解法就是当年那个做法——
**后端把 ID 序列化成 string**，JS 收到 `"186429408913933113"` 按字符串处理，一位不差。这比让前端全局
上大整数库简单得多。

### 2.2 于是雪花的「短」缩水了

绕了一圈，为了前端安全，雪花 ID 被迫从数字转成了字符串。这时候再回头看它的卖点：

- 「有序」还在；
- 「短」呢？在前端 JSON 传输层，它从「一个数字」变成了「19 字符的字符串」，相对 UUID 的 36 字符，
  差距只剩 **19 vs 36**——而它那串机器 ID 协调、碰撞隐患、init panic、时钟成本**一分没省**。

**雪花「短、有序」的好处，为了前端安全被迫转 string 后已经缩水，而伺候成本原封不动。** 这是整个选型
里最关键的一次「重新算账」。

## 3. UUID 版本澄清：为什么是 v7（不是 v4 / v8）

一提 UUID，很多人的第一反应是 v4——那个全随机的版本，也是拖累它名声的元凶。但 UUID 不止一种。
**RFC 9562**（2024-05 发布）一次性定义了 v6 / v7 / v8，其中做主键该用的是 **v7**：

| 版本 | 结构 | 有序？ | 做主键 |
|------|------|--------|--------|
| v4 | 全随机 128 位 | 否 | ❌ 随机键频繁页分裂、拖慢写入、膨胀索引 |
| **v7** | **48 位毫秒时间戳 + 随机位** | **是** | **✅ 就是它** |
| v8 | 自定义/实验布局（厂商自填 128 位） | 看你怎么填 | ❌ 无标准生成算法、无开箱生成器 |

- **v7 时间有序**，做主键时新记录的键落在 **B-tree 最右端（像自增一样）**，不像 v4 全随机那样到处插、
  频繁页分裂。有序性带来的收益是实打实的：v7 主键索引比 v4 随机键**小约 26–27%**，有序扫描也快数倍。
- **v8 不是 v7 的升级**，是 RFC 留给厂商「更自由地自定义布局」用的实验版——PostgreSQL 没有 `uuidv8()`
  函数、`google/uuid` 也不提供开箱生成（要你自己填满 128 位布局）。所以「记得好像有个更新的 v8」这个
  印象是对的，但它不是主键要走的方向。

**论主键，v7 就是终点**，不存在「再等等有没有更好的版本」。

## 4. 纠正「UUID 又长又占地方」的直觉

换 UUID 时脑子里冒出的第一个担忧通常是：36 个字符，比雪花那串数字还长，多占地方吧？

这个直觉**只在前端 JSON 传输层成立**。主键真正的成本大头在**数据库怎么存、索引怎么比**，而这一层的
结论恰恰相反。核心一句话：

> **「36 字符」只是 UUID 序列化成 JSON 给前端时的样子，不是它在数据库里的样子。**

PostgreSQL 有**原生 `uuid` 类型**，它在磁盘上是 **16 字节二进制**（一个 128 位整数），不是 36 字符文本。
把四种存法摆到一张表上算清楚：

| 存法 | 磁盘/行 | 索引比较 | 写入局部性 | 前端 JS 安全 |
|------|---------|----------|------------|--------------|
| `varchar(20)` 雪花串（现状） | ~21 字节（含长度头） | 文本逐字节 | 有序 | 需转 string |
| **原生 `uuid`(v7)** | **16 字节** | **二进制整数，更快** | **有序** | **天生 string** |
| `varchar(36)` 存 UUID（反例） | ~37 字节 | 文本逐字节 | 有序 | 天生 string |
| `bigint`（对照） | 8 字节 | 整数，最快 | 有序 | 有 2^53 问题 |

两个反直觉的结论：

- **原生 `uuid`（16 字节）比现状 `varchar(20)` 雪花串（~21 字节）还小、索引还更快**（二进制整数比较
  vs 文本逐字节）。换 v7 不是「花复杂度换省一半长度」，是**库内更小更快**。
- **千万别用 `varchar(36)` 存 UUID**。hex 字符串每个字节只编码 4 bit 信息，用 36 字符文本去存 128 位
  数据要占 ~37 字节，**白白浪费近一倍空间**，还退回文本比较——这是最差的组合。

「长」这件事，只体现在前端 JSON 传输和日志字节上，每个 ID 多十几字节，对一个后台管理系统可以忽略。

## 5. 字段怎么设计：列写 `uuid`，别写 `varchar`

上一节的结论落到 DDL 上就一句：**列类型写 `uuid`，不写 `varchar(36)`**。这引出两个最容易误解的点。

### 5.1 「字符串怎么就变成了 16 字节？」——PG 自动转，应用层无感

这是最普遍的直觉困惑：我 INSERT 的明明是字符串，怎么就存成 16 字节数字了？难道要自己转？

**不用。`uuid` 类型「对外是文本、对内是二进制」，两者互转由 PostgreSQL 自动完成，应用层完全碰不到那
16 字节。** 数据流是这样的：

```
应用层 INSERT 传入字符串   "0192f8c2-7c3e-7xxx-xxxx-xxxxxxxxxxxx"
        │  PG 解析该文本，压缩成 16 字节二进制
        ▼
磁盘 / 索引：  16 字节  （128 位整数，二进制比较）
        │  SELECT 时 PG 把 16 字节格式化回 36 字符文本
        ▼
应用层收到字符串   "0192f8c2-7c3e-7xxx-xxxx-xxxxxxxxxxxx"
```

你 INSERT 的是字符串，SELECT 出来还是格式一模一样的字符串。**「16 字节」只发生在 PG 内部**，是它对
`uuid` 类型的存储优化，省空间、加快索引比较全是**数据库白送的，不用写一行转换代码**。

> ⚠️ 前提是列类型必须是 `uuid`。如果错用 `varchar(36)`，PG 就**不做**这层压缩，老老实实按文本存 37
> 字节——省空间的好处直接没了。这就是 5.1 的标题为什么强调「别写 varchar」。

### 5.2 以后换 PG17 及以下，要退回 varchar 吗？

不用。这里要把两样东西分清楚——**「`uuid` 列类型」和「`uuidv7()` 生成函数」不是一回事**：

| 东西 | 起始版本 | PG17 及以下 |
|------|----------|-------------|
| `uuid` **列类型** | 远早于 18（PG 很早就内建） | ✅ 正常，照样 16 字节二进制存储 |
| `gen_random_uuid()`（v4 生成函数） | PG13+ | ✅ 有 |
| `uuidv7()` **生成函数** | **仅 PG18+** | ❌ 报错 `function uuidv7() does not exist` |

```sql
-- ❌ 拿到 PG17 会炸：function uuidv7() does not exist
config_id uuid PRIMARY KEY DEFAULT uuidv7()

-- ✅ PG13+ 都能跑：列仍是原生 uuid，只是不让库自动生成、改由应用层生成
config_id uuid PRIMARY KEY
```

**炸的是 `DEFAULT uuidv7()` 这个默认值表达式，不是 `uuid` 列类型本身。** 降级只需去掉 DEFAULT、改由
应用层生成 ID，**列类型永远是 `uuid`，绝不退回 `varchar`**。`varchar(36)` 只有那种根本不支持 `uuid`
类型的异构数据库才被迫用——PostgreSQL 任何在用版本都不属于此列。

## 6. 落地：两条生成路径

UUIDv7 有序、列用 `uuid` 省空间快索引，这些是「为什么换」。接下来是「怎么生成」——两条路都行、都是 v7，
按需要选。

### 6.1 方式 A：数据库端 `DEFAULT uuidv7()`（PG18，Go 侧啥都不写）

```sql
-- backend/migrations/000001_ddl_init_schema.up.sql
CREATE TABLE system_config (
    config_id uuid PRIMARY KEY DEFAULT uuidv7(),   -- UUIDv7 主键，PG18 库端自动生成
    config_key VARCHAR(100) NOT NULL,
    config_value TEXT NOT NULL,
    -- ...
);
```

INSERT 不用传 `config_id`，数据库自己填。**最省心**，代价是**绑 PG18**（`uuidv7()` 函数是 PG18 新增）。

### 6.2 方式 B：应用层生成（不绑数据库版本）

```go
// backend/pkg/utils/idgen/idgen.go
package idgen

import "github.com/google/uuid"

// NewV7String 生成一个 UUIDv7 字符串（时间有序、做主键索引友好）。
func NewV7String() string {
    return uuid.Must(uuid.NewV7()).String()
}
```

列类型仍是 `uuid`，只是 ID 由 Go 生成后插入。这样换到 PG17 及以下也不受影响——反正库端不管生成，只管
存储与返回。

`google/uuid` **v1.6.0+** 对同一毫秒内的多次 `NewV7()` 调用加了**互斥锁保证单调递增**（通过递增
随机位里的 12 位计数器），批量生成天然有序，不用自己操心时钟回拨或重复问题。

### 6.3 可移植性与本项目选择

跨 PG 版本兼容性对比：

| 方案 | 绑 PG18？ | 降级到 PG17 怎么办 |
|------|----------|-------------------|
| 方式 A：`DEFAULT uuidv7()` | **是** | 去掉 DEFAULT，走方式 B |
| 方式 B：应用层 `uuid.NewV7()` | **否** | 无需任何改动（列一直是 `uuid`） |
| `uuid` **列类型本身** | **否** | PG 很早就内建，任何在用版本都支持 |

**本项目已落地的选择**：主键走**方式 A**（PG18 库端 `DEFAULT uuidv7()` 自动生成），`idgen` 保留**方式 B**
的一行封装做应用层备用——**「用的时候再调，不用就不调」**。

数据库已锁定 PG18、切换窗口最佳（单表单 ID 列无外键、业务代码全空），直接上库端生成最省心。万一哪天
需要在应用层预生成 ID（如批量种子、跨表外键预填），idgen 那一行备用方法立刻能用。

## 7. Go / GORM 落地与值流向

现在来看 Go 侧怎么对接——核心是 GORM Gen 的 database-first 代码生成。

### 7.1 生成的 model 字段类型

方式 A（库端 `DEFAULT uuidv7()`）下，`backend/migrations/000001` 建完表、跑 `make gen-db` 后，
GORM Gen 反射 `system_config` 表的 `config_id uuid DEFAULT uuidv7()` 列，生成的结构是：

```go
// backend/internal/dal/model/system_config.gen.go
type SystemConfig struct {
    ConfigID    string `gorm:"column:config_id;type:uuid;primaryKey;default:uuidv7()" json:"config_id"`
    ConfigKey   string `gorm:"column:config_key;type:varchar(100);not null" json:"config_key"`
    ConfigValue string `gorm:"column:config_value;type:text;not null" json:"config_value"`
    // ...
}
```

注意字段类型是 `string`（不是 `uuid.UUID`）——这符合本项目硬规则 `dto-id-type.md`：**所有 ID 字段用
`string`，禁止 `int64`。** GORM tag 里的 `type:uuid` 告诉 GORM「列类型是数据库原生 uuid」，但 Go
结构字段是 `string`。

### 7.2 为什么会自动生成 string？——gen 的类型映射

PG 原生 `uuid` 列**不会像 `varchar` 那样自动映射成 Go `string`**。默认行为下，Gen 会把 `uuid` 列
生成为 `uuid.UUID` 类型（github.com/google/uuid 的 `[16]byte` 包装），违反 `dto-id-type` 规则。

本项目在 `backend/scripts/gen-from-db/main.go` 里加了一条**全局类型映射**：

```go
// backend/scripts/gen-from-db/main.go（片段）
genCfg := gen.Config{
    OutPath:      "./internal/dal/query",
    ModelPkgPath: "./internal/dal/model",
    // ...
}

// PG 原生 uuid 列默认不映射为 Go string，显式映射以符合 dto-id-type 规则（所有 ID 用 string）。
genCfg.WithDataTypeMap(map[string]func(gorm.ColumnType) string{
    "uuid": func(gorm.ColumnType) string { return "string" },
})

g := gen.NewGenerator(genCfg)
g.UseDB(db)
// 反射活库所有表，逐表 GenerateModel 后 ApplyBasic + Execute
// ...
```

**有了这一行，所有 `uuid` 列都生成为 Go `string` 字段**。`make gen-db` 实际执行的是这个脚本，所以
每次改表结构后重跑，生成的 model 自动符合项目规则。

### 7.3 完整值流向（DB → 前端，全程 string）

数据库列 `uuid` + 方式 A（`DEFAULT uuidv7()`）下，一条记录从插入到前端接收的完整链路：

```
1. INSERT 不传 config_id
   → PG 触发 DEFAULT uuidv7()，库内生成 16 字节 v7
   → 存进磁盘（16 字节二进制）

2. SELECT
   → PG 从磁盘读 16 字节，格式化回 36 字符文本
   → pgx（GORM 底层驱动）Scan 成 Go string（已是文本，直接赋值）
   → GORM 填进 model.SystemConfig.ConfigID（string 字段）

3. JSON 序列化
   → json.Marshal(response) 把 ConfigID 输出为 JSON 字符串 "0192f8c2-..."
   → axios 收到 {"config_id": "0192f8c2-..."}
   → TypeScript 解析后 config_id: string（前端类型声明本就是 string）
```

**全链路碰不到那 16 字节**。前端和 Go 代码从头到尾只看到字符串，PG 的二进制压缩是库内优化，透明。
因为 `uuid` 天生就是字符串格式（不像雪花那样从数字转过来），所以「前端零改动」——原本就按 string 收发，
现在还是按 string 收发，只是 ID 的内容从 19 位数字串换成了 36 字符 UUID 串，**逻辑上没有任何区别**。

## 8. 要不要封 idgen 包？（YAGNI 与诚实结论）

走到这里，你可能会问：那我还用不用一个 idgen 包？还是直接裸调 `google/uuid`？

先看换芯前后的代码量对比。

### 8.1 换芯前后对比

**换芯前**（sonyflake）：

```go
// backend/pkg/utils/idgen/idgen.go（老版）
package idgen

import (
    "github.com/sony/sonyflake"
    "strconv"
)

var sf *sonyflake.Sonyflake

func init() {
    // 三级兜底：环境变量 → /etc/machine-id → MAC 哈希
    machineID := GetMachineID()
    // epoch 固定为 2026-01-01 00:00:00 UTC，上线后永不能改
    sf = sonyflake.NewSonyflake(sonyflake.Settings{
        MachineID: func() (uint16, error) { return machineID, nil },
        StartTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
    })
    if sf == nil { panic("failed to create sonyflake") }
}

func GenerateUUID() string {
    id, err := sf.NextID()
    if err != nil { panic("failed to generate ID: " + err.Error()) }
    return strconv.FormatUint(id, 10)
}

// machine_id.go（另一个文件，60+ 行三级兜底逻辑）
```

**换芯后**（UUIDv7）：

```go
// backend/pkg/utils/idgen/idgen.go（新版）
package idgen

import "github.com/google/uuid"

// NewV7String 生成一个 UUIDv7 字符串（时间有序、做主键索引友好）。
//
// 备用封装：主键默认走数据库端 uuid 列的 DEFAULT uuidv7()（PG18 库端自动生成），
// 插入时无需传 ID；仅当需要在应用层预生成 ID（如批量种子、跨表外键预填）时才调用本方法。
//
// 无 error 返回：uuid.NewV7 仅在读系统随机源（crypto/rand）失败时返回 error，
// 正常运行的机器上几乎不可能发生，故用官方 uuid.Must 做 fail-fast。
// 相比原 sonyflake 方案，这里没有机器 ID 协调 / 时钟回拨 / init panic 这些概念。
func NewV7String() string {
    return uuid.Must(uuid.NewV7()).String()
}
```

机器 ID 协调、`init()` panic、epoch、时钟、`machine_id.go` 60+ 行**全部消失**，塌缩成**一行函数体**。

### 8.2 那还留着干嘛？

诚实地说，**全新项目直接裸调 `google/uuid` 即可，不必封这一层**。`uuid` 库本身就是那层封装，一行的
东西再包一层是过度设计。

**本项目保留 `idgen` 这个薄壳，是务实考虑**：方式 A（库端 `DEFAULT uuidv7()`）是主路径，Go 侧本来
啥都不用写，但万一哪天需要在应用层预生成 ID（批量种子、跨表外键预填、或者降级到 PG17 以下），这一行
封装立刻能用——不必临时去翻 `google/uuid` 文档、想「怎么一行取 string」。

顺带说个命名细节：老版函数叫 `GenerateUUID`，是雪花时代留下的名字——那会儿它返回的其实是雪花数字串，
叫「UUID」本就名不副实。换芯后正好把它改成 `NewV7String`（生成 v7 的 string 形式，和 `google/uuid`
自己的 `NewString()` 命名习惯对齐）。**敢直接改名而不做兼容别名，正是因为前面确认过 backend 零业务
调用者**——锁定程度低，改名零成本。换作老项目那种几十处调用的场景，就得掂量是留旧名还是全局改了。

开发友好度对比：

| 方案 | 机器 ID | init panic | 时钟/epoch | 单调性 | 代码行数 |
|------|---------|------------|-----------|--------|---------|
| sonyflake | 需三级兜底 + 碰撞隐患 | 有（init 失败挂） | 需管 | 靠时钟 + 序列 | ~100 行（含 machine_id.go） |
| v7 应用层 | 无 | 无 | 无 | v1.6.0 mutex 自动保证 | 1 行 |
| v7 库端 `DEFAULT` | 无 | 无 | 无 | PG 内部保证 | **0 行**（Go 侧啥都不写） |

**结论**：封不封 idgen 不影响 v7 本身的优势——它已经把「伺候 ID」的那些事从代码里删干净了。留一行
备用是实用主义；全新起步直接裸调是 YAGNI 原则。两个都对，看项目情况。


## 9. 落地改动清单与验证

本项目 2026-07 在 `backend/` 实际改动的文件（**范围仅 `backend/`，`backend-rbac/` 未动**）：

| 文件 | 改动 |
|------|------|
| `migrations/000001_ddl_init_schema.up.sql` | `config_id VARCHAR(20)` → `config_id uuid PRIMARY KEY DEFAULT uuidv7()` |
| `migrations/000002_data_base_config.up.sql` | 种子 INSERT 去掉硬编码 `config_id`，交给库端 `DEFAULT` 填 |
| `scripts/gen-from-db/main.go` | 加 `WithDataTypeMap`：`uuid` → `string` 全局映射 |
| `pkg/utils/idgen/idgen.go` | 换芯 sonyflake → `uuid.Must(uuid.NewV7()).String()` |
| `pkg/utils/idgen/machine_id.go` | **删除**（三级机器 ID 兜底不再需要） |
| `go.mod` | `google/uuid` v1.3.0 → v1.6.0；移除 `sony/sonyflake` |

种子去掉硬编码 ID 值得单独说一句。项目规则 `migration-data-convention.md` 要求「ID 硬编码进 SQL」，
前提是**跨表外键需要一个稳定已知的 ID**。而 `config_id` 无任何外键引用它，所以这里**合理豁免**——
让库端 `DEFAULT uuidv7()` 自动生成，反而正好演示了「方式 A」：

```sql
-- backend/migrations/000002_data_base_config.up.sql
-- config_id 由库端 DEFAULT uuidv7() 生成，不再硬编码（该列无外键引用，合理豁免 ID 硬编码规则）
INSERT INTO system_config (config_key, config_value, remark, created_at, updated_at, deleted_at) VALUES
  ('site_name', 'Admin', '站点名称', ..., ..., 0),
  ('site_status', 'active', '站点状态', ..., ..., 0)
ON CONFLICT (config_key, deleted_at) DO UPDATE SET
  config_value = EXCLUDED.config_value, remark = EXCLUDED.remark, updated_at = EXCLUDED.updated_at;
```

仍然幂等（冲突键是 `config_key, deleted_at`），可重复跑。

验证步骤（本会话实际执行、全绿）：

```bash
# 1. DB 往返：重建库，验证 uuidv7() DDL + 无 config_id 的 seed 能跑通
cd backend && make migrate-reset

# 2. 重新反射生成 model，确认 config_id 生成为 string
make gen-db
grep config_id internal/dal/model/system_config.gen.go
# → ConfigID string `gorm:"...;type:uuid;primaryKey;default:uuidv7()"`

# 3. 编译 + 静态检查
go build ./... && go vet ./...

# 4. 依赖干净
grep sonyflake go.mod        # 应无输出
grep 'google/uuid' go.mod    # → v1.6.0
```

DB 往返跑通后，seed 两行拿到了库端生成的 v7 config_id（形如 `019f4a92-d3ef-785f-...`，版本位确认为 7）。

## 10. 承认的 tradeoff（不回避）

换 UUIDv7 不是没有代价，只是这些代价对一个后台管理系统可以接受：

- **库内 16 字节 > bigint 的 8 字节**。如果极致追求存储与索引效率、且能接受在应用层解决 JS 精度问题，
  纯 `bigint` 自增仍是最省的。但那会把「2^53 前端精度」和「分布式不友好」两个问题带回来，对多租户
  SaaS 不划算。
- **前端 JSON 与日志每个 ID 多十几字节**（36 vs 19 字符）。对传输量、日志体积有极微影响，可忽略。
- **可读性**：雪花数字串肉眼能粗略看出「谁先谁后」（数字递增）；UUIDv7 的前缀也是毫秒时间戳，但
  肉眼看 `0192f8c2-...` 不直观。对调试影响很小——两者都能按主键排序看顺序——但值得诚实一提。
- **方式 A 绑 PG18 引擎**。库端 `DEFAULT uuidv7()` 用了 PG18 专有函数，换数据库引擎（如 MySQL 8 是
  `UUID_TO_BIN()` 一套）或降级 PG17 需要改。缓解办法就是 §6 的方式 B：改走应用层生成，列类型不变。

## 踩坑 / 决策清单

| 坑 / 决策 | 表现 | 最终做法 |
|-----------|------|----------|
| 机器 ID 协调 | 雪花三级兜底 + 「300 台 50% 碰撞」隐患 | v7 零协调，删掉 machine_id.go |
| init panic | 机器 ID / 实例创建失败，服务起不来 | v7 无此启动失败面，Must fail-fast 仅随机源失败 |
| JS 2^53 精度 | 64 位 ID 前端 `JSON.parse` 静默丢尾数 → 404 / 删错行 | ID 天生 string，隐患消失 |
| 「UUID 又长又占地方」 | 直觉以为 36 字符很占空间 | 库内原生 `uuid` 16 字节，比雪花 varchar(20) 还小 |
| 用 `varchar(36)` 存 UUID | 浪费近一倍空间 + 退回文本比较 | 列类型必须写 `uuid`，别写 varchar |
| `DEFAULT uuidv7()` 绑版本 | 拿到 PG17 报 `function uuidv7() does not exist` | 炸的是 DEFAULT 不是列类型，降级去掉 DEFAULT 走应用层生成 |
| `uuid` 列不自动映射 string | Gen 默认生成 `uuid.UUID`，违反 dto-id-type | gen-from-db 加 `WithDataTypeMap` uuid→string |
| 要不要封 idgen 包 | 一行的东西再包一层是过度设计 | 全新起步裸调；本项目留一行备用并借零调用者之机改名 GenerateUUID → NewV7String |
| v7 error 处理 | `NewV7()` 返回 error | 仅随机源失败（几乎不可能），`uuid.Must` fail-fast |

## 参考

- [01 ID生成方案选型：雪花 vs UUID](../saas-backend/research/idgen/01-ID生成方案选型-雪花vs-UUID.md)
- [02 UUIDv7 落地细节：前端影响与生成方式](../saas-backend/research/idgen/02-UUIDv7落地细节-前端影响与生成方式.md)
- [03 UUIDv7 存储机制与落地：PG18 字段设计与文本二进制转换](../saas-backend/research/idgen/03-UUIDv7存储机制与落地-PG18字段设计与文本二进制转换.md)
- [RFC 9562 — UUIDs (IETF, 2024-05)](https://datatracker.ietf.org/doc/rfc9562/)
- [PostgreSQL 18 Release Notes](https://www.postgresql.org/docs/current/release-18.html)
- [google/uuid](https://github.com/google/uuid)

*系列教程：[saas-backend](../saas-backend/README.md)*
