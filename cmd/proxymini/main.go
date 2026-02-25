package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mishankov/proxymini/internal/config"
	"github.com/mishankov/proxymini/internal/db"
	"github.com/mishankov/proxymini/internal/handlers"
	"github.com/mishankov/proxymini/internal/services"
	frontend "github.com/mishankov/proxymini/webapp"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	rlDB, err := db.Connect(conf.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer rlDB.Close()

	err = db.Init(rlDB)
	if err != nil {
		log.Fatal(err)
	}

	rlSvc := services.NewRequestLogService(rlDB)

	// Start retention scheduler if retention is configured
	if conf.Retention > 0 {
		go startRetentionScheduler(rlSvc, conf.Retention)
	}

	proxyHandler := handlers.NewProxyHandler(rlSvc, conf)
	rlHandler := handlers.NewRequestLogHandler(rlSvc)
	appFileServer := http.FileServer(http.FS(frontend.Assets()))

	server := http.NewServeMux()
	server.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
	})
	server.Handle("/app/", http.StripPrefix("/app", appFileServer))
	server.Handle("/api/logs", rlHandler)
	server.Handle("/", proxyHandler)

	log.Println("Starting server at: http://localhost:" + conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, server))
}

func startRetentionScheduler(rlSvc *services.RequestLogService, retentionSeconds int) {
	ticker := time.NewTicker(time.Duration(retentionSeconds) * time.Second)
	defer ticker.Stop()

	log.Printf("Starting logs retention scheduler (retention: %d seconds)\n", retentionSeconds)

	for range ticker.C {
		if err := rlSvc.DeleteOlderThan(int64(retentionSeconds)); err != nil {
			log.Println("Error deleting old logs:", err)
		} else {
			log.Println("Old logs cleaned up successfully")
		}
	}
}
