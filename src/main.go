package main

import (
	"fmt"
	forumUtils "forumUtils/utils"
	"log"
	"net/http"
	"text/template"
	"time"
)

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
		db, err := forumUtils.ConnectDatabase()
		if err != nil {
			log.Printf("Erreur de connexion à la base de données: %v", err)
			http.Error(w, "Database error", 500)
			return
		}
		defer db.Close()

		threadList, err := forumUtils.GetThreadList(db)
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

		userID, valid := forumUtils.ValidateToken(db, cookie.Value)
		if !valid {
			log.Printf("Token invalide ou expiré")
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		var user forumUtils.ConnectedUser
		err = db.QueryRow("SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", userID).
			Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Erreur lors de la récupération du profil: %v", err)
			http.Error(w, "Utilisateur non trouvé", 404)
			return
		}

		data := struct {
			Threads forumUtils.ThreadList
			User    forumUtils.ConnectedUser
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

		db, err := forumUtils.ConnectDatabase()
		if err != nil {
			log.Printf("Erreur de connexion à la base de données: %v", err)
			http.Error(w, "Database error", 500)
			return
		}

		defer db.Close()
		userID, valid := forumUtils.ValidateToken(db, cookie.Value)
		if !valid {
			log.Printf("Token invalide ou expiré")
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		// Récupérer les informations utilisateur
		var user forumUtils.ConnectedUser
		err = db.QueryRow("SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", userID).
			Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Erreur lors de la récupération du profil: %v", err)
			http.Error(w, "Utilisateur non trouvé", 404)
			return
		}
		temp.ExecuteTemplate(w, "profile", user)
	})

	// Thread
	http.HandleFunc("/createThread", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Aucun token trouvé: %v", err)
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		db, err := forumUtils.ConnectDatabase()
		if err != nil {
			log.Printf("Erreur de connexion à la base de données: %v", err)
			http.Error(w, "Database error", 500)
			return
		}

		defer db.Close()
		userID, valid := forumUtils.ValidateToken(db, cookie.Value)
		if !valid {
			log.Printf("Token invalide ou expiré")
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		// Récupérer les informations utilisateur
		var user forumUtils.ConnectedUser
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

			cookie, err := r.Cookie("session_token")
			if err != nil {
				log.Printf("Aucun token trouvé: %v", err)
				http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
				return
			}

			db, err := forumUtils.ConnectDatabase()
			if err != nil {
				log.Printf("Erreur de connexion à la base de données: %v", err)
				http.Error(w, "Database error", 500)
				return
			}

			defer db.Close()

			userID, valid := forumUtils.ValidateToken(db, cookie.Value)
			if !valid {
				log.Printf("Token invalide ou expiré")
				http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
				return
			}

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

	http.HandleFunc("/thread", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Aucun token trouvé: %v", err)
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		db, err := forumUtils.ConnectDatabase()
		if err != nil {
			log.Printf("Erreur de connexion à la base de données: %v", err)
			http.Error(w, "Database error", 500)
			return
		}
		defer db.Close()

		userID, valid := forumUtils.ValidateToken(db, cookie.Value)
		if !valid {
			log.Printf("Token invalide ou expiré")
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}

		var user forumUtils.ConnectedUser

		err = db.QueryRow("SELECT id, username, email, created_at, updated_at FROM users WHERE id = ?", userID).
			Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Erreur lors de la récupération du profil: %v", err)
			http.Error(w, "Utilisateur non trouvé", 404)
			return
		}

		threadID := r.URL.Query().Get("id")

		if threadID == "" {
			log.Println("ID du thread manquant dans la requête")
			http.Error(w, "ID du thread manquant", 400)
			return
		}

		var thread forumUtils.Thread

		err = db.QueryRow("SELECT id, title, description, tags, author_id, state, created_at FROM threads WHERE id = ?", threadID).
			Scan(&thread.ID, &thread.Title, &thread.Description, &thread.Tag, &thread.AuthorID, &thread.State, &thread.CreatedAt)

		if err != nil {
			log.Printf("Erreur lors de la récupération du thread: %v", err)
			http.Error(w, "Thread non trouvé", 404)
			return
		}

		// Récupérer l'objet User pour l'auteur
		author, err := forumUtils.GetUserByID(db, thread.AuthorID)
		if err != nil {
			log.Printf("Erreur lors de la récupération de l'auteur: %v", err)
			http.Error(w, "Erreur lors de la récupération de l'auteur", 500)
			return
		}
		thread.Author = *author

		var messages []forumUtils.Message
		rows, err := db.Query("SELECT id, thread_id, author_id, content, created_at FROM messages WHERE thread_id = ?", threadID)
		if err != nil {
			log.Printf("Erreur lors de la récupération des messages: %v", err)
			http.Error(w, "Erreur lors de la récupération des messages", 500)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var message forumUtils.Message
			if err := rows.Scan(&message.ID, &message.ThreadID, &message.AuthorID, &message.Content, &message.CreatedAt); err != nil {
				log.Printf("Erreur lors du scan des messages: %v", err)
				http.Error(w, "Erreur lors de la récupération des messages", 500)
				return
			}
			// Récupérer l'objet User pour l'auteur du message
			author, err := forumUtils.GetUserByID(db, message.AuthorID)
			if err != nil {
				log.Printf("Erreur lors de la récupération de l'auteur du message: %v", err)
				http.Error(w, "Erreur lors de la récupération de l'auteur du message", 500)
				return
			}
			message.Author = *author
			messages = append(messages, message)
		}
		if err := rows.Err(); err != nil {
			log.Printf("Erreur lors de la lecture des messages: %v", err)
			http.Error(w, "Erreur lors de la récupération des messages", 500)
			return
		}

		data := struct {
			Thread   forumUtils.Thread
			User     forumUtils.ConnectedUser
			Messages []forumUtils.Message
		}{
			Thread:   thread,
			User:     user,
			Messages: messages,
		}

		if err := temp.ExecuteTemplate(w, "thread", data); err != nil {
			log.Printf("Erreur lors de l'exécution du template: %v", err)
			http.Error(w, "Erreur lors de l'affichage de la page", 500)
			return
		}
	})

	http.HandleFunc("/thread/postMessage", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.ParseForm()
			threadID := r.FormValue("thread_id")
			content := r.FormValue("message_content")
			if threadID == "" || content == "" {
				log.Println("Données manquantes dans le formulaire")
				http.Error(w, "Tous les champs sont requis", 400)
				return
			}
			cookie, err := r.Cookie("session_token")
			if err != nil {
				log.Printf("Aucun token trouvé: %v", err)
				http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
				return
			}
			db, err := forumUtils.ConnectDatabase()
			if err != nil {
				log.Printf("Erreur de connexion à la base de données: %v", err)
				http.Error(w, "Database error", 500)
				return
			}
			defer db.Close()
			userID, valid := forumUtils.ValidateToken(db, cookie.Value)
			if !valid {
				log.Printf("Token invalide ou expiré")
				http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
				return
			}
			stmt, err := db.Prepare("INSERT INTO messages (thread_id, author_id, content) VALUES (?, ?, ?)")
			if err != nil {
				log.Printf("Erreur de préparation de la requête: %v", err)
				http.Error(w, "Erreur serveur", 500)
				return
			}
			defer stmt.Close()
			_, err = stmt.Exec(threadID, userID, content)
			if err != nil {
				log.Printf("Erreur lors de l'insertion du message: %v", err)
				http.Error(w, "Erreur lors de la création du message", 500)
				return
			}
			log.Printf("Message créé avec succès par l'utilisateur %d dans le thread %s", userID, threadID)
			http.Redirect(w, r, "/thread?id="+threadID, http.StatusSeeOther)
			return
		}
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/thread/like", func(w http.ResponseWriter, r *http.Request) {

	})

	http.HandleFunc("/thread/dislike", func(w http.ResponseWriter, r *http.Request) {
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

			hashedPassword := forumUtils.HashPassword(password)

			db, err := forumUtils.ConnectDatabase()
			if err != nil {
				log.Printf("Erreur de connexion à la base de données: %v", err)
				http.Error(w, "Database error", 500)
				return
			}
			defer db.Close()

			var user forumUtils.ConnectedUser

			err = db.QueryRow("SELECT id, username FROM users WHERE username = ? AND password = ?", username, hashedPassword).
				Scan(&user.ID, &user.Username)
			if err != nil {
				log.Printf("Erreur lors de la récupération de l'utilisateur: %v", err)
				http.Error(w, "Identifiants incorrects", 401)
				return
			}

			token, err := forumUtils.CreateSessionToken(db, user.ID)
			if err != nil {
				log.Printf("Erreur création token: %v", err)
				http.Error(w, "Erreur serveur", 500)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Expires:  time.Now().Add(24 * time.Hour), // Expire dans 24h
				HttpOnly: true,
				Secure:   false,
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
			hashedPassword := forumUtils.HashPassword(password)

			db, err := forumUtils.ConnectDatabase()
			if err != nil {
				log.Printf("Erreur de connexion à la base de données: %v", err)
				http.Error(w, "Database error", 500)
				return
			}
			defer db.Close()

			err = forumUtils.RegisterUser(db, username, email, hashedPassword)
			if err != nil {
				log.Printf("Erreur lors de l'enregistrement: %v", err)
				http.Error(w, "Failed to register: "+err.Error(), 500)
				return
			}
			log.Printf("Utilisateur %s enregistré avec succès", username)

			http.Redirect(w, r, "/loginPage", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil {
			db, err := forumUtils.ConnectDatabase()
			if err == nil {
				forumUtils.DeleteSessionToken(db, cookie.Value)
				db.Close()
			}
		}

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

func RunServer() {
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir(".templates/styles"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir(".templates/images"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir(".templates/fonts"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir(".templates/scripts"))))

	fmt.Println("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
