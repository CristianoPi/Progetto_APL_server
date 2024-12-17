package main

import (
	"Progetto_APL/config"
	"Progetto_APL/routes"
	"log"
	"net/http"
)

func main() {
	config.LoadConfig()
	router := routes.SetupRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
