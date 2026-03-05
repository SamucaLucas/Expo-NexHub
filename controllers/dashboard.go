package controllers

import (
	"log"
	"net/http"
	"nexhub/models"
	"nexhub/structs"
)

// Estrutura de dados composta (Usuário + Projetos)
// Usaremos isso para passar tudo o que a página precisa de uma vez
type PageData struct {
	Usuario     structs.Usuario
	Projetos []structs.Projeto
	NaoLidas int
}

// ---------------------------------------------------------
// 1. DASHBOARD (Tela Inicial / Resumo)
// Rota: /dev/dashboard
// ---------------------------------------------------------
func DashboardDev(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticação (Padrão)
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)

	// 2. [NOVO] Busca os números reais do banco
	stats, err := models.BuscarStatsUsuario(user.Id)
	if err != nil {
		// Se der erro, zeramos para não quebrar a tela
		stats = structs.DashboardStats{}
	}

	// 3. Monta os dados para o HTML
	dados := struct {
		Usuario  structs.Usuario
		Stats    structs.DashboardStats
		NaoLidas int
	}{
		Usuario:  user,
		Stats:    stats,
		NaoLidas: totalNaoLidas,
	}

	temp.ExecuteTemplate(w, "DashboardDev", dados)
}

// ---------------------------------------------------------
// 2. MEUS PROJETOS (Lista Completa)
// Rota: /dev/meus-projetos
// ---------------------------------------------------------
func MeusProjetos(w http.ResponseWriter, r *http.Request) {
	// 1. Auth e Busca de Usuário (Precisamos do usuário para a Navbar/Sidebar)
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 2. Buscar a Lista de Projetos deste Dev
	projetos, err := models.BuscarProjetosDoDev(user.Id)
	if err != nil {
		log.Println("Erro ao buscar projetos:", err)
		// Não trava, apenas a lista vai vazia
	}
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)
	// 3. Empacota tudo (User + Projetos) para mandar pro HTML
	dados := PageData{
		Usuario:     user,
		Projetos: projetos,
		NaoLidas: totalNaoLidas,
	}

	// 4. Renderiza a página "MeusProjetos"
	// IMPORTANTE: Verifique se o seu HTML "meus_projetos.html" tem {{ define "MeusProjetos" }} no topo
	err = temp.ExecuteTemplate(w, "MeusProjetos", dados)
	if err != nil {
		log.Println("❌ Erro ao renderizar MeusProjetos:", err)
	}
}
