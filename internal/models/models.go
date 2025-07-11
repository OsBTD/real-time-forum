package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID           int
	FirstName    string
	LastName     string
	Email        string
	Gender       string
	Age          int
	Nickname     string
	Password     string
	SessionToken sql.NullString
	CreatedAt    time.Time
}

type Post struct {
	ID        int
	UserID    int
	Title     string
	Content   string
	CreatedAt time.Time
}

type Category struct {
	ID     int
	PostID int
	Name   string
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Content   string
	CreatedAt time.Time
}

type Reaction struct {
	ID          int
	UserID      int
	ContentType string
	ContentID   int
	Reaction    string
	CreatedAt   time.Time
}

type Message struct {
	ID         int
	SenderID   int
	ReceiverID int
	Content    string
	CreatedAt  time.Time
}

type WebSocketMessage struct {
	Type    string      `json:"type"` // "message", "typing", "online"
	Payload interface{} `json:"payload"`
}

type ChatMessage struct {
	Sender    string `json:"sender"`
	Receiver  string `json:"receiver"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}
