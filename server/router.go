package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"forum/internal/auth"
	db "forum/internal/database"

	H "forum/internal/handlers"
)

func NewRouter() *http.ServeMux {
	// Create router
	router := http.NewServeMux()

	// Public routes
	router.HandleFunc("/", H.HomeHandler)
	router.HandleFunc("/login", auth.LoginHandler)
	router.HandleFunc("/register", auth.RegisterHandler)
	router.HandleFunc("/logout", auth.LogoutHandler)

	// Protected routes
	router.Handle("/add-post", auth.RequireAuth(http.HandlerFunc(H.AddPostHandler)))
	router.Handle("/add-comment", auth.RequireAuth(http.HandlerFunc(H.CommentHandler)))
	router.Handle("/like-post", auth.RequireAuth(http.HandlerFunc(H.LikePostHandler)))
	router.Handle("/dislike-post", auth.RequireAuth(http.HandlerFunc(H.DislikePostHandler)))
	router.Handle("/like-comment", auth.RequireAuth(http.HandlerFunc(H.LikeCommentHandler)))
	router.Handle("/dislike-comment", auth.RequireAuth(http.HandlerFunc(H.DislikeCommentHandler)))

	// Static files
	// router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// static files handler plus checks for directories and ".." and forbids users from accessing
	router.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		subPath := strings.TrimPrefix(r.URL.Path, "/static/")
		filePath := filepath.Join("static", subPath)

		info, err := os.Stat(filePath)
		if err != nil {
			db.HandleError(w, http.StatusNotFound, "Page not found")
			return
		}
		if info.IsDir() {
			db.HandleError(w, http.StatusForbidden, "Access is forbidden")
			return
		}

		http.ServeFile(w, r, filePath)
	})

	return router
}
