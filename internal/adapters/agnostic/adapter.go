package agnostic

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/prashantluhar/testpay/internal/engine"
)

type Adapter struct{}

func New() *Adapter         { return &Adapter{} }
func (a *Adapter) Name() string { return "agnostic" }

func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	headers := map[string]string{"Content-Type": "application/json"}
	status := "success"
	if result.HTTPStatus >= 400 {
		status = "failed"
	}
	if result.IsPending {
		status = "pending"
	}
	resp, _ := json.Marshal(map[string]any{
		"id":         fmt.Sprintf("txn_%d", time.Now().UnixNano()),
		"status":     status,
		"error_code": result.ErrorCode,
	})
	return result.HTTPStatus, resp, headers
}

func (a *Adapter) BuildWebhookPayload(result *engine.Result, chargeID string, amount int64, currency string) map[string]any {
	event := "transaction.success"
	if result.HTTPStatus >= 400 {
		event = "transaction.failed"
	}
	return map[string]any{
		"event":      event,
		"id":         chargeID,
		"amount":     amount,
		"currency":   currency,
		"timestamp":  time.Now().Unix(),
		"error_code": result.ErrorCode,
	}
}
