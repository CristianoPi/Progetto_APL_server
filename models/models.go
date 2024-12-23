// le tabelle vengono mappate in strutture
// considernado ad esempio:
// [Nome  string `json:"nome"`] significa [NomeColonna tipo 'json:nome nel json per serializzazione/deserializzazione']
package models

import (
	"gorm.io/gorm"
)

type User struct {
	ID    uint   `gorm:"primarykey"`
	Nome  string `json:"nome"`
	Pwd   string `json:"pwd" gorm:"size:60"` // Campo per la password hashata
	Email string `json:"email" gorm:"unique"`
}

type Project struct {
	gorm.Model
	Descrizione string `json:"descrizione"`
	AutoreID    uint   `json:"autore"` // Chiave esterna che fa riferimento a User
	Autore      User   `gorm:"foreignKey:AutoreID"`
}

type Code struct {
	gorm.Model
	Codice      string `json:"codice"`
	Descrizione string `json:"descrizione"`
	Statistiche string `json:"statistiche"`
}

type File struct {
	gorm.Model
	Link        string `json:"link"`
	Descrizione string `json:"descrizione"`
}

type Task struct {
	gorm.Model
	Nome         string `json:"nome"`
	Descrizione  string `json:"descrizione"`
	Commenti     string `json:"commenti"`
	AutoreID     uint   `json:"autore"` // Chiave esterna che fa riferimento a User
	Autore       User   `gorm:"foreignKey:AutoreID"`
	IncaricatoID uint   `json:"incaricato"` // Chiave esterna che fa riferimento a User
	Incaricato   User   `gorm:"foreignKey:IncaricatoID"`
	CodeID       uint   `json:"codice_sorgente_id"` // Chiave esterna che fa riferimento a Code
	Code         Code   `gorm:"foreignKey:CodeID"`
}

type Attached struct {
	IDFile uint `json:"id_file"` // Chiave esterna che fa riferimento a File
	IDTask uint `json:"id_task"` // Chiave esterna che fa riferimento a Task
	File   File `gorm:"foreignKey:IDFile"`
	Task   Task `gorm:"foreignKey:IDTask"`
}
