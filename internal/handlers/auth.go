package handlers

import (
	"net/http"
)

type AuthHandler struct {
	token string
}

func NewAuthHandler(token string) *AuthHandler {
	return &AuthHandler{token: token}
}

func (h *AuthHandler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no auth token is configured, allow access
		if h.token == "" {
			next(w, r)
			return
		}

		// Check for auth cookie
		cookie, err := r.Cookie("auth_token")
		if err != nil || cookie.Value != h.token {
			// Redirect to login page
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		next(w, r)
	}
}

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		token := r.FormValue("token")
		if token == h.token {
			// Set auth cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "auth_token",
				Value: h.token,
				Path:  "/",
			})
			http.Redirect(w, r, "/app/", http.StatusTemporaryRedirect)
			return
		}

		// Invalid token - show error
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Show login page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - ProxyMini</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
            background: #f5f5f5;
        }
        .login-container {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            width: 100%;
            max-width: 400px;
        }
        h1 {
            margin-top: 0;
            color: #333;
        }
        .form-group {
            margin-bottom: 1rem;
        }
        label {
            display: block;
            margin-bottom: 0.5rem;
            color: #666;
        }
        input[type="password"] {
            width: 100%;
            padding: 0.5rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
            font-size: 1rem;
        }
        button {
            width: 100%;
            padding: 0.75rem;
            background: #007bff;
            color: white;
            border: none;
            border-radius: 4px;
            font-size: 1rem;
            cursor: pointer;
        }
        button:hover {
            background: #0056b3;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>ProxyMini Login</h1>
        <form method="POST" action="/login">
            <div class="form-group">
                <label for="token">Auth Token</label>
                <input type="password" id="token" name="token" required autofocus>
            </div>
            <button type="submit">Login</button>
        </form>
    </div>
</body>
</html>`))
}
