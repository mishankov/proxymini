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

	proxyHandler := handlers.NewProxyHandler(rlSvc, conf)
	rlHandler := handlers.NewRequestLogHandler(rlSvc)
	authHandler := handlers.NewAuthHandler(conf.AuthToken)
	appFileServer := http.FileServer(http.FS(frontend.Assets()))

	server := http.NewServeMux()
	server.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
	})
	server.Handle("/app/", authHandler.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		appFileServer.ServeHTTP(w, r)
	}))
	server.HandleFunc("/login", authHandler.LoginHandler)
	server.Handle("/api/logs", rlHandler)
	server.Handle("/", proxyHandler)

	log.Println("Starting server at: http://localhost:" + conf.Port)
	log.Fatal(http.ListenAndServe(":"+conf.Port, server))
}
