package auth

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	db "forum/internal/database"
)

// Checks the session cookie and adds user data to the context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Default context with no user
		ctx := r.Context()
		userData := ContextUser{
			LoggedIn: false,
			UserID:   0,
			Username: "",
		}

		// Check for session cookie
		cookie, err := r.Cookie(SessionCookieName)
		if err == nil {
			// Use a prepared statement to securely query the database for the session data
			prep, err := db.DB.Prepare(`
                SELECT s.user_id, s.expires_at, u.username 
                FROM sessions s
                JOIN users u ON s.user_id = u.id 
                WHERE s.id = ? AND s.expires_at > ?
            `)
			if err != nil {
				log.Printf("DB prepare error: %v\n", err)
				db.HandleError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
			defer prep.Close()

			// Query the session data based on the cookie value and current time
			var session Sessions
			err = prep.QueryRow(cookie.Value, time.Now()).Scan(
				&session.UserID,
				&session.ExpiresAt,
				&session.Username,
			)

			// Handle session query results
			if err != nil {
				if err == sql.ErrNoRows {
					log.Printf("No session found for the provided cookie. Invalid session.\n")
				} else {
					log.Printf("Error during session lookup: %v\n", err)
					db.HandleError(w, http.StatusInternalServerError, "Internal server error")
					return
				}

				// Handle invalid or expired session: remove the cookie
				http.SetCookie(w, &http.Cookie{
					Name:     SessionCookieName,
					Value:    "",
					Path:     "/",
					Expires:  time.Now().Add(-1 * time.Hour), // Expire the cookie immediately
					MaxAge:   -1,                             // Immediately expire the cookie
					HttpOnly: true,
					Secure:   *secureCookie,
					SameSite: http.SameSiteLaxMode,
				})
			} else {
				// Session is valid, set user data
				userData.LoggedIn = true
				userData.UserID = session.UserID
				userData.Username = session.Username
				log.Printf("Session set to user: %v", session.Username)
			}
		}

		// Add the user data to the request context
		ctx = context.WithValue(ctx, UserKey, userData)

		// Continue the request chain
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Ensures the user is logged in before allowing access to a route
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userData, ok := r.Context().Value(UserKey).(ContextUser)
		if !ok || !userData.LoggedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Continue with the request if the user is logged in
		next.ServeHTTP(w, r)
	})
}
