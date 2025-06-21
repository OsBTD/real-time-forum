package auth

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	db "forum/internal/database"

	"golang.org/x/crypto/bcrypt"
)

// Validate the registration credentials
func validateRegistrationCredentials(cred Credentials) (bool, map[string]string) {
	errors := make(map[string]string)
	valid := true

	if !usernameRegex.MatchString(cred.Username) {
		errors["Username"] = "Username must be 3-20 characters (letters, digits, underscores)"
		valid = false
	}

	if !emailRegex.MatchString(cred.Email) {
		errors["Email"] = "Invalid email format"
		valid = false
	}

	if !passwordRegex.MatchString(cred.Password) {
		errors["Password"] = "Password must be 8-32 chars, 1 letter & 1 number"
		valid = false
	}

	return valid, errors
}

// Check if an email is already in use
func isEmailAlreadyInUse(email string) (bool, error) {
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("error checking email uniqueness: %v", err)
	}
	return count > 0, nil
}

// Hash the password
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hashedPassword), nil
}

// Register the new user in the database
func registerNewUser(username, email, hashedPassword string) error {
	prep, err := db.DB.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing query: %v", err)
	}
	defer prep.Close()

	_, err = prep.Exec(username, email, hashedPassword)
	if err != nil {
		return fmt.Errorf("error executing query: %v", err)
	}

	return nil
}

// Handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	userData, _ := r.Context().Value(UserKey).(ContextUser)
	if userData.LoggedIn {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	if r.URL.Path != "/register" {
		db.HandleError(w, http.StatusNotFound, "Page not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		db.RenderTemplate(w, "register", map[string]interface{}{"Title": "Registration"})
		return

	case http.MethodPost:
		r.ParseForm()
		cred := Credentials{
			Username: html.EscapeString(strings.TrimSpace(r.FormValue("username"))),
			Email:    html.EscapeString(strings.ToLower(strings.TrimSpace(r.FormValue("email")))),
			Password: html.EscapeString(strings.TrimSpace(r.FormValue("password"))),
		}

		// Validate registration credentials
		valid, errors := validateRegistrationCredentials(cred)
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			db.RenderTemplate(w, "register", map[string]interface{}{
				"Title":       "Registration",
				"Credentials": cred,
				"Errors":      errors,
			})
			return
		}

		// Check if the email is already in use
		emailInUse, err := isEmailAlreadyInUse(cred.Email)
		if err != nil {
			log.Printf("Error checking email uniqueness: %v", err)
			db.HandleError(w, http.StatusInternalServerError, "Database error")
			return
		}
		if emailInUse {
			errors["Email"] = "Email already in use"
			w.WriteHeader(http.StatusBadRequest)
			db.RenderTemplate(w, "register", map[string]interface{}{
				"Title":       "Registration",
				"Credentials": cred,
				"Errors":      errors,
			})
			return
		}

		// Hash the password
		hashedPassword, err := hashPassword(cred.Password)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			db.HandleError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Register the new user in the database
		if err := registerNewUser(cred.Username, cred.Email, hashedPassword); err != nil {
			log.Printf("Error registering user: %v", err)
			db.HandleError(w, http.StatusInternalServerError, "Registration failed")
			return
		}

		fmt.Printf("User: %v successfully registered\n", cred.Username)

		// Redirect to login page
		http.Redirect(w, r, "/login", http.StatusSeeOther)

	default:
		db.HandleError(w, http.StatusMethodNotAllowed, "Invalid method")
	}
}
