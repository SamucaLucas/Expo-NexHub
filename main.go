package main

import (
	"log"
	"net/http"
	"nexhub/db"
	"nexhub/routers" // <--- Importe seu pacote de rotas

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis do sistema")
	}
	// Carregar a conexão com o banco de dados
	db.ConectaComBanco()
	// Carrega todas as rotas definidas na pasta routers
	routers.CarregarRotas()

	log.Println("Servidor NexHub rodando em: http://localhost:8080")

	// Inicia o servidor
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
