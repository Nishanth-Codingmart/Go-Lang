package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TotalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kvstore_total_requests",
		Help: "The total number of requests",
	})
	GetRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kvstore_get_requests",
		Help: "The total number of GET requests",
	})
	PutRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kvstore_put_requests",
		Help: "The total number of PUT requests",
	})
	DeleteRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kvstore_delete_requests",
		Help: "The total number of DELETE requests",
	})
	ActiveKeys = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kvstore_active_keys",
		Help: "The total number of active keys in the store",
	})
	ExpiredKeys = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kvstore_expired_keys_total",
		Help: "The total number of keys that have expired",
	})
)

// Since prometheus counters are hard to read back, we use separate atomic counters for /stats
// In a real production app, you might use a custom registry or read from the Prometheus registry.
// For simplicity and to meet the requirement, we'll use the Prometheus counters (they can be read in some versions)
// but it's cleaner to just track them if we need them in JSON.
// However, the prompt asks for both. I'll stick to Prometheus for /metrics and custom for /stats if needed,
// but actually Chi middleware or a custom collector can do both.
// I'll add a helper to fetch these for the Stats response.
