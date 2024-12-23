// SONO ARRIVATO AL LOGINHANDLER, SI DEVE SOLO PROVARE
package routes

import (
	"Progetto_APL/controllers"
	"Progetto_APL/middleware"
	"net/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	//rotte senza accesso
	mux.HandleFunc("/users", controllers.CreateUser) //bisogna dare in input nelle post il json --> guardare client.go per esempio
	mux.HandleFunc("/users/", controllers.GetUser)   //get user data la mail tipo: http://localhost:8080/users/mario.rossi1@example.com

	//rotte per gestire l'accesso
	mux.HandleFunc("/login", controllers.LoginHandler)
	mux.HandleFunc("/logout", controllers.LogoutHandler)

	// rotte protette, Applica il middleware alla rotta /profile
	mux.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(controllers.ProfileHandler)))
	return mux
}
