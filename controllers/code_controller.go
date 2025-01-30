package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"Progetto_APL/config"
	"Progetto_APL/models"

	"github.com/google/uuid"
)

// CreateCode gestisce l'upload e la creazione di un nuovo codice
func CreateCode(w http.ResponseWriter, r *http.Request) {
	log.Println("CODE CONTROLLER: Inizio funzione CreateCode")

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del file: %v", err)
		http.Error(w, "Code file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Genera un nome univoco per il file
	uniqueFileName := uuid.New().String() + filepath.Ext(header.Filename)
	filePath := filepath.Join("uploads_code", uniqueFileName)

	// Salva il file nel server
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("CODE CONTROLLER: Errore nella creazione del file: %v", err)
		http.Error(w, "Unable to save code file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		log.Printf("CODE CONTROLLER: Errore nella copia del file: %v", err)
		http.Error(w, "Unable to save code file", http.StatusInternalServerError)
		return
	}

	// Salva i dettagli del codice nel database
	newCode := models.Code{
		Codice:      filePath,
		Descrizione: "",
		Statistiche: "",
	}

	result := config.DB.Create(&newCode)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel salvataggio del codice nel database: %v", result.Error)
		http.Error(w, "Unable to save code details", http.StatusInternalServerError)
		return
	}

	log.Println("CODE CONTROLLER: Codice caricato con successo e dettagli salvati nel database")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newCode); err != nil {
		log.Printf("CODE CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("CODE CONTROLLER: Risposta inviata con successo")
}

// RunCode esegue il codice Python in modo asincrono e raccoglie le statistiche e l'output
func RunCode(w http.ResponseWriter, r *http.Request) {
	log.Println("CODE CONTROLLER: Inizio funzione RunCode")

	// Ottieni l'ID del codice dai parametri della richiesta
	idStr := r.URL.Query().Get("code_id")
	if idStr == "" {
		http.Error(w, "ID del codice mancante", http.StatusBadRequest)
		return
	}
	log.Printf("CODE CONTROLLER: ID del codice ricevuto: %v", idStr)

	// Recupera il codice dal database utilizzando l'ID
	var code models.Code
	result := config.DB.First(&code, idStr)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del codice dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Crea un nuovo record di esecuzione
	execution := models.Execution{
		CodeID: code.ID,
		Status: "in corso",
	}
	result = config.DB.Create(&execution)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nella creazione del record di esecuzione: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Esegui il codice Python in modo asincrono
	go func(executionID uint, codePath string) {
		cmd := exec.Command("python", "executor.py", codePath)
		output, err := cmd.CombinedOutput()

		// Aggiorna lo stato dell'esecuzione nel database
		var status string
		var errorMsg string
		if err != nil {
			status = "fallito"
			errorMsg = err.Error()
		} else {
			status = "completato"
		}

		config.DB.Model(&models.Execution{}).Where("id = ?", executionID).Updates(models.Execution{
			Status: status,
			Output: string(output),
			Error:  errorMsg,
		})
	}(execution.ID, code.Codice)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	response := map[string]interface{}{
		"message":      "Code execution started",
		"execution_id": execution.ID,
	}
	json.NewEncoder(w).Encode(response)
	log.Println("CODE CONTROLLER: Esecuzione del codice avviata")
}

// GetExecutionStatus recupera lo stato e i risultati di un'esecuzione
func GetExecutionStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("CODE CONTROLLER: Inizio funzione GetExecutionStatus")

	// Ottieni l'ID dell'esecuzione dai parametri della richiesta
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "ID dell'esecuzione mancante", http.StatusBadRequest)
		return
	}
	log.Printf("CODE CONTROLLER: ID dell'esecuzione ricevuto: %v", idStr)

	// Recupera l'esecuzione dal database utilizzando l'ID
	var execution models.Execution
	result := config.DB.First(&execution, idStr)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero dell'esecuzione dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(execution)
	log.Println("CODE CONTROLLER: Stato dell'esecuzione recuperato con successo")
}
