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
	log.Println("Inizio creazione progetto")

	// Recupera la sessione
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok {
		log.Println("Utente non trovato nella sessione")
		http.Error(w, "Utente non autenticato", http.StatusUnauthorized)
		return
	}

	var project models.Project
	err = json.NewDecoder(r.Body).Decode(&project)
	if err != nil {
		log.Printf("Errore nella decodifica del payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("Payload decodificato con successo: %+v", project)

	// Associa l'utente al progetto
	project.AutoreID = user.ID

	result := config.DB.Create(&project)
	if result.Error != nil {
		log.Printf("Errore nella creazione del progetto nel database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Progetto creato con successo: %+v", project)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(project); err != nil {
		log.Printf("Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("Risposta inviata con successo")
}

// ListProjects elenca tutti i progetti dell'utente della sessione
func ListProjects(w http.ResponseWriter, r *http.Request) {
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
	result := config.DB.Where("autore_id = ?", user.ID).Find(&projects)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(projects)
}

// DeleteProject elimina un progetto in base all'ID
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/delete_project/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

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

	result = config.DB.Delete(&project)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Project deleted")
}
