package controllers

import (
	"PROJET_FORUM/models"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type BaseController struct {
	Templates *template.Template
}

func NewBaseController() *BaseController {
	temp, err := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
	}).ParseGlob("views/templates/*.html")
	if err != nil {
		log.Fatal("Erreur chargement templates:", err)
	}

	return &BaseController{
		Templates: temp,
	}
}

func (bc *BaseController) RenderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	if err := bc.Templates.ExecuteTemplate(w, templateName, data); err != nil {
		log.Printf("Erreur rendu template: %v", err)
		http.Error(w, "Erreur interne", 500)
	}
}

func (bc *BaseController) GetCurrentUser(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	userID, valid := models.ValidateSession(cookie.Value)
	if !valid {
		return nil, fmt.Errorf("session invalide")
	}

	return models.GetUserByID(userID)
}

func (bc *BaseController) RequireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := bc.GetCurrentUser(r)
		if err != nil {
			http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
			return
		}
		handler(w, r)
	}
}
