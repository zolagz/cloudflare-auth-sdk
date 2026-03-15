package cloudflare_auth_sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// GetCreditsBalance 查询用户的积分账户余额。
// 若账户不存在（lazy-create 场景），返回 ErrCreditsAccountNotFound。
//
// 示例：
//
//	account, err := client.GetCreditsBalance(ctx, userID)
//	if err != nil { ... }
//	fmt.Println("balance:", account.Balance)
func (c *Client) GetCreditsBalance(ctx context.Context, userID string) (*CreditsAccount, error) {
	const op = "Client.GetCreditsBalance"

	if userID == "" {
		return nil, NewAppError(op, ErrInvalidInput, "userID is required", 400)
	}

	rows, err := c.D1Query(ctx,
		"SELECT user_id, balance, reserved, total_earned, total_spent FROM credits_accounts WHERE user_id = ?",
		userID,
	)
	if err != nil {
		return nil, NewAppError(op, err, fmt.Sprintf("failed to query credits account: %v", err), 500)
	}

	if len(rows) == 0 {
		return nil, NewAppError(op, ErrCreditsAccountNotFound,
			fmt.Sprintf("credits account not found for user %s", userID), 404)
	}

	return rowToCreditsAccount(rows[0])
}

// CheckCredits 检查用户积分是否足够。
// 若余额 < required，返回 ErrInsufficientCredits（HTTP 402）。
//
// 示例：
//
//	if err := client.CheckCredits(ctx, userID, 100); err != nil {
//	    // 积分不足，拒绝执行任务
//	}
func (c *Client) CheckCredits(ctx context.Context, userID string, required int64) error {
	const op = "Client.CheckCredits"

	account, err := c.GetCreditsBalance(ctx, userID)
	if err != nil {
		return err
	}

	available := account.Balance - account.Reserved
	if available < required {
		return NewAppError(op, ErrInsufficientCredits,
			fmt.Sprintf("insufficient credits: have %d available, need %d", available, required), 402)
	}
	return nil
}

// DeductCredits 原子性地扣减用户积分，并写入流水记录。
// 若账户余额不足，UPDATE 的 WHERE 条件不满足，RowsAffected=0，返回 ErrInsufficientCredits (HTTP 402)。
// 采用两次独立 D1Exec（而非 batch）以确保 RowsAffected 被正确返回。
//
// 示例：
//
//	result, err := client.DeductCredits(ctx, DeductCreditsParams{
//	    UserID:      userID,
//	    Amount:      50,
//	    Service:     "anthropic",
//	    Model:       "claude-3-5-sonnet",
//	    TaskID:      uuid.New().String(), // 每次任务生成唯一 ID
//	    Description: "AI chat completion",
//	})
func (c *Client) DeductCredits(ctx context.Context, p DeductCreditsParams) (*CreditsAccount, error) {
	const op = "Client.DeductCredits"

	if p.UserID == "" {
		return nil, NewAppError(op, ErrInvalidInput, "UserID is required", 400)
	}
	if p.Amount <= 0 {
		return nil, NewAppError(op, ErrInvalidInput, "Amount must be positive", 400)
	}
	if p.Service == "" {
		return nil, NewAppError(op, ErrInvalidInput, "Service is required", 400)
	}
	if p.TaskID == "" {
		p.TaskID = uuid.New().String()
	}

	amountStr := strconv.FormatInt(p.Amount, 10)
	deltaStr := strconv.FormatInt(-p.Amount, 10)
	nowStr := strconv.FormatInt(time.Now().UnixMilli(), 10)
	ledgerID := uuid.New().String()

	// ── 第一步：原子扣减 ─────────────────────────────────────────────────────
	// WHERE 条件同时充当余额校验：(balance - reserved) >= amount
	// 若条件不满足（余额不足或账户不存在），RowsAffected = 0。
	updateResult, err := c.D1Exec(ctx,
		`UPDATE credits_accounts
		    SET balance     = balance - ?,
		        total_spent = total_spent + ?,
		        updated_at  = ?
		  WHERE user_id = ? AND (balance - reserved) >= CAST(? AS INTEGER)`,
		amountStr, amountStr, nowStr, p.UserID, amountStr,
	)
	if err != nil {
		return nil, NewAppError(op, err, fmt.Sprintf("failed to deduct credits: %v", err), 500)
	}
	if updateResult.RowsAffected == 0 {
		// 判断是「账户不存在」还是「余额不足」，给出精确错误信息
		account, aerr := c.GetCreditsBalance(ctx, p.UserID)
		if aerr != nil {
			return nil, aerr
		}
		available := account.Balance - account.Reserved
		return nil, NewAppError(op, ErrInsufficientCredits,
			fmt.Sprintf("insufficient credits: have %d available (balance=%d reserved=%d), need %d",
				available, account.Balance, account.Reserved, p.Amount), 402)
	}

	// ── 第二步：写流水（INSERT OR IGNORE 保证幂等，TaskID 重复不会报错）────────
	_, err = c.D1Exec(ctx,
		`INSERT OR IGNORE INTO credits_ledger
		     (id, user_id, delta, balance_after, type, service, model, task_id, description, created_at)
		 SELECT ?, ?, ?,
		        (SELECT balance FROM credits_accounts WHERE user_id = ?),
		        'spend', ?, ?, ?, ?, ?`,
		ledgerID, p.UserID, deltaStr, p.UserID,
		p.Service, p.Model, p.TaskID, p.Description, nowStr,
	)
	if err != nil {
		return nil, NewAppError(op, err, fmt.Sprintf("credits deducted but ledger write failed: %v", err), 500)
	}

	return c.GetCreditsBalance(ctx, p.UserID)
}

