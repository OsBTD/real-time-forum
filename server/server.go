package server

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"forum/internal/auth"
	db "forum/internal/database"

	_ "github.com/mattn/go-sqlite3"
)

func Run() error {
	flag.Parse()

	// Initialize database
	if err := db.InitDatabase("forum.db"); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.DB.Close()

	// Initialize templates
	if err := db.InitTemplates(); err != nil {
		log.Fatal("Failed to initialize templates:", err)
	}

	// Start session cleanup goroutine
	go func() {
		for {
			db.CleanSessions()
			time.Sleep(1 * time.Hour)
		}
	}()

	middlewareRouter := auth.AuthMiddleware(NewRouter())

	server := &http.Server{
		Addr:         ":9090",
		Handler:      middlewareRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("Server is running on http://localhost:9090")
	return server.ListenAndServe()
}
