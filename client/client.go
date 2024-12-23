package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar" //il client deve essere in grado di gestire i cookie per la sessione
	"os"

	"golang.org/x/crypto/bcrypt"
)

type Utente struct {
	ID    uint   `json:"id"` // Campo ID per identificare l'utente
	Nome  string `json:"nome"`
	Pwd   string `json:"pwd"`
	Email string `json:"email"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	baseURL := "http://localhost:8080"

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Errore nella creazione del cookie jar: %v", err)
	}

	client := &http.Client{
		Jar: jar,
	}

	for {
		fmt.Println("Menu:")
		fmt.Println("1. Crea un nuovo utente")
		fmt.Println("2. Recupera un utente tramite email")
		fmt.Println("3. Login")
		fmt.Println("4. Esci")
		fmt.Print("Scegli un'opzione: ")

		var choice int
		fmt.Scan(&choice)

		switch choice {
		case 1:
			createUser(client, baseURL)
		case 2:
			getUser(client, baseURL)
		case 3:
			login(client, baseURL)
		case 4:
			fmt.Println("Uscita...")
			os.Exit(0)
		default:
			fmt.Println("Opzione non valida. Riprova.")
		}
	}
}

func createUser(client *http.Client, baseURL string) {
	var newUser Utente

	fmt.Print("Inserisci il nome dell'utente: ")
	fmt.Scan(&newUser.Nome)

	fmt.Print("Inserisci la password dell'utente: ")
	var password string
	fmt.Scan(&password)

	// Cripta la password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Errore nella criptazione della password: %v", err)
	}
	newUser.Pwd = string(hashedPwd)

	fmt.Print("Inserisci l'email dell'utente: ")
	fmt.Scan(&newUser.Email)

	userJSON, err := json.Marshal(newUser)
	if err != nil {
		log.Fatalf("Errore nella serializzazione dell'utente: %v", err)
	}

	resp, err := client.Post(baseURL+"/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		log.Fatalf("Errore nella richiesta POST: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Creazione utente fallita. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Utente creato con successo.")
}

func getUser(client *http.Client, baseURL string) {
	fmt.Print("Inserisci l'email dell'utente: ")
	var email string
	fmt.Scan(&email)

	url := fmt.Sprintf("%s/users/%s", baseURL, email)
	getResp, err := client.Get(url)
	if err != nil {
		log.Fatalf("Errore nella richiesta GET: %v", err)
	}
	defer getResp.Body.Close()

	getBody, err := io.ReadAll(getResp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if getResp.StatusCode != http.StatusOK {
		log.Fatalf("Recupero utente fallito. Stato HTTP: %d, Messaggio: %s", getResp.StatusCode, string(getBody))
	}

	fmt.Printf("Risposta al recupero dell'utente: %s\n", getBody)

	var fetchedUser Utente
	if err := json.Unmarshal(getBody, &fetchedUser); err != nil {
		log.Fatalf("Errore nel parsing della risposta JSON: %v", err)
	}

	fmt.Printf("Utente recuperato: %+v\n", fetchedUser)
}

func login(client *http.Client, baseURL string) {
	var creds Credentials

	fmt.Print("Inserisci l'email: ")
	fmt.Scan(&creds.Email)

	fmt.Print("Inserisci la password: ")
	fmt.Scan(&creds.Password)

	credsJSON, err := json.Marshal(creds)
	if err != nil {
		log.Fatalf("Errore nella serializzazione delle credenziali: %v", err)
	}

	resp, err := client.Post(baseURL+"/login", "application/json", bytes.NewBuffer(credsJSON))
	if err != nil {
		log.Fatalf("Errore nella richiesta POST: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Login fallito. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Login effettuato con successo.")
	fmt.Printf("Messaggio: %s\n", string(body))
}
