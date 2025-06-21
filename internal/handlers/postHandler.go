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

// Modified AddPostHandler
func AddPostHandler(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(auth.UserKey).(auth.ContextUser)

	if r.URL.Path != "/add-post" {
		db.HandleError(w, http.StatusNotFound, "Pge not found")
	}

	categories, err := getCategories()
	if err != nil {
		log.Printf("err: %v\n", err)
		db.HandleError(w, http.StatusInternalServerError, "Error loading categories")
		return
	}

	if r.Method == http.MethodGet {
		db.RenderTemplate(w, "add_post", map[string]interface{}{
			"Title":              "Add Post",
			"Categories":         categories,
			"LoggedIn":           userData.LoggedIn,
			"Username":           userData.Username,
			"SelectedCategories": []int{},
		})
		return
	}

	if r.Method == http.MethodPost {
		// Begin transaction
		tx, err := db.DB.Begin()
		if err != nil {
			log.Printf("err: %v\n", err)
			db.HandleError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		defer tx.Rollback()

		// Insert post
		Title := r.FormValue("title")
		Content := r.FormValue("content")
		categoryStrs := r.Form["categories"]

		escapedTitle := html.EscapeString(Title)
		escapedContent := html.EscapeString(Content)

		var selectedCategories []int
		for _, catID := range categoryStrs {
			if id, err := strconv.Atoi(catID); err == nil {
				if id < 1 || id > 5 {
					// If category is out of range, skip it.
					continue
				}
				selectedCategories = append(selectedCategories, id)
			}
		}

		if strings.TrimSpace(escapedTitle) == "" || strings.TrimSpace(escapedContent) == "" || len(selectedCategories) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			db.RenderTemplate(w, "add_post", map[string]interface{}{
				"Title":              "Add Post",
				"Categories":         categories,
				"LoggedIn":           userData.LoggedIn,
				"Username":           userData.Username,
				"SelectedCategories": selectedCategories,
				"Error":              "Please fill in all fields and select at least one category.",
			})
			return
		}

		if len(Title) > 50 || len(Content) > 1000 {
			w.WriteHeader(http.StatusBadRequest)
			db.RenderTemplate(w, "add_post", map[string]interface{}{
				"Title":              "Add Post",
				"Categories":         categories,
				"LoggedIn":           userData.LoggedIn,
				"Username":           userData.Username,
				"SelectedCategories": selectedCategories,
				"Error":              "Title or content length exceeded. Title must be <= 50 characters and content <= 1000 characters.",
			})
			return
		}
		result, err := tx.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)",
			userData.UserID, escapedTitle, escapedContent)
		if err != nil {
			log.Printf("err: %v\n", err)
			db.HandleError(w, http.StatusInternalServerError, "Failed to create post")
			return
		}

		postID, err := result.LastInsertId()
		if err != nil {
			log.Printf("err: %v\n", err)
			db.HandleError(w, http.StatusInternalServerError, "Failed to get post id")
			return
		}

		// Handle categories
		for _, catID := range selectedCategories {
			_, err = tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)",
				postID, catID)
			if err != nil {
				log.Printf("err: %v\n", err)
				db.HandleError(w, http.StatusInternalServerError, "Failed to associate categories")
				return
			}
		}

		if err = tx.Commit(); err != nil {
			log.Printf("err: %v\n", err)
			db.HandleError(w, http.StatusInternalServerError, "Failed to complete post creation")
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		log.Printf("err: %v\n", err)
		db.HandleError(w, http.StatusMethodNotAllowed, "Invalid method")
		return
	}
}

// LikePostHandler handles liking a post.
func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(auth.UserKey).(auth.ContextUser)

	postID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("Error converting post ID: %v\n", err)
		db.HandleError(w, http.StatusBadRequest, "Invalid post ID")
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

	// Process a like reaction for the post.
	if err := ProcessReaction(tx, "post", postID, userData.UserID, true); err != nil {
		log.Printf("Error processing post reaction: %v\n", err)
		if err.Error() == "content not found" {
			db.HandleError(w, http.StatusBadRequest, "Post not found")
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

// DislikePostHandler handles disliking a post.
func DislikePostHandler(w http.ResponseWriter, r *http.Request) {
	userData := r.Context().Value(auth.UserKey).(auth.ContextUser)

	postID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Printf("Error converting post ID: %v\n", err)
		db.HandleError(w, http.StatusBadRequest, "Invalid post ID")
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

	// Process a dislike reaction for the post.
	if err := ProcessReaction(tx, "post", postID, userData.UserID, false); err != nil {
		log.Printf("Error processing post reaction: %v\n", err)
		if err.Error() == "content not found" {
			db.HandleError(w, http.StatusBadRequest, "Post not found")
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
