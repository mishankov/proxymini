package main

import (
	"context"
	"os"

	"github.com/mishankov/proxymini/internal/app"
	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/platforma-dev/platforma/log"
)

func main() {
	ctx := context.Background()

	conf, err := config.New()
	if err != nil {
		log.ErrorContext(ctx, "failed to load config", "error", err)
		os.Exit(1)
	}

	rlDB, err := db.Connect(conf.DBPath)
	if err != nil {
		log.ErrorContext(ctx, "failed to connect to request log database", "error", err)
		os.Exit(1)
	}
	defer rlDB.Close()

	app, err := app.Build(conf, rlDB)
	if err != nil {
		log.ErrorContext(ctx, "failed to build runtime", "error", err)
		os.Exit(1)
	}

	if err := app.Run(ctx); err != nil {
		log.ErrorContext(ctx, "application run failed", "error", err)
		os.Exit(1)
	}
}
