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
		go func() {
			ticker := time.NewTicker(time.Duration(conf.Retention) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				cutoff := time.Now().UTC().Add(-time.Duration(conf.Retention) * time.Second).Unix()
				err := rlSvc.DeleteOlderThan(cutoff)
				if err != nil {
					log.Error("failed to delete old logs", "error", err)
				}
			}
		}()
	}

	app.RegisterService("api", server)

	return app, nil
}
