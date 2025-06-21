package handlers

import (
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"

	"forum/internal/auth"

	db "forum/internal/database"
)

// CommentHandler handles displaying the comment form (GET) and processing new comments (POST).
func CommentHandler(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(auth.UserKey).(auth.ContextUser)
	if r.URL.Path != "/add-comment" {
		db.HandleError(w, http.StatusNotFound, "Pge not found")
	}
	if r.Method == http.MethodGet {
		postID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			db.HandleError(w, http.StatusBadRequest, "Invalid comment id")
			return
		}

		db.RenderTemplate(w, "add_comment", map[string]interface{}{
			"Title":    "Add Comment",
			"PostID":   postID,
			"LoggedIn": userData.LoggedIn,
			"Username": userData.Username,
		})
		return
	}
	if r.Method == http.MethodPost {
		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			db.HandleError(w, http.StatusBadRequest, "Invalid post id")
			return
		}
		// check if the post exists
		var exists bool
		err = db.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM posts WHERE id = ?)", postID).Scan(&exists)
		if err != nil {
			db.HandleError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		if !exists {
			db.HandleError(w, http.StatusBadRequest, "Post not found")
			return
		}

		content := r.FormValue("content")
		escapedContent := html.EscapeString(content)

		if strings.TrimSpace(escapedContent) == "" {
			w.WriteHeader(http.StatusBadRequest)
			db.RenderTemplate(w, "add_comment", map[string]interface{}{
				"Title":    "Add Comment",
				"PostID":   postID,
				"LoggedIn": userData.LoggedIn,
				"Username": userData.Username,
				"Error":    "Content cannot be empty",
			})
			return
		}

		_, err = db.DB.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userData.UserID, escapedContent)
		if err != nil {
			db.HandleError(w, http.StatusInternalServerError, "Failed to add comment")
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		db.HandleError(w, http.StatusMethodNotAllowed, "Invalid method")
		return
	}
}

// LikeCommentHandler handles liking a comment.
func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(auth.UserKey).(auth.ContextUser)

	commentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("Error converting comment ID: %v\n", err)
		db.HandleError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	// Begin transaction.
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v\n", err)
		db.HandleError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback()

	// Process a like reaction for the comment.
	if err := ProcessReaction(tx, "comment", commentID, userData.UserID, true); err != nil {
		log.Printf("Error processing comment reaction: %v\n", err)
		if err.Error() == "content not found" {
			db.HandleError(w, http.StatusBadRequest, "Comment not found")
		} else {
			db.HandleError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v\n", err)
		db.HandleError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// DislikeCommentHandler handles disliking a comment.
func DislikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(auth.UserKey).(auth.ContextUser)

	commentID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("Error converting comment ID: %v\n", err)
		db.HandleError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	// Begin transaction.
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v\n", err)
		db.HandleError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer tx.Rollback()

	// Process a dislike reaction for the comment.
	if err := ProcessReaction(tx, "comment", commentID, userData.UserID, false); err != nil {
		log.Printf("Error processing comment reaction: %v\n", err)
		if err.Error() == "content not found" {
			db.HandleError(w, http.StatusBadRequest, "Comment not found")
		} else {
			db.HandleError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v\n", err)
		db.HandleError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
