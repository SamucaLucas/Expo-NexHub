package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"nexhub/models"
	"nexhub/structs"
	"os"
	"strconv"
	"time"
)

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticação e Segurança
	user := GetUserFromSession(r)

	// Se o usuário não estiver logado ou não for ADMIN, redireciona para a home
	if user.Id == 0 || user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)
	// 2. Busca Estatísticas Reais do Banco de Dados
	// A função retorna: TotalDevs, TotalEmpresas, TotalProjetos, Labels(Meses), Dados(QtdUsers)
	devs, empresas, projs, banidos, meses, usersGrafico, err := models.BuscarEstatisticasAdmin()
	if err != nil {
		// Logamos o erro no servidor, mas deixamos a página carregar (com zeros) para não quebrar a UI
		log.Println("Erro crítico ao carregar dashboard admin:", err)
	}

	// 3. Monta a Struct de Dados para o HTML
	dados := structs.AdminDashboardData{
		Usuario:         user,
		TotalDevs:       devs,
		TotalEmpresas:   empresas,
		TotalProjetos:   projs,
		TotalBanidos:    banidos, // Implementar futuramente
		ChartMeses:      meses,
		ChartNovosUsers: usersGrafico,
		NaoLidas:        totalNaoLidas,
	}

	// 4. Renderiza o Template
	temp.ExecuteTemplate(w, "AdminDashboard", dados)
}

// Handler da Página de Usuários
func AdminUsuariosHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	// Segurança: Só Admin entra
	if user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)
	filtro := r.URL.Query().Get("q")
	listaUsuarios, _ := models.ListarTodosUsuarios(filtro)

	dados := struct {
		Usuario  interface{}
		Usuarios interface{} // Lista de usuários
		Filtro   string
		NaoLidas int
	}{
		Usuario:  user,
		Usuarios: listaUsuarios,
		Filtro:   filtro,
		NaoLidas: totalNaoLidas,
	}

	temp.ExecuteTemplate(w, "AdminUsuarios", dados)
}

// Handler da Ação de Promover a Admin
func AdminPromoverHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Pega o ID da URL (?id=10)
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	if id > 0 {
		models.TornarUsuarioAdmin(id)
	}

	// Recarrega a página
	http.Redirect(w, r, "/admin/usuarios", http.StatusSeeOther)
}

func AdminRemoverAdminHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	if id > 0 {
		// Evita que você remova seu próprio admin sem querer
		if id == user.Id {
			// (Opcional) Adicionar mensagem de erro "Você não pode se remover"
		} else {
			models.RemoverAdmin(id)
		}
	}
	http.Redirect(w, r, "/admin/usuarios", http.StatusSeeOther)
}

// Ação: Banir ou Desbanir
func AdminBanirHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	if id > 0 && id != user.Id { // Não pode banir a si mesmo
		models.AlternarBanimento(id)
	}

	http.Redirect(w, r, "/admin/usuarios", http.StatusSeeOther)
}

// controllers/admin.go

func AdminProjetosHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)
	busca := r.URL.Query().Get("q")
	status := r.URL.Query().Get("status")

	// ALTERAÇÃO AQUI: Capturamos o erro em vez de ignorar
	listaProjetos, err := models.ListarTodosProjetos(busca, status)
	if err != nil {
		log.Println("ERRO AO LISTAR PROJETOS NO CONTROLLER:", err)
	}

	dados := struct {
		Usuario  interface{}
		Projetos interface{}
		Busca    string
		Status   string
		NaoLidas int
	}{
		Usuario:  user,
		Projetos: listaProjetos,
		Busca:    busca,
		Status:   status,
		NaoLidas: totalNaoLidas,
	}

	temp.ExecuteTemplate(w, "AdminProjetos", dados)
}

