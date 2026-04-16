package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

// CapturedRequest holds a snapshot of the incoming request.
type CapturedRequest struct {
	Headers map[string]string
	Body    []byte
	Time    time.Time
}

// CapturedResponse holds a snapshot of the outgoing response.
type CapturedResponse struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

type captureWriter struct {
	http.ResponseWriter
	Status int
	buf    bytes.Buffer
}

func (cw *captureWriter) WriteHeader(code int) {
	cw.Status = code
	cw.ResponseWriter.WriteHeader(code)
}

func (cw *captureWriter) Write(b []byte) (int, error) {
	cw.buf.Write(b)
	return cw.ResponseWriter.Write(b)
}

// Capture wraps the handler so the response is recorded into *CapturedResponse.
func Capture(out *CapturedResponse) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cw := &captureWriter{ResponseWriter: w, Status: 200}
			next.ServeHTTP(cw, r)
			out.Status = cw.Status
			out.Body = cw.buf.Bytes()
			out.Headers = headerMap(w.Header())
		})
	}
}

// ReadBody reads and restores r.Body, returning the bytes.
func ReadBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	return b
}

func headerMap(h http.Header) map[string]string {
	m := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return m
}
