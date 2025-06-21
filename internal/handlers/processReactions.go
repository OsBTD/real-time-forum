package handlers

import (
	"database/sql"
	"fmt"
)

// ProcessReaction handles toggling a like/dislike reaction for different content types.
// Parameters:
//   - tx: the active transaction.
//   - contentType: a string indicating the type of content ("post" or "comment").
//   - contentID: the ID of the post or comment.
//   - userID: the ID of the user reacting.
//   - liked: true for a like reaction, false for a dislike.
//
// Returns an error if any database operation fails or if the content doesn't exist.
func ProcessReaction(tx *sql.Tx, contentType string, contentID, userID int, liked bool) error {
	var existsQuery, reactionQuery, insertQuery, deleteQuery, updateQuery string

	switch contentType {
	case "post":
		existsQuery = "SELECT EXISTS(SELECT 1 FROM posts WHERE id = ?)"
		reactionQuery = "SELECT liked FROM post_reactions WHERE post_id = ? AND user_id = ?"
		insertQuery = "INSERT INTO post_reactions (post_id, user_id, liked) VALUES (?, ?, ?)"
		deleteQuery = "DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?"
		updateQuery = "UPDATE post_reactions SET liked = ? WHERE post_id = ? AND user_id = ?"
	case "comment":
		existsQuery = "SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)"
		reactionQuery = "SELECT liked FROM comment_reactions WHERE comment_id = ? AND user_id = ?"
		insertQuery = "INSERT INTO comment_reactions (comment_id, user_id, liked) VALUES (?, ?, ?)"
		deleteQuery = "DELETE FROM comment_reactions WHERE comment_id = ? AND user_id = ?"
		updateQuery = "UPDATE comment_reactions SET liked = ? WHERE comment_id = ? AND user_id = ?"
	default:
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	// Check if the content exists.
	var exists bool
	if err := tx.QueryRow(existsQuery, contentID).Scan(&exists); err != nil {
		return fmt.Errorf("error checking existence: %v", err)
	}
	if !exists {
		return fmt.Errorf("content not found")
	}

	// Check for an existing reaction from the user.
	var currentReaction bool
	err := tx.QueryRow(reactionQuery, contentID, userID).Scan(&currentReaction)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("error querying reaction: %v", err)
	}

	if err == sql.ErrNoRows {
		// No reaction exists: insert the new reaction.
		if _, err := tx.Exec(insertQuery, contentID, userID, liked); err != nil {
			return fmt.Errorf("error inserting reaction: %v", err)
		}
	} else {
		// Reaction exists: if it's the same as the new one, toggle it off (delete it);
		// otherwise, update it to the new value.
		if currentReaction == liked {
			if _, err := tx.Exec(deleteQuery, contentID, userID); err != nil {
				return fmt.Errorf("error deleting reaction: %v", err)
			}
		} else {
			if _, err := tx.Exec(updateQuery, liked, contentID, userID); err != nil {
				return fmt.Errorf("error updating reaction: %v", err)
			}
		}
	}

	return nil
}
