package services

import (
	"database/sql"
	"fmt"
	"real-time-forum/internal/models"
	"time"
)

type ChatService struct {
	DB *sql.DB
}

func (s *ChatService) SaveMessage(senderID, receiverID int, content string) error {
	if content == "" {
		return fmt.Errorf("message content cannot be empty")
	}

	stmt := `INSERT INTO messages (sender_id, receiver_id, content, created_at) 
	         VALUES (?, ?, ?, ?)`
	
	_, err := s.DB.Exec(stmt, senderID, receiverID, content, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func (s *ChatService) GetMessages(senderID, receiverID, limit, offset int) ([]models.Message, error) {
	query := `SELECT id, sender_id, receiver_id, content, created_at 
	          FROM messages 
	          WHERE (sender_id = ? AND receiver_id = ?) 
	             OR (sender_id = ? AND receiver_id = ?)
	          ORDER BY created_at DESC 
	          LIMIT ? OFFSET ?`
	
	rows, err := s.DB.Query(query, senderID, receiverID, receiverID, senderID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		messages = append(messages, msg)
	}
	return messages, nil
}