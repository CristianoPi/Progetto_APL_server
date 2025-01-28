package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

// CreateProject crea un nuovo progetto
func CreateProject(w http.ResponseWriter, r *http.Request) {
	log.Println("PROJECT CONTROLLER: Inizio funzione CreateProject")

	// Recupera la sessione
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok {
		log.Println("PROJECT CONTROLLER: Utente non trovato nella sessione")
		http.Error(w, "Utente non autenticato", http.StatusUnauthorized)
		return
	}

	var project models.Project
	err = json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nella decodifica del payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("PROJECT CONTROLLER: Payload decodificato con successo: %+v", project)

	// Associa l'utente al progetto
	project.AutoreID = user.ID

	// Salva il progetto nel database
	err = config.DB.Create(&project).Error
	if err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nella creazione del progetto nel database: %v", err)
		http.Error(w, "Errore nella creazione del progetto nel database", http.StatusInternalServerError)
		return
	}
	log.Printf("PROJECT CONTROLLER: Progetto creato con successo: %+v", project)

	// Imposta l'header Content-Type a application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("PROJECT CONTROLLER: Risposta inviata con successo")
}

// ListProjects elenca tutti i progetti
func ListProjects(w http.ResponseWriter, r *http.Request) {
	log.Println("PROJECT CONTROLLER: Inizio funzione ListProjects")

	// Recupera la sessione
	session, err := store.Get(r, "session-name")
	if err != nil {
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		http.Error(w, "Non sei autenticato", http.StatusUnauthorized)
		return
	}

	var projects []models.Project
	result := config.DB.Find(&projects)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("PROJECT CONTROLLER: i progetti sono: %+v", projects)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(projects); err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("PROJECT CONTROLLER: Progetti inviati con successo")
}

// DeleteProject elimina un progetto in base all'ID
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	log.Println("PROJECT CONTROLLER: Inizio funzione DeleteProject")

	// Recupera la sessione
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok {
		log.Println("PROJECT CONTROLLER: Utente non trovato nella sessione")
		http.Error(w, "Utente non autenticato", http.StatusUnauthorized)
		return
	}

	// Ottieni l'ID del progetto dai parametri della richiesta
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID del progetto mancante", http.StatusBadRequest)
		return
	}
	log.Printf("PROJECT CONTROLLER: ID del progetto ricevuto: %v", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID del progetto non valido", http.StatusBadRequest)
		return
	}
	log.Printf("PROJECT CONTROLLER: ID del progetto convertito con successo: %v", id)

	// Recupera il progetto dal database utilizzando l'ID
	var project models.Project
	result := config.DB.First(&project, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Project not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Verifica se l'utente è l'autore del progetto
	if project.AutoreID != user.ID {
		log.Println("PROJECT CONTROLLER: Utente non autorizzato a eliminare questo progetto")
		http.Error(w, "Utente non autorizzato a eliminare questo progetto", http.StatusForbidden)
		return
	}

	// Elimina il progetto dal database
	result = config.DB.Delete(&project)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Project deleted")
	log.Println("PROJECT CONTROLLER: Progetto eliminato con successo")
}