// EnsureCreditsAccount 确保用户积分账户存在（首次使用时 lazy-create）。
// 若账户已存在则直接返回，不做修改。
//
// 示例：
//
//	_, err := client.EnsureCreditsAccount(ctx, userID)
func (c *Client) EnsureCreditsAccount(ctx context.Context, userID string) (*CreditsAccount, error) {
	const op = "Client.EnsureCreditsAccount"

	if userID == "" {
		return nil, NewAppError(op, ErrInvalidInput, "userID is required", 400)
	}

	now := strconv.FormatInt(time.Now().UnixMilli(), 10)

	if _, err := c.D1Exec(ctx,
		`INSERT OR IGNORE INTO credits_accounts (user_id, balance, reserved, total_earned, total_spent, created_at, updated_at)
		 VALUES (?, 0, 0, 0, 0, ?, ?)`,
		userID, now, now,
	); err != nil {
		return nil, NewAppError(op, err, "failed to ensure credits account", 500)
	}

	return c.GetCreditsBalance(ctx, userID)
}

// GrantCredits 向用户账户充入积分，并写入流水。
//
// 示例：
//
//	account, err := client.GrantCredits(ctx, userID, 1000, "admin-grant", "refID")
func (c *Client) GrantCredits(ctx context.Context, userID string, amount int64, description, refID string) (*CreditsAccount, error) {
	const op = "Client.GrantCredits"

	if userID == "" {
		return nil, NewAppError(op, ErrInvalidInput, "userID is required", 400)
	}
	if amount <= 0 {
		return nil, NewAppError(op, ErrInvalidInput, "Amount must be positive", 400)
	}

	// 确保账户存在（不需要读取余额，UPDATE 时账户必须存在）
	if _, err := c.GetCreditsBalance(ctx, userID); err != nil {
		if isAppError(err, ErrCreditsAccountNotFound) {
			if _, cerr := c.EnsureCreditsAccount(ctx, userID); cerr != nil {
				return nil, cerr
			}
		} else {
			return nil, err
		}
	}

	now := time.Now().UnixMilli()
	ledgerID := uuid.New().String()
	taskID := uuid.New().String()

	amountStr := strconv.FormatInt(amount, 10)
	nowStr := strconv.FormatInt(now, 10)

	if _, err := c.D1Exec(ctx,
		`UPDATE credits_accounts
		    SET balance      = balance + ?,
		        total_earned = total_earned + ?,
		        updated_at   = ?
		  WHERE user_id = ?`,
		amountStr, amountStr, nowStr, userID,
	); err != nil {
		return nil, NewAppError(op, err, fmt.Sprintf("failed to grant credits: %v", err), 500)
	}

	if _, err := c.D1Exec(ctx,
		`INSERT OR IGNORE INTO credits_ledger
		     (id, user_id, delta, balance_after, type, service, task_id, ref_id, description, created_at)
		 SELECT ?, ?, ?,
		        (SELECT balance FROM credits_accounts WHERE user_id = ?),
		        'grant', 'admin', ?, ?, ?, ?`,
		ledgerID, userID, amountStr, userID, taskID, refID, description, nowStr,
	); err != nil {
		return nil, NewAppError(op, err, fmt.Sprintf("credits granted but ledger write failed: %v", err), 500)
	}

	return c.GetCreditsBalance(ctx, userID)
}

// ─── 内部辅助 ──────────────────────────────────────────────────────────────────

// rowToCreditsAccount 将 D1 查询行解析为 CreditsAccount。
func rowToCreditsAccount(row map[string]interface{}) (*CreditsAccount, error) {
	b, err := json.Marshal(row)
	if err != nil {
		return nil, err
	}
	var raw struct {
		UserID      string      `json:"user_id"`
		Balance     json.Number `json:"balance"`
		Reserved    json.Number `json:"reserved"`
		TotalEarned json.Number `json:"total_earned"`
		TotalSpent  json.Number `json:"total_spent"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}

	balance, _ := raw.Balance.Int64()
	reserved, _ := raw.Reserved.Int64()
	earned, _ := raw.TotalEarned.Int64()
	spent, _ := raw.TotalSpent.Int64()

	return &CreditsAccount{
		UserID:      raw.UserID,
		Balance:     balance,
		Reserved:    reserved,
		TotalEarned: earned,
		TotalSpent:  spent,
	}, nil
}

// isAppError 检查 err 是否包含目标哨兵错误。
func isAppError(err error, target error) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*AppError); ok {
		return ae.Err == target
	}
	return false
}
