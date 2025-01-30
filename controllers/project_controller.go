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

	// Elimina i task associati al progetto
	result = config.DB.Where("progetto_id = ?", id).Delete(&models.Task{})
	if result.Error != nil {
		log.Printf("PROJECT CONTROLLER: Errore nell'eliminazione dei task associati al progetto: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("PROJECT CONTROLLER: Task associati al progetto eliminati con successo")

	// Elimina il progetto dal database
	result = config.DB.Delete(&project)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Project and associated tasks deleted")
	log.Println("PROJECT CONTROLLER: Progetto e task associati eliminati con successo")
}

func CountTasksInProject(w http.ResponseWriter, r *http.Request) {
	log.Println("PROJECT CONTROLLER: Inizio funzione CountTasksInProject")

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

	// Conta il numero totale di task nel progetto
	var totalTasks int64
	result := config.DB.Model(&models.Task{}).Where("progetto_id = ?", id).Count(&totalTasks)
	if result.Error != nil {
		log.Printf("PROJECT CONTROLLER: Errore nel conteggio dei task: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Conta il numero di task completati nel progetto
	var completedTasks int64
	result = config.DB.Model(&models.Task{}).Where("progetto_id = ? AND completato = ?", id, true).Count(&completedTasks)
	if result.Error != nil {
		log.Printf("PROJECT CONTROLLER: Errore nel conteggio dei task completati: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("PROJECT CONTROLLER: n task: %v", totalTasks)
	log.Printf("PROJECT CONTROLLER: n task completati: %v", completedTasks)

	// Crea la risposta JSON
	response := map[string]int64{
		"total_tasks":     totalTasks,
		"completed_tasks": completedTasks,
	}

	log.Printf("PROJECT CONTROLLER: response: %v", response)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("PROJECT CONTROLLER: Conteggio task inviato con successo")
}

func GetProjectByID(w http.ResponseWriter, r *http.Request) {
	log.Println("PROJECT CONTROLLER: Inizio funzione GetProjectByID")

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
	if err := config.DB.First(&project, id).Error; err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nel recupero del progetto: %v", err)
		http.Error(w, "Errore nel recupero del progetto", http.StatusInternalServerError)
		return
	}

	log.Printf("PROJECT CONTROLLER: GetProjectByID torna %v", project)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("PROJECT CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("PROJECT CONTROLLER: Progetto recuperato con successo")
}
