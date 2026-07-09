# Repository 与事务约定

> 2026-07-08 定型。repo 层与事务的唯一约定，AI 建 repo/service 照此。
> 选型论证（为什么重建派、Ent 对比、社区无多数派）见 `docs/saas-backend/research/database/07-repo层与事务范式选型.md`，日常不必读。

## 结论：重建派

- repo 是结构体，构造吃 `*gorm.DB`，内部 `query.Use(db)` 得到 `q`（与 老项目 B 一致）
- 不定义 interface（不 mock，靠集成测试）
- 所有事务：service 层 `s.db.Transaction(func(tx *gorm.DB))` + 闭包内 `NewXxxRepo(tx)` 重建，调 repo 封装好的方法。repo 内部不开事务

## 规则 1：repo 构造吃 db，方法签名不带 q

```go
type CustomResourcePersonRepo struct {
    db *gorm.DB
    q  *query.Query
}

func NewCustomResourcePersonRepo(db *gorm.DB) *CustomResourcePersonRepo {
    return &CustomResourcePersonRepo{db: db, q: query.Use(db)}
}

func (r *CustomResourcePersonRepo) BatchDeleteByIDs(ctx context.Context, ids []string) error {
    _, err := r.q.CustomResourcePerson.WithContext(ctx).Where(r.q.CustomResourcePerson.ID.In(ids...)).Delete()
    return err
}
```

## 规则 2：service 持 db + 一组 base repo

```go
type CustomResourceService struct {
    db         *gorm.DB
    personRepo *repository.CustomResourcePersonRepo
    imageRepo  *repository.CustomResourceImageRepo
}

func NewCustomResourceService(db *gorm.DB) *CustomResourceService {
    return &CustomResourceService{
        db:         db,
        personRepo: repository.NewCustomResourcePersonRepo(db),
        imageRepo:  repository.NewCustomResourceImageRepo(db),
    }
}
```

## 规则 3：非事务操作直接用 base repo

```go
return s.personRepo.ListByCategory(ctx, categoryID)
```

## 规则 4：所有事务都在 service 层重建 repo，不在 repo 内开事务

无论单 repo 多步（如「先删后插」）还是跨多 repo，一律在 service 层 `s.db.Transaction` + 重建 repo，**调用 repo 已封装好的原子方法**。

**不要在 repo 里写 `r.q.Transaction` 的组合方法**（如 `Replace`）——repo 方法已经是封装好的单元（`DeleteByCategory` / `BatchCreate`），组合它们是 service 的事，在 repo 里用裸 `tx.Xxx` 重写一遍等于重复封装。

单 repo 多步也走同一套（先删后插）：

```go
func (s *CustomResourceService) ReplacePersons(ctx context.Context, categoryID string, persons []*model.CustomResourcePerson) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        txPersonRepo := repository.NewCustomResourcePersonRepo(tx)
        if err := txPersonRepo.DeleteByCategory(ctx, categoryID); err != nil {
            return err
        }
        return txPersonRepo.BatchCreate(ctx, persons)
    })
}
```

## 规则 5：跨多 repo 事务同样在 service 层重建 repo（核心）

```go
func (s *CustomResourceService) DeleteCategory(ctx context.Context, categoryID string, personIDs []string) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        txPersonRepo := repository.NewCustomResourcePersonRepo(tx)
        txImageRepo := repository.NewCustomResourceImageRepo(tx)
        if len(personIDs) > 0 {
            if err := txImageRepo.BatchDeleteByPersonIDs(ctx, personIDs); err != nil {
                return xerr.Wrap(xerr.ErrInternal.Code, "删除照片失败", err)
            }
        }
        return txPersonRepo.BatchDeleteByIDs(ctx, personIDs)
    })
}
```

## 规则 6：事务闭包内禁用 s.xxxRepo（最易踩的坑）

闭包内全部用 `tx` 重建的 repo。混用 `s.imageRepo`（base 版）会让该操作**静默跑在事务外**，回滚时不回滚，数据不一致。**约定：事务版 repo 变量名以 `tx` 开头。**

## 规则 7：事务闭包内 error 必须透传

任何一步 error 必须 `return`，中途吞掉会导致本该回滚的事务被提交。用 `xerr.Wrap` 包装（见 service-patterns.md 规则 8）。

---

**最后更新**：2026-07-08
