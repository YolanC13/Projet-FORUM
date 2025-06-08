package controllers

import (
	"PROJET_FORUM/models"
	"log"
	"net/http"
	"strconv"
)

type UserController struct {
	*BaseController
}

func NewUserController() *UserController {
	return &UserController{
		BaseController: NewBaseController(),
	}
}

// Affiche le profil de l'utilisateur dont l'ID est passé en query (?id=...)
func (uc *UserController) ShowProfile(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID utilisateur manquant", 400)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID utilisateur invalide", 400)
		return
	}

	user, err := models.GetUserByID(id)
	if err != nil {
		log.Printf("Erreur récupération profil: %v", err)
		http.Error(w, "Utilisateur non trouvé", 404)
		return
	}

	uc.RenderTemplate(w, "profile", user)
}
