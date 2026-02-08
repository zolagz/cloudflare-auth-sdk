# Cloudflare Authentication SDK

一个基于 Cloudflare Workers KV 的 Golang 身份认证 SDK，提供用户注册、登录和 JWT 令牌管理功能。

## 功能特性

- ✅ 用户注册和登录
- ✅ 密码安全哈希 (bcrypt)
- ✅ JWT 令牌生成和验证
- ✅ Cloudflare Workers KV 存储管理
- ✅ 完整的错误处理
- ✅ RESTful API 服务器示例
- ✅ 类型安全的 API 设计

## 项目结构

```
.
├── cmd/
│   ├── example/          # 使用示例
│   │   └── main.go
│   └── server/           # HTTP 服务器
│       └── main.go
├── internal/
│   ├── auth/            # 认证模块
│   │   ├── models.go    # 数据模型
│   │   └── service.go   # 认证服务
│   ├── config/          # 配置管理
│   │   └── config.go
│   ├── errors/          # 错误处理
│   │   └── errors.go
│   └── kv/              # KV 存储客户端
│       └── client.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 安装

### 前置要求

- Go 1.22 或更高版本
- Cloudflare 账号
- Workers KV 命名空间

### 依赖安装

```bash
go mod download
```

## 配置

创建 `.env` 文件并设置以下环境变量：

```bash
# Cloudflare API 认证（选择其一）
# 方式1: 使用 API Token (推荐)
CLOUDFLARE_API_TOKEN=your_api_token_here

# 方式2: 使用 API Key + Email
CLOUDFLARE_API_KEY=your_api_key_here
CLOUDFLARE_EMAIL=your_email@example.com

# Cloudflare 账号信息
CLOUDFLARE_ACCOUNT_ID=your_account_id
CLOUDFLARE_NAMESPACE_ID=your_namespace_id

# JWT 配置
JWT_SECRET=your_secret_key_here

# 服务器配置
SERVER_PORT=8080
```

### 获取 Cloudflare 凭证

1. **API Token** (推荐):
   - 访问 [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens)
   - 创建一个新的 API Token
   - 确保包含 "Workers KV Storage" 权限

2. **Account ID**:
   - 在 Cloudflare Dashboard 中，选择你的账户
   - 在右侧栏可以找到 Account ID

3. **Namespace ID**:
   - 进入 Workers & Pages → KV
   - 创建一个新的 KV 命名空间或使用现有的
   - 复制 Namespace ID

## 使用方法

### 1. 快速开始示例

运行示例代码查看所有功能：

```bash
make run
# 或
go run cmd/example/main.go
```

### 2. 启动 HTTP 服务器

```bash
go run cmd/server/main.go
```

服务器将在 `http://localhost:8080` 上运行。

### 3. API 端点

#### 注册用户

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

响应:
```json
{
  "message": "User registered successfully",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  }
}
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123!"
  }'
```

响应:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-02-09T12:00:00Z",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  }
}
```

#### 验证令牌

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

#### 获取用户信息

```bash
curl -X GET http://localhost:8080/user \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 健康检查

```bash
curl http://localhost:8080/health
```

## 代码示例

### 初始化客户端

```go
package main

import (
    "context"
    "log"
    
    "github.com/cloudflare/cloudflare-go"
    "github.com/zolagz/cloudflare-auth-sdk/internal/auth"
    "github.com/zolagz/cloudflare-auth-sdk/internal/config"
    "github.com/zolagz/cloudflare-auth-sdk/internal/kv"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建 Cloudflare API 客户端
    api, err := cloudflare.NewWithAPIToken(cfg.CloudflareAPIToken)
    if err != nil {
        log.Fatal(err)
    }
    
    // 初始化 KV 客户端
    kvClient := kv.NewClient(api, cfg.AccountID, cfg.NamespaceID)
    
    // 初始化认证服务
    authService := auth.NewService(kvClient, cfg.JWTSecret, cfg.JWTExpiration)
}
```

### 用户注册

```go
registerReq := &auth.RegisterRequest{
    Email:    "user@example.com",
    Password: "SecurePassword123!",
}

user, err := authService.Register(context.Background(), registerReq)
if err != nil {
    log.Fatal(err)
}

log.Printf("User registered: %s", user.Email)
```

### 用户登录

```go
loginReq := &auth.LoginRequest{
    Email:    "user@example.com",
    Password: "SecurePassword123!",
}

loginResp, err := authService.Login(context.Background(), loginReq)
if err != nil {
    log.Fatal(err)
}

log.Printf("Token: %s", loginResp.Token)
log.Printf("Expires: %s", loginResp.ExpiresAt)
```

### 验证令牌

```go
claims, err := authService.ValidateToken(token)
if err != nil {
    log.Fatal(err)
}

log.Printf("User ID: %s", claims.UserID)
log.Printf("Email: %s", claims.Email)
```

### KV 操作

```go
// 存储键值对
err := kvClient.Set(ctx, "mykey", []byte("myvalue"), nil)

// 带过期时间的存储
err = kvClient.Set(ctx, "mykey", []byte("myvalue"), &kv.WriteOptions{
    ExpirationTTL: 3600, // 1小时
})

// 读取值
value, err := kvClient.Get(ctx, "mykey")

// 列出键
keys, err := kvClient.List(ctx, "prefix:", 100)

// 删除键
err = kvClient.Delete(ctx, "mykey")

// 批量删除
err = kvClient.DeleteBulk(ctx, []string{"key1", "key2", "key3"})
```

## 开发

### 运行测试

```bash
make test
```

### 代码格式化

```bash
make fmt
```

### 代码检查

```bash
make lint
```

### 构建

```bash
make build
```

生成的二进制文件将位于 `bin/` 目录。

## 安全考虑

1. **密码存储**: 使用 bcrypt 进行密码哈希，成本因子为默认值（10）
2. **JWT 密钥**: 确保 `JWT_SECRET` 足够强且保密
3. **HTTPS**: 生产环境中务必使用 HTTPS
4. **令牌过期**: JWT 令牌默认 24 小时过期
5. **环境变量**: 不要将 `.env` 文件提交到版本控制

## 错误处理

SDK 使用自定义错误类型提供详细的错误信息：

```go
import apperrors "github.com/zolagz/cloudflare-auth-sdk/internal/errors"

// 检查特定错误
if errors.Is(err, apperrors.ErrUserNotFound) {
    // 处理用户不存在的情况
}

// 获取应用错误详情
var appErr *apperrors.AppError
if errors.As(err, &appErr) {
    log.Printf("Operation: %s, Code: %d, Message: %s", 
        appErr.Op, appErr.Code, appErr.Message)
}
```

## 性能优化

- 使用连接池管理 API 连接
- KV 操作支持批量删除
- JWT 令牌签名使用 HMAC-SHA256
- 密码哈希使用 bcrypt 默认成本（平衡安全性和性能）

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [Cloudflare Workers KV](https://developers.cloudflare.com/workers/runtime-apis/kv/)
- [Cloudflare Go SDK](https://github.com/cloudflare/cloudflare-go)
- [JWT Go 实现](https://github.com/golang-jwt/jwt)

## 作者

Difyz Team

## 更新日志

### v1.0.0 (2026-02-08)
- ✅ 初始版本
- ✅ 用户注册和登录
- ✅ JWT 令牌管理
- ✅ Cloudflare KV 集成
- ✅ HTTP 服务器示例
