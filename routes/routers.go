package routes

import (
	"Progetto_APL/controllers"
	"net/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/users", controllers.CreateUser)
	mux.HandleFunc("/users/", controllers.GetUser)
	return mux
}
