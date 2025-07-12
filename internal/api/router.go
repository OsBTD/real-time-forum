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
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/register", api.RegisterHandler).Methods("POST")
	apiRouter.HandleFunc("/login", api.LoginHandler).Methods("POST")
	apiRouter.HandleFunc("/logout", api.LogoutHandler).Methods("POST")
	apiRouter.HandleFunc("/session", api.SessionCheckHandler).Methods("GET") // Add this new route
	apiRouter.HandleFunc("/posts", api.CreatePostHandler).Methods("POST")
	apiRouter.HandleFunc("/posts", api.GetPostsHandler).Methods("GET")
	// Add other API routes here...
	apiRouter.HandleFunc("/comments", api.CreateCommentHandler).Methods("POST")
	apiRouter.HandleFunc("/messages", api.GetMessagesHandler).Methods("GET")
	apiRouter.HandleFunc("/messages", api.SendMessageHandler).Methods("POST")
	apiRouter.HandleFunc("/users", api.GetOnlineUsersHandler).Methods("GET")

	// WebSocket endpoint
	router.HandleFunc("/ws", api.WebSocketHandler)

	// Serve static files (CSS, JS)
	// Note the directory is "front", not "frontend"
	fs := http.FileServer(http.Dir("./front/"))
	router.PathPrefix("/js/").Handler(fs)
	router.Handle("/style.css", http.FileServer(http.Dir("./front")))

	// Serve the main index.html for the root path
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./front/index.html")
	})

	return router
}