// Ação: Ocultar/Aprovar Projeto (Toggle simples ou específico)
func AdminAlterarStatusProjetoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		return
	}

	idStr := r.URL.Query().Get("id")
	acao := r.URL.Query().Get("acao") // "ocultar" ou "aprovar"
	id, _ := strconv.Atoi(idStr)

	if id > 0 {
		if acao == "ocultar" {
			models.AlterarStatusProjeto(id, "Oculto")
		} else if acao == "aprovar" {
			// Volta para "Em Andamento" ou o padrão do sistema
			models.AlterarStatusProjeto(id, "Em Andamento")
		}
	}
	http.Redirect(w, r, "/admin/projetos", http.StatusSeeOther)
}

// Ação: Excluir Projeto
func AdminExcluirProjetoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		return
	}

	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	if id > 0 {
		models.ExcluirProjeto(id)
	}
	http.Redirect(w, r, "/admin/projetos", http.StatusSeeOther)
}

func AdminPerfilHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	totalNaoLidas, _ := models.ContarTotalNaoLidas(user.Id)
	// Criamos uma struct wrapper para passar ao template
	dados := struct {
		Usuario  interface{}
		NaoLidas int
	}{
		Usuario:  user,
		NaoLidas: totalNaoLidas,
	}
	temp.ExecuteTemplate(w, "AdminPerfil", dados)
}

// POST: Processa o form
func AdminSalvarPerfilHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.TipoUsuario != "ADMIN" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Importante: Aumentar limite de memória para upload
	r.ParseMultipartForm(10 << 20) // 10MB

	// 1. Coleta dados de texto
	nome := r.FormValue("nome")
	email := r.FormValue("email")
	novaSenha := r.FormValue("nova_senha")
	confSenha := r.FormValue("confirmar_senha")

	// 2. Validação de Senha (básica)
	if novaSenha != "" && novaSenha != confSenha {
		http.Redirect(w, r, "/admin/perfil?erro=senhas_nao_conferem", http.StatusSeeOther)
		return
	}

	// 3. UPLOAD DA FOTO (Lógica igual a do seu perfil de empresa)
	fotoFinal := user.FotoPerfil // Começa com a foto atual. Se não enviar nova, mantém essa.

	// ATENÇÃO: O nome aqui deve ser "foto_perfil" igual ao name="..." do seu HTML
	file, handler, err := r.FormFile("foto_perfil")

	if err == nil {
		defer file.Close()

		// Cria nome único: admin_ID_TIMESTAMP_NOMEORIGINAL
		nomeArquivo := fmt.Sprintf("admin_%d_%d_%s", user.Id, time.Now().Unix(), handler.Filename)

		caminhoDisco := "static/uploads/" + nomeArquivo  // Onde salva no PC
		caminhoBanco := "/static/uploads/" + nomeArquivo // O que salva no Banco (URL)

		// Garante que a pasta existe
		os.MkdirAll("static/uploads", os.ModePerm)

		// Cria o arquivo
		dst, errCreate := os.Create(caminhoDisco)
		if errCreate != nil {
			fmt.Println("Erro ao criar arquivo no disco:", errCreate)
		} else {
			defer dst.Close()
			// Copia os bytes do upload para o arquivo no disco
			_, errCopy := io.Copy(dst, file)
			if errCopy != nil {
				fmt.Println("Erro ao copiar arquivo:", errCopy)
			} else {
				// Só atualiza a variável se deu tudo certo
				fotoFinal = caminhoBanco
			}
		}
	}

	// 4. Chama o Model para atualizar no banco
	// Passamos 'novaSenha' em texto puro. O Model decide se gera hash (se preenchida) ou ignora (se vazia).
	err = models.AtualizarPerfilAdmin(user.Id, nome, email, novaSenha, fotoFinal)

	if err != nil {
		fmt.Println("Erro ao atualizar banco:", err)
		http.Redirect(w, r, "/admin/perfil?erro=banco", http.StatusSeeOther)
		return
	}

	// Sucesso
	http.Redirect(w, r, "/admin/perfil?sucesso=true", http.StatusSeeOther)
}
