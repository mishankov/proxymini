package handlers

import (
	"net/http"
	"strings"

	"github.com/mishankov/proxymini/internal/config"
)

type AuthHandler struct {
	config    *config.Config
	appFS     http.Handler
}

func NewAuthHandler(config *config.Config, appFS http.Handler) *AuthHandler {
	return &AuthHandler{
		config:    config,
		appFS:     appFS,
	}
}

const authCookieName = "proxymini_auth"

func (ah *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If no auth token configured, allow all requests
		if ah.config.AuthToken == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check for valid auth cookie
		cookie, err := r.Cookie(authCookieName)
		if err == nil && cookie.Value == ah.config.AuthToken {
			next.ServeHTTP(w, r)
			return
		}

		// Redirect to login if not authenticated
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	})
}

func (ah *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// If no auth token configured, redirect to app
	if ah.config.AuthToken == "" {
		http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		token := r.FormValue("token")
		if token == ah.config.AuthToken {
			// Set auth cookie
			http.SetCookie(w, &http.Cookie{
				Name:     authCookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteStrictMode,
			})
			http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
			return
		}

		// Invalid token, show login page with error
		ah.renderLoginPage(w, true)
		return
	}

	// Check if already logged in
	cookie, err := r.Cookie(authCookieName)
	if err == nil && cookie.Value == ah.config.AuthToken {
		http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
		return
	}

	ah.renderLoginPage(w, false)
}

func (ah *AuthHandler) renderLoginPage(w http.ResponseWriter, showError bool) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	var errorMsg string
	if showError {
		errorMsg = `<div style="color: #dc3545; margin-bottom: 15px; text-align: center;">Invalid token</div>`
	}

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ProxyMini Login</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .login-container {
            background: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            width: 100%;
            max-width: 360px;
        }
        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
            font-size: 24px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #555;
            font-size: 14px;
            font-weight: 500;
        }
        input[type="password"] {
            width: 100%;
            padding: 12px 15px;
            border: 2px solid #e0e0e0;
            border-radius: 6px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
        }
        button {
            width: 100%;
            padding: 14px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
        }
        button:active {
            transform: translateY(0);
        }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>🔒 ProxyMini</h1>
        ` + errorMsg + `
        <form method="POST" action="/login">
            <div class="form-group">
                <label for="token">Auth Token</label>
                <input type="password" id="token" name="token" placeholder="Enter your auth token" required autofocus>
            </div>
            <button type="submit">Login</button>
        </form>
    </div>
</body>
</html>`

	w.Write([]byte(html))
}

func (ah *AuthHandler) ServeApp(w http.ResponseWriter, r *http.Request) {
	// Strip /app prefix and serve files
	path := strings.TrimPrefix(r.URL.Path, "/app")
	if path == "" {
		path = "/"
	}
	r.URL.Path = path
	ah.appFS.ServeHTTP(w, r)
}
