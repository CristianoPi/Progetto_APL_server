package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"encoding/json"
	"fmt"
	"io"

	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
)

// CreateFile gestisce l'upload e la creazione di un nuovo file
func CreateFile(w http.ResponseWriter, r *http.Request) {
	log.Println("FILE CONTROLLER: Inizio funzione CreateFile")

	// Recupera il file dal form data
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nel recupero del file: %v", err)
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Recupera l'ID del task dal form data
	taskIDStr := r.FormValue("taskId")
	if taskIDStr == "" {
		log.Println("FILE CONTROLLER: ID del task mancante")
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nella conversione dell'ID del task: %v", err)
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	log.Printf("FILE CONTROLLER: header filename %v", header.Filename)

	// Genera un nome univoco per il file
	uniqueFileName := uuid.New().String() + filepath.Ext(header.Filename)
	filePath := filepath.Join("file_code_py", uniqueFileName)

	// Salva il file nel server
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nella creazione del file: %v", err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nella copia del file: %v", err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	// Usa il nome del file originale come descrizione
	descrizione := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	newFile := models.File{
		Link:        filePath,
		Descrizione: descrizione,
		TaskID:      uint(taskID),
	}

	result := config.DB.Create(&newFile)
	if result.Error != nil {
		log.Printf("FILE CONTROLLER: Errore nel salvataggio del file nel database: %v", result.Error)
		http.Error(w, "Unable to save file details", http.StatusInternalServerError)
		return
	}

	log.Println("FILE CONTROLLER: File caricato con successo e dettagli salvati nel database")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newFile); err != nil {
		log.Printf("FILE CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Errore nella codifica della risposta JSON", http.StatusInternalServerError)
		return
	}
	log.Println("FILE CONTROLLER: Risposta inviata con successo")
}

// // ListFiles elenca tutti i file caricati
// func ListFiles(w http.ResponseWriter, r *http.Request) {
// 	log.Println("FILE CONTROLLER: Inizio funzione ListFiles")

// 	files, err := ioutil.ReadDir("uploads")
// 	if err != nil {
// 		log.Printf("FILE CONTROLLER: Errore nella lettura della directory: %v", err)
// 		http.Error(w, "Unable to list files", http.StatusInternalServerError)
// 		return
// 	}

// 	var fileList []string
// 	for _, file := range files {
// 		fileList = append(fileList, file.Name())
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	if err := json.NewEncoder(w).Encode(fileList); err != nil {
// 		log.Printf("FILE CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
// 		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Println("FILE CONTROLLER: Lista dei file inviata con successo")
// }

// DeleteFile elimina un file tramite ID
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	log.Println("FILE CONTROLLER: Inizio funzione DeleteFile")

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		log.Println("FILE CONTROLLER: ID del file mancante")
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("file_code_py", id)
	if err := os.Remove(filePath); err != nil {
		log.Printf("FILE CONTROLLER: Errore nell'eliminazione del file: %v", err)
		http.Error(w, "Unable to delete file", http.StatusInternalServerError)
		return
	}

	log.Println("FILE CONTROLLER: File eliminato con successo")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File deleted successfully")
	log.Println("FILE CONTROLLER: Risposta inviata con successo")
}

// GetFilesByTaskID recupera i file associati a un task tramite ID
func GetFilesByTaskID(w http.ResponseWriter, r *http.Request) {
	log.Println("FILE CONTROLLER: Inizio funzione GetFilesByTaskID")

	taskIDStr := r.URL.Query().Get("task_id")
	if taskIDStr == "" {
		log.Println("FILE CONTROLLER: ID del task mancante")
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	log.Printf("FILE CONTROLLER: ID del task ricevuto: %v", taskIDStr)

	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nella conversione dell'ID del task: %v", err)
		http.Error(w, "Invalid Task ID", http.StatusBadRequest)
		return
	}

	// Recupera i dettagli del task
	var task models.Task
	result := config.DB.First(&task, taskID)
	if result.Error != nil {
		log.Printf("FILE CONTROLLER: Errore nel recupero dei dettagli del task: %v", result.Error)
		http.Error(w, "Unable to retrieve task details", http.StatusInternalServerError)
		return
	}

	var fileList []models.File

	// Verifica se il task ha un CodeID valido
	if task.CodeID != 0 {
		// Recupera i dettagli del codice associato al task
		var code models.Code
		result = config.DB.First(&code, task.CodeID)
		if result.Error != nil {
			log.Printf("FILE CONTROLLER: Errore nel recupero dei dettagli del codice: %v", result.Error)
		} else {
			// Crea un oggetto File con i dettagli del codice
			codeFile := models.File{
				ID:          0,
				Link:        code.Codice,
				Descrizione: code.Descrizione,
				TaskID:      uint(taskID),
			}
			fileList = append(fileList, codeFile)
		}
	}

	// Recupera i file associati al task
	var files []models.File
	result = config.DB.Where("task_id = ?", taskID).Find(&files)
	if result.Error != nil {
		log.Printf("FILE CONTROLLER: Errore nel recupero dei file per il task ID %v: %v", taskID, result.Error)
		http.Error(w, "Unable to retrieve files", http.StatusInternalServerError)
		return
	}

	// Aggiungi i file associati al task alla lista
	fileList = append(fileList, files...)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fileList); err != nil {
		log.Printf("FILE CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
		return
	}

	log.Println("FILE CONTROLLER: Lista dei file inviata con successo")
}

// DownloadFile gestisce il download di un file tramite il percorso fornito dal client
func DownloadFile(w http.ResponseWriter, r *http.Request) {
	log.Println("FILE CONTROLLER: Inizio funzione DownloadFile")

	// Recupera i campi dal form data
	idStr := r.FormValue("id")
	link := r.FormValue("link")
	descrizione := r.FormValue("descrizione")
	taskIDStr := r.FormValue("taskId")

	if idStr == "" || link == "" || descrizione == "" || taskIDStr == "" {
		log.Println("FILE CONTROLLER: Campi mancanti nel form data")
		http.Error(w, "Missing form data", http.StatusBadRequest)
		return
	}

	log.Printf("FILE CONTROLLER: Campi ricevuti - id: %v, link: %v, descrizione: %v, taskId: %v", idStr, link, descrizione, taskIDStr)

	// Apri il file
	filePath := link
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nell'apertura del file: %v", err)
		http.Error(w, "Unable to open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	// Imposta gli header della risposta
	w.Header().Set("Content-Disposition", "attachment; filename="+descrizione)
	w.Header().Set("Content-Type", "application/octet-stream")

	// Copia il contenuto del file nella risposta
	if _, err := io.Copy(w, f); err != nil {
		log.Printf("FILE CONTROLLER: Errore nella copia del file nella risposta: %v", err)
		http.Error(w, "Unable to send file", http.StatusInternalServerError)
		return
	}

	log.Println("FILE CONTROLLER: File inviato con successo")
}
