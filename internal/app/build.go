package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/platforma-dev/platforma/application"
	"github.com/platforma-dev/platforma/httpserver"
	"github.com/platforma-dev/platforma/log"

	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/proxy"
	"github.com/mishankov/proxymini/internal/requestlog"
	frontend "github.com/mishankov/proxymini/webui"
)

const httpShutdownTimeout = 2 * time.Second

// cleanupInterval is the frequency at which old logs are checked and deleted
const cleanupInterval = 1 * time.Hour

func Build(conf *config.Config, rlDB *sqlx.DB) (*application.Application, error) {
	if conf == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Request logs
	rlSvc := requestlog.NewRequestLogService(rlDB)
	rlHandler := requestlog.NewRequestLogHandler(rlSvc)

	// Proxy
	proxyHandler := proxy.NewProxyHandler(rlSvc, conf)

	// WebUI file server
	appFileServer := http.FileServer(http.FS(frontend.Assets()))

	// HTTP Server
	server := httpserver.New(conf.Port, httpShutdownTimeout)
	server.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
	})
	server.Handle("/app/", http.StripPrefix("/app", appFileServer))
	server.Handle("/api/logs", rlHandler)
	server.Handle("/", proxyHandler)

	// App
	app := application.New()

	app.OnStartFunc(func(_ context.Context) error {
		return db.Init(rlDB)
	}, application.StartupTaskConfig{
		Name:         "sqlite-migrate",
		AbortOnError: true,
	})

	// Start retention scheduler if retention is configured
	if conf.Retention > 0 {
		app.OnStartFunc(func(ctx context.Context) error {
			go startRetentionScheduler(ctx, rlSvc, conf.Retention)
			return nil
		}, application.StartupTaskConfig{
			Name:         "logs-retention-scheduler",
			AbortOnError: false,
		})
	}

	app.RegisterService("api", server)

	return app, nil
}

func startRetentionScheduler(ctx context.Context, rlSvc *requestlog.RequestLogService, retentionSeconds int) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			threshold := time.Now().UTC().Unix() - int64(retentionSeconds)
			if err := rlSvc.DeleteOlderThan(threshold); err != nil {
				log.Error("failed to delete old request logs", "error", err)
			}
		}
	}
}
