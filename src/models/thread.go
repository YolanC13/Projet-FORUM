package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Thread struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Tag         string    `json:"tag"`
	AuthorID    int       `json:"author_id"`
	Author      User      `json:"author"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	LikeCount   int       `json:"like_count"`
}

type ThreadList struct {
	Threads []Thread `json:"threads"`
}

func (t *Thread) Create() error {
	stmt, err := DB.Prepare("INSERT INTO threads (title, description, tags, author_id) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("erreur de préparation: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(t.Title, t.Description, t.Tag, t.AuthorID)
	if err != nil {
		return fmt.Errorf("erreur d'insertion: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID: %v", err)
	}
	t.ID = int(id)
	return nil
}

func GetThreadByID(id int) (*Thread, error) {
	var thread Thread
	err := DB.QueryRow("SELECT id, title, description, tags, author_id, state, created_at FROM threads WHERE id = ?", id).
		Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("thread non trouvé: %v", err)
	}

	// Charger l'auteur
	author, err := GetUserByID(thread.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération auteur: %v", err)
	}
	thread.Author = *author
	return &thread, nil
}

func GetAllThreads() (ThreadList, error) {
	var threads []Thread
	rows, err := DB.Query("SELECT id, title, description, tags, author_id, state, created_at FROM threads")
	if err != nil {
		return ThreadList{}, fmt.Errorf("erreur récupération threads: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt); err != nil {
			return ThreadList{}, fmt.Errorf("erreur scan thread: %v", err)
		}

		author, err := GetUserByID(thread.AuthorID)
		if err != nil {
			return ThreadList{}, fmt.Errorf("erreur récupération auteur: %v", err)
		}
		thread.Author = *author

		likeCount, _ := CountLikes(thread.ID)
		thread.LikeCount = likeCount

		threads = append(threads, thread)
	}
	return ThreadList{Threads: threads}, nil
}

func GetAllThreadsSortedByRecent() (ThreadList, error) {
	var threads []Thread
	rows, err := DB.Query("SELECT id, title, description, tags, author_id, state, created_at FROM threads ORDER BY created_at DESC")
	if err != nil {
		return ThreadList{}, fmt.Errorf("erreur récupération threads: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt); err != nil {
			return ThreadList{}, fmt.Errorf("erreur scan thread: %v", err)
		}
		author, err := GetUserByID(thread.AuthorID)
		if err != nil {
			return ThreadList{}, fmt.Errorf("erreur récupération auteur: %v", err)
		}
		thread.Author = *author
		likeCount, _ := CountLikes(thread.ID)
		thread.LikeCount = likeCount
		threads = append(threads, thread)
	}
	return ThreadList{Threads: threads}, nil
}

func GetAllThreadsSortedByPopularity() (ThreadList, error) {
	var threads []Thread
	rows, err := DB.Query(`
        SELECT t.id, t.title, t.description, t.tags, t.author_id, t.state, t.created_at, 
        COUNT(l.id) as like_count
        FROM threads t
        LEFT JOIN likes l ON t.id = l.thread_id
        GROUP BY t.id
        ORDER BY like_count DESC, t.created_at DESC`)
	if err != nil {
		return ThreadList{}, fmt.Errorf("erreur récupération threads: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		var likeCount int
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt, &likeCount); err != nil {
			return ThreadList{}, fmt.Errorf("erreur scan thread: %v", err)
		}
		author, err := GetUserByID(thread.AuthorID)
		if err != nil {
			return ThreadList{}, fmt.Errorf("erreur récupération auteur: %v", err)
		}
		thread.Author = *author
		thread.LikeCount = likeCount
		threads = append(threads, thread)
	}
	return ThreadList{Threads: threads}, nil
}

func (t *Thread) Delete() error {
	_, err := DB.Exec("DELETE FROM threads WHERE id = ?", t.ID)
	return err
}

func AdminDeleteThread(threadID int) error {
	_, err := DB.Exec("DELETE FROM threads WHERE id = ?", threadID)
	return err
}

func GetThreadsByTags(tags []string) (ThreadList, error) {
	var threads []Thread
	var rows *sql.Rows
	var err error

	if len(tags) == 0 {
		return GetAllThreads()
	}

	query := "SELECT id, title, description, tags, author_id, state, created_at FROM threads WHERE "
	args := make([]interface{}, len(tags))
	for i, tag := range tags {
		if i > 0 {
			query += " OR "
		}
		query += "tags = ?"
		args[i] = tag
	}
	rows, err = DB.Query(query, args...)
	if err != nil {
		return ThreadList{}, fmt.Errorf("erreur récupération threads: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt); err != nil {
			return ThreadList{}, fmt.Errorf("erreur scan thread: %v", err)
		}
		author, err := GetUserByID(thread.AuthorID)
		if err != nil {
			return ThreadList{}, fmt.Errorf("erreur récupération auteur: %v", err)
		}
		thread.Author = *author
		likeCount, _ := CountLikes(thread.ID)
		thread.LikeCount = likeCount
		threads = append(threads, thread)
	}
	return ThreadList{Threads: threads}, nil
}

func SearchThreads(query string, tags []string) (ThreadList, error) {
	var threads []Thread
	var rows *sql.Rows
	var err error

	baseQuery := "SELECT id, title, description, tags, author_id, state, created_at FROM threads WHERE (title LIKE ? OR description LIKE ?)"
	args := []interface{}{"%" + query + "%", "%" + query + "%"}

	if len(tags) > 0 {
		baseQuery += " AND ("
		for i, tag := range tags {
			if i > 0 {
				baseQuery += " OR "
			}
			baseQuery += "tags = ?"
			args = append(args, tag)
		}
		baseQuery += ")"
	}

	rows, err = DB.Query(baseQuery, args...)
	if err != nil {
		return ThreadList{}, fmt.Errorf("erreur recherche threads: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt); err != nil {
			return ThreadList{}, fmt.Errorf("erreur scan thread: %v", err)
		}
		author, err := GetUserByID(thread.AuthorID)
		if err != nil {
			return ThreadList{}, fmt.Errorf("erreur récupération auteur: %v", err)
		}
		thread.Author = *author
		likeCount, _ := CountLikes(thread.ID)
		thread.LikeCount = likeCount
		threads = append(threads, thread)
	}
	return ThreadList{Threads: threads}, nil
}
