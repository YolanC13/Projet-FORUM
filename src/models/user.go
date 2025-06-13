package models

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Salt      string    `json:"-"`
	IsAdmin   bool      `json:"is_admin"`
	IsBanned  bool      `json:"is_banned"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) Create() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", u.Username, u.Email).Scan(&count)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification: %v", err)
	}
	if count > 0 {
		return fmt.Errorf("utilisateur ou email déjà existant")
	}

	salt, err := generateSalt()
	if err != nil {
		return fmt.Errorf("erreur génération salt: %v", err)
	}
	u.Salt = salt
	u.Password = hashPassword(u.Password, salt)

	stmt, err := DB.Prepare("INSERT INTO users (username, email, password, salt) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("erreur de préparation: %v", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(u.Username, u.Email, u.Password, u.Salt)
	if err != nil {
		return fmt.Errorf("erreur d'exécution: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID: %v", err)
	}
	u.ID = int(id)
	return nil
}

func GetUserByID(id int) (*User, error) {
	var user User
	err := DB.QueryRow("SELECT id, username, email, password, salt, is_admin, is_banned, created_at, updated_at FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.Salt, &user.IsAdmin, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("utilisateur non trouvé")
		}
		return nil, fmt.Errorf("erreur lors de la récupération: %v", err)
	}
	return &user, nil
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.QueryRow("SELECT id, username, password, salt FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Salt)

	if err != nil {
		return nil, fmt.Errorf("utilisateur non trouvé")
	}
	return &user, nil
}

func (u *User) ValidatePassword(password string) bool {
	hashedPassword := hashPassword(password, u.Salt)
	return hashedPassword == u.Password
}

func generateSalt() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashPassword(password, salt string) string {
	hash := sha512.Sum512([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

func BanUser(userID int) error {
	_, err := DB.Exec("UPDATE users SET is_banned = 1 WHERE id = ?", userID)
	return err
}

func UnbanUser(userID int) error {
	_, err := DB.Exec("UPDATE users SET is_banned = 0 WHERE id = ?", userID)
	return err
}

func GetAllUsers() ([]User, error) {
	var users []User
	rows, err := DB.Query("SELECT id, username, email, is_admin, is_banned, created_at, updated_at FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}
