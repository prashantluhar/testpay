// Package payletter mocks Payletter (KR) payment API.
// Real-world: HMAC-SHA256 signing, hex signature header.
package payletter

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter             { return &Adapter{} }
func (a *Adapter) Name() string { return "payletter" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	h := map[string]string{"Content-Type": "application/json"}
	tid := fmt.Sprintf("PL%d", time.Now().UnixNano())
	if result.HTTPStatus >= 400 {
		b, _ := json.Marshal(map[string]any{
			"code":    -1,
			"message": string(result.Mode),
			"tid":     tid,
		})
		return result.HTTPStatus, b, h
	}
	amt, cur := jsonAmt(body, 5000, "KRW")
	b, _ := json.Marshal(map[string]any{
		"code":       0,
		"message":    "OK",
		"tid":        tid,
		"cid":        fmt.Sprintf("CID_%d", time.Now().UnixNano()),
		"amount":     amt,
		"currency":   cur,
		"pg_code":    "payletter",
		"online_url": fmt.Sprintf("https://payletter.com/mock/%s", tid),
		"trans_date": time.Now().Format("20060102150405"),
	})
	return 200, b, h
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string, req map[string]any) map[string]any {
	out := map[string]any{
		"event":    pickEvent(result),
		"tid":      chargeID,
		"cid":      fmt.Sprintf("CID_%d", time.Now().UnixNano()),
		"code":     ifFailed(result, -1, 0),
		"amount":   amount,
		"currency": currency,
		"trans_date": time.Now().Format("20060102150405"),
	}
	if req != nil {
		out["request_echo"] = req
	}
	return out
}

func pickEvent(r *engine.Result) string {
	if r.HTTPStatus >= 400 {
		return "payment.failed"
	}
	if r.IsPending {
		return "payment.pending"
	}
	return "payment.paid"
}
func ifFailed(r *engine.Result, onFail, onSuccess int) int {
	if r.HTTPStatus >= 400 {
		return onFail
	}
	return onSuccess
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
