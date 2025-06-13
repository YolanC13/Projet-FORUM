package controllers

import (
	"PROJET_FORUM/models"
	"net/http"
	"strconv"
)

type AdminController struct {
	*BaseController
}

func NewAdminController() *AdminController {
	return &AdminController{
		BaseController: NewBaseController(),
	}
}

func (ac *AdminController) RequireAdmin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := ac.GetCurrentUser(r)
		if err != nil || !user.IsAdmin {
			http.Error(w, "Accès refusé", http.StatusForbidden)
			return
		}
		handler(w, r)
	}
}

func (ac *AdminController) AdminPanel(w http.ResponseWriter, r *http.Request) {
	users, _ := models.GetAllUsers()
	threads, _ := models.GetAllThreads()
	messages, _ := models.GetAllMessages()
	data := struct {
		Users    []models.User
		Threads  []models.Thread
		Messages []models.Message
	}{
		Users:    users,
		Threads:  threads.Threads,
		Messages: messages,
	}
	ac.RenderTemplate(w, "adminPanel", data)
}

func (ac *AdminController) DeleteThread(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.Atoi(r.FormValue("thread_id"))
	models.AdminDeleteThread(threadID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (ac *AdminController) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	messageID, _ := strconv.Atoi(r.FormValue("message_id"))
	models.AdminDeleteMessage(messageID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (ac *AdminController) BanUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	models.BanUser(userID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (ac *AdminController) UnbanUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	models.UnbanUser(userID)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
