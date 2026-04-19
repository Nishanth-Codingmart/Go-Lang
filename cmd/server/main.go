package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/kvstore/internal/api"
	"github.com/user/kvstore/internal/cleanup"
	"github.com/user/kvstore/internal/config"
	"github.com/user/kvstore/internal/store"
	"github.com/user/kvstore/pkg/logger"
)

func main() {
	// 1. Setup Logger
	l := logger.New()

	// 2. Load Config
	cfg := config.Load()
	l.Info("Starting KV Store Server",
		"port", cfg.Port,
		"cleanup_interval", cfg.CleanupInterval,
		"shard_count", cfg.ShardCount,
	)

	// 3. Initialize Store
	kvStore := store.NewShardedStore(cfg.ShardCount)

	// 4. Start Cleanup Worker
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleanupWorker := cleanup.NewWorker(kvStore, cfg.CleanupInterval, l)
	go cleanupWorker.Start(ctx)

	// 5. Setup API
	h := api.NewHandler(kvStore)
	router := api.NewRouter(h, l)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	// 6. Graceful Shutdown Handling
	serverErrors := make(chan error, 1)
	go func() {
		l.Info("Server listening", "addr", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			l.Error("Server error", "error", err)
			os.Exit(1)
		}

	case sig := <-shutdown:
		l.Info("Shutdown signal received", "signal", sig)
		cancel() // Stop cleanup worker

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			l.Error("Could not stop server gracefully", "error", err)
			if err := srv.Close(); err != nil {
				l.Error("Could not close server", "error", err)
			}
		}
	}

	l.Info("Server stopped")
}
