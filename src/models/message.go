package models

import (
	"fmt"
	"time"
)

type Message struct {
	ID        int       `json:"id"`
	ThreadID  int       `json:"thread_id"`
	AuthorID  int       `json:"author_id"`
	Content   string    `json:"content"`
	Author    User      `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *Message) Create() error {
	stmt, err := DB.Prepare("INSERT INTO messages (thread_id, author_id, content) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("erreur préparation: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(m.ThreadID, m.AuthorID, m.Content)
	if err != nil {
		return fmt.Errorf("erreur insertion: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID: %v", err)
	}
	m.ID = int(id)
	return nil
}

func GetMessagesByThreadID(threadID int) ([]Message, error) {
	var messages []Message
	rows, err := DB.Query("SELECT id, thread_id, author_id, content, created_at FROM messages WHERE thread_id = ?", threadID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération messages: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.ID, &message.ThreadID, &message.AuthorID, &message.Content, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("erreur scan message: %v", err)
		}

		author, err := GetUserByID(message.AuthorID)
		if err != nil {
			return nil, fmt.Errorf("erreur récupération auteur: %v", err)
		}
		message.Author = *author
		messages = append(messages, message)
	}
	return messages, nil
}

func GetMessagesByThreadIDPaginated(threadID, page, pageSize int) ([]Message, int, error) {
	var messages []Message
	var total int

	// Compte le nombre total de messages pour la pagination
	err := DB.QueryRow("SELECT COUNT(*) FROM messages WHERE thread_id = ?", threadID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage messages: %v", err)
	}

	offset := (page - 1) * pageSize
	rows, err := DB.Query("SELECT id, thread_id, author_id, content, created_at FROM messages WHERE thread_id = ? ORDER BY created_at ASC LIMIT ? OFFSET ?", threadID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération messages: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.ID, &message.ThreadID, &message.AuthorID, &message.Content, &message.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("erreur scan message: %v", err)
		}
		author, err := GetUserByID(message.AuthorID)
		if err == nil {
			message.Author = *author
		}
		messages = append(messages, message)
	}
	return messages, total, nil
}

func AdminDeleteMessage(messageID int) error {
	_, err := DB.Exec("DELETE FROM messages WHERE id = ?", messageID)
	return err
}

func GetAllMessages() ([]Message, error) {
	var messages []Message
	rows, err := DB.Query("SELECT id, thread_id, author_id, content, created_at FROM messages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.ID, &message.ThreadID, &message.AuthorID, &message.Content, &message.CreatedAt); err != nil {
			continue
		}
		author, err := GetUserByID(message.AuthorID)
		if err == nil {
			message.Author = *author
		}
		messages = append(messages, message)
	}
	return messages, nil
}
