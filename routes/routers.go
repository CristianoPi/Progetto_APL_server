package routes

import (
	"Progetto_APL/controllers"
	"Progetto_APL/middleware"
	"net/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Rotte senza accesso
	mux.HandleFunc("/users", controllers.CreateUser) // bisogna dare in input nelle post il json --> guardare client.go per esempio
	mux.HandleFunc("/users/", controllers.GetUser)   // get user data la mail tipo: http://localhost:8080/users/mario.rossi1@example.com

	// Rotte per gestire l'accesso
	mux.HandleFunc("/login", controllers.LoginHandler)
	mux.HandleFunc("/logout", controllers.LogoutHandler)

	// Rotte protette con middleware
	mux.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(controllers.ProfileHandler)))

	// Rotte per gestione dei progetti
	mux.Handle("/project", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateProject)))
	mux.Handle("/projects/", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListProjects)))
	mux.Handle("/delete_project/{id}", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteProject)))

	//rotte per gestione dei task
	mux.Handle("/task", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateTask)))
	mux.Handle("/tasks/", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListTasks)))
	mux.Handle("/delete_task/{id}", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteTask)))
	//_____________________________________________________________________________
	// Rotte per gestione dei file
	mux.Handle("/files", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateFile)))
	mux.Handle("/files/list", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListFiles)))
	mux.Handle("/files/get", middleware.AuthMiddleware(http.HandlerFunc(controllers.GetFile)))
	mux.Handle("/files/delete", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteFile)))
	return mux
}
