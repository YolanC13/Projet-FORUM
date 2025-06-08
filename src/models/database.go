package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDatabase() (*sql.DB, error) {
	dsn := "root:@tcp(127.0.0.1:3306)/forum?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de l'ouverture de la base: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de se connecter à la base: %v", err)
	}

	DB = db
	log.Println("Connexion à la base de données réussie")
	return db, nil
}

func GetDB() *sql.DB {
	return DB
}
