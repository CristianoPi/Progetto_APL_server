// le tabelle vengono mappate in strutture
// considernado ad esempio:
// [Nome  string `json:"nome"`] significa [NomeColonna tipo 'json:nome nel json per serializzazione/deserializzazione']
package models

type User struct {
	ID       uint   `gorm:"primarykey"`
	Username string `json:"username" gorm:"unique"`
	Pwd      string `json:"pwd"` // Campo per la password hashata
	Email    string `json:"email" gorm:"unique"`
}

type Project struct {
	ID          uint   `gorm:"primarykey"`
	Descrizione string `json:"descrizione"`
	AutoreID    uint   `json:"autoreID"` // Chiave esterna che fa riferimento a User
}

type Code struct {
	ID          uint   `gorm:"primarykey"`
	Codice      string `json:"codice"`
	Descrizione string `json:"descrizione"`
	Statistiche string `json:"statistiche"`
}

type Execution struct {
	ID     uint   `gorm:"primarykey"`
	CodeID uint   `json:"code_id"`
	Status string `json:"status"` // "in corso", "completato", "fallito"
	Output string `json:"output"`
	Error  string `json:"error"`
}

type Task struct {
	ID           uint   `gorm:"primarykey"`
	Descrizione  string `json:"descrizione"`
	Commenti     string `json:"commenti"`
	Completato   bool   `json:"completato"`
	AutoreID     uint   `json:"autoreID"`           // Chiave esterna che fa riferimento a User
	IncaricatoID uint   `json:"incaricatoID"`       // Chiave esterna che fa riferimento a User
	CodeID       uint   `json:"codice_sorgente_id"` // Chiave esterna che fa riferimento a Code
	ProgettoID   uint   `json:"progettoID"`         // Chiave esterna che fa riferimento a Project

}

type File struct {
	ID          uint   `gorm:"primarykey"`
	Link        string `json:"link"`
	Descrizione string `json:"descrizione"`
	TaskID      uint   `json:"taskID"` // Chiave esterna che fa riferimento a task
}
