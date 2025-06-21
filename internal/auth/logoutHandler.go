package auth

import (
	"fmt"
	"log"
	"net/http"
	"time"

	db "forum/internal/database"
)

// Handles user logout and session removal
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		log.Printf("Error retrieving cookie: %v\n", err)
		db.HandleError(w, http.StatusBadRequest, "Not logged in")
		return
	}

	// Prepare SQL statement to delete session
	prep, err := db.DB.Prepare("DELETE FROM sessions WHERE id = ?")
	if err != nil {
		log.Printf("Error preparing statement: %v\n", err)
		db.HandleError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer prep.Close()

	// Execute the prepared statement
	if _, err := prep.Exec(cookie.Value); err != nil {
		log.Printf("Error executing query: %v\n", err)
		db.HandleError(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	// Expire the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour), // Set cookie expiration to the past to expire it
		MaxAge:   -1,                             // MaxAge set to -1 for immediate expiration
		HttpOnly: true,
		Secure:   *secureCookie,        // Ensure the cookie is sent only over secure channels (https)
		SameSite: http.SameSiteLaxMode, // Or SameSiteStrictMode depending on your preference
	})

	fmt.Println("User logged out successfully")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
