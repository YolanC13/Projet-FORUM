package controllers

import (
	"PROJET_FORUM/models"
	"log"
	"net/http"
	"strconv"
)

type ThreadController struct {
	*BaseController
}

func NewThreadController() *ThreadController {
	return &ThreadController{
		BaseController: NewBaseController(),
	}
}

func (tc *ThreadController) ShowMainMenu(w http.ResponseWriter, r *http.Request) {
	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	threadList, err := models.GetAllThreads()
	if err != nil {
		log.Printf("Erreur récupération threads: %v", err)
		http.Error(w, "Erreur récupération threads", 500)
		return
	}

	data := struct {
		Threads models.ThreadList
		User    *models.User
		Tags    []string
		Search  string
	}{
		Threads: threadList,
		User:    user,
		Tags:    nil,
		Search:  "",
	}

	tc.RenderTemplate(w, "mainMenu", data)
}

func (tc *ThreadController) ShowMainMenuSortedByRecent(w http.ResponseWriter, r *http.Request) {
	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}
	threadList, err := models.GetAllThreadsSortedByRecent()
	if err != nil {
		http.Error(w, "Erreur récupération threads", 500)
		return
	}
	data := struct {
		Threads models.ThreadList
		User    *models.User
		Tags    []string
		Search  string
	}{
		Threads: threadList,
		User:    user,
		Tags:    nil,
		Search:  "",
	}
	tc.RenderTemplate(w, "mainMenu", data)
}

func (tc *ThreadController) ShowMainMenuSortedByPopularity(w http.ResponseWriter, r *http.Request) {
	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}
	threadList, err := models.GetAllThreadsSortedByPopularity()
	if err != nil {
		http.Error(w, "Erreur récupération threads", 500)
		return
	}
	data := struct {
		Threads models.ThreadList
		User    *models.User
		Tags    []string
		Search  string
	}{
		Threads: threadList,
		User:    user,
		Tags:    nil,
		Search:  "",
	}
	tc.RenderTemplate(w, "mainMenu", data)
}

func (tc *ThreadController) ShowCreateThread(w http.ResponseWriter, r *http.Request) {
	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	tc.RenderTemplate(w, "createThread", user)
}

func (tc *ThreadController) CreateThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	title := r.FormValue("thread_title")
	description := r.FormValue("thread_desc")
	tags := r.FormValue("thread_tag")

	if title == "" || description == "" || tags == "" {
		http.Error(w, "Tous les champs sont requis", 400)
		return
	}

	thread := &models.Thread{
		Title:       title,
		Description: description,
		Tag:         tags,
		AuthorID:    user.ID,
	}

	if err := thread.Create(); err != nil {
		log.Printf("Erreur création thread: %v", err)
		http.Error(w, "Erreur création thread", 500)
		return
	}

	log.Printf("Thread créé par utilisateur %d", user.ID)
	http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
}

func (tc *ThreadController) ShowThread(w http.ResponseWriter, r *http.Request) {
	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	threadIDStr := r.URL.Query().Get("id")
	if threadIDStr == "" {
		http.Error(w, "ID thread manquant", 400)
		return
	}
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "ID thread invalide", 400)
		return
	}

	// Pagination
	const pageSize = 10
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if pInt, err := strconv.Atoi(p); err == nil && pInt > 0 {
			page = pInt
		}
	}

	thread, err := models.GetThreadByID(threadID)
	if err != nil {
		log.Printf("Erreur récupération thread: %v", err)
		http.Error(w, "Thread non trouvé", 404)
		return
	}

	messages, total, err := models.GetMessagesByThreadIDPaginated(threadID, page, pageSize)
	if err != nil {
		log.Printf("Erreur récupération messages: %v", err)
		http.Error(w, "Erreur récupération messages", 500)
		return
	}

	totalPages := (total + pageSize - 1) / pageSize

	data := struct {
		Thread     *models.Thread
		User       *models.User
		Messages   []models.Message
		Page       int
		TotalPages int
	}{
		Thread:     thread,
		User:       user,
		Messages:   messages,
		Page:       page,
		TotalPages: totalPages,
	}

	tc.RenderTemplate(w, "thread", data)
}

func (tc *ThreadController) PostMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	threadIDStr := r.FormValue("thread_id")
	content := r.FormValue("message_content")

	if threadIDStr == "" || content == "" {
		http.Error(w, "Tous les champs sont requis", 400)
		return
	}

	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "ID thread invalide", 400)
		return
	}

	message := &models.Message{
		ThreadID: threadID,
		AuthorID: user.ID,
		Content:  content,
	}

	if err := message.Create(); err != nil {
		log.Printf("Erreur création message: %v", err)
		http.Error(w, "Erreur création message", 500)
		return
	}

	log.Printf("Message créé par utilisateur %d", user.ID)
	http.Redirect(w, r, "/thread?id="+threadIDStr, http.StatusSeeOther)
}

func (tc *ThreadController) DeleteThread(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	threadIDStr := r.FormValue("thread_id")
	if threadIDStr == "" {
		http.Error(w, "ID thread manquant", 400)
		return
	}
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "ID thread invalide", 400)
		return
	}

	thread, err := models.GetThreadByID(threadID)
	if err != nil {
		http.Error(w, "Thread non trouvé", 404)
		return
	}

	// Optionnel : vérifier que l'utilisateur est bien l'auteur
	if thread.AuthorID != user.ID {
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	if err := thread.Delete(); err != nil {
		log.Printf("Erreur suppression thread: %v", err)
		http.Error(w, "Erreur suppression thread", 500)
		return
	}

	log.Printf("Thread %d supprimé par utilisateur %d", threadID, user.ID)
	http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
}

func (tc *ThreadController) ShowMainMenuWithFilters(w http.ResponseWriter, r *http.Request) {
	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	// Récupère les tags cochés dans l'URL (ex: ?tag=Sport&tag=Music)
	tags := r.URL.Query()["tag"]

	threadList, err := models.GetThreadsByTags(tags)
	if err != nil {
		http.Error(w, "Erreur récupération threads", 500)
		return
	}

	data := struct {
		Threads models.ThreadList
		User    *models.User
		Tags    []string
		Search  string
	}{
		Threads: threadList,
		User:    user,
		Tags:    tags,
		Search:  "",
	}
	tc.RenderTemplate(w, "mainMenu", data)
}

func (tc *ThreadController) SearchThreads(w http.ResponseWriter, r *http.Request) {
	user, err := tc.GetCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/connexionPage", http.StatusSeeOther)
		return
	}

	// Récupère la recherche et les tags cochés
	query := r.FormValue("search")
	tags := r.URL.Query()["tag"]
	if len(tags) == 0 {
		// Si la recherche vient d'un POST, récupère les tags du formulaire
		r.ParseForm()
		tags = r.Form["tag"]
	}

	threadList, err := models.SearchThreads(query, tags)
	if err != nil {
		http.Error(w, "Erreur recherche threads", 500)
		return
	}

	data := struct {
		Threads models.ThreadList
		User    *models.User
		Tags    []string
		Search  string
	}{
		Threads: threadList,
		User:    user,
		Tags:    tags,
		Search:  query,
	}
	tc.RenderTemplate(w, "mainMenu", data)
}
