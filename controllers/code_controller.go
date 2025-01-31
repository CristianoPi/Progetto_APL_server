package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

	// Recupera l'ID del task dal form data
	taskIDStr := r.FormValue("taskId")
	if taskIDStr == "" {
		log.Println("CODE CONTROLLER: ID del task mancante")
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		log.Printf("CODE CONTROLLER: Errore nella conversione dell'ID del task: %v", err)
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	// Genera un nome univoco per il file
	uniqueFileName := uuid.New().String() + filepath.Ext(header.Filename)
	filePath := filepath.Join("file_code_py", uniqueFileName)

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
		Descrizione: header.Filename,
		Statistiche: "",
	}

	result := config.DB.Create(&newCode)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel salvataggio del codice nel database: %v", result.Error)
		http.Error(w, "Unable to save code details", http.StatusInternalServerError)
		return
	}

	// Aggiorna il task con il nuovo CodeID
	task := models.Task{}
	if err := config.DB.First(&task, taskID).Error; err != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del task: %v", err)
		http.Error(w, "Unable to retrieve task", http.StatusInternalServerError)
		return
	}

	task.CodeID = newCode.ID
	if err := config.DB.Save(&task).Error; err != nil {
		log.Printf("CODE CONTROLLER: Errore nell'aggiornamento del task con il nuovo CodeID: %v", err)
		http.Error(w, "Unable to update task with new CodeID", http.StatusInternalServerError)
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

// RunCode esegue il codice Python in modo asincrono
func RunCode(w http.ResponseWriter, r *http.Request) {
	log.Println("CODE CONTROLLER: Inizio funzione RunCode")

	// Ottieni l'ID del task dai parametri della richiesta
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		log.Println("CODE CONTROLLER: ID del task mancante")
		http.Error(w, "ID del task mancante", http.StatusBadRequest)
		return
	}
	log.Printf("CODE CONTROLLER: ID del task ricevuto: %v", taskIDStr)

	// Recupera il task dal database utilizzando l'ID
	var task models.Task
	result := config.DB.First(&task, taskIDStr)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del task dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Recupera il codice dal database utilizzando il CodeID del task
	var code models.Code
	result = config.DB.First(&code, task.CodeID)
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

	// Rimuovi "file_code_py\" da codePath
	log.Printf("CODE CONTROLLER: Codice da eseguire: %s", code.Codice)
	relativeCodePath := strings.TrimPrefix(code.Codice, "file_code_py\\")
	log.Printf("CODE CONTROLLER: Percorso relativo del codice: %s", relativeCodePath)

	// Esegui il codice Python in modo asincrono
	go func(executionID uint, codePath string) {
		log.Printf("CODE CONTROLLER: Esecuzione del comando: python file_code_py/executor.py %s", codePath)
		cmd := exec.Command("python", "executor.py", codePath)
		cmd.Dir = "C:/Users/asus/OneDrive - Università degli Studi di Catania/UNI/MAGISTRALE/SECONDO ANNO/Advanced Programming Languages/Progetto/Progetto_APL/file_code_py" // Assicurati che questo sia il percorso corretto
		output, err := cmd.CombinedOutput()

		// Log dell'output e dell'errore
		log.Printf("CODE CONTROLLER: Output del comando: %s", string(output))
		if err != nil {
			log.Printf("CODE CONTROLLER: Errore del comando: %v", err)
		}

		// Aggiorna lo stato dell'esecuzione nel database
		var status string
		var errorMsg string
		if err != nil {
			status = "fallito"
			errorMsg = err.Error()
		} else {
			status = "completato"
		}

		// Parse the output to extract execution time, errors, and created files
		var outputData map[string]interface{}
		if err := json.Unmarshal(output, &outputData); err != nil {
			log.Printf("CODE CONTROLLER: Errore nel parsing dell'output JSON: %v", err)
			status = "fallito"
			errorMsg = "Errore nel parsing dell'output JSON"
		}

		// Aggiorna le statistiche del codice
		statistics := "Execution completed successfully"
		if status == "fallito" {
			statistics = "Execution failed"
		}

		config.DB.Model(&models.Execution{}).Where("id = ?", executionID).Updates(models.Execution{
			Status: status,
			Output: string(output),
			Error:  errorMsg,
		})

		// Aggiorna il campo Statistiche del codice
		config.DB.Model(&models.Code{}).Where("id = ?", code.ID).Update("Statistiche", statistics)

		// Aggiungi i file creati alla tabella File
		if createdFiles, ok := outputData["created_files"].([]interface{}); ok {
			log.Printf("CODE CONTROLLER: Numero di file creati: %d", len(createdFiles))
			for _, file := range createdFiles {
				if filePath, ok := file.(string); ok {
					log.Printf("CODE CONTROLLER: Aggiungendo file: %s", filePath)
					newFile := models.File{
						TaskID:      task.ID,
						Link:        "file_code_py/" + filePath,
						Descrizione: filePath,
					}
					if err := config.DB.Create(&newFile).Error; err != nil {
						log.Printf("CODE CONTROLLER: Errore nella creazione del record del file: %v", err)
					} else {
						log.Printf("CODE CONTROLLER: File creato con successo: %v", filePath)
					}
				} else {
					log.Printf("CODE CONTROLLER: Errore nel casting del filePath: %v", file)
				}
			}
		} else {
			log.Printf("CODE CONTROLLER: Nessun file creato trovato nell'output JSON")
		}
	}(execution.ID, relativeCodePath)

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

	// Ottieni l'ID del task dai parametri della richiesta
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		http.Error(w, "ID del task mancante", http.StatusBadRequest)
		return
	}
	log.Printf("CODE CONTROLLER: ID del task ricevuto: %v", taskIDStr)

	// Recupera il task dal database utilizzando l'ID
	var task models.Task
	result := config.DB.First(&task, taskIDStr)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del task dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Recupera il codice dal database utilizzando il CodeID del task
	var code models.Code
	result = config.DB.First(&code, task.CodeID)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del codice dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Recupera l'ultima esecuzione associata al codice
	var execution models.Execution
	result = config.DB.Where("code_id = ?", code.ID).Order("id DESC").First(&execution)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero dell'esecuzione dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Restituisci solo lo stato dell'esecuzione
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": execution.Status})
	log.Println("CODE CONTROLLER: Stato dell'esecuzione recuperato con successo")
}

