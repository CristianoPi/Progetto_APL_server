USE db;
CREATE TABLE IF NOT EXISTS utente (
            id INT AUTO_INCREMENT,
            nome VARCHAR(255) NOT NULL,
            pwd VARCHAR(255) NOT NULL,
            email VARCHAR(255) NOT NULL UNIQUE,
            PRIMARY KEY (id)
        );

CREATE TABLE IF NOT EXISTS progetto (
            id INT AUTO_INCREMENT,	
            descrizione TEXT NOT NULL,
            autore INT,
            PRIMARY KEY (id),
            FOREIGN KEY (autore) REFERENCES utente(id)
        );

CREATE TABLE IF NOT EXISTS codice_sorgente (
            id INT AUTO_INCREMENT,
            codice VARCHAR(255) NOT NULL,
            descrizione TEXT,
            statistiche TEXT,
            PRIMARY KEY (id)
        );
CREATE TABLE IF NOT EXISTS documento (
			id INT AUTO_INCREMENT,
            link VARCHAR(255) NOT NULL,
            descrizione TEXT,
            PRIMARY KEY (id)
		);

CREATE TABLE IF NOT EXISTS task (
            id INT AUTO_INCREMENT,
            nome VARCHAR(255) NOT NULL,
            descrizione TEXT NOT NULL,
            commenti TEXT,
            autore INT,
            incaricato INT,
            codice_sorgente_id INT,
            PRIMARY KEY (id),
            FOREIGN KEY (autore) REFERENCES utente(id),
            FOREIGN KEY (incaricato) REFERENCES utente(id),
            FOREIGN KEY (codice_sorgente_id) REFERENCES codice_sorgente(id)
        );
CREATE TABLE IF NOT EXISTS allegato (
			id_doc INT NOT NULL,
            id_task INT NOT NULL,
			FOREIGN KEY (id_doc) REFERENCES documento(id),
            FOREIGN KEY (id_task) REFERENCES task(id),
            PRIMARY KEY (id_doc, id_task)
		);
