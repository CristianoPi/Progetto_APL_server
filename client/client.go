package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar" //il client deve essere in grado di gestire i cookie per la sessione
	"os"

	"golang.org/x/crypto/bcrypt"
)

// !!DEVO CAPIRE COME RENDERE VISIBILE MODELS
type User struct {
	ID    uint   `gorm:"primarykey"`
	Nome  string `json:"nome"`
	Pwd   string `json:"pwd" gorm:"size:60"` // Campo per la password hashata
	Email string `json:"email" gorm:"unique"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Project struct {
	ID          uint   `gorm:"primarykey"`
	Descrizione string `json:"descrizione"`
	AutoreID    uint   `json:"autore"` // Chiave esterna che fa riferimento a User
	Autore      User   `gorm:"foreignKey:AutoreID"`
}

type Code struct {
	ID          uint   `gorm:"primarykey"`
	Codice      string `json:"codice"`
	Descrizione string `json:"descrizione"`
	Statistiche string `json:"statistiche"`
}

type Task struct {
	//Per permettere che alcune istanze di Task non abbiano un collegamento con Code, puoi rendere il campo CodeID opzionale.
	//In Go, puoi fare questo utilizzando un puntatore al tipo uint per il campo CodeID. In questo modo, il campo può essere nil se non è presente un collegamento con Code.
	ID           uint    `gorm:"primarykey"`
	Nome         string  `json:"nome"`
	Descrizione  string  `json:"descrizione"`
	Commenti     string  `json:"commenti"`
	AutoreID     uint    `json:"autore"` // Chiave esterna che fa riferimento a User
	Autore       User    `gorm:"foreignKey:AutoreID"`
	IncaricatoID *uint   `json:"incaricato"` // Chiave esterna che fa riferimento a User
	Incaricato   User    `gorm:"foreignKey:IncaricatoID"`
	CodeID       *uint   `json:"codice_sorgente_id"` // Chiave esterna che fa riferimento a Code
	Code         Code    `gorm:"foreignKey:CodeID"`
	ProjectID    *uint   `json:"progetto"` // Chiave esterna che fa riferimento a Project
	Project      Project `gorm:"foreignKey:ProjectID"`
}

func main() {
	baseURL := "http://localhost:8080"

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Errore nella creazione del cookie jar: %v", err)
	}

	client := &http.Client{
		Jar: jar,
	}

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
		fmt.Println("10. Esci")
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
			fmt.Println("Uscita...")
			os.Exit(0)
		default:
			fmt.Println("Opzione non valida. Riprova.")
		}
	}
}

func createUser(client *http.Client, baseURL string) {
	var newUser User

	fmt.Print("Inserisci il nome dell'utente: ")
	fmt.Scan(&newUser.Nome)

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

	userJSON, err := json.Marshal(newUser)
	if err != nil {
		log.Fatalf("Errore nella serializzazione dell'utente: %v", err)
	}

	resp, err := client.Post(baseURL+"/users", "application/json", bytes.NewBuffer(userJSON))
	if err != nil {
		log.Fatalf("Errore nella richiesta POST: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Creazione utente fallita. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Utente creato con successo.")
}

func getUser(client *http.Client, baseURL string) {
	fmt.Print("Inserisci l'email dell'utente: ")
	var email string
	fmt.Scan(&email)

	url := fmt.Sprintf("%s/users/%s", baseURL, email)
	getResp, err := client.Get(url)
	if err != nil {
		log.Fatalf("Errore nella richiesta GET: %v", err)
	}
	defer getResp.Body.Close()

	getBody, err := io.ReadAll(getResp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if getResp.StatusCode != http.StatusOK {
		log.Fatalf("Recupero utente fallito. Stato HTTP: %d, Messaggio: %s", getResp.StatusCode, string(getBody))
	}

	fmt.Printf("Risposta al recupero dell'utente: %s\n", getBody)

	var fetchedUser User
	if err := json.Unmarshal(getBody, &fetchedUser); err != nil {
		log.Fatalf("Errore nel parsing della risposta JSON: %v", err)
	}

	fmt.Printf("Utente recuperato: %+v\n", fetchedUser)
}

func login(client *http.Client, baseURL string) {
	var creds Credentials

	fmt.Print("Inserisci l'email: ")
	fmt.Scan(&creds.Email)

	fmt.Print("Inserisci la password: ")
	fmt.Scan(&creds.Password)

	credsJSON, err := json.Marshal(creds)
	if err != nil {
		log.Fatalf("Errore nella serializzazione delle credenziali: %v", err)
	}

	resp, err := client.Post(baseURL+"/login", "application/json", bytes.NewBuffer(credsJSON))
	if err != nil {
		log.Fatalf("Errore nella richiesta POST: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Login fallito. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Login effettuato con successo.")
	fmt.Printf("Messaggio: %s\n", string(body))
}

func createProject(client *http.Client, baseURL string) {
	var newProject Project

	fmt.Print("Inserisci la descrizione del progetto: ")
	fmt.Scan(&newProject.Descrizione)

	projectJSON, err := json.Marshal(newProject)
	if err != nil {
		log.Fatalf("Errore nella serializzazione del progetto: %v", err)
	}

	resp, err := client.Post(baseURL+"/project", "application/json", bytes.NewBuffer(projectJSON))
	if err != nil {
		log.Fatalf("Errore nella richiesta POST: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Creazione progetto fallita. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Progetto creato con successo.")
}

func listProjects(client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/projects")
	if err != nil {
		log.Fatalf("Errore nella richiesta GET: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Elenco progetti fallito. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Risposta all'elenco dei progetti: %s\n", body)

	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		log.Fatalf("Errore nel parsing della risposta JSON: %v", err)
	}

	fmt.Println("Progetti recuperati:")
	for _, project := range projects {
		fmt.Printf("ID: %d, Descrizione: %s\n", project.ID, project.Descrizione)
	}
}

func deleteProject(client *http.Client, baseURL string) {
	fmt.Print("Inserisci l'ID del progetto da eliminare: ")
	var id int
	fmt.Scan(&id)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/delete_project/%d", baseURL, id), nil)

	if err != nil {
		log.Fatalf("Errore nella creazione della richiesta DELETE: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Errore nella richiesta DELETE: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Eliminazione progetto fallita. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Progetto eliminato con successo.")
}

func createTask(client *http.Client, baseURL string) {
	var newTask Task

	fmt.Print("Inserisci il nome del task: ")
	fmt.Scan(&newTask.Nome)

	fmt.Print("Inserisci la descrizione del task: ")
	fmt.Scan(&newTask.Descrizione)

	fmt.Print("Inserisci i commenti del task: ")
	fmt.Scan(&newTask.Commenti)

	fmt.Print("Inserisci l'ID dell'incaricato del task (può essere vuoto): ")
	var incaricatoID uint
	_, err := fmt.Scan(&incaricatoID)
	if err == nil {
		newTask.IncaricatoID = &incaricatoID
	}

	fmt.Print("Inserisci l'ID del codice sorgente del task (può essere vuoto): ")
	var codeID uint
	_, err = fmt.Scan(&codeID)
	if err == nil {
		newTask.CodeID = &codeID
	}

	fmt.Print("Inserisci l'ID del progetto del task (può essere vuoto): ")
	var projectID uint
	_, err = fmt.Scan(&projectID)
	if err == nil {
		newTask.ProjectID = &projectID
	}

	taskJSON, err := json.Marshal(newTask)
	if err != nil {
		log.Fatalf("Errore nella serializzazione del task: %v", err)
	}

	resp, err := client.Post(baseURL+"/task", "application/json", bytes.NewBuffer(taskJSON))
	if err != nil {
		log.Fatalf("Errore nella richiesta POST: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("Creazione task fallita. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Task creato con successo.")
}

func listTasks(client *http.Client, baseURL string) {
	resp, err := client.Get(baseURL + "/tasks")
	if err != nil {
		log.Fatalf("Errore nella richiesta GET: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Elenco tasks fallito. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}

	fmt.Printf("Risposta all'elenco dei tasks: %s\n", body)

	var tasks []Task
	if err := json.Unmarshal(body, &tasks); err != nil {
		log.Fatalf("Errore nel parsing della risposta JSON: %v", err)
	}

	fmt.Println("Tasks recuperati:")
	for _, task := range tasks {
		fmt.Printf("ID: %d, Nome: %s, Descrizione: %s\n", task.ID, task.Nome, task.Descrizione)
	}
}

func deleteTask(client *http.Client, baseURL string) {
	fmt.Print("Inserisci l'ID del task da eliminare: ")
	var id int
	fmt.Scan(&id)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/delete_task/%d", baseURL, id), nil)
	if err != nil {
		log.Fatalf("Errore nella creazione della richiesta DELETE: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Errore nella richiesta DELETE: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Errore nella lettura del corpo della risposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Eliminazione task fallita. Stato HTTP: %d, Messaggio: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Task eliminato con successo.")
}
