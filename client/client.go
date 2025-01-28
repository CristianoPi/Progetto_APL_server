package main

import (
	"Progetto_APL/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("4f8b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b"))

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	// Create a new HTTP client with cookie jar
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}

	baseURL := "http://localhost:8080"

	for {
		fmt.Println("Menu:")
		fmt.Println("1. Crea un nuovo utente")
		fmt.Println("2. Recupera un utente tramite email")
		fmt.Println("3. Login")
		fmt.Println("4. Crea un nuovo progetto")
		fmt.Println("5. Elenca i progetti dell'utente")
		fmt.Println("6. Elimina un progetto")
		fmt.Println("7. Crea un nuovo task")
		fmt.Println("8. Elenca i tasks")
		fmt.Println("9. Elimina un task")
		fmt.Println("10. Crea un nuovo file")
		fmt.Println("11. Elenca i file")
		fmt.Println("12. Recupera un file tramite ID")
		fmt.Println("13. Elimina un file")
		fmt.Println("14. Esci")
		fmt.Print("Scegli un'opzione: ")

		var choice int
		fmt.Scan(&choice)

		switch choice {
		case 1:
			createUser(client, baseURL)
		case 2:
			getUser(client, baseURL)
		case 3:
			login(client, baseURL)
		case 4:
			createProject(client, baseURL)
		case 5:
			listProjects(client, baseURL)
		case 6:
			deleteProject(client, baseURL)
		case 7:
			createTask(client, baseURL)
		case 8:
			listTasks(client, baseURL)
		case 9:
			deleteTask(client, baseURL)
		case 10:
			createFile(client, baseURL)
		case 11:
			listFiles(client, baseURL)
		case 12:
			getFile(client, baseURL)
		case 13:
			deleteFile(client, baseURL)
		case 14:
			fmt.Println("Uscita...")
			return
		default:
			fmt.Println("Opzione non valida. Riprova.")
		}
	}
}

func createUser(client *http.Client, baseURL string) {
	var newUser models.User
	fmt.Print("Inserisci il nome dell'utente: ")
	fmt.Scan(&newUser.Username)
	fmt.Print("Inserisci l'email dell'utente: ")
	fmt.Scan(&newUser.Email)
	fmt.Print("Inserisci la password dell'utente: ")
	var password string
	fmt.Scan(&password)

	// Cripta la password
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Errore nella criptazione della password: %v", err)
	}
	newUser.Pwd = string(hashedPwd)

	body, err := json.Marshal(newUser)
	if err != nil {
		log.Fatalf("Unable to marshal user: %v", err)
	}

	resp, err := client.Post(baseURL+"/users", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to create user: %v", resp.Status)
	}

	fmt.Println("User created successfully")
}

func getUser(client *http.Client, baseURL string) {
	var email string
	fmt.Print("Inserisci l'email dell'utente: ")
	fmt.Scan(&email)

	resp, err := client.Get(fmt.Sprintf("%s/users/%s", baseURL, email))
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get user: %v", resp.Status)
	}

	var user models.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		log.Fatalf("Unable to decode response: %v", err)
	}

	fmt.Printf("User: %+v\n", user)
}

func login(client *http.Client, baseURL string) {
	var creds Credentials
	fmt.Print("Inserisci l'email: ")
	fmt.Scan(&creds.Email)
	fmt.Print("Inserisci la password: ")
	fmt.Scan(&creds.Password)

	body, err := json.Marshal(creds)
	if err != nil {
		log.Fatalf("Unable to marshal credentials: %v", err)
	}

	resp, err := client.Post(baseURL+"/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to login: %v", resp.Status)
	}

	fmt.Println("Login successful")
}

func createProject(client *http.Client, baseURL string) {
	var newProject models.Project
	fmt.Print("Inserisci la descrizione del progetto: ")
	fmt.Scan(&newProject.Descrizione)

	// Recupera l'ID dell'utente dalla sessione
	userID, err := getUserIDFromSession(client, baseURL)
	if err != nil {
		log.Fatalf("Unable to get user ID from session: %v", err)
	}
	// ...existing code...

	newProject.AutoreID = userID

	body, err := json.Marshal(newProject)
	if err != nil {
		log.Fatalf("Unable to marshal project: %v", err)
	}

	resp, err := client.Post(baseURL+"/project", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	// Leggi il corpo della risposta
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Unable to read response body: %v", err)
	}

	// Logga il corpo della risposta
	log.Printf("Response body: %s", responseBody)

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to create project: %v", resp.Status)
	}

	// Tenta di decodificare la risposta
	var result map[string]interface{}
	if err := json.Unmarshal(responseBody, &result); err != nil {
		log.Fatalf("Unable to decode response: %v", err)
	}

	// ...existing code...
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to create project: %v", resp.Status)
	}

	fmt.Println("Project created successfully")
}

