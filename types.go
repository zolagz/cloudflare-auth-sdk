package cloudflare_auth_sdk

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User represents a user in the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserInfo represents public user information (without sensitive data).
type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// LoginResponse represents the response from a successful login.
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

// Claims represents JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// KVKey represents a key in the KV namespace with metadata.
type KVKey struct {
	Name       string      `json:"name"`
	Expiration float64     `json:"expiration,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
}

// KVWriteOptions contains options for writing KV pairs.
type KVWriteOptions struct {
	ExpirationTTL int    // Time to live in seconds
	Metadata      string // Optional metadata
}

// toJSON converts User to JSON bytes
func (u *User) toJSON() ([]byte, error) {
	return json.Marshal(u)
}

// userFromJSON parses User from JSON bytes
func userFromJSON(data []byte) (*User, error) {
	var user User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ToUserInfo converts User to UserInfo
func (u *User) ToUserInfo() UserInfo {
	return UserInfo{
		ID:    u.ID,
		Email: u.Email,
	}
}

// ─── D1 通用类型 ───────────────────────────────────────────────────────────────

// D1QueryResult 是执行 D1 SQL 查询后的结果。
type D1QueryResult struct {
	// Results 是查询返回的行，每行为 map[string]interface{}。
	Results []map[string]interface{}
	// RowsAffected 是被 INSERT/UPDATE/DELETE 影响的行数。
	RowsAffected int
	// Success 标识 Cloudflare API 是否认为此查询成功。
	Success bool
}

// D1Statement 表示 D1BatchExec 中的单条 SQL 语句及其参数。
type D1Statement struct {
	// SQL 是要执行的 SQL 语句，参数用 ? 占位。
	SQL string
	// Params 是按位置顺序对应 ? 的参数列表，全部以字符串传入。
	Params []string
}

// ─── 积分相关类型 ──────────────────────────────────────────────────────────────

// CreditsAccount 对应 credits_accounts 表的一行。
type CreditsAccount struct {
	UserID      string `json:"user_id"`
	Balance     int64  `json:"balance"`
	Reserved    int64  `json:"reserved"`
	TotalEarned int64  `json:"total_earned"`
	TotalSpent  int64  `json:"total_spent"`
}

// DeductCreditsParams 是扣减积分所需的参数。
type DeductCreditsParams struct {
	// UserID 是要扣减积分的用户 ID。
	UserID string
	// Amount 是要扣减的积分数量（正整数）。
	Amount int64
	// Service 是发起扣减的服务名称（如 "openai"、"anthropic"）。
	Service string
	// Model 是使用的模型名称（可选）。
	Model string
	// TaskID 是任务唯一标识，用于幂等性控制（可选）。
	TaskID string
	// Description 是本次扣减的说明（可选）。
	Description string
}
