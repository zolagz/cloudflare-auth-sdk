package cloudflare_auth_sdk

import (
	"context"
	"encoding/json"
	"fmt"

	cloudflare "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/d1"
)

// D1Query 执行一条 SELECT 类 SQL，返回行列表。
// 每行以 map[string]interface{} 表示，键为列名。
//
// 示例：
//
//	rows, err := client.D1Query(ctx, "SELECT * FROM users WHERE id = ?", userID)
func (c *Client) D1Query(ctx context.Context, sql string, params ...string) ([]map[string]interface{}, error) {
	const op = "Client.D1Query"

	if err := c.requireD1(); err != nil {
		return nil, err
	}

	res, err := c.cfClient.D1.Database.Query(ctx, c.d1DatabaseID, d1.DatabaseQueryParams{
		AccountID: cloudflare.F(c.accountID),
		Body: d1.DatabaseQueryParamsBodyD1SingleQuery{
			Sql:    cloudflare.F(sql),
			Params: cloudflare.F(params),
		},
	})
	if err != nil {
		return nil, NewAppError(op, fmt.Errorf("%w: %v", ErrD1QueryFailed, err), fmt.Sprintf("D1 query error: %v", err), 500)
	}

	var rows []map[string]interface{}
	for _, qr := range res.Result {
		if !qr.Success {
			return nil, NewAppError(op, ErrD1QueryFailed, "D1 query returned success=false", 500)
		}
		for _, raw := range qr.Results {
			row, err := toStringMap(raw)
			if err != nil {
				return nil, NewAppError(op, ErrD1QueryFailed, fmt.Sprintf("failed to parse row: %v", err), 500)
			}
			rows = append(rows, row)
		}
	}
	return rows, nil
}

// D1Exec 执行一条写操作 SQL（INSERT / UPDATE / DELETE）。
// 返回受影响行数及完整查询结果。
//
// 示例：
//
//	result, err := client.D1Exec(ctx,
//	    "UPDATE credits_accounts SET balance = balance - ? WHERE user_id = ?",
//	    "100", userID,
//	)
func (c *Client) D1Exec(ctx context.Context, sql string, params ...string) (*D1QueryResult, error) {
	const op = "Client.D1Exec"

	if err := c.requireD1(); err != nil {
		return nil, err
	}

	res, err := c.cfClient.D1.Database.Query(ctx, c.d1DatabaseID, d1.DatabaseQueryParams{
		AccountID: cloudflare.F(c.accountID),
		Body: d1.DatabaseQueryParamsBodyD1SingleQuery{
			Sql:    cloudflare.F(sql),
			Params: cloudflare.F(params),
		},
	})
	if err != nil {
		return nil, NewAppError(op, fmt.Errorf("%w: %v", ErrD1QueryFailed, err), fmt.Sprintf("D1 exec error: %v", err), 500)
	}

	result := &D1QueryResult{Success: true}
	for _, qr := range res.Result {
		if !qr.Success {
			result.Success = false
			return result, NewAppError(op, ErrD1QueryFailed, "D1 exec returned success=false", 500)
		}
		result.RowsAffected += int(qr.Meta.Changes)
		for _, raw := range qr.Results {
			row, err := toStringMap(raw)
			if err != nil {
				return nil, NewAppError(op, ErrD1QueryFailed, fmt.Sprintf("failed to parse row: %v", err), 500)
			}
			result.Results = append(result.Results, row)
		}
	}
	return result, nil
}

// D1BatchExec 在单次 HTTP 请求中以 batch 方式执行多条 SQL。
// 每条语句独立执行，结果按顺序返回，适合需要同时写入多张表的场景。
//
// 示例：
//
//	results, err := client.D1BatchExec(ctx, []D1Statement{
//	    {SQL: "UPDATE credits_accounts SET balance = balance - ? WHERE user_id = ?", Params: []string{"50", userID}},
//	    {SQL: "INSERT INTO credits_ledger (...) VALUES (?,?,?)", Params: []string{...}},
//	})
func (c *Client) D1BatchExec(ctx context.Context, stmts []D1Statement) ([]*D1QueryResult, error) {
	const op = "Client.D1BatchExec"

	if err := c.requireD1(); err != nil {
		return nil, err
	}
	if len(stmts) == 0 {
		return nil, NewAppError(op, ErrInvalidInput, "no statements provided", 400)
	}

	// 构建 batch 参数：每条语句独立提供 SQL+Params
	batch := make([]d1.DatabaseQueryParamsBodyMultipleQueriesBatch, len(stmts))
	for i, s := range stmts {
		batch[i] = d1.DatabaseQueryParamsBodyMultipleQueriesBatch{
			Sql:    cloudflare.F(s.SQL),
			Params: cloudflare.F(s.Params),
		}
	}

	res, err := c.cfClient.D1.Database.Query(ctx, c.d1DatabaseID, d1.DatabaseQueryParams{
		AccountID: cloudflare.F(c.accountID),
		Body: d1.DatabaseQueryParamsBodyMultipleQueries{
			Batch: cloudflare.F(batch),
		},
	})
	if err != nil {
		return nil, NewAppError(op, fmt.Errorf("%w: %v", ErrD1QueryFailed, err), fmt.Sprintf("D1 batch exec error: %v", err), 500)
	}

	var results []*D1QueryResult
	for _, qr := range res.Result {
		r := &D1QueryResult{
			Success:      qr.Success,
			RowsAffected: int(qr.Meta.Changes),
		}
		for _, raw := range qr.Results {
			row, err := toStringMap(raw)
			if err != nil {
				return nil, NewAppError(op, ErrD1QueryFailed, fmt.Sprintf("failed to parse row: %v", err), 500)
			}
			r.Results = append(r.Results, row)
		}
		results = append(results, r)
	}
	return results, nil
}

// requireD1 验证 D1 数据库 ID 已配置。
func (c *Client) requireD1() error {
	if c.d1DatabaseID == "" {
		return NewAppError("Client.requireD1", ErrD1NotConfigured, "D1DatabaseID is required for D1 operations", 500)
	}
	return nil
}

// toStringMap 将 interface{} 转成 map[string]interface{}（通过 JSON round-trip）。
func toStringMap(v interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
