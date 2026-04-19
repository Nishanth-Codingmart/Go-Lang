package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/kvstore/internal/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(s store.Store) *Handler {
	return &Handler{store: s}
}

type SetRequest struct {
	Value      string `json:"value"`
	TTLSeconds int    `json:"ttl_seconds"`
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	h.store.Set(key, req.Value, time.Duration(req.TTLSeconds)*time.Second)
	Success(w, "Key set successfully", nil)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	val, expiresAt, exists := h.store.Get(key)
	if !exists {
		Error(w, http.StatusNotFound, "Key not found")
		return
	}

	data := map[string]interface{}{
		"value":      val,
		"expires_at": expiresAt,
	}
	Success(w, "Key retrieved successfully", data)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	h.store.Delete(key)
	Success(w, "Key deleted successfully", nil)
}

func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := h.store.Stats()
	Success(w, "Stats retrieved successfully", stats)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
