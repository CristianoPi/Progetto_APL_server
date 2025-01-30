package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
)

// CreateFile gestisce l'upload e la creazione di un nuovo file
func CreateFile(w http.ResponseWriter, r *http.Request) {
	log.Println("FILE CONTROLLER: Inizio funzione CreateFile")

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nel recupero del file: %v", err)
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Genera un nome univoco per il file
	uniqueFileName := uuid.New().String() + filepath.Ext(header.Filename)
	filePath := filepath.Join("uploads", uniqueFileName)

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

	// Salva i dettagli del file nel database
	descrizione := r.FormValue("descrizione")
	newFile := models.File{
		Link:        filePath,
		Descrizione: descrizione,
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

// ListFiles elenca tutti i file caricati
func ListFiles(w http.ResponseWriter, r *http.Request) {
	log.Println("FILE CONTROLLER: Inizio funzione ListFiles")

	files, err := ioutil.ReadDir("uploads")
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nella lettura della directory: %v", err)
		http.Error(w, "Unable to list files", http.StatusInternalServerError)
		return
	}

	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(fileList); err != nil {
		log.Printf("FILE CONTROLLER: Errore nella codifica della risposta JSON: %v", err)
		http.Error(w, "Unable to encode response", http.StatusInternalServerError)
		return
	}

	log.Println("FILE CONTROLLER: Lista dei file inviata con successo")
}

// GetFileByID recupera un file tramite ID
func GetFileByID(w http.ResponseWriter, r *http.Request) {
	log.Println("FILE CONTROLLER: Inizio funzione GetFileByID")

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		log.Println("FILE CONTROLLER: ID del file mancante")
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join("uploads", id)
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("FILE CONTROLLER: Errore nell'apertura del file: %v", err)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+id)
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("FILE CONTROLLER: Errore nella copia del file: %v", err)
		http.Error(w, "Unable to send file", http.StatusInternalServerError)
		return
	}

	log.Println("FILE CONTROLLER: File inviato con successo")
}

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

	filePath := filepath.Join("uploads", id)
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
