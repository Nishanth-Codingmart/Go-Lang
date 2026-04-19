package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/user/kvstore/internal/metrics"
)

func NewRouter(h *Handler, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.TotalRequests.Inc()
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/health", h.Health)
	r.Get("/stats", h.Stats)
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/kv", func(r chi.Router) {
		r.Put("/{key}", h.Set)
		r.Get("/{key}", h.Get)
		r.Delete("/{key}", h.Delete)
	})

	return r
}
