package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateFile gestisce l'upload e la creazione di un nuovo file
func CreateFile(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Salva il file nel server
	filePath := filepath.Join("uploads", header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	// Crea un nuovo record nel database
	newFile := models.File{
		Link:        filePath,
		Descrizione: r.FormValue("descrizione"),
	}
	result := config.DB.Create(&newFile)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(newFile)
}

// ListFiles elenca tutti i file
func ListFiles(w http.ResponseWriter, r *http.Request) {
	var files []models.File
	if err := config.DB.Find(&files).Error; err != nil {
		http.Error(w, "Unable to list files", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(files)
}

// GetFile ottiene un file per ID
func GetFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(file)
}

// DeleteFile elimina un file per ID
func DeleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	var file models.File
	if err := config.DB.First(&file, id).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Elimina il file dal server
	if err := os.Remove(file.Link); err != nil {
		http.Error(w, "Unable to delete file", http.StatusInternalServerError)
		return
	}

	// Elimina il record dal database
	if err := config.DB.Delete(&file).Error; err != nil {
		http.Error(w, "Unable to delete file record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "File deleted"})
}
