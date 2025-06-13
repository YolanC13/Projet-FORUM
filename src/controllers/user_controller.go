package controllers

import (
	"PROJET_FORUM/models"
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
		http.Error(w, "Utilisateur non trouvé", 404)
		return
	}
	currentUser, _ := uc.GetCurrentUser(r)
	data := struct {
		User        *models.User
		CurrentUser *models.User
	}{
		User:        user,
		CurrentUser: currentUser,
	}
	uc.RenderTemplate(w, "profile", data)
}
