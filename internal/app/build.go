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

const loginPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ProxyMini - Login</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f172a;
            color: #e2e8f0;
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            padding: 20px;
        }
        .login-container {
            background: #1e293b;
            border-radius: 8px;
            padding: 32px;
            width: 100%;
            max-width: 400px;
            box-shadow: 0 4px 6px -1px rgba(0,0,0,0.3);
        }
        h1 {
            font-size: 24px;
            margin-bottom: 8px;
            color: #f8fafc;
        }
        p {
            color: #94a3b8;
            margin-bottom: 24px;
            font-size: 14px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 6px;
            font-size: 14px;
            color: #cbd5e1;
        }
        input[type="password"] {
            width: 100%;
            padding: 10px 12px;
            border: 1px solid #334155;
            border-radius: 6px;
            background: #0f172a;
            color: #f8fafc;
            font-size: 14px;
        }
        input[type="password"]:focus {
            outline: none;
            border-color: #3b82f6;
        }
        button {
            width: 100%;
            padding: 10px;
            background: #3b82f6;
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: background 0.2s;
        }
        button:hover {
            background: #2563eb;
        }
        .error {
            color: #ef4444;
            font-size: 13px;
            margin-top: 12px;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>ProxyMini</h1>
        <p>Enter your access token to continue</p>
        <form method="POST" action="/auth">
            <div class="form-group">
                <label for="token">Token</label>
                <input type="password" id="token" name="token" required autofocus>
            </div>
            <button type="submit">Sign In</button>
            <div id="error" class="error"></div>
        </form>
    </div>
    <script>
        const params = new URLSearchParams(window.location.search);
        if (params.get('error')) {
            document.getElementById('error').textContent = 'Invalid token';
        }
    </script>
</body>
</html>`

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

	// Auth middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if conf.AuthToken == "" {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(authCookieName)
			if err != nil || cookie.Value != conf.AuthToken {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// HTTP Server
	server := httpserver.New(conf.Port, httpShutdownTimeout)

	// Login page
	server.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if conf.AuthToken == "" {
			http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(loginPageHTML))
	})

	// Auth endpoint
	server.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/login?error=1", http.StatusTemporaryRedirect)
			return
		}

		token := r.FormValue("token")
		if token != conf.AuthToken {
			http.Redirect(w, r, "/login?error=1", http.StatusTemporaryRedirect)
			return
		}

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
