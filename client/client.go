package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Utente struct {
	ID    uint   `json:"id"` // Campo ID per identificare l'utente
	Nome  string `json:"nome"`
	Pwd   string `json:"pwd"`
	Email string `json:"email"`
}

func main() {
	// URL del server
	baseURL := "http://localhost:8080"

	// Creazione di un nuovo utente
	newUser := Utente{
		Nome:  "Mario Rossi",
		Pwd:   "$2a$12$XkZc8mDcyDneCjmuSI5ZMuNtrNfLuqcon90kMO63c1b2ifNCD47zC", //hash di "password"
		Email: "mario.rossi1@example.com",
	}

	// Converti l'utente in JSON
	userJSON, err := json.Marshal(newUser)
	if err != nil {
		log.Fatalf("Errore nella serializzazione dell'utente: %v", err)
	}

	// Invia la richiesta POST per creare un nuovo utente
	resp, err := http.Post(baseURL+"/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		log.Fatalf("Errore nella richiesta POST: %v", err)
	}
	defer resp.Body.Close()

	// Controlla lo stato della risposta
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Creazione utente fallita. Stato HTTP: %d", resp.StatusCode)
	}
	fmt.Println("Utente creato con successo.")

	// Recupero dell'utente tramite l'email
	email := newUser.Email
	stringa := fmt.Sprintf("%s/users/%s", baseURL, email)
	print(stringa)
	getResp, err := http.Get(stringa)
	if err != nil {
		log.Fatalf("Errore nella richiesta GET: %v", err)
	}
	defer getResp.Body.Close()

	// Leggi la risposta della richiesta GET
	getBody, err := io.ReadAll(getResp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	// Controlla lo stato della risposta GET
	if getResp.StatusCode != http.StatusOK {
		log.Fatalf("Recupero utente fallito. Stato HTTP: %d, Messaggio: %s", getResp.StatusCode, string(getBody))
	}

	fmt.Printf("Risposta al recupero dell'utente: %s\n", getBody)

	// Decodifica la risposta JSON per verificare i dettagli dell'utente recuperato
	var fetchedUser Utente
	if err := json.Unmarshal(getBody, &fetchedUser); err != nil {
		log.Fatalf("Errore nel parsing della risposta JSON: %v", err)
	}

	fmt.Printf("Utente recuperato: %+v\n", fetchedUser)
}
