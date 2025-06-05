package forumUtils

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type ConnectedUser struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Session struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
}

type Thread struct {
	ID          int
	Title       string
	Description string
	Tag         string
	AuthorID    int
	Author      User
	State       string
	CreatedAt   time.Time
}

type Message struct {
	ID        int
	ThreadID  int
	AuthorID  int
	Content   string
	Author    User
	CreatedAt time.Time
}

type ThreadList struct {
	Threads []Thread
}

type Like struct {
	ID        int
	UserID    int
	ContentID int
	IsLike    bool
	Scope     bool
}

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GenerateToken() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func GetUserByID(db *sql.DB, userID int) (*User, error) {
	var user User

	err := db.QueryRow("SELECT id, username, email, password, created_at, updated_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("utilisateur non trouvé")
		}
		return nil, fmt.Errorf("erreur lors de la récupération de l'utilisateur: %v", err)
	}

	return &user, nil
}

func CreateSessionToken(db *sql.DB, userID int) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour) // Token valide 24h

	stmt, err := db.Prepare("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	_, err = stmt.Exec(token, userID, expiresAt)
	if err != nil {
		return "", err
	}
	return token, nil
}

func ValidateToken(db *sql.DB, token string) (int, bool) {
	var userID int
	var expiresAt time.Time

	err := db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE token = ?", token).
		Scan(&userID, &expiresAt)
	if err != nil {
		log.Printf("Token non trouvé: %v", err)
		return 0, false
	}

	if time.Now().After(expiresAt) {
		log.Printf("Token expiré")
		DeleteSessionToken(db, token)
		return 0, false
	}
	return userID, true
}

func DeleteSessionToken(db *sql.DB, token string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func CleanExpiredTokens(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM sessions WHERE expires_at < ?", time.Now())
	return err
}

func ConnectDatabase() (*sql.DB, error) {
	dsn := "root:@tcp(127.0.0.1:3306)/forum?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à la base: %v", err)
	}

	log.Println("Connexion à la base de données réussie")

	return db, nil
}

func RegisterUser(db *sql.DB, username, email, hashedPassword string) error {
	var count int

	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("utilisateur ou email déjà existant")
	}

	stmt, err := db.Prepare("INSERT INTO users (username, email, password) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("erreur de préparation: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(username, email, hashedPassword)
	if err != nil {
		return fmt.Errorf("erreur d'exécution: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID: %v", err)
	}

	log.Printf("Utilisateur inséré avec l'ID: %d", id)
	return nil
}

func HashPassword(password string) string {
	hash := sha512.Sum512([]byte(password))
	return hex.EncodeToString(hash[:])
}

func GetThreadList(db *sql.DB) (ThreadList, error) {
	var threads []Thread
	rows, err := db.Query("SELECT id, title, description, tags, author_id, state, created_at FROM threads")

	if err != nil {
		return ThreadList{}, fmt.Errorf("erreur lors de la récupération des threads: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var thread Thread
		if err := rows.Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt); err != nil {
			return ThreadList{}, fmt.Errorf("erreur lors du scan des threads: %v", err)
		}
		userPtr, err := GetUserByID(db, thread.AuthorID)
		if err != nil {
			return ThreadList{}, fmt.Errorf("erreur lors de la récupération de l'auteur du thread: %v", err)
		}
		thread.Author = *userPtr
		threads = append(threads, thread)
	}
	return ThreadList{Threads: threads}, nil
}
