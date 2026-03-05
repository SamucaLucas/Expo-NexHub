package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"nexhub/config"
	"nexhub/models"
	"nexhub/structs"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// 1. DASHBOARD DA EMPRESA
func DashboardEmpresa(w http.ResponseWriter, r *http.Request) {
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Busca estatísticas (Quantos salvos)
	stats, _ := models.BuscarStatsEmpresa(user.Id)
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)

	dados := struct {
		Usuario  structs.Usuario
		Stats    structs.DashboardStats
		NaoLidas int
	}{
		Usuario:  user,
		Stats:    stats,
		NaoLidas: totalNaoLidas,
	}

	temp.ExecuteTemplate(w, "DashboardEmpresa", dados)
}

// 2. PERFIL DA EMPRESA (Visualizar e Editar)

func DetalheEmpresaHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Pega o ID da URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 2. Busca os dados da Empresa (Dona do Perfil)
	empresa, err := models.BuscarUsuarioPorID(id)
	if err != nil {
		http.Error(w, "Empresa não encontrada", 404)
		return
	}

	// Validação simples para garantir que é uma empresa
	if empresa.TipoUsuario != "EMPRESA" {
		// Se for um Dev, redireciona para o detalhe de talento
		http.Redirect(w, r, fmt.Sprintf("/talento/detalhes?id=%d", id), http.StatusSeeOther)
		return
	}

	// 3. Lógica de Quem está Visitando (Usuario Logado)
	var usuarioLogado structs.Usuario
	role := "Visitante"
	estaSalvo := false

	session, _ := config.Store.Get(r, "nexhub-session")
	if userId, ok := session.Values["userId"].(int); ok {
		usuarioEncontrado, err := models.BuscarUsuarioPorID(userId)
		if err == nil {
			usuarioLogado = usuarioEncontrado

			if usuarioLogado.TipoUsuario == "DEV" {
				role = "Dev"
				// Verifica se o Dev favoritou esta empresa
				estaSalvo = models.ChecarSeFavoritou(usuarioLogado.Id, empresa.Id, "EMPRESA")
			} else if usuarioLogado.TipoUsuario == "EMPRESA" {
				role = "Empresa"
			}
		}
	}

	// 4. Monta o pacote de dados
	dados := struct {
		Usuario structs.Usuario // Quem visita (Navbar/Notificações)
		Empresa structs.Usuario // Perfil sendo visto
		Role    string
		Salvou  bool
	}{
		Usuario: usuarioLogado,
		Empresa: empresa,
		Role:    role,
		Salvou:  estaSalvo,
	}

	// 5. Renderiza
	temp.ExecuteTemplate(w, "DetalheEmpresa", dados)
}

