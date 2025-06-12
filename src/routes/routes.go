package routes

import (
	"PROJET_FORUM/controllers"
	"net/http"
)

func SetupRoutes() {
	// Contrôleurs
	authController := controllers.NewAuthController()
	threadController := controllers.NewThreadController()
	userController := controllers.NewUserController()
	likeController := controllers.NewLikeController()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/mainMenu", http.StatusSeeOther)
	})

	http.HandleFunc("/connexionPage", authController.ShowConnectionPage)
	http.HandleFunc("/loginPage", authController.ShowLoginPage)
	http.HandleFunc("/registerPage", authController.ShowRegisterPage)
	http.HandleFunc("/userlogin", authController.Login)
	http.HandleFunc("/userRegister", authController.Register)
	http.HandleFunc("/logout", authController.Logout)

	http.HandleFunc("/mainMenu", authController.RequireAuth(threadController.ShowMainMenu))
	http.HandleFunc("/createThread", authController.RequireAuth(threadController.ShowCreateThread))
	http.HandleFunc("/createThread/process", authController.RequireAuth(threadController.CreateThread))
	http.HandleFunc("/thread", authController.RequireAuth(threadController.ShowThread))
	http.HandleFunc("/thread/postMessage", authController.RequireAuth(threadController.PostMessage))
	http.HandleFunc("/thread/like", authController.RequireAuth(likeController.LikeThread))
	http.HandleFunc("/thread/delete", authController.RequireAuth(threadController.DeleteThread))
	http.HandleFunc("/mainMenu/sort/recent", authController.RequireAuth(threadController.ShowMainMenuSortedByRecent))
	http.HandleFunc("/mainMenu/sort/popularity", authController.RequireAuth(threadController.ShowMainMenuSortedByPopularity))

	http.HandleFunc("/profile", authController.RequireAuth(userController.ShowProfile))

	// Fichiers statiques
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("views/templates/styles"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("views/templates/images"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("views/templates/fonts"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("views/templates/scripts"))))
}
