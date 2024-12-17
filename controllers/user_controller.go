package controllers

import (
	"Progetto_APL/config"
	"Progetto_APL/models"
	"database/sql"
	"encoding/json"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := "INSERT INTO users (nome, pwd, email) VALUES (?, ?, ?)"
	_, err = config.DB.Exec(query, user.Nome, user.Pwd, user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/users/"):]
	query := "SELECT id, nome, email FROM users WHERE id = ?"
	row := config.DB.QueryRow(query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Nome, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(user)
}
