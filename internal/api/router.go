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

	// Create a file server for the 'front' directory
	staticFileServer := http.FileServer(http.Dir("./front"))

	// Serve JS files. http.StripPrefix is used to remove the '/js/' part
	// so the file server can find the files in the root of the 'front' directory.
	jsFileServer := http.FileServer(http.Dir("./front/js"))
	router.PathPrefix("/js/").Handler(http.StripPrefix("/js/", jsFileServer))
	// Serve the style.css file directly
	router.Handle("/style.css", staticFileServer)

	// Serve the main index.html for the root path
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./front/index.html")
	})

	return router
}
