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
	log.Println("TASK CONTROLLER: Inizio creazione task")

	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("TASK CONTROLLER: Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	user, ok := session.Values["user"].(models.User)
	if !ok {
		log.Println("TASK CONTROLLER: Utente non trovato nella sessione")
		http.Error(w, "Utente non autenticato", http.StatusUnauthorized)
		return
	}

	var task models.Task
	err = json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		log.Printf("TASK CONTROLLER: Errore nella decodifica del payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: Payload decodificato con successo: %+v", task)

	// Associa l'utente al task e imposta Completato a false
	task.AutoreID = user.ID
	task.Completato = false

	result := config.DB.Create(&task)
	if result.Error != nil {
		log.Printf("TASK CONTROLLER: Errore nella creazione del task nel database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("TASK CONTROLLER: Task creato con successo: %+v", task)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("TASK CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("TASK CONTROLLER: Risposta inviata con successo")
}

func ListTasks(w http.ResponseWriter, r *http.Request) {
	log.Println("TASK CONTROLLER: Inizio recupero tasks")

	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("TASK CONTROLLER: Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok || user.ID == 0 {
		log.Println("TASK CONTROLLER: Non sei autenticato")
		http.Error(w, "Non sei autenticato", http.StatusUnauthorized)
		return
	}

	var tasks []models.Task
	result := config.DB.Find(&tasks)
	if result.Error != nil {
		log.Printf("TASK CONTROLLER: Errore nel recupero dei tasks dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("TASK CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("TASK CONTROLLER: Tasks inviati con successo")
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {
	log.Println("TASK CONTROLLER: Inizio eliminazione task")

	// Recupera la sessione
	session, err := store.Get(r, "session-name")
	if err != nil {
		log.Printf("TASK CONTROLLER: Errore nel recupero della sessione: %v", err)
		http.Error(w, "Errore nel recupero della sessione", http.StatusInternalServerError)
		return
	}

	// Recupera i dettagli dell'utente dalla sessione
	user, ok := session.Values["user"].(models.User)
	if !ok {
		log.Println("TASK CONTROLLER: Utente non trovato nella sessione")
		http.Error(w, "Utente non autenticato", http.StatusUnauthorized)
		return
	}

	// Ottieni l'ID del task dai parametri della richiesta
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID del task mancante", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: ID del task ricevuto: %v", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID del task non valido", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: ID del task convertito con successo: %v", id)

	// Recupera il task dal database utilizzando l'ID
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

	// Verifica se l'utente è l'autore del task
	if task.AutoreID != user.ID {
		log.Println("TASK CONTROLLER: Utente non autorizzato a eliminare questo task")
		http.Error(w, "Utente non autorizzato a eliminare questo task", http.StatusForbidden)
		return
	}

	// Elimina il task dal database
	result = config.DB.Delete(&task)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Task deleted")
	log.Println("TASK CONTROLLER: Task eliminato con successo")
}

// UpdateTask aggiorna lo stato di completamento di un task
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	log.Println("TASK CONTROLLER: Inizio aggiornamento task")

	// Decodifica il payload JSON
	var updatedTask models.Task
	err := json.NewDecoder(r.Body).Decode(&updatedTask)
	if err != nil {
		log.Printf("TASK CONTROLLER: Errore nella decodifica del payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: Payload decodificato con successo: %+v", updatedTask)

	// Recupera il task dal database utilizzando l'ID
	var task models.Task
	result := config.DB.First(&task, updatedTask.ID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Aggiorna il campo Completato
	task.Completato = updatedTask.Completato

	// Salva le modifiche nel database
	result = config.DB.Save(&task)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
	log.Println("TASK CONTROLLER: Task aggiornato con successo")
}

func ListTasksByProject(w http.ResponseWriter, r *http.Request) {
	log.Println("TASK CONTROLLER: Inizio recupero tasks per progetto")

	// Recupera l'ID del progetto dai parametri della richiesta
	projectIDStr := r.URL.Query().Get("progetto_id")
	if projectIDStr == "" {
		http.Error(w, "ID del progetto mancante", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: ID del progetto ricevuto: %v", projectIDStr)

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		http.Error(w, "ID del progetto non valido", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: ID del progetto convertito con successo: %v", projectID)

	// Recupera i task dal database utilizzando l'ID del progetto
	var tasks []models.Task
	result := config.DB.Where("progetto_id = ?", projectID).Find(&tasks)
	if result.Error != nil {
		log.Printf("TASK CONTROLLER: Errore nel recupero dei tasks dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		log.Printf("TASK CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}

	log.Printf("TASK CONTROLLER: Tasks per progetto inviati con successo: %+v", tasks)
}

func GetTaskByID(w http.ResponseWriter, r *http.Request) {
	log.Println("TASK CONTROLLER: Inizio funzione GetTaskByID")

	// Ottieni l'ID del task dai parametri della richiesta
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID del task mancante", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: ID del task ricevuto: %v", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID del task non valido", http.StatusBadRequest)
		return
	}
	log.Printf("TASK CONTROLLER: ID del task convertito con successo: %v", id)

	// Recupera il task dal database utilizzando l'ID
	var task models.Task
	if err := config.DB.First(&task, id).Error; err != nil {
		log.Printf("TASK CONTROLLER: Errore nel recupero del task: %v", err)
		http.Error(w, "Errore nel recupero del task", http.StatusInternalServerError)
		return
	}

	log.Printf("TASK CONTROLLER: GetTaskByID torna %v", task)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		log.Printf("TASK CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("TASK CONTROLLER: Task recuperato con successo")
}