func listProjects(client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/projects")
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to list projects: %v", resp.Status)
	}

	var projects []models.Project
	err = json.NewDecoder(resp.Body).Decode(&projects)
	if err != nil {
		log.Fatalf("Unable to decode response: %v", err)
	}

	fmt.Printf("Projects: %+v\n", projects)
}

func deleteProject(client *http.Client, baseURL string) {
	var id uint
	fmt.Print("Inserisci l'ID del progetto: ")
	fmt.Scan(&id)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/projects/%d", baseURL, id), nil)
	if err != nil {
		log.Fatalf("Unable to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to delete project: %v", resp.Status)
	}

	fmt.Println("Project deleted successfully")
}

func createTask(client *http.Client, baseURL string) {
	var newTask models.Task
	fmt.Print("Inserisci la descrizione del task: ")
	fmt.Scan(&newTask.Descrizione)

	// Recupera l'ID dell'utente dalla sessione
	userID, err := getUserIDFromSession(client, baseURL)
	if err != nil {
		log.Fatalf("Unable to get user ID from session: %v", err)
	}
	newTask.AutoreID = userID

	body, err := json.Marshal(newTask)
	if err != nil {
		log.Fatalf("Unable to marshal task: %v", err)
	}

	resp, err := client.Post(baseURL+"/tasks", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Failed to create task: %v", resp.Status)
	}

	fmt.Println("Task created successfully")
}

func listTasks(client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/tasks")
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to list tasks: %v", resp.Status)
	}

	var tasks []models.Task
	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		log.Fatalf("Unable to decode response: %v", err)
	}

	fmt.Printf("Tasks: %+v\n", tasks)
}

func deleteTask(client *http.Client, baseURL string) {
	var id uint
	fmt.Print("Inserisci l'ID del task: ")
	fmt.Scan(&id)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/tasks/%d", baseURL, id), nil)
	if err != nil {
		log.Fatalf("Unable to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to delete task: %v", resp.Status)
	}

	fmt.Println("Task deleted successfully")
}

func createFile(client *http.Client, baseURL string) {
	var filePath string
	fmt.Print("Inserisci il percorso del file: ")
	fmt.Scan(&filePath)

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Unable to open file: %v", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		log.Fatalf("Unable to create form file: %v", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Fatalf("Unable to copy file: %v", err)
	}
	fmt.Print("Inserisci la descrizione del file: ")
	var descrizione string
	fmt.Scan(&descrizione)
	writer.WriteField("descrizione", descrizione)
	writer.Close()

	req, err := http.NewRequest("POST", baseURL+"/files", body)
	if err != nil {
		log.Fatalf("Unable to create request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to create file: %v", resp.Status)
	}

	var newFile models.File
	err = json.NewDecoder(resp.Body).Decode(&newFile)
	if err != nil {
		log.Fatalf("Unable to decode response: %v", err)
	}

	fmt.Printf("File created: %+v\n", newFile)
}

func listFiles(client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/files")
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to list files: %v", resp.Status)
	}

	var files []models.File
	err = json.NewDecoder(resp.Body).Decode(&files)
	if err != nil {
		log.Fatalf("Unable to decode response: %v", err)
	}

	fmt.Printf("Files: %+v\n", files)
}

func getFile(client *http.Client, baseURL string) {
	var id uint
	fmt.Print("Inserisci l'ID del file: ")
	fmt.Scan(&id)

	resp, err := client.Get(fmt.Sprintf("%s/files/%d", baseURL, id))
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to get file: %v", resp.Status)
	}

	var file models.File
	err = json.NewDecoder(resp.Body).Decode(&file)
	if err != nil {
		log.Fatalf("Unable to decode response: %v", err)
	}

	fmt.Printf("File: %+v\n", file)
}

func deleteFile(client *http.Client, baseURL string) {
	var id uint
	fmt.Print("Inserisci l'ID del file: ")
	fmt.Scan(&id)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/files/%d", baseURL, id), nil)
	if err != nil {
		log.Fatalf("Unable to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Failed to delete file: %v", resp.Status)
	}

	fmt.Println("File deleted successfully")
}

func getUserIDFromSession(client *http.Client, baseURL string) (uint, error) {
	resp, err := client.Get(baseURL + "/profile")
	if err != nil {
		return 0, fmt.Errorf("unable to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get user profile: %v", resp.Status)
	}

	var user models.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return 0, fmt.Errorf("unable to decode response: %v", err)
	}

	return user.ID, nil
}
