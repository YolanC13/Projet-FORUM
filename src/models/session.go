package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Session struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *Session) Create() error {
	token, err := generateToken()
	if err != nil {
		return err
	}
	s.Token = token
	s.ExpiresAt = time.Now().Add(24 * time.Hour)

	stmt, err := DB.Prepare("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(s.Token, s.UserID, s.ExpiresAt)
	return err
}

func ValidateSession(token string) (int, bool) {
	var userID int
	var expiresAt time.Time

	err := DB.QueryRow("SELECT user_id, expires_at FROM sessions WHERE token = ?", token).
		Scan(&userID, &expiresAt)
	if err != nil {
		return 0, false
	}

	if time.Now().After(expiresAt) {
		DeleteSession(token)
		return 0, false
	}
	return userID, true
}

func DeleteSession(token string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	return err
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
