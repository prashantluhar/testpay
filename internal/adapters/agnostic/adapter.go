package agnostic

import "github.com/prashantluhar/testpay/internal/engine"

type Adapter struct{}

func New() *Adapter { return &Adapter{} }
func (a *Adapter) Name() string { return "agnostic" }
func (a *Adapter) BuildResponse(result *engine.Result, body []byte) (int, []byte, map[string]string) {
	return result.HTTPStatus, []byte(`{}`), nil
}
func (a *Adapter) BuildWebhookPayload(result *engine.Result, id string, amount int64, currency string) map[string]any {
	return map[string]any{}
}
