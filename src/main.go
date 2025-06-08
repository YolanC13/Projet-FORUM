package main

import (
	"PROJET_FORUM/models"
	"PROJET_FORUM/routes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Initialiser la base de données
	db, err := models.InitDatabase()
	if err != nil {
		log.Fatal("Erreur initialisation base:", err)
	}
	defer db.Close()

	// Configurer les routes
	routes.SetupRoutes()

	// Démarrer le serveur
	fmt.Println("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
