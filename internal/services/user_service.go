package services

import (
	"database/sql"
	"errors"
	"fmt"
	"real-time-forum/internal/auth"
	"real-time-forum/internal/models"
)

type UserService struct {
	DB *sql.DB
}

func (s *UserService) Register(user *models.User) error {
	// Validate input
	if user.Nickname == "" || user.Email == "" || user.Password == "" {
		return errors.New("all fields are required")
	}

	// Check if email or nickname exists
	var count int
	err := s.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR nickname = ?", 
		user.Email, user.Nickname).Scan(&count)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}
	if count > 0 {
		return errors.New("email or nickname already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}
	user.Password = hashedPassword

	// Insert user
	stmt := `INSERT INTO users (first_name, last_name, email, gender, age, nickname, password) 
	         VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	_, err = s.DB.Exec(stmt,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Gender,
		user.Age,
		user.Nickname,
		user.Password,
	)
	
	if err != nil {
		return fmt.Errorf("user creation failed: %w", err)
	}
	return nil
}

func (s *UserService) Login(emailOrNickname, password string) (string, error) {
	var user models.User
	query := `SELECT id, password FROM users WHERE email = ? OR nickname = ?`
	
	err := s.DB.QueryRow(query, emailOrNickname, emailOrNickname).Scan(
		&user.ID,
		&user.Password,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid credentials")
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	// Verify password
	if !auth.ComparePasswords(user.Password, password) {
		return "", errors.New("invalid credentials")
	}

	// Generate session token
	token, err := auth.GenerateSessionToken()
	if err != nil {
		return "", fmt.Errorf("session creation failed: %w", err)
	}

	// Update session token
	_, err = s.DB.Exec("UPDATE users SET session_token = ? WHERE id = ?", token, user.ID)
	if err != nil {
		return "", fmt.Errorf("session update failed: %w", err)
	}

	return token, nil
}

func (s *UserService) Logout(token string) error {
	_, err := s.DB.Exec("UPDATE users SET session_token = NULL WHERE session_token = ?", token)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	return nil
}

func (s *UserService) GetAllUsers() ([]string, error) {
	rows, err := s.DB.Query("SELECT nickname FROM users ORDER BY nickname ASC")
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var nickname string
		if err := rows.Scan(&nickname); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		users = append(users, nickname)
	}
	return users, nil
}