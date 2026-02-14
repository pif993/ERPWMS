package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

type Step struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Result struct {
	OK    bool   `json:"ok"`
	Steps []Step `json:"steps"`
}

// Run executes an in-process end-to-end suite through the existing router.
func Run(router http.Handler, adminEmail, adminPassword string) Result {
	res := Result{OK: true}

	do := func(req *http.Request) (int, []byte, error) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code, w.Body.Bytes(), nil
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		code, _, _ := do(req)
		ok := code == http.StatusOK
		res.Steps = append(res.Steps, Step{Name: "GET /health", OK: ok, Status: code})
		res.OK = res.OK && ok
	}

	var access string
	{
		body, _ := json.Marshal(map[string]string{"email": adminEmail, "password": adminPassword})
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		code, b, _ := do(req)
		ok := code == http.StatusOK
		if ok {
			var m map[string]string
			_ = json.Unmarshal(b, &m)
			access = m["access_token"]
			if access == "" {
				ok = false
			}
		}
		res.Steps = append(res.Steps, Step{Name: "POST /api/auth/login", OK: ok, Status: code})
		res.OK = res.OK && ok
	}

	{
		req := httptest.NewRequest(http.MethodGet, "/api/stock/balances?limit=1&offset=0", nil)
		req.Header.Set("Authorization", "Bearer "+access)
		code, _, _ := do(req)
		ok := code == http.StatusOK
		res.Steps = append(res.Steps, Step{Name: "GET /api/stock/balances", OK: ok, Status: code})
		res.OK = res.OK && ok
	}

	return res
}
