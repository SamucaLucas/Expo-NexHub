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

	"golang.org/x/crypto/bcrypt"
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

	r.ParseMultipartForm(10 << 20) // 10MB limite para a foto

	nome := r.FormValue("nome")
	email := r.FormValue("email")
	novaSenha := r.FormValue("nova_senha")
	confSenha := r.FormValue("confirmar_senha")

	if novaSenha != "" && novaSenha != confSenha {
		http.Redirect(w, r, "/admin/perfil?erro=senhas_nao_conferem", http.StatusSeeOther)
		return
	}

	// -----------------------------------------------------
	// UPLOAD DA FOTO DO ADMIN (Pasta: /uploads/admins/)
	// -----------------------------------------------------
	file, handler, err := r.FormFile("foto_perfil")
	if err == nil {
		defer file.Close()
		nomeArquivo := fmt.Sprintf("admin_%d_%s", time.Now().Unix(), handler.Filename)
		pastaDestino := "static/uploads/admins"

		os.MkdirAll(pastaDestino, os.ModePerm)

		caminhoDisco := pastaDestino + "/" + nomeArquivo
		caminhoBanco := "/" + pastaDestino + "/" + nomeArquivo

		dst, errCreate := os.Create(caminhoDisco)
		if errCreate == nil {
			defer dst.Close()
			if _, errCopy := io.Copy(dst, file); errCopy == nil {
				user.FotoPerfil = caminhoBanco
			}
		}
	}

	user.NomeCompleto = nome
	user.Email = email

	// --- CORREÇÃO DA SENHA AQUI ---
	// Se a pessoa digitou uma nova senha, a gente criptografa.
	// Se não digitou (""), o user.SenhaHash continua intacto com a senha antiga!
	if novaSenha != "" {
		hash, errHash := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
		if errHash == nil {
			user.SenhaHash = string(hash)
		}
	}

	err = models.AtualizarPerfil(user)
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
	cursos, _ := models.ListarTodosCursos() // Busca a lista de cursos para o formulário

	dados := struct {
		Usuario structs.Usuario
		Alunos  []structs.Aluno
		Cursos  []structs.Curso // Adicionamos os cursos aqui
	}{
		Usuario: user,
		Alunos:  listaAlunos,
		Cursos:  cursos,
	}

	temp.ExecuteTemplate(w, "AdminAlunos", dados)
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

	// -----------------------------------------------------
	// UPLOAD DA FOTO DO ALUNO (Pasta: /uploads/alunos/)
	// -----------------------------------------------------
	file, handler, err := r.FormFile("foto_perfil")
	if err == nil {
		defer file.Close()
		nomeArquivo := fmt.Sprintf("aluno_%d_%s", time.Now().Unix(), handler.Filename)
		pastaDestino := "static/uploads/alunos" // Nova pasta

		os.MkdirAll(pastaDestino, os.ModePerm)

		caminhoDisco := pastaDestino + "/" + nomeArquivo
		caminhoBanco := "/" + pastaDestino + "/" + nomeArquivo

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

	listaProjetos, _ := models.ListarTodosProjetosAdmin()
	cursos, _ := models.ListarTodosCursos() // Busca os cursos para o Modal
	areas, _ := models.ListarTodasAreas()

	dados := struct {
		Usuario  structs.Usuario
		Projetos []structs.Projeto
		Cursos   []structs.Curso
		Areas    []structs.Area
	}{
		Usuario:  user,
		Projetos: listaProjetos,
		Cursos:   cursos,
		Areas:    areas,
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

	var ptrArea *int
	if idArea > 0 {
		ptrArea = &idArea
	}

	projeto := structs.Projeto{
		Titulo:              r.FormValue("titulo"),
		Descricao:           r.FormValue("descricao"),
		IdCurso:             &idCurso,
		IdArea:              ptrArea,
		SemestreLetivo:      r.FormValue("semestre_letivo"),
		ProfessorOrientador: r.FormValue("professor_orientador"),
		StatusProjeto:       r.FormValue("status_projeto"),
		LinkRepositorio:     r.FormValue("link_repositorio"),
		CadastradoPor:       &user.IdUsuario,
	}

	// 1. SALVA O PROJETO NO BANCO PRIMEIRO (para gerar o ID)
	idProjetoGerado, err := models.CriarProjeto(projeto)

	if err == nil && idProjetoGerado > 0 {
		// -----------------------------------------------------
		// UPLOAD DA CAPA DO PROJETO (Pasta: /uploads/projetos/projeto_{ID}/)
		// -----------------------------------------------------
		file, handler, errFile := r.FormFile("imagem_capa")
		if errFile == nil {
			defer file.Close()
			nomeArquivo := fmt.Sprintf("capa_%d_%s", time.Now().Unix(), handler.Filename)

			// Cria a pasta ESPECÍFICA deste projeto!
			pastaDestino := fmt.Sprintf("static/uploads/projetos/projeto_%d", idProjetoGerado)
			os.MkdirAll(pastaDestino, os.ModePerm)

			caminhoDisco := pastaDestino + "/" + nomeArquivo
			caminhoBanco := "/" + pastaDestino + "/" + nomeArquivo

			dst, errCreate := os.Create(caminhoDisco)
			if errCreate == nil {
				defer dst.Close()
				if _, errCopy := io.Copy(dst, file); errCopy == nil {
					// 2. ATUALIZA A CAPA NO BANCO
					models.AtualizarCapaProjeto(idProjetoGerado, caminhoBanco)
				}
			}
		}

		// (No futuro, aqui vai a lógica similar para percorrer os PDFs daquele projeto e salvar na mesma pasta)

		// Salvar Link do YouTube
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

func AdminEditarProjetoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idProjeto, _ := strconv.Atoi(r.URL.Query().Get("id"))

	projeto, _ := models.BuscarProjetoCompletoPorId(idProjeto)
	alunos, _ := models.ListarAlunos()
	cursos, _ := models.ListarTodosCursos()
	areas, _ := models.ListarTodasAreas()

	// --- RESOLVENDO O PROBLEMA DOS PONTEIROS AQUI ---
	projetoIdCurso := 0
	if projeto.IdCurso != nil {
		projetoIdCurso = *projeto.IdCurso
	}

	projetoIdArea := 0
	if projeto.IdArea != nil {
		projetoIdArea = *projeto.IdArea
	}
	// ------------------------------------------------

	dados := struct {
		Usuario        structs.Usuario
		Projeto        structs.Projeto
		Alunos         []structs.Aluno
		Cursos         []structs.Curso
		Areas          []structs.Area
		ProjetoIdCurso int // Novo campo para o HTML
		ProjetoIdArea  int // Novo campo para o HTML
	}{
		Usuario:        user,
		Projeto:        projeto,
		Alunos:         alunos,
		Cursos:         cursos,
		Areas:          areas,
		ProjetoIdCurso: projetoIdCurso,
		ProjetoIdArea:  projetoIdArea,
	}

	temp.ExecuteTemplate(w, "AdminProjetoEdit", dados)
}

// 1. Atualizar Dados Básicos do Projeto
func AdminAtualizarProjetoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		return
	}

	// Aumentando o limite de memória para suportar imagens
	r.ParseMultipartForm(10 << 20)

	idProjeto, _ := strconv.Atoi(r.FormValue("id_projeto"))
	idCurso, _ := strconv.Atoi(r.FormValue("id_curso"))
	idArea, _ := strconv.Atoi(r.FormValue("id_area"))

	var ptrCurso, ptrArea *int
	if idCurso > 0 {
		ptrCurso = &idCurso
	}
	if idArea > 0 {
		ptrArea = &idArea
	}

	projeto := structs.Projeto{
		IdProjeto:           idProjeto,
		Titulo:              r.FormValue("titulo"),
		Descricao:           r.FormValue("descricao"),
		IdCurso:             ptrCurso,
		IdArea:              ptrArea,
		SemestreLetivo:      r.FormValue("semestre_letivo"),
		ProfessorOrientador: r.FormValue("professor_orientador"),
		StatusProjeto:       r.FormValue("status_projeto"),
		LinkRepositorio:     r.FormValue("link_repositorio"),
	}

	// 1. Atualiza as informações de texto
	models.AtualizarProjeto(projeto)

	// 2. Verifica se o usuário enviou uma NOVA foto de capa
	file, handler, err := r.FormFile("imagem_capa")
	if err == nil {
		defer file.Close()

		pastaDestino := fmt.Sprintf("static/uploads/projetos/projeto_%d", idProjeto)
		os.MkdirAll(pastaDestino, os.ModePerm)

		nomeArquivo := fmt.Sprintf("capa_%d_%s", time.Now().Unix(), handler.Filename)
		caminhoDisco := pastaDestino + "/" + nomeArquivo
		caminhoBanco := "/" + caminhoDisco

		dst, errCreate := os.Create(caminhoDisco)
		if errCreate == nil {
			defer dst.Close()
			if _, errCopy := io.Copy(dst, file); errCopy == nil {
				// Atualiza o caminho da imagem no banco de dados
				models.AtualizarCapaProjeto(idProjeto, caminhoBanco)
			}
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/projetos/editar?id=%d&sucesso=base", idProjeto), http.StatusSeeOther)
}

// 2. Adicionar/Remover Membros da Equipe
func AdminProjetoAdicionarEquipeHandler(w http.ResponseWriter, r *http.Request) {
	idProjeto, _ := strconv.Atoi(r.FormValue("id_projeto"))
	idAluno, _ := strconv.Atoi(r.FormValue("id_aluno"))
	funcao := r.FormValue("funcao")

	if idProjeto > 0 && idAluno > 0 {
		models.AdicionarMembroEquipe(idProjeto, idAluno, funcao)
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/projetos/editar?id=%d", idProjeto), http.StatusSeeOther)
}

func AdminProjetoRemoverEquipeHandler(w http.ResponseWriter, r *http.Request) {
	idProjeto, _ := strconv.Atoi(r.URL.Query().Get("id_projeto"))
	idAluno, _ := strconv.Atoi(r.URL.Query().Get("id_aluno"))

	if idProjeto > 0 && idAluno > 0 {
		models.RemoverMembroEquipe(idProjeto, idAluno)
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/projetos/editar?id=%d", idProjeto), http.StatusSeeOther)
}

// 3. Adicionar/Remover Links
func AdminProjetoAdicionarLinkHandler(w http.ResponseWriter, r *http.Request) {
	idProjeto, _ := strconv.Atoi(r.FormValue("id_projeto"))
	link := structs.ProjetoLink{
		IdProjeto: idProjeto,
		TipoLink:  r.FormValue("tipo_link"),
		Url:       r.FormValue("url"),
	}

	if idProjeto > 0 && link.Url != "" {
		models.AdicionarLinkProjeto(link)
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/projetos/editar?id=%d", idProjeto), http.StatusSeeOther)
}

func AdminProjetoRemoverLinkHandler(w http.ResponseWriter, r *http.Request) {
	idLink, _ := strconv.Atoi(r.URL.Query().Get("id_link"))
	idProjeto := r.URL.Query().Get("id_projeto")

	if idLink > 0 {
		models.RemoverLinkProjeto(idLink)
	}
	http.Redirect(w, r, "/admin/projetos/editar?id="+idProjeto, http.StatusSeeOther)
}

// 4. Upload de Arquivos PDF
func AdminProjetoUploadArquivoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(30 << 20) // Limite de 30MB para PDFs

	idProjeto, _ := strconv.Atoi(r.FormValue("id_projeto"))
	if idProjeto == 0 {
		return
	}

	file, handler, err := r.FormFile("arquivo_pdf")
	if err == nil {
		defer file.Close()

		pastaDestino := fmt.Sprintf("static/uploads/projetos/projeto_%d", idProjeto)
		os.MkdirAll(pastaDestino, os.ModePerm)

		nomeArquivoSeguro := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
		caminhoDisco := pastaDestino + "/" + nomeArquivoSeguro
		caminhoBanco := "/" + caminhoDisco

		dst, errCreate := os.Create(caminhoDisco)
		if errCreate == nil {
			defer dst.Close()
			io.Copy(dst, file)
			// Salva no Banco
			models.SalvarArquivoProjeto(idProjeto, handler.Filename, caminhoBanco)
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/admin/projetos/editar?id=%d", idProjeto), http.StatusSeeOther)
}

func AdminProjetoRemoverArquivoHandler(w http.ResponseWriter, r *http.Request) {
	idArquivo, _ := strconv.Atoi(r.URL.Query().Get("id_arquivo"))
	idProjeto := r.URL.Query().Get("id_projeto")

	if idArquivo > 0 {
		models.RemoverArquivoProjeto(idArquivo)
		// Nota: Para ser 100% limpo, você pode usar os.Remove() aqui para apagar do HD também!
	}
	http.Redirect(w, r, "/admin/projetos/editar?id="+idProjeto, http.StatusSeeOther)
}

// ==========================================
// SUPER ADMIN: GESTÃO DE ANALISTAS
// ==========================================

func AdminAnalistasHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// SEGURANÇA: Se tiver um curso vinculado, é Especialista, logo NÃO É Admin Geral.
	if user.IdCursoAnalista != nil {
		http.Redirect(w, r, "/admin/dashboard?erro=AcessoNegado", http.StatusSeeOther)
		return
	}

	analistas, _ := models.ListarTodosAnalistas()

	dados := struct {
		Usuario   structs.Usuario
		Analistas []models.AnalistaAdmin
	}{
		Usuario:   user,
		Analistas: analistas,
	}

	temp.ExecuteTemplate(w, "AdminUsuarios", dados)
}

func AdminExcluirAnalistaHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	// Bloqueia se não estiver logado ou se não for o Admin Geral
	if user.IdUsuario == 0 || user.IdCursoAnalista != nil {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	// Proteção: O admin não pode excluir a si mesmo
	if id > 0 && id != user.IdUsuario {
		models.DeletarAnalista(id)
	}
	http.Redirect(w, r, "/admin/analistas", http.StatusSeeOther)
}

// --- TELA DE EDITAR ALUNO ---
func AdminEditarAlunoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idAluno, _ := strconv.Atoi(r.URL.Query().Get("id"))
	aluno, err := models.BuscarAlunoPorID(idAluno)
	if err != nil {
		http.Redirect(w, r, "/admin/alunos", http.StatusSeeOther)
		return
	}

	cursos, _ := models.ListarTodosCursos()

	// Tratando os ponteiros para o HTML não quebrar
	alunoIdCurso := 0
	if aluno.IdCurso != nil {
		alunoIdCurso = *aluno.IdCurso
	}
	alunoSemestre := 0
	if aluno.SemestreAtual != nil {
		alunoSemestre = *aluno.SemestreAtual
	}

	dados := struct {
		Usuario       structs.Usuario
		Aluno         structs.Aluno
		Cursos        []structs.Curso
		AlunoIdCurso  int
		AlunoSemestre int
	}{
		Usuario:       user,
		Aluno:         aluno,
		Cursos:        cursos,
		AlunoIdCurso:  alunoIdCurso,
		AlunoSemestre: alunoSemestre,
	}

	temp.ExecuteTemplate(w, "AdminAlunoEdit", dados)
}

// --- SALVAR EDIÇÃO DO ALUNO ---
func AdminAtualizarAlunoHandler(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)
	if user.IdUsuario == 0 {
		return
	}

	r.ParseMultipartForm(10 << 20) // Limite 10MB para foto

	idAluno, _ := strconv.Atoi(r.FormValue("id_aluno"))
	idCurso, _ := strconv.Atoi(r.FormValue("id_curso"))
	semestre, _ := strconv.Atoi(r.FormValue("semestre_atual"))

	var ptrCurso, ptrSemestre *int
	if idCurso > 0 {
		ptrCurso = &idCurso
	}
	if semestre > 0 {
		ptrSemestre = &semestre
	}

	aluno := structs.Aluno{
		IdAluno:       idAluno,
		NomeCompleto:  r.FormValue("nome_completo"),
		IdCurso:       ptrCurso,
		SemestreAtual: ptrSemestre,
		Biografia:     r.FormValue("biografia"),
		EmailContato:  r.FormValue("email_contato"),
		LinkedinLink:  r.FormValue("linkedin_link"),
		GithubLink:    r.FormValue("github_link"),
		PortfolioLink: r.FormValue("portfolio_link"),
	}

	// Salva as informações de texto
	models.AtualizarAluno(aluno)

	// Verifica se enviou uma foto de perfil nova
	file, handler, err := r.FormFile("foto_perfil")
	if err == nil {
		defer file.Close()

		pastaDestino := "static/uploads/alunos"
		os.MkdirAll(pastaDestino, os.ModePerm)

		nomeArquivo := fmt.Sprintf("aluno_%d_%s", time.Now().Unix(), handler.Filename)
		caminhoDisco := pastaDestino + "/" + nomeArquivo
		caminhoBanco := "/" + caminhoDisco

		dst, errCreate := os.Create(caminhoDisco)
		if errCreate == nil {
			defer dst.Close()
			if _, errCopy := io.Copy(dst, file); errCopy == nil {
				models.AtualizarFotoAluno(idAluno, caminhoBanco)
			}
		}
	}

	http.Redirect(w, r, "/admin/alunos?sucesso=editado", http.StatusSeeOther)
}
