# Cloudflare Go SDK v6 升级指南

## 升级概述

本项目已成功升级到 `cloudflare-go v6.6.0`，使用最新的 API 接口。

## 主要变更

### 1. 依赖更新
```go
// 之前 (v0.86.0)
github.com/cloudflare/cloudflare-go v0.86.0

// 现在 (v6.6.0)
github.com/cloudflare/cloudflare-go/v6 v6.6.0
```

### 2. 客户端初始化

**之前:**
```go
var api *cloudflare.API
if cfg.CloudflareAPIToken != "" {
    api, err = cloudflare.NewWithAPIToken(cfg.CloudflareAPIToken)
} else {
    api, err = cloudflare.New(cfg.CloudflareAPIKey, cfg.CloudflareEmail)
}
```

**现在:**
```go
import (
    cloudflare "github.com/cloudflare/cloudflare-go/v6"
    "github.com/cloudflare/cloudflare-go/v6/option"
)

var client *cloudflare.Client
if cfg.CloudflareAPIToken != "" {
    client = cloudflare.NewClient(
        option.WithAPIToken(cfg.CloudflareAPIToken),
    )
} else {
    client = cloudflare.NewClient(
        option.WithAPIKey(cfg.CloudflareAPIKey),
        option.WithAPIEmail(cfg.CloudflareEmail),
    )
}
```

### 3. KV 操作 API 变更

**之前:**
```go
// Get
value, err := api.GetWorkersKV(ctx, cloudflare.AccountIdentifier(accountID), 
    cloudflare.GetWorkersKVParams{
        NamespaceID: namespaceID,
        Key:         key,
    })

// Set
_, err := api.SetWorkersKV(ctx, cloudflare.AccountIdentifier(accountID), 
    cloudflare.SetWorkersKVParams{
        NamespaceID:   namespaceID,
        Key:           key,
        Value:         value,
        ExpirationTTL: ttl,
    })

// List
keys, _, err := api.ListWorkersKVs(ctx, cloudflare.AccountIdentifier(accountID),
    cloudflare.ListWorkersKVsParams{
        NamespaceID: namespaceID,
    })
```

**现在:**
```go
import (
    cloudflare "github.com/cloudflare/cloudflare-go/v6"
    "github.com/cloudflare/cloudflare-go/v6/kv"
)

// Get - 返回 *http.Response，需要读取 Body
resp, err := client.KV.Namespaces.Values.Get(ctx, namespaceID, key, 
    kv.NamespaceValueGetParams{
        AccountID: cloudflare.F(accountID),
    })
defer resp.Body.Close()
value, _ := io.ReadAll(resp.Body)

// Set - 使用 cloudflare.F() 包装参数
_, err := client.KV.Namespaces.Values.Update(ctx, namespaceID, key, 
    kv.NamespaceValueUpdateParams{
        AccountID:     cloudflare.F(accountID),
        Value:         cloudflare.F(string(value)),
        ExpirationTTL: cloudflare.F(float64(ttl)),
    })

// List - 使用新的列表接口
resp, err := client.KV.Namespaces.Keys.List(ctx, namespaceID, 
    kv.NamespaceKeyListParams{
        AccountID: cloudflare.F(accountID),
        Prefix:    cloudflare.F(prefix),
        Limit:     cloudflare.F(float64(limit)),
    })
```

### 4. 类型变更

- `*cloudflare.API` → `*cloudflare.Client`
- 参数需要使用 `cloudflare.F()` 包装
- `ExpirationTTL` 从 `int` 变为 `float64`
- `Key.Expiration` 从 `*int64` 变为 `float64`
- Get 方法返回 `*http.Response` 而不是直接返回数据

## 测试验证

### 编译测试
```bash
go build ./cmd/example
go build ./cmd/server
```

### 依赖检查
```bash
go list -m github.com/cloudflare/cloudflare-go/v6
# 输出: github.com/cloudflare/cloudflare-go/v6 v6.6.0
```

## 参考示例

查看 `internal/kv/main.go.example` 获取完整的 v6 API 使用示例。

## 注意事项

1. **参数包装**: 所有参数值都需要使用 `cloudflare.F()` 进行包装
2. **响应读取**: Get 方法现在返回 HTTP Response，需要手动读取 Body
3. **类型转换**: 注意 TTL 等参数从 int 变为 float64
4. **错误处理**: 保持原有的错误处理逻辑不变

## 升级时间

- 升级日期: 2026年2月8日
- 版本: v0.86.0 → v6.6.0
