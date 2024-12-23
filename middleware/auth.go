package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("super-secret-key"))

type User struct {
	ID    uint
	Nome  string
	Email string
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")

		// Controlla se l'utente è loggato
		user, ok := session.Values["user"].(User)
		if !ok || user.ID == 0 {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Se l'utente è loggato, passa al prossimo handler
		next.ServeHTTP(w, r)
	})
}
