package routes

import (
	"Progetto_APL/controllers"
	"Progetto_APL/middleware"
	"net/http"
)

func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Rotte senza accesso
	mux.HandleFunc("/register", controllers.CreateUser) // bisogna dare in input nelle post il json --> guardare client.go per esempio
	mux.HandleFunc("/users/", controllers.GetUser)      // get user data la mail tipo: http://localhost:8080/users/mario.rossi1@example.com
	mux.HandleFunc("/author/", controllers.GetUserByID)
	mux.HandleFunc("/check-username/", controllers.CheckUsername)
	mux.HandleFunc("/forgot-password/", controllers.ForgotPassword)

	// Rotte per gestire l'accesso
	mux.HandleFunc("/login", controllers.LoginHandler)
	mux.HandleFunc("/logout", controllers.LogoutHandler)

	// Rotte protette con middleware
	mux.Handle("/all_users", middleware.AuthMiddleware(http.HandlerFunc(controllers.GetAllUsers)))
	mux.Handle("/profile", middleware.AuthMiddleware(http.HandlerFunc(controllers.ProfileHandler)))
	mux.Handle("/changepwd", middleware.AuthMiddleware(http.HandlerFunc(controllers.Changepwd)))

	// Rotte per gestione dei progetti
	mux.Handle("/project", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateProject)))
	mux.Handle("/projects/", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListProjects)))
	mux.Handle("/delete_project/", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteProject)))
	mux.Handle("/count_tasks_in_project/", middleware.AuthMiddleware(http.HandlerFunc(controllers.CountTasksInProject)))
	//al solito id in input, output così:
	//response := map[string]int64{
	// 	"total_tasks":     totalTasks,
	// 	"completed_tasks": completedTasks,
	// }
	mux.Handle("/get_project_by_id/", middleware.AuthMiddleware(http.HandlerFunc(controllers.GetProjectByID)))

	//rotte per gestione dei task
	mux.Handle("/task", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateTask)))
	mux.Handle("/tasks/", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListTasks)))
	mux.Handle("/delete_task/", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteTask))) //get - input id task da eliminare(fa solo se autore...)
	mux.Handle("/update_task/", middleware.AuthMiddleware(http.HandlerFunc(controllers.UpdateTask)))
	mux.Handle("/tasks_by_project/", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListTasksByProject)))
	mux.Handle("/get_task_by_id/", middleware.AuthMiddleware(http.HandlerFunc(controllers.GetTaskByID)))
	//_____________________________________________________________________________

	// Rotte per gestione dei file
	mux.Handle("/files", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateFile)))
	mux.Handle("/files/list", middleware.AuthMiddleware(http.HandlerFunc(controllers.ListFiles)))
	mux.Handle("/files/get", middleware.AuthMiddleware(http.HandlerFunc(controllers.GetFileByID)))
	mux.Handle("/files/delete", middleware.AuthMiddleware(http.HandlerFunc(controllers.DeleteFile)))

	// Rotte per gestione dei codici
	mux.Handle("/code", middleware.AuthMiddleware(http.HandlerFunc(controllers.CreateCode)))
	mux.Handle("/run_code/", middleware.AuthMiddleware(http.HandlerFunc(controllers.RunCode)))
	mux.Handle("/get_status_code/", middleware.AuthMiddleware(http.HandlerFunc(controllers.GetExecutionStatus)))

	return mux
}
