package controllers

import (
	"fmt"
	"io"
	"net/http"
	"nexhub/models"
	"nexhub/structs"
	"os"
	"strconv"
	"time"
)

// ==========================================
// 1. DASHBOARD E PERFIL (ADMIN)
// ==========================================

func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Aqui futuramente você pode chamar models.BuscarEstatisticasAdmin()
	// adaptado para contar Total de Alunos, Total de Projetos, etc.
	dados := struct {
		Usuario structs.Usuario
	}{
		Usuario: user,
	}

	temp.ExecuteTemplate(w, "AdminDashboard", dados)
}

func AdminPerfilHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	dados := struct {
		Usuario structs.Usuario
	}{
		Usuario: user,
	}
	temp.ExecuteTemplate(w, "AdminPerfil", dados)
}

func AdminSalvarPerfilHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	nome := r.FormValue("nome")
	email := r.FormValue("email")
	novaSenha := r.FormValue("nova_senha")
	confSenha := r.FormValue("confirmar_senha")

	if novaSenha != "" && novaSenha != confSenha {
		http.Redirect(w, r, "/admin/perfil?erro=senhas_nao_conferem", http.StatusSeeOther)
		return
	}

	// O Admin da V2 não tem mais foto_perfil na tabela, então removemos a lógica de upload daqui.
	// Atualizamos apenas dados básicos.
	user.NomeCompleto = nome
	user.Email = email
	if novaSenha != "" {
		user.SenhaHash = novaSenha // O model deve hashear isso (ou você criar uma func específica)
	}

	err := models.AtualizarPerfil(user)
	if err != nil {
		http.Redirect(w, r, "/admin/perfil?erro=banco", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/perfil?sucesso=true", http.StatusSeeOther)
}

// ==========================================
// 2. GESTÃO DE ALUNOS (VITRINE)
// ==========================================

func AdminAlunosHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	listaAlunos, _ := models.ListarAlunos()

	dados := struct {
		Usuario structs.Usuario
		Alunos  []structs.Aluno
	}{
		Usuario: user,
		Alunos:  listaAlunos,
	}

	temp.ExecuteTemplate(w, "AdminUsuarios", dados) // Pode renomear o template HTML depois para AdminAlunos
}

func AdminSalvarAlunoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10MB limite

	idCurso, _ := strconv.Atoi(r.FormValue("id_curso"))
	semestre, _ := strconv.Atoi(r.FormValue("semestre_atual"))

	aluno := structs.Aluno{
		NomeCompleto:  r.FormValue("nome_completo"),
		IdCurso:       &idCurso,
		SemestreAtual: &semestre,
		Biografia:     r.FormValue("biografia"),
		EmailContato:  r.FormValue("email_contato"),
		LinkedinLink:  r.FormValue("linkedin_link"),
		GithubLink:    r.FormValue("github_link"),
		PortfolioLink: r.FormValue("portfolio_link"),
		CadastradoPor: &user.IdUsuario,
	}

	// Lógica de Upload da Foto do Aluno
	file, handler, err := r.FormFile("foto_perfil")
	if err == nil {
		defer file.Close()
		nomeArquivo := fmt.Sprintf("aluno_%d_%s", time.Now().Unix(), handler.Filename)
		caminhoDisco := "static/uploads/" + nomeArquivo
		caminhoBanco := "/static/uploads/" + nomeArquivo

		os.MkdirAll("static/uploads", os.ModePerm)
		dst, errCreate := os.Create(caminhoDisco)
		if errCreate == nil {
			defer dst.Close()
			if _, errCopy := io.Copy(dst, file); errCopy == nil {
				aluno.FotoPerfil = caminhoBanco
			}
		}
	}

	_, err = models.CriarAluno(aluno)
	if err != nil {
		fmt.Println("Erro ao cadastrar aluno:", err)
	}

	http.Redirect(w, r, "/admin/alunos", http.StatusSeeOther)
}

func AdminExcluirAlunoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id > 0 {
		models.DeletarAluno(id)
	}
	http.Redirect(w, r, "/admin/alunos", http.StatusSeeOther)
}

// ==========================================
// 3. GESTÃO DE PROJETOS
// ==========================================

func AdminProjetosHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Busca projetos (Adapte para listar todos da nova estrutura)
	// listaProjetos, _ := models.ListarTodosProjetos("", "")

	dados := struct {
		Usuario  structs.Usuario
		Projetos []structs.Projeto
	}{
		Usuario: user,
		// Projetos: listaProjetos,
	}

	temp.ExecuteTemplate(w, "AdminProjetos", dados)
}

func AdminSalvarProjetoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	r.ParseMultipartForm(20 << 20) // 20MB para suportar os PDFs

	idCurso, _ := strconv.Atoi(r.FormValue("id_curso"))
	idArea, _ := strconv.Atoi(r.FormValue("id_area"))

	projeto := structs.Projeto{
		Titulo:              r.FormValue("titulo"),
		Descricao:           r.FormValue("descricao"),
		IdCurso:             &idCurso,
		IdArea:              &idArea,
		SemestreLetivo:      r.FormValue("semestre_letivo"),
		ProfessorOrientador: r.FormValue("professor_orientador"),
		StatusProjeto:       r.FormValue("status_projeto"),
		LinkRepositorio:     r.FormValue("link_repositorio"),
		CadastradoPor:       &user.IdUsuario,
	}

	// Lógica de Upload da Capa do Projeto
	file, handler, err := r.FormFile("imagem_capa")
	if err == nil {
		defer file.Close()
		nomeArquivo := fmt.Sprintf("proj_%d_%s", time.Now().Unix(), handler.Filename)
		caminhoDisco := "static/uploads/" + nomeArquivo
		caminhoBanco := "/static/uploads/" + nomeArquivo

		os.MkdirAll("static/uploads", os.ModePerm)
		dst, errCreate := os.Create(caminhoDisco)
		if errCreate == nil {
			defer dst.Close()
			if _, errCopy := io.Copy(dst, file); errCopy == nil {
				projeto.ImagemCapa = caminhoBanco
			}
		}
	}

	idProjetoGerado, err := models.CriarProjeto(projeto)

	// Se salvou o projeto com sucesso, aqui você processaria os links extras e PDFs
	if err == nil && idProjetoGerado > 0 {
		// Exemplo: Salvar Link do YouTube
		urlYoutube := r.FormValue("url_youtube")
		if urlYoutube != "" {
			models.AdicionarLinkProjeto(structs.ProjetoLink{
				IdProjeto: idProjetoGerado,
				TipoLink:  "YOUTUBE",
				Url:       urlYoutube,
			})
		}
	} else {
		fmt.Println("Erro ao cadastrar projeto:", err)
	}

	http.Redirect(w, r, "/admin/projetos", http.StatusSeeOther)
}

func AdminAlterarStatusProjetoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	acao := r.URL.Query().Get("acao")

	if id > 0 {
		if acao == "ocultar" {
			// models.AtualizarStatusProjeto(id, "OCULTO")
		} else if acao == "aprovar" {
			// models.AtualizarStatusProjeto(id, "EM_ANDAMENTO")
		}
	}
	http.Redirect(w, r, "/admin/projetos", http.StatusSeeOther)
}

func AdminExcluirProjetoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id > 0 {
		models.DeletarProjeto(id)
	}
	http.Redirect(w, r, "/admin/projetos", http.StatusSeeOther)
}
