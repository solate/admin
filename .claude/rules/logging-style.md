# 日志风格规范

## 规则
所有日志调用必须写在一行，禁止多行链式调用

## 代码示例

### ❌ 错误
```go
log.Info().
    Str("url", url).
    Interface("request", req).
    Int("count", len(items)).
    Msg("操作完成")
```

### ✅ 正确
```go
log.Info().Str("url", url).Interface("request", req).Int("count", len(items)).Msg("操作完成")
```

## 适用范围
- 所有使用 zerolog 的日志调用
- `log.Info()`, `log.Error()`, `log.Warn()`, `log.Debug()` 等

---

**最后更新**：2026-03-27
