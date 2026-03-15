package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    sdk "github.com/zolagz/cloudflare-auth-sdk"
)

func main() {
    client, err := sdk.NewClient(&sdk.ClientOptions{
        APIToken:     "j-aaaaaaaaaaaaaaaaaaaa",
        AccountID:    "888888888888888",
        NamespaceID:  "your-kv-namespace-id",
        D1DatabaseID: "fffff-xxxx-4563-a137-dddd", // D1 数据库 UUID
        JWTSecret:    "xxxxxxxx-x-xxx-xx-xx",
    })
    if err != nil {
        log.Fatalf("init client: %v", err)
    }

    ctx := context.Background()
    userID := "sss-ss-ssss-8e2f-ss"
    taskCost := int64(1) // 本次任务消耗 50 积分

    // ── 第一步：检查积分是否充足（快速拦截，不写库）──────────────────────────
    if err := client.CheckCredits(ctx, userID, taskCost); err != nil {
        var appErr *sdk.AppError
        if errors.As(err, &appErr) && errors.Is(appErr.Err, sdk.ErrInsufficientCredits) {
            fmt.Printf("❌ 积分不足，拒绝执行任务：%s\n", appErr.Message)
            return
        }
        log.Fatalf("check credits: %v", err)
    }

    // ── 第二步：执行业务逻辑 ──────────────────────────────────────────────────
    fmt.Println("✅ 积分充足，开始执行任务...")
    // ... 调用 AI、执行计算等 ...

    // ── 第三步：原子性扣减积分 + 写流水 ─────────────────────────────────────
    account, err := client.DeductCredits(ctx, sdk.DeductCreditsParams{
        UserID:      userID,
        Amount:      taskCost,
        Service:     "anthropic",
        Model:       "claude-3-5-sonnet",
        TaskID:      "task-uuid-abc123", // 幂等 key，相同 TaskID 不会重复扣减
        Description: "AI chat completion",
    })
    if err != nil {
        var appErr *sdk.AppError
        if errors.As(err, &appErr) {
            switch {
            case errors.Is(appErr.Err, sdk.ErrInsufficientCredits):
                // 并发场景：两步之间余额被其他请求消耗
                fmt.Printf("❌ 积分已被消耗，HTTP %d：%s\n", appErr.Code, appErr.Message)
            default:
                fmt.Printf("❌ 扣减失败，HTTP %d：%s\n", appErr.Code, appErr.Message)
            }
        }
        return
    }

    fmt.Printf("✅ 扣减成功！剩余余额：%d，已消费：%d\n",
        account.Balance, account.TotalSpent)
}