package routes

import (
	"Progetto_APL/controllers"
	"Progetto_APL/middleware"
	"net/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	//rotte senza accesso
	mux.HandleFunc("/users", controllers.CreateUser) //bisogna dare in input nelle post il json --> guardare client.go per esempioq
	mux.HandleFunc("/users/", controllers.GetUser)   //get user data la mail tipo: http://localhost:8080/users/mario.rossi1@example.com

	//rotte per gestire l'accesso
	mux.HandleFunc("/login", controllers.LoginHandler)
	mux.HandleFunc("/logout", controllers.LogoutHandler)

	//_______________________ROTTE PROTETTE CON MINDDLEWARE______________________
	mux.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(controllers.ProfileHandler)))

	//rotte per gestione dei progetti
	mux.Handle("/project", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateProject)))
	mux.Handle("/projects/", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListProjects)))
	mux.Handle("/delete_project/{id}", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteProject)))

	//rotte per gestione dei task
	mux.Handle("/task", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateTask)))
	mux.Handle("/tasks/", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListTasks)))
	mux.Handle("/delete_task/{id}", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteTask)))
	//_____________________________________________________________________________

	return mux

}
