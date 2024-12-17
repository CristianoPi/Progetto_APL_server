package models

type User struct {
	ID    int    `json:"id"`
	Nome  string `json:"nome"`
	Pwd   string `json:"pwd"`
	Email string `json:"email"`
}

type Progetto struct {
	ID        int    `json:"id"`
	Obiettivi string `json:"obiettivi"`
	CreatoDa  int    `json:"creato_da"`
}

type Task struct {
	ID               int    `json:"id"`
	Descrizione      string `json:"descrizione"`
	Commenti         string `json:"commenti"`
	Modifiche        string `json:"modifiche"`
	CreatoDa         int    `json:"creato_da"`
	AssegnatoDa      int    `json:"assegnato_da"`
	CodiceSorgenteID int    `json:"codice_sorgente_id"`
	AllegatoID       int    `json:"allegato_id"`
}

type CodiceSorgente struct {
	ID          int    `json:"id"`
	Codice      string `json:"codice"`
	Descrizione string `json:"descrizione"`
	Statistiche string `json:"statistiche"`
}

type Allegato struct {
	ID          int    `json:"id"`
	Link        string `json:"link"`
	Descrizione string `json:"descrizione"`
}
