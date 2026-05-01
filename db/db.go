package db

import (
	"database/sql"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// IMPORTANTE: Ajuste o caminho abaixo para a pasta real do seu projeto!
	// Exemplo: "github.com/samucalucas/expo-nexhub/structs"
	"nexhub/structs"
)

// A sua variável clássica! Nenhum model vai quebrar, eles continuarão usando essa conexão.
var DB *sql.DB

// Variável para uso do Gorm caso decida usar os recursos avançados dele depois
var GormDB *gorm.DB

func ConectaComBanco() {
	// Lembre-se de substituir pela sua string de conexão real, ou puxar do .env
	dsn := "user=samucael dbname=nexhubexpo host=localhost password=2784 sslmode=disable TimeZone=America/Sao_Paulo"

	var err error

	// 1. Inicia a conexão poderosa usando o GORM
	GormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("🚨 Erro fatal ao conectar no banco via GORM: ", err)
	}

	// 2. Extrai a conexão *sql.DB padrão do GORM para popular a sua variável global
	DB, err = GormDB.DB()
	if err != nil {
		log.Fatal("🚨 Erro ao extrair conexão nativa (sql.DB) do GORM: ", err)
	}

	// Testa para ter certeza de que o banco está vivo
	if err = DB.Ping(); err != nil {
		log.Fatal("🚨 Banco de dados falhou ao responder ao Ping: ", err)
	}

	fmt.Println("✅ Banco de Dados conectado com sucesso!")

	// 3. ✨ O AUTO-MIGRATE ✨
	// O Gorm vai ler todas as structs e garantir que as tabelas existam no banco!
	err = GormDB.AutoMigrate(
		&structs.Area{},
		&structs.Curso{},
		&structs.Habilidade{},
		&structs.Usuario{},
		&structs.RecuperacaoSenha{},
		&structs.Aluno{},
		&structs.Projeto{},
		&structs.ProjetoArquivo{},
		&structs.ProjetoLink{},
		&structs.ProjetoImagem{},
		&structs.Avaliacao{},
	)

	if err != nil {
		log.Fatal("🚨 Erro ao sincronizar as tabelas (AutoMigrate): ", err)
	} else {
		fmt.Println("🚀 Estrutura do banco de dados verificada e sincronizada pelo GORM!")
	}
}
