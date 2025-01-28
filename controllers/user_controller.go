package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID       uint
	Username string
	Email    string
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

// Funzione per generare una password randomica
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Configura le impostazioni email
var emailConf = map[string]string{
	"smtp_server":   "smtp.gmail.com",
	"smtp_port":     "587",
	"smtp_user":     "cristiano.pistorio@gmail.com",
	"smtp_password": "iiujymduizmhgmvf",
	"from_email":    "cristiano.pistorio@gmail.com",
}

// Funzione per inviare un'email
func sendEmail(to, subject, body string) error {
	from := emailConf["from_email"]
	password := emailConf["smtp_password"]

	// Configura il server SMTP
	smtpHost := emailConf["smtp_server"]
	smtpPort := emailConf["smtp_port"]

	// Configura il messaggio
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	// Autenticazione SMTP
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Connessione al server SMTP
	log.Println("Connessione al server SMTP...")
	server, err := smtp.Dial(smtpHost + ":" + smtpPort)
	if err != nil {
		log.Printf("Errore nella connessione al server SMTP: %v", err)
		return err
	}
	defer server.Close()

	// Avvia TLS
	if err = server.StartTLS(&tls.Config{ServerName: smtpHost}); err != nil {
		log.Printf("Errore nell'avvio di TLS: %v", err)
		return err
	}

	// Autenticazione
	if err = server.Auth(auth); err != nil {
		log.Printf("Errore nell'autenticazione SMTP: %v", err)
		return err
	}

	// Imposta il mittente e il destinatario
	if err = server.Mail(from); err != nil {
		log.Printf("Errore nell'impostazione del mittente: %v", err)
		return err
	}
	if err = server.Rcpt(to); err != nil {
		log.Printf("Errore nell'impostazione del destinatario: %v", err)
		return err
	}

	// Scrivi il messaggio
	wc, err := server.Data()
	if err != nil {
		log.Printf("Errore nell'invio dei dati: %v", err)
		return err
	}
	defer wc.Close()
	if _, err = wc.Write([]byte(msg)); err != nil {
		log.Printf("Errore nella scrittura del messaggio: %v", err)
		return err
	}

	log.Printf("Email inviata a %s: %s", to, body)
	return nil
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
		log.Printf("USER CONTROLLER: Errore nel salvataggio della sessione: %v", err)
		http.Error(w, fmt.Sprintf("Errore nel salvataggio della sessione: %v", err), http.StatusInternalServerError)
		return
	}
	log.Println("USER CONTROLLER: Utente autenticato con successo")
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
	log.Println("USER CONTROLLER: Utente disconnesso con successo")
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("USER CONTROLLER: Inizio gestione profilo")

	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}
	log.Println("USER CONTROLLER: Sessione recuperata con successo")

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok {
		http.Error(w, "Utente non trovato nella sessione", http.StatusUnauthorized)
		return
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nella serializzazione dell'utente: %v", err)
		http.Error(w, "Errore nella serializzazione dell'utente", http.StatusInternalServerError)
		return
	}

	// Imposta l'header Content-Type a application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJSON)
	log.Println("USER CONTROLLER: Profilo utente inviato con successo")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("USER CONTROLLER: Inizio creazione utente")

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nella decodifica del corpo della richiesta: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("USER CONTROLLER: Dati utente decodificati: %+v", user)

	// Hash della password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nell'hashing della password: %v", err)
		http.Error(w, "Errore nell'hashing della password", http.StatusInternalServerError)
		return
	}
	user.Pwd = string(hashedPassword)

	result := config.DB.Create(&user)
	if result.Error != nil {
		log.Printf("USER CONTROLLER: Errore nella creazione dell'utente nel database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("USER CONTROLLER: Utente creato con successo")

	w.WriteHeader(http.StatusCreated)
	log.Println("USER CONTROLLER: Risposta inviata con successo")
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
	log.Println("USER CONTROLLER: Utente recuperato con successo")
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {

	// Ottieni l'ID utente dai parametri della richiesta
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID utente mancante", http.StatusBadRequest)
		return
	}
	log.Printf("USER CONTROLLER: Inizio funzione GetUserByID cerco user con id: %v", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID utente non valido", http.StatusBadRequest)
		return
	}
	log.Printf("USER CONTROLLER: ID utente convertito con successo: %v", id)

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		log.Printf("USER CONTROLLER: Errore nel recupero dell'utente: %v", err)
		http.Error(w, "Errore nel recupero dell'utente", http.StatusInternalServerError)
		return
	}

	log.Printf(`USER CONTROLLER: GetUserByID torna %v`, user)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("USER CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("USER CONTROLLER: Utente recuperato con successo")
}

func Changepwd(w http.ResponseWriter, r *http.Request) {
	log.Println("USER CONTROLLER: Inizio funzione ChangePassword")

	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nella decodifica del payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("USER CONTROLLER: Payload decodificato con successo: %+v", user)

	// Verifica se l'utente esiste nel database
	var existingUser models.User
	if err := config.DB.First(&existingUser, user.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("USER CONTROLLER: Utente non trovato: %v", user.ID)
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("USER CONTROLLER: Errore nel recupero dell'utente: %v", err)
			http.Error(w, "Errore nel recupero dell'utente", http.StatusInternalServerError)
		}
		return
	}

	// Hash della nuova password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nell'hashing della password: %v", err)
		http.Error(w, "Errore nell'hashing della password", http.StatusInternalServerError)
		return
	}
	existingUser.Pwd = string(hashedPassword)

	// Aggiorna la password dell'utente nel database
	if err := config.DB.Save(&existingUser).Error; err != nil {
		log.Printf("USER CONTROLLER: Errore nell'aggiornamento della password: %v", err)
		http.Error(w, "Errore nell'aggiornamento della password", http.StatusInternalServerError)
		return
	}
	log.Println("USER CONTROLLER: Password aggiornata con successo")

	w.WriteHeader(http.StatusOK)
	log.Println("USER CONTROLLER: Risposta inviata con successo")
}

