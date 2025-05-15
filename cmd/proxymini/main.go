package main

import (
	"log"
	"net/http"

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

	proxyHandler := handlers.NewProxyHandler(rlSvc)
	rlHandler := handlers.NewRequestLogHandler(rlSvc)

	server := http.NewServeMux()
	server.Handle("/app", http.StripPrefix("/app", http.FileServer(http.FS(frontend.Assets()))))
	server.Handle("/api/logs", rlHandler)
	server.Handle("/", proxyHandler)

	log.Println("Starting server at: http://localhost:" + conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, server))
}
