package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/platforma-dev/platforma/application"
	"github.com/platforma-dev/platforma/httpserver"

	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/proxy"
	"github.com/mishankov/proxymini/internal/requestlog"
	frontend "github.com/mishankov/proxymini/webapp"
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

	app.RegisterService("api", server)

	return app, nil
}
