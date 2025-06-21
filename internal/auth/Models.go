package auth

import (
	"flag"
	"regexp"
	"time"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	passwordRegex = regexp.MustCompile(`^[A-Za-z\d!@#$%^&*()_+]{8,32}$`)
)

type contextKey string

const (
	SessionCookieName            = "session_token01"
	UserKey           contextKey = "userID"
)

var secureCookie = flag.Bool("secure-cookie01", false, "Set secure cookie flag")

type users struct {
	userID     int
	storedHash string
	dbEmail    string
	username   string
}

type Credentials struct {
	Username string
	Email    string
	Password string
}

type Sessions struct {
	UserID    int
	ExpiresAt time.Time
	Username  string
}

type ContextUser struct {
	LoggedIn bool
	UserID   int
	Username string
}