func PerfilEmpresa(w http.ResponseWriter, r *http.Request) {
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// --- SALVAR (POST) ---
	if r.Method == "POST" {
		r.ParseMultipartForm(10 << 20) // 10MB

		// Coleta dados
		user.NomeFantasia = r.FormValue("nome_fantasia")
		user.NomeCompleto = r.FormValue("responsavel")
		user.RamoAtuacao = r.FormValue("ramo")
		user.SiteEmpresa = r.FormValue("site")
		user.Cidade = r.FormValue("cidade")
		user.Biografia = r.FormValue("sobre")
		user.Email = r.FormValue("email") // Permite mudar email

		// Lógica de Senha
		novaSenha := r.FormValue("nova_senha")
		confirmarSenha := r.FormValue("confirmar_senha")

		if novaSenha != "" {
			if novaSenha != confirmarSenha {
				http.Redirect(w, r, "/empresa/perfil?erro=senhas_nao_conferem", http.StatusSeeOther)
				return
			}
			// Gera Hash
			hash, _ := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
			user.SenhaHash = string(hash)
		}
		// Se novaSenha for vazia, o user.SenhaHash continua com o valor antigo que veio do banco (user),
		// então o AtualizarPerfilEmpresa vai salvar o hash antigo de novo (sem problemas).

		// Upload Logo (Mantido igual)
		file, handler, err := r.FormFile("logo_empresa")
		if err == nil {
			defer file.Close()
			nomeArquivo := fmt.Sprintf("company_%d_%d_%s", user.Id, time.Now().Unix(), handler.Filename)
			caminhoDisco := "static/uploads/" + nomeArquivo
			caminhoBanco := "/static/uploads/" + nomeArquivo
			os.MkdirAll("static/uploads", os.ModePerm)
			dst, _ := os.Create(caminhoDisco)
			defer dst.Close()
			io.Copy(dst, file)
			user.FotoPerfil = caminhoBanco
		}

		// Salva no Banco (Essa função precisa aceitar SenhaHash no UPDATE, verifique seu model)
		err = models.AtualizarPerfilEmpresa(user)
		if err != nil {
			log.Println("Erro ao atualizar:", err)
			http.Redirect(w, r, "/empresa/perfil?erro=banco", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/empresa/perfil?sucesso=true", http.StatusSeeOther)
		return
	}
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)

	// --- VISUALIZAR (GET) ---
	dados := struct {
		Usuario  structs.Usuario
		NaoLidas int
	}{
		Usuario:  user,
		NaoLidas: totalNaoLidas,
	}
	temp.ExecuteTemplate(w, "EmpresaPerfil", dados)
}

// Rota: /favoritar?id=10&tipo=PROJETO
func ToggleFavoritoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Segurança: Só quem está logado pode salvar
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 2. Pega os dados da URL
	idItem, _ := strconv.Atoi(r.URL.Query().Get("id"))
	tipo := r.URL.Query().Get("tipo") // Deve ser 'PROJETO' ou 'DEV'

	if idItem == 0 || (tipo != "PROJETO" && tipo != "DEV") {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 3. Chama o Model
	models.AlternarFavorito(user.Id, idItem, tipo)

	// 4. Redireciona de volta para a página que o usuário estava (Refresh)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func EmpresaProjetos(w http.ResponseWriter, r *http.Request) {
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	termo := r.URL.Query().Get("q")
	filtro := r.URL.Query().Get("filter") // 'salvos' ou vazio
	cidade := r.URL.Query().Get("cidade") // Cidade
	area := r.URL.Query().Get("area")

	cidadeNoHTML := cidade // Variável para manter o select marcado no HTML

	if cidade == "" {
		cidade = user.Cidade // Padrão: Cidade da Empresa
		cidadeNoHTML = user.Cidade
	} else if cidade == "Todas" {
		cidade = "" // Para o SQL entender que não tem filtro
	}

	var projetos []structs.Projeto

	if filtro == "salvos" {
		projetos, _ = models.BuscarProjetosSalvos(user.Id)
		for i := range projetos {
			projetos[i].EstaSalvo = true
		}
	} else {
		// Busca TODOS os projetos (vitrine geral) - Reutiliza função do Public
		projetos, _ = models.BuscarTodosProjetos(termo, area, cidade, user.Id)
	}
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)

	dados := struct {
		Usuario      structs.Usuario
		Projetos     []structs.Projeto
		Termo        string
		Filtro       string
		FiltroArea   string
		FiltroCidade string
		NaoLidas     int
	}{
		Usuario:      user,
		Projetos:     projetos,
		Termo:        termo,
		Filtro:       filtro,
		FiltroArea:   area,
		FiltroCidade: cidadeNoHTML,
		NaoLidas:     totalNaoLidas,
	}

	temp.ExecuteTemplate(w, "EmpresaProjetos", dados)
}

func EmpresaTalentos(w http.ResponseWriter, r *http.Request) {
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	termo := r.URL.Query().Get("q")
	nivel := r.URL.Query().Get("nivel") // Junior/Pleno/Senior
	cidade := r.URL.Query().Get("cidade")
	filtro := r.URL.Query().Get("filter")

	// --- LÓGICA DE CIDADE PADRÃO ---
	cidadeSelecionadaNoHTML := cidade
	if cidade == "" {
		cidade = user.Cidade
		cidadeSelecionadaNoHTML = user.Cidade
	} else if cidade == "Todas" {
		cidade = ""
	}
	// -------------------------------

	var talentos []structs.Usuario

	if filtro == "salvos" {
		talentos, _ = models.BuscarTalentosSalvos(user.Id)
		for i := range talentos {
			talentos[i].EstaSalvo = true
		}
	} else {
		talentos, _ = models.BuscarTodosTalentos(termo, nivel, cidade, user.Id)
	}
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)

	dados := struct {
		Talentos     []structs.Usuario
		Usuario      structs.Usuario
		Termo        string
		Filtro       string
		FiltroNivel  string
		FiltroCidade string
		NaoLidas     int
	}{
		Talentos:     talentos,
		Usuario:      user,
		Termo:        termo,
		Filtro:       filtro,
		FiltroNivel:  nivel,
		FiltroCidade: cidadeSelecionadaNoHTML,
		NaoLidas:     totalNaoLidas,
	}

	temp.ExecuteTemplate(w, "EmpresaTalentos", dados)
}
