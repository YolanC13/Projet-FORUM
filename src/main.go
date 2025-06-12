package main

import (
	"PROJET_FORUM/models"
	"PROJET_FORUM/routes"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db, err := models.InitDatabase()
	if err != nil {
		log.Fatal("Erreur initialisation base:", err)
	}
	defer db.Close()

	routes.SetupRoutes()
	fmt.Println("Server started on localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
