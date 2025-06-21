package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	db "forum/internal/database"

	"golang.org/x/crypto/bcrypt"
)

// Helper function to validate the login credentials
func validateCredentials(cred Credentials) (bool, map[string]string) {
	errors := make(map[string]string)
	valid := true

	// Validate email format
	if !emailRegex.MatchString(cred.Email) {
		errors["Email"] = "Invalid email format"
		valid = false
	}

	// Validate password format
	if !passwordRegex.MatchString(cred.Password) {
		errors["Password"] = "Password must be 8-32 chars, 1 letter & 1 number"
		valid = false
	}

	return valid, errors
}

// Helper function to retrieve user by email
func getUserByEmail(email string) (users, error) {
	prep, err := db.DB.Prepare("SELECT id, email, username, password FROM users WHERE email = ?")
	if err != nil {
		return users{}, fmt.Errorf("error preparing query: %v", err)
	}
	defer prep.Close()

	user := users{}
	if err = prep.QueryRow(email).Scan(&user.userID, &user.dbEmail, &user.username, &user.storedHash); err != nil {
		return users{}, fmt.Errorf("error executing query: %v", err)
	}

	return user, nil
}

// Helper function to delete existing sessions
func deleteExistingSession(userID int) error {
	_, err := db.DB.Exec("DELETE FROM sessions WHERE user_id = ?", userID)
	return err
}

// Helper function to create a new session
func createSession(userID int, sessionID string, expiresAt time.Time) error {
	prep, err := db.DB.Prepare("INSERT INTO sessions (id, user_id, expires_at) VALUES (?,?,?)")
	if err != nil {
		return fmt.Errorf("error preparing insert query: %v", err)
	}
	defer prep.Close()

	_, err = prep.Exec(sessionID, userID, expiresAt)
	return err
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	userData, _ := r.Context().Value(UserKey).(ContextUser)
	if userData.LoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	if r.URL.Path != "/login" {
		db.HandleError(w, http.StatusNotFound, "Page not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		db.RenderTemplate(w, "login", map[string]interface{}{"Title": "Login Page"})
		return

	case http.MethodPost:
		r.ParseForm()
		cred := Credentials{
			Email:    strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
			Password: strings.TrimSpace(r.FormValue("password")),
		}

		// Validate input credentials
		valid, errors := validateCredentials(cred)
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			db.RenderTemplate(w, "login", map[string]interface{}{
				"Title":       "Login",
				"Credentials": cred,
				"Errors":      errors,
			})
			return
		}

		// Fetch user from the database
		user, err := getUserByEmail(cred.Email)
		if err != nil {
			// Prevent timing attacks by always hashing a dummy password
			_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$dummy"), []byte(cred.Password))

			// Show generic error message
			errors["Email"] = "Invalid email or password"
			db.RenderTemplate(w, "login", map[string]interface{}{
				"Title":  "Login",
				"Errors": errors,
			})
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.storedHash), []byte(cred.Password)); err != nil {
			errors["Password"] = "Invalid email or password"
			db.RenderTemplate(w, "login", map[string]interface{}{
				"Title":  "Login",
				"Errors": errors,
			})
			return
		}

		// Remove existing sessions for the user
		if err := deleteExistingSession(user.userID); err != nil {
			db.HandleError(w, http.StatusInternalServerError, "Failed to delete existing session")
			return
		}

		// Create new session
		genSessionID := GenerateSessionID()
		expiresAt := time.Now().Add(24 * time.Hour)
		if genSessionID == "" {
			log.Println("Empty session value")
			db.HandleError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Insert the session into the database
		if err := createSession(user.userID, genSessionID, expiresAt); err != nil {
			db.HandleError(w, http.StatusInternalServerError, "Failed to create session")
			return
		}

		// Set session cookie and redirect
		setSessionCookie(w, genSessionID, expiresAt)
		fmt.Printf("User: %v successfully logged in\n", user.username)
		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		db.HandleError(w, http.StatusMethodNotAllowed, "Invalid method")
	}
}
