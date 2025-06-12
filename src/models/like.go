package models

import (
	"fmt"
)

type Like struct {
	ID       int `json:"id"`
	UserID   int `json:"user_id"`
	ThreadID int `json:"thread_id"`
}

func ToggleLike(userID, threadID int) error {
	var exists int
	err := DB.QueryRow("SELECT COUNT(*) FROM likes WHERE user_id = ? AND thread_id = ?", userID, threadID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("erreur vérification like: %v", err)
	}
	if exists > 0 {
		_, err := DB.Exec("DELETE FROM likes WHERE user_id = ? AND thread_id = ?", userID, threadID)
		return err
	}
	_, err = DB.Exec("INSERT INTO likes (user_id, thread_id) VALUES (?, ?)", userID, threadID)
	return err
}

func CountLikes(threadID int) (int, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM likes WHERE thread_id = ?", threadID).Scan(&count)
	return count, err
}
