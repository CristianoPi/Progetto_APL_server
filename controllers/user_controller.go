package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID    uint
	Nome  string
	Email string
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	var user models.User
	result := config.DB.Where("email = ?", creds.Email).First(&user)
	if result.Error != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}
	// Confronta la password fornita con quella hashata memorizzata nel database
	err = bcrypt.CompareHashAndPassword([]byte(user.Pwd), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	// Crea una nuova sessione e salva i dettagli dell'utente
	session, _ := store.Get(r, "session-name")
	session.Values["user"] = user
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Errore nel salvataggio della sessione: %v", err)
		http.Error(w, fmt.Sprintf("Errore nel salvataggio della sessione: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("Utente autenticato con successo")
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	// Elimina i dati della sessione
	session.Values["user"] = nil
	session.Options.MaxAge = -1 // Imposta MaxAge a -1 per eliminare il cookie
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Errore nel salvataggio della sessione", http.StatusInternalServerError)
		return
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Inizio gestione profilo")

	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}
	log.Println("Sessione recuperata con successo")

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok {
		log.Println("Errore nella conversione del tipo utente")
		http.Error(w, "Non sei autenticato", http.StatusUnauthorized)
		return
	}
	if user.ID == 0 {
		log.Println("ID utente non valido")
		http.Error(w, "Non sei autenticato", http.StatusUnauthorized)
		return
	}
	log.Printf("Utente autenticato: %s", user.Nome)

	_, err = w.Write([]byte("Benvenuto, " + user.Nome))
	if err != nil {
		log.Printf("Errore nella scrittura della risposta: %v", err)
		http.Error(w, "Errore nella scrittura della risposta", http.StatusInternalServerError)
		return
	}
	log.Println("Risposta inviata con successo")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Inizio creazione utente")

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("Errore nella decodifica del corpo della richiesta: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Dati utente decodificati: %+v", user)

	result := config.DB.Create(&user)
	if result.Error != nil {
		log.Printf("Errore nella creazione dell'utente nel database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Utente creato con successo")

	w.WriteHeader(http.StatusCreated)
	log.Println("Risposta inviata con successo")
}
func GetUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Path[len("/users/"):]
	var user models.User
	print("questa è la mail: ", email)
	result := config.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(user)
}
