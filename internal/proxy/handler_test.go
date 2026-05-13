package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerProxiesRequestAndPreservesStatus(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/resource" || r.URL.RawQuery != "a=1" {
			t.Fatalf("url = %s, want /resource?a=1", r.URL.String())
		}
		if got := r.Header.Get("X-Test"); got != "yes" {
			t.Fatalf("X-Test = %q, want yes", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if string(body) != "payload" {
			t.Fatalf("body = %q, want payload", string(body))
		}

		w.Header().Set("X-Upstream", "ok")
		w.Header().Set("Connection", "X-Remove")
		w.Header().Set("X-Remove", "remove me")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer upstream.Close()

	req := httptest.NewRequest(http.MethodPost, "/"+upstream.URL+"/resource?a=1", strings.NewReader("payload"))
	req.Header.Set("X-Test", "yes")
	w := httptest.NewRecorder()

	Handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	if got := resp.Header.Get("X-Upstream"); got != "ok" {
		t.Fatalf("X-Upstream = %q, want ok", got)
	}
	if got := resp.Header.Get("X-Remove"); got != "" {
		t.Fatalf("X-Remove = %q, want removed", got)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want *", got)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != "proxied" {
		t.Fatalf("response body = %q, want proxied", string(body))
	}
}

func TestHandlerAcceptsSingleSlashSchemeURL(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer upstream.Close()

	target := strings.Replace(upstream.URL, "http://", "http:/", 1)
	req := httptest.NewRequest(http.MethodGet, "/"+target+"/single-slash", nil)
	w := httptest.NewRecorder()

	Handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != "/single-slash" {
		t.Fatalf("response body = %q, want /single-slash", string(body))
	}
}

func TestHandlerRejectsInvalidTargetURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/not-a-url", nil)
	w := httptest.NewRecorder()

	Handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlerOptionsPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/anything", nil)
	w := httptest.NewRecorder()

	Handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "OPTIONS") {
		t.Fatalf("Access-Control-Allow-Methods = %q, want OPTIONS", got)
	}
}
