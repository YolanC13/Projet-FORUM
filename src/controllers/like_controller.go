package controllers

import (
	"PROJET_FORUM/models"
	"net/http"
	"strconv"
)

type LikeController struct {
	*BaseController
}

func NewLikeController() *LikeController {
	return &LikeController{
		BaseController: NewBaseController(),
	}
}

func (lc *LikeController) LikeThread(w http.ResponseWriter, r *http.Request) {
	user, err := lc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}
	threadIDStr := r.URL.Query().Get("id")
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "ID thread invalide", 400)
		return
	}
	err = models.ToggleLike(user.ID, threadID)
	if err != nil {
		http.Error(w, "Erreur like: "+err.Error(), 500)
		return
	}
	http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
}
