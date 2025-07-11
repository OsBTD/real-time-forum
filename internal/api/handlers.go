package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"real-time-forum/internal/models"
	"real-time-forum/internal/services"
	"real-time-forum/internal/websocket"
	"time"
)

type API struct {
	DB  *sql.DB
	Hub *websocket.Hub
}

func (a *API) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userService := services.UserService{DB: a.DB}
	if err := userService.Register(&user); err != nil {
		log.Printf("Registration failed: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
}

func (a *API) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		EmailOrNickname string `json:"emailOrNickname"`
		Password        string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userService := services.UserService{DB: a.DB}
	token, err := userService.Login(credentials.EmailOrNickname, credentials.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (a *API) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	userService := services.UserService{DB: a.DB}
	if err := userService.Logout(cookie.Value); err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

func (a *API) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	websocket.ServeWs(a.Hub, w, r, a.DB)
}

// Similar handlers for posts, comments, chat, etc.
