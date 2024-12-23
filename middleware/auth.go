package middleware

import (
	"Progetto_APL/models"
	"encoding/gob"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("4f8b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,  // Durata della sessione in secondi
		HttpOnly: true,  // Impedisce l'accesso ai cookie tramite JavaScript
		Secure:   false, // Imposta su true se usi HTTPS
	}
	gob.Register(models.User{}) // Registra il tipo models.User
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("AuthMiddleware: Inizio gestione richiesta")

		session, err := store.Get(r, "session-name")
		if err != nil {
			log.Printf("Errore nel recupero della sessione: %v", err)
			http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
			return
		}

		// Controlla se l'utente è loggato
		user, ok := session.Values["user"].(models.User)
		log.Printf("Valore della sessione recuperato: %v, OK: %v", user, ok)
		if !ok || user.ID == 0 {
			log.Println("AuthMiddleware: Utente non autenticato, reindirizzamento a /login")
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		log.Printf("AuthMiddleware: Utente autenticato: %s", user.Nome)
		// Se l'utente è loggato, passa al prossimo handler
		next.ServeHTTP(w, r)
	})
}
