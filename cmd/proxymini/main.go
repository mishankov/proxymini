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

	// Start retention cleanup scheduler if retention is configured
	if conf.Retention > 0 {
		log.Printf("Logs retention enabled: %d seconds", conf.Retention)
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()

			// Run cleanup immediately on startup
			if err := rlSvc.DeleteOlderThan(conf.Retention); err != nil {
				log.Printf("Error deleting old logs: %v", err)
			} else {
				log.Printf("Deleted logs older than %d seconds", conf.Retention)
			}

			for range ticker.C {
				if err := rlSvc.DeleteOlderThan(conf.Retention); err != nil {
					log.Printf("Error deleting old logs: %v", err)
				} else {
					log.Printf("Deleted logs older than %d seconds", conf.Retention)
				}
			}
		}()
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