// GetCodeStatistics recupera le statistiche del codice associato a un task tramite ID
func GetCodeStatistics(w http.ResponseWriter, r *http.Request) {
	log.Println("CODE CONTROLLER: Inizio funzione GetCodeStatistics")

	// Ottieni l'ID del task dai parametri della richiesta
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		http.Error(w, "ID del task mancante", http.StatusBadRequest)
		return
	}
	log.Printf("CODE CONTROLLER: ID del task ricevuto: %v", taskIDStr)

	// Recupera il task dal database utilizzando l'ID
	var task models.Task
	result := config.DB.First(&task, taskIDStr)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del task dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Recupera l'ultima esecuzione del codice associato al task
	var execution models.Execution
	result = config.DB.Where("code_id = ?", task.CodeID).First(&execution)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero dell'esecuzione dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the output to extract execution time and errors
	var outputData map[string]interface{}
	if err := json.Unmarshal([]byte(execution.Output), &outputData); err != nil {
		log.Printf("CODE CONTROLLER: Errore nel parsing dell'output JSON: %v", err)
		http.Error(w, "Errore nel parsing dell'output JSON", http.StatusInternalServerError)
		return
	}

	// Restituisci il tempo di esecuzione e gli errori
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"execution_time": outputData["execution_time"],
		"errors":         outputData["errors"],
	})
	log.Println("CODE CONTROLLER: Statistiche dell'esecuzione recuperate con successo")
}

// GetResults recupera l'output dell'esecuzione del codice associato a un task tramite ID
func GetResults(w http.ResponseWriter, r *http.Request) {
	log.Println("CODE CONTROLLER: Inizio funzione GetResults")

	// Ottieni l'ID del task dai parametri della richiesta
	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		http.Error(w, "ID del task mancante", http.StatusBadRequest)
		return
	}
	log.Printf("CODE CONTROLLER: ID del task ricevuto: %v", taskIDStr)

	// Recupera il task dal database utilizzando l'ID
	var task models.Task
	result := config.DB.First(&task, taskIDStr)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero del task dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Recupera l'ultima esecuzione del codice associato al task
	var execution models.Execution
	result = config.DB.Where("code_id = ?", task.CodeID).First(&execution)
	if result.Error != nil {
		log.Printf("CODE CONTROLLER: Errore nel recupero dell'esecuzione dal database: %v", result.Error)
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Parse the output to extract the output and created files
	var outputData map[string]interface{}
	if err := json.Unmarshal([]byte(execution.Output), &outputData); err != nil {
		log.Printf("CODE CONTROLLER: Errore nel parsing dell'output JSON: %v", err)
		http.Error(w, "Errore nel parsing dell'output JSON", http.StatusInternalServerError)
		return
	}

	// Restituisci l'output e i file creati
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"output":        outputData["stdout"],
		"created_files": outputData["created_files"],
	})
	log.Println("CODE CONTROLLER: Output e file creati dell'esecuzione recuperati con successo")
}
