# Progetto APL

Questo progetto è un'applicazione web scritta in Go che utilizza MySQL come database. 
L'applicazione gestisce e salva i progetti che possono essere creati e assegnati a diversi utenti.
I progetti sono composti da diversi task. I task hanno sempre collegato del codice in python.
L'applicazione consente di caricare ed eseguire codice Python in modo asincrono, salvando i risultati nel database.

## Requisiti

- Go 1.20 o superiore
- MySQL

## Configurazione

Assicurati di avere Go e MySQL installati sul tuo sistema.

1. Clonare la repository
2. Crea un database MySQL chiamato `DB`:
3. Eseguire go run main.go
