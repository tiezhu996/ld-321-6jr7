package dto

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mustJSONPost(t *testing.T, body map[string]interface{}) *http.Request {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	return req
}
