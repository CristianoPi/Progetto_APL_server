func createTables(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id INT AUTO_INCREMENT,
            nome VARCHAR(255) NOT NULL,
            pwd VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL UNIQUE,
            PRIMARY KEY (id)
        );`,
		`CREATE TABLE IF NOT EXISTS progetti (
            id INT AUTO_INCREMENT,
            obiettivi TEXT NOT NULL,
            creato_da INT,
            PRIMARY KEY (id),
            FOREIGN KEY (creato_da) REFERENCES users(id)
        );`,
		`CREATE TABLE IF NOT EXISTS tasks (
            id INT AUTO_INCREMENT,
            descrizione TEXT NOT NULL,
            commenti TEXT,
            modifiche TEXT,
            creato_da INT,
            assegnato_da INT,
            codice_sorgente_id INT,
            allegato_id INT,
            PRIMARY KEY (id),
            FOREIGN KEY (creato_da) REFERENCES users(id),
            FOREIGN KEY (assegnato_da) REFERENCES users(id),
            FOREIGN KEY (codice_sorgente_id) REFERENCES codice_sorgente(id),
            FOREIGN KEY (allegato_id) REFERENCES allegati(id)
        );`,
		`CREATE TABLE IF NOT EXISTS codice_sorgente (
            id INT AUTO_INCREMENT,
            codice TEXT NOT NULL,
            descrizione TEXT,
            statistiche TEXT,
            PRIMARY KEY (id)
        );`,
		`CREATE TABLE IF NOT EXISTS allegati (
            id INT AUTO_INCREMENT,
            link TEXT NOT NULL,
            descrizione TEXT,
            PRIMARY KEY (id)
        );`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal("Failed to execute query:", err)
		}
	}

	fmt.Println("Tables created successfully!")
}