// Package tappay mocks TapPay (TW) payment API.
package tappay

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "tappay" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	h := map[string]string{"Content-Type": "application/json"}
	recTradeID := fmt.Sprintf("TAP%d", time.Now().UnixNano())
	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"status":                    10003,
			"msg":                       string(result.Mode),
			"rec_trade_id":              recTradeID,
			"bank_result_code":          "DECLINED",
			"bank_result_msg":           "Issuer declined",
			"transaction_time_millis":   time.Now().UnixMilli(),
		})
		return result.HTTPStatus, b, h
	}
	amt, cur := jsonAmt(body, 5000, "TWD")
	b, _ := json.Marshal(map[string]any{
		"status":                  0,
		"msg":                     "Success",
		"rec_trade_id":            recTradeID,
		"bank_transaction_id":     fmt.Sprintf("BANK_%d", time.Now().UnixNano()),
		"acquirer":                "TPN_CTBC",
		"amount":                  amt,
		"currency":                cur,
		"transaction_time_millis": time.Now().UnixMilli(),
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	statusCode := 0
	if result.HTTPStatus >= 400 {
		statusCode = 10003
	}
	out := map[string]any{
		"status":                  statusCode,
		"rec_trade_id":            chargeID,
		"amount":                  amount,
		"currency":                currency,
		"transaction_time_millis": time.Now().UnixMilli(),
		"msg":                     pickMsg(result),
	}
	if req != nil {
		if meta, ok := req["metadata"].(map[string]any); ok {
			out["metadata"] = meta
		}
		out["request_echo"] = req
	}
	return out
}

func pickMsg(r *engine.Result) string {
	if r.HTTPStatus >= 400 {
		return "failed"
	}
	if r.IsPending {
		return "processing"
	}
	return "Success"
}
func jsonAmt(body []byte, defAmt int64, defCur string) (int64, string) {
	amt, cur := defAmt, defCur
	var m map[string]any
	if len(body) == 0 || json.Unmarshal(body, &m) != nil {
		return amt, cur
	}
	if v, ok := m["amount"].(float64); ok {
		amt = int64(v)
	}
	if v, ok := m["currency"].(string); ok && v != "" {
		cur = v
	}
	return amt, cur
}
