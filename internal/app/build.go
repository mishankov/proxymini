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
	frontend "github.com/mishankov/proxymini/webui"
)

const httpShutdownTimeout = 2 * time.Second
const authCookieName = "proxymini_auth"

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

	// Auth middleware - protects routes when AuthToken is configured
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if no token configured
			if conf.AuthToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Check auth cookie
			cookie, err := r.Cookie(authCookieName)
			if err != nil || cookie.Value != conf.AuthToken {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Login handler
	server.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// If no auth token configured, redirect to app
		if conf.AuthToken == "" {
			http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
			return
		}

		if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form", http.StatusBadRequest)
				return
			}

			token := r.FormValue("token")
			if token == conf.AuthToken {
				// Set auth cookie
				http.SetCookie(w, &http.Cookie{
					Name:     authCookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteStrictMode,
					MaxAge:   86400 * 30, // 30 days
				})
				http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
				return
			}

			// Invalid token - show login page with error
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(loginPageHTML(true)))
			return
		}

		// GET request - show login page
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(loginPageHTML(false)))
	})

	// Logout handler
	server.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     authCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			MaxAge:   -1,
		})
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	})

	// App routes with auth middleware
	server.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
	})
	server.Handle("/app/", authMiddleware(http.StripPrefix("/app", appFileServer)))
	server.Handle("/api/logs", authMiddleware(rlHandler))
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

func loginPageHTML(showError bool) string {
	errorMsg := ""
	if showError {
		errorMsg = `<p style="color: #ef4444; margin: 10px 0; font-size: 14px;">Invalid token. Please try again.</p>`
	}

	return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>ProxyMini - Login</title>
	<style>
		* { box-sizing: border-box; margin: 0; padding: 0; }
		body {
			font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
			padding: 20px;
		}
		.login-container {
			background: white;
			padding: 40px;
			border-radius: 12px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
			width: 100%;
			max-width: 400px;
		}
		h1 {
			color: #1f2937;
			margin-bottom: 8px;
			font-size: 24px;
		}
		.subtitle {
			color: #6b7280;
			margin-bottom: 24px;
			font-size: 14px;
		}
		input[type="password"] {
			width: 100%;
			padding: 12px 16px;
			border: 2px solid #e5e7eb;
			border-radius: 8px;
			font-size: 16px;
			transition: border-color 0.2s;
			margin-bottom: 16px;
		}
		input[type="password"]:focus {
			outline: none;
			border-color: #667eea;
		}
		button {
			width: 100%;
			padding: 12px;
			background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
			color: white;
			border: none;
			border-radius: 8px;
			font-size: 16px;
			font-weight: 600;
			cursor: pointer;
			transition: transform 0.1s, box-shadow 0.2s;
		}
		button:hover {
			transform: translateY(-1px);
			box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
		}
		button:active {
			transform: translateY(0);
		}
	</style>
</head>
<body>
	<div class="login-container">
		<h1>🔒 ProxyMini</h1>
		<p class="subtitle">Enter your access token to continue</p>
		<form method="POST" action="/login">
			<input type="password" name="token" placeholder="Access token" required autofocus>
			` + errorMsg + `
			<button type="submit">Sign In</button>
		</form>
	</div>
</body>
</html>`
}
