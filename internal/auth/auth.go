package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"real-time-forum/internal/models"

	"golang.org/x/crypto/bcrypt"
)

const (
	SessionTokenLength = 32
	SessionDuration    = 24 * time.Hour
)

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

func ComparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func GenerateSessionToken() (string, error) {
	token := make([]byte, SessionTokenLength)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(token), nil
}

func ValidateSession(db *sql.DB, token string) (*models.User, error) {
	if token == "" {
		return nil, errors.New("empty session token")
	}

	query := `SELECT id, first_name, last_name, email, gender, age, nickname 
          FROM users WHERE session_token = ?`

	row := db.QueryRow(query, token)
	user := &models.User{}
	err := row.Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Gender,
		&user.Age,
		&user.Nickname,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}
	return user, nil
}
