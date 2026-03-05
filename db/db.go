package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Importa o driver do Postgres
)

var DB *sql.DB

// ConectaComBanco abre a conexão com o PostgreSQL
func ConectaComBanco() {
	// CONFIGURAÇÃO: Ajuste user, password e dbname conforme o seu computador!
	// Se tiver senha, coloque password=sua_senha
	connStr := "user=samucael dbname=nexhub host=localhost password=2784 sslmode=disable"

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Erro ao abrir conexão com o banco:", err)
	}

	// Testa se a conexão está viva
	err = DB.Ping()
	if err != nil {
		log.Fatal("Erro ao conectar (Ping) no banco:", err)
	}

	fmt.Println("✅ Conexão com o Banco de Dados (PostgreSQL) realizada com sucesso!")
}
