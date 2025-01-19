package main

import (
	"Progetto_APL/config"
	"Progetto_APL/routes"
	"net/http"

	"log"
)

func main() {
	config.LoadConfig() // Connetti al database utilizzando GORM
	router := routes.SetupRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
