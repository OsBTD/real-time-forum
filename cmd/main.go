package main

import (
	"real-time-forum/internal/api"
	"real-time-forum/internal/database"
	"real-time-forum/internal/websocket"
	"log"
	"net/http"
)

func main() {
	// Initialize database
	db, err := database.InitDB("database/forum.db")
	if err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	defer db.Close()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Set up routes
	router := api.SetupRouter(db, hub)

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Server failed:", err)
	}
}
