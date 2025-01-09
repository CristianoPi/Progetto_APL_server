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

func CreateTask(w http.ResponseWriter, r *http.Request) {
	log.Println("Inizio creazione task")

	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	user, ok := session.Values["user"].(models.User)
	if !ok {
		log.Println("Utente non trovato nella sessione")
		http.Error(w, "Utente non autenticato", http.StatusUnauthorized)
		return
	}

	var task models.Task
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Printf("Errore nella decodifica del payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("Payload decodificato con successo: %+v", task)

	// Associa l'utente al progetto
	task.AutoreID = user.ID

	result := config.DB.Create(&task)
	if result.Error != nil {
		log.Printf("Errore nella creazione del task nel database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Task creato con successo: %+v", task)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("Risposta inviata con successo")
}

func ListTasks(w http.ResponseWriter, r *http.Request) {
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

	var tasks []models.Task
	result := config.DB.Where("autore_id = ?", user.ID).Find(&tasks)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	idStr1 := r.URL.Path[len("/delete_task/"):]
	id, err := strconv.Atoi(idStr1)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	var task models.Task
	result := config.DB.First(&task, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	result = config.DB.Delete(&task)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Project deleted")
}
