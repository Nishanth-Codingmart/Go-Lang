package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/kvstore/internal/store"
	"github.com/user/kvstore/pkg/logger"
)

func TestAPI_KV_Flow(t *testing.T) {
	l := logger.New()
	s := store.NewShardedStore(4)
	h := NewHandler(s)
	router := NewRouter(h, l)

	// 1. PUT key
	putReq := SetRequest{Value: "hello", TTLSeconds: 60}
	body, _ := json.Marshal(putReq)
	req := httptest.NewRequest("PUT", "/kv/test-key", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	// 2. GET key
	req = httptest.NewRequest("GET", "/kv/test-key", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	if data["value"] != "hello" {
		t.Errorf("Expected hello, got %v", data["value"])
	}

	// 3. DELETE key
	req = httptest.NewRequest("DELETE", "/kv/test-key", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	// 4. GET deleted key
	req = httptest.NewRequest("GET", "/kv/test-key", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}
}

func TestAPI_Stats(t *testing.T) {
	l := logger.New()
	s := store.NewShardedStore(4)
	h := NewHandler(s)
	router := NewRouter(h, l)

	s.Set("k1", "v1", 0)

	req := httptest.NewRequest("GET", "/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp Response
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp.Data.(map[string]interface{})
	if data["TotalKeys"].(float64) != 1 {
		t.Errorf("Expected 1 key, got %v", data["TotalKeys"])
	}
}
