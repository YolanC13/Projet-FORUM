package main

import (
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"text/template"
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
	State       string
	CreatedAt   time.Time
}

type ThreadList struct {
	Threads []Thread
}

func main() {
	InitialiseServer()
}

func InitialiseServer() {
	temp, errTemp := template.ParseGlob("templates/*.html")
	if errTemp != nil {
		log.Printf("Error parsing template: %v\n", errTemp)
		return
	}

	//Menu principal
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
	})

	http.HandleFunc("/mainMenu", func(w http.ResponseWriter, r *http.Request) {
		db, err := ConnectDatabase()
		if err != nil {
			log.Printf("Erreur de connexion à la base de données: %v", err)
			http.Error(w, "Database error", 500)
			return
		}
		defer db.Close()
		threadList, err := GetThreadList(db)
		if err != nil {
			log.Printf("Erreur lors de la récupération des threads: %v", err)
			http.Error(w, "Erreur lors de la récupération des threads", 500)
			return
		}

		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Aucun token trouvé: %v", err)
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		userID, valid := ValidateToken(db, cookie.Value)
		if !valid {
			log.Printf("Token invalide ou expiré")

			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		var user ConnectedUser
		err = db.QueryRow("SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", userID).
			Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Erreur lors de la récupération du profil: %v", err)
			http.Error(w, "Utilisateur non trouvé", 404)
			return
		}

		data := struct {
			Threads ThreadList
			User    ConnectedUser
		}{
			Threads: threadList,
			User:    user,
		}

		if err := temp.ExecuteTemplate(w, "mainMenu", data); err != nil {
			log.Printf("Erreur lors de l'exécution du template: %v", err)
			http.Error(w, "Erreur lors de l'affichage de la page", 500)
			return
		}
	})

	// Profil utilisateur
	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		// Récupérer le token depuis le cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Aucun token trouvé: %v", err)
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		db, err := ConnectDatabase()
		if err != nil {
			log.Printf("Erreur de connexion à la base de données: %v", err)
			http.Error(w, "Database error", 500)
			return
		}
		defer db.Close()

		userID, valid := ValidateToken(db, cookie.Value)
		if !valid {
			log.Printf("Token invalide ou expiré")
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		// Récupérer les informations utilisateur
		var user ConnectedUser
		err = db.QueryRow("SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", userID).
			Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Erreur lors de la récupération du profil: %v", err)
			http.Error(w, "Utilisateur non trouvé", 404)
			return
		}

		temp.ExecuteTemplate(w, "profile", user)
	})

	// Création de thread
	http.HandleFunc("/createThread", func(w http.ResponseWriter, r *http.Request) {
		// Récupérer le token depuis le cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Aucun token trouvé: %v", err)
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		db, err := ConnectDatabase()
		if err != nil {
			log.Printf("Erreur de connexion à la base de données: %v", err)
			http.Error(w, "Database error", 500)
			return
		}
		defer db.Close()

		userID, valid := ValidateToken(db, cookie.Value)
		if !valid {
			log.Printf("Token invalide ou expiré")
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		// Récupérer les informations utilisateur
		var user ConnectedUser
		err = db.QueryRow("SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", userID).
			Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Erreur lors de la récupération du profil: %v", err)
			http.Error(w, "Utilisateur non trouvé", 404)
			return
		}

		temp.ExecuteTemplate(w, "createThread", user)
	})

	http.HandleFunc("/createThread/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			title := r.FormValue("thread_title")
			description := r.FormValue("thread_desc")
			tags := r.FormValue("thread_tag")

			if title == "" || description == "" || tags == "" {
				log.Println("Données manquantes dans le formulaire")
				http.Error(w, "Tous les champs sont requis", 400)
				return
			}

			// Récupérer le token depuis le cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				log.Printf("Aucun token trouvé: %v", err)
				http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
				return
			}

			db, err := ConnectDatabase()
			if err != nil {
				log.Printf("Erreur de connexion à la base de données: %v", err)
				http.Error(w, "Database error", 500)
				return
			}
			defer db.Close()

			// Vérifier si le token est valide et récupérer l'userID
			userID, valid := ValidateToken(db, cookie.Value)
			if !valid {
				log.Printf("Token invalide ou expiré")
				http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
				return
			}

			// Insérer le thread dans la base de données
			stmt, err := db.Prepare("INSERT INTO threads (title, description, tags, author_id) VALUES (?, ?, ?, ?)")
			if err != nil {
				log.Printf("Erreur de préparation de la requête: %v", err)
				http.Error(w, "Erreur serveur", 500)
				return
			}
			defer stmt.Close()

			_, err = stmt.Exec(title, description, tags, userID)
			if err != nil {
				log.Printf("Erreur lors de l'insertion du thread: %v", err)
				http.Error(w, "Erreur lors de la création du thread", 500)
				return
			}

			log.Printf("Thread créé avec succès par l'utilisateur %d", userID)

			http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
			return
		}
	})

	//Login / Register
	http.HandleFunc("/connexionPage", func(w http.ResponseWriter, r *http.Request) {
		temp.ExecuteTemplate(w, "connexionPage", nil)
	})

	http.HandleFunc("/loginPage", func(w http.ResponseWriter, r *http.Request) {
		temp.ExecuteTemplate(w, "loginPage", nil)
	})

	http.HandleFunc("/registerPage", func(w http.ResponseWriter, r *http.Request) {
		temp.ExecuteTemplate(w, "registerPage", nil)
	})

	http.HandleFunc("/userlogin", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == "" || password == "" {
				log.Println("Données manquantes dans le formulaire")
				http.Error(w, "Tous les champs sont requis", 400)
				return
			}

			log.Printf("Tentative de connexion pour: %s", username)

			hashedPassword := HashPassword(password)

			db, err := ConnectDatabase()
			if err != nil {
				log.Printf("Erreur de connexion à la base de données: %v", err)
				http.Error(w, "Database error", 500)
				return
			}
			defer db.Close()

			var user ConnectedUser
			err = db.QueryRow("SELECT id, username FROM users WHERE username = ? AND password = ?", username, hashedPassword).
				Scan(&user.ID, &user.Username)
			if err != nil {
				log.Printf("Erreur lors de la récupération de l'utilisateur: %v", err)
				http.Error(w, "Identifiants incorrects", 401)
				return
			}

			// Créer un token de session
			token, err := CreateSessionToken(db, user.ID)
			if err != nil {
				log.Printf("Erreur création token: %v", err)
				http.Error(w, "Erreur serveur", 500)
				return
			}

			// Définir le cookie avec le token
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Expires:  time.Now().Add(24 * time.Hour), // Expire dans 24h
				HttpOnly: true,
				Secure:   false, // Mettre à true en production avec HTTPS
				Path:     "/",
			})

			log.Printf("Utilisateur %s connecté avec succès", user.Username)
			http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/userRegister", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")

			if username == "" || email == "" || password == "" {
				log.Println("Données manquantes dans le formulaire")
				http.Error(w, "Tous les champs sont requis", 400)
				return
			}

			log.Printf("Tentative d'enregistrement pour: %s, %s", username, email)

			hashedPassword := HashPassword(password)

			db, err := ConnectDatabase()
			if err != nil {
				log.Printf("Erreur de connexion à la base de données: %v", err)
				http.Error(w, "Database error", 500)
				return
			}
			defer db.Close()

			err = RegisterUser(db, username, email, hashedPassword)
			if err != nil {
				log.Printf("Erreur lors de l'enregistrement: %v", err)
				http.Error(w, "Failed to register: "+err.Error(), 500)
				return
			}

			log.Printf("Utilisateur %s enregistré avec succès", username)
			http.Redirect(w, r, "/loginPage", http.StatusSeeOther)
		}
	})

	// Route de déconnexion
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil {
			db, err := ConnectDatabase()
			if err == nil {
				DeleteSessionToken(db, cookie.Value)
				db.Close()
			}
		}

		// Supprimer le cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Now().Add(-1 * time.Hour),
			HttpOnly: true,
			Path:     "/",
		})

		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
	})

	//Lance le serveur
	RunServer()
}

func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
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

	// Vérifier si le token n'est pas expiré
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

func RunServer() {
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir(".templates/styles"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(".templates/images"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir(".templates/fonts"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir(".templates/scripts"))))

	fmt.Println("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func RegisterUser(db *sql.DB, username, email, hashedPassword string) error {
	// Vérifier si l'utilisateur existe déjà
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
		threads = append(threads, thread)
	}

	return ThreadList{Threads: threads}, nil
}
