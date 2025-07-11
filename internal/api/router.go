package api

import (
	"database/sql"
	"net/http"
	"real-time-forum/internal/websocket"

	"github.com/gorilla/mux"
)

func SetupRouter(db *sql.DB, hub *websocket.Hub) *mux.Router {
	api := &API{DB: db, Hub: hub}

	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/api/register", api.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/login", api.LoginHandler).Methods("POST")
	router.HandleFunc("/api/logout", api.LogoutHandler).Methods("POST")
	router.HandleFunc("/api/posts", api.CreatePostHandler).Methods("POST")
	router.HandleFunc("/api/posts", api.GetPostsHandler).Methods("GET")
	router.HandleFunc("/api/comments", api.CreateCommentHandler).Methods("POST")
	router.HandleFunc("/api/messages", api.GetMessagesHandler).Methods("GET")
	router.HandleFunc("/api/messages", api.SendMessageHandler).Methods("POST")
	router.HandleFunc("/api/users", api.GetOnlineUsersHandler).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", api.WebSocketHandler)

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	return router
}
