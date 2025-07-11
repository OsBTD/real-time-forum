package services

import (
	"database/sql"
	"errors"
	"fmt"
	"real-time-forum/internal/models"
	"time"
)

type PostService struct {
	DB *sql.DB
}

func (s *PostService) CreatePost(userID int, title, content string, categories []string) error {
	if title == "" || content == "" || len(categories) == 0 {
		return errors.New("all fields are required")
	}

	// Start transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback()

	// Insert post
	postStmt := `INSERT INTO posts (user_id, title, content, created_at) VALUES (?, ?, ?, ?)`
	res, err := tx.Exec(postStmt, userID, title, content, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("post creation failed: %w", err)
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get post ID: %w", err)
	}

	// Insert categories
	categoryStmt := `INSERT INTO categories (post_id, category) VALUES (?, ?)`
	for _, category := range categories {
		if _, err := tx.Exec(categoryStmt, postID, category); err != nil {
			return fmt.Errorf("category insertion failed: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("transaction commit failed: %w", err)
	}

	return nil
}

func (s *PostService) GetPosts(limit, offset int) ([]models.Post, error) {
	query := `SELECT id, user_id, title, content, created_at 
	          FROM posts ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := s.DB.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// Similar methods for comments and reactions
func (s *PostService) CreateComment(postID, userID int, content string) error {
	if content == "" {
		return errors.New("comment cannot be empty")
	}

	stmt := `INSERT INTO comments (post_id, user_id, comment, created_at) 
	         VALUES (?, ?, ?, ?)`

	_, err := s.DB.Exec(stmt, postID, userID, content, time.Now().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}
