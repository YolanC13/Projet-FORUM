package controllers

import (
	"PROJET_FORUM/models"
	"log"
	"net/http"
	"time"
)

type AuthController struct {
	*BaseController
}

func NewAuthController() *AuthController {
	return &AuthController{
		BaseController: NewBaseController(),
	}
}

func (ac *AuthController) ShowConnectionPage(w http.ResponseWriter, r *http.Request) {
	ac.RenderTemplate(w, "connexionPage", nil)
}

func (ac *AuthController) ShowLoginPage(w http.ResponseWriter, r *http.Request) {
	ac.RenderTemplate(w, "loginPage", nil)
}

func (ac *AuthController) ShowRegisterPage(w http.ResponseWriter, r *http.Request) {
	ac.RenderTemplate(w, "registerPage", nil)
}

func (ac *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Tous les champs sont requis", 400)
		return
	}

	user, err := models.GetUserByUsername(username)
	if err != nil || !user.ValidatePassword(password) {
		http.Error(w, "Identifiants incorrects", 401)
		return
	}

	session := &models.Session{UserID: user.ID}
	if err := session.Create(); err != nil {
		log.Printf("Erreur création session: %v", err)
		http.Error(w, "Erreur serveur", 500)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})

	log.Printf("Utilisateur %s connecté", user.Username)
	http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
}

func (ac *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		http.Error(w, "Tous les champs sont requis", 400)
		return
	}

	user := &models.User{
		Username: username,
		Email:    email,
		Password: password,
	}

	if err := user.Create(); err != nil {
		log.Printf("Erreur enregistrement: %v", err)
		http.Error(w, "Erreur lors de l'enregistrement: "+err.Error(), 500)
		return
	}

	log.Printf("Utilisateur %s enregistré", username)
	http.Redirect(w, r, "/loginPage", http.StatusSeeOther)
}

func (ac *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		models.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
}