func CheckUsername(w http.ResponseWriter, r *http.Request) {
	log.Println("USER CONTROLLER: Inizio funzione CheckUsername")

	// Ottieni l'username dai parametri della richiesta
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username mancante", http.StatusBadRequest)
		return
	}
	log.Printf("USER CONTROLLER: Cerco username: %v", username)

	var user models.User
	result := config.DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Username non trovato, rispondi con true
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(true)
			log.Println("USER CONTROLLER: Username disponibile")
			return
		}
		log.Printf("USER CONTROLLER: Errore nel recupero dell'utente: %v", result.Error)
		http.Error(w, "Errore nel recupero dell'utente", http.StatusInternalServerError)
		return
	}

	// Username trovato, rispondi con false
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(false)
	log.Println("USER CONTROLLER: Username non disponibile")
}

func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	log.Println("USER CONTROLLER: Inizio funzione ForgotPassword")

	// Ottieni l'email dai parametri della richiesta
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email mancante", http.StatusBadRequest)
		return
	}
	log.Printf("USER CONTROLLER: Email ricevuta: %v", email)

	// Verifica se l'utente esiste nel database
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("USER CONTROLLER: Utente non trovato con email: %v", email)
			// Non fare nulla se l'email non è nel database
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(false)
			return
		}
		log.Printf("USER CONTROLLER: Errore nel recupero dell'utente: %v", err)
		http.Error(w, "Errore nel recupero dell'utente", http.StatusInternalServerError)
		return
	}

	// Genera una nuova password randomica
	newPassword := generateRandomPassword(12)
	log.Printf("USER CONTROLLER: Nuova password generata: %v", newPassword)

	// Hash della nuova password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nell'hashing della password: %v", err)
		http.Error(w, "Errore nell'hashing della password", http.StatusInternalServerError)
		return
	}
	user.Pwd = string(hashedPassword)

	// Aggiorna la password dell'utente nel database
	if err := config.DB.Save(&user).Error; err != nil {
		log.Printf("USER CONTROLLER: Errore nell'aggiornamento della password: %v", err)
		http.Error(w, "Errore nell'aggiornamento della password", http.StatusInternalServerError)
		return
	}
	log.Println("USER CONTROLLER: Password aggiornata con successo")

	// Invia l'email con la nuova password
	err = sendEmail(user.Email, "Password Reset", "La tua nuova password è: "+newPassword)
	if err != nil {
		log.Printf("USER CONTROLLER: Errore nell'invio dell'email: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(false)
		return
	}
	log.Println("USER CONTROLLER: Email inviata con successo")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(true)
	log.Println("USER CONTROLLER: Risposta inviata con successo")
}
