package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var store = sessions.NewCookieStore([]byte("super-secret-key"))

type User struct {
	ID    uint
	Nome  string
	Email string
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
	session.Save(r, w)

	http.Redirect(w, r, "/profile", http.StatusFound)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Elimina i dati della sessione
	session.Values["user"] = nil
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(User)
	if !ok || user.ID == 0 {
		http.Error(w, "Non sei autenticato", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Benvenuto, " + user.Nome))
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := config.DB.Create(&user)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
