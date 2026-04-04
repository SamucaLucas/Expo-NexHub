package controllers

import (
	"html/template"
	"net/http"
	"nexhub/models"
	"nexhub/structs"
	"strconv"
)

var temp = template.Must(template.ParseGlob("templates/**/*.html"))

// ==========================================
// 1. PÁGINAS PRINCIPAIS (VITRINE)
// ==========================================

// IndexHandler renderiza a página inicial (Home)
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Futuramente, você pode criar uma função no model para buscar apenas os 3 últimos projetos
	// projetos, _ := models.ListarUltimosProjetos(3)
	// talentos, _ := models.ListarUltimosAlunos(4)

	dados := struct {
		// Projetos []structs.Projeto
		// Talentos []structs.Aluno
	}{}

	temp.ExecuteTemplate(w, "Index", dados)
}

// SobreHandler renderiza a página "Sobre a Plataforma" e lista os Analistas
func SobreHandler(w http.ResponseWriter, r *http.Request) {
	analistas, _ := models.ListarAnalistasParaSobre()

	dados := struct {
		Analistas []models.AnalistaCard
	}{
		Analistas: analistas,
	}

	temp.ExecuteTemplate(w, "Sobre", dados)
}

// ==========================================
// 2. VITRINE DE PROJETOS MULTIDISCIPLINARES
// ==========================================

func ProjetosHandler(w http.ResponseWriter, r *http.Request) {
	// Aqui você pode capturar filtros da URL (ex: ?curso=direito&status=concluido)
	// busca := r.URL.Query().Get("q")
	// idCursoStr := r.URL.Query().Get("curso")

	// projetos, _ := models.ListarProjetosPublicos(busca, idCursoStr)

	dados := struct {
		// Projetos []structs.Projeto
	}{}

	temp.ExecuteTemplate(w, "ProjetosPublicos", dados)
}

func DetalheProjetoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	if id == 0 {
		http.Redirect(w, r, "/projetos", http.StatusSeeOther)
		return
	}

	// 1. Busca todos os dados do projeto (arquivos, links, equipe e galeria)
	projeto, err := models.BuscarDetalhesProjeto(id)
	if err != nil {
		http.Redirect(w, r, "/projetos", http.StatusSeeOther)
		return
	}

	// 2. Buscar avaliações/comentários de visitantes (se você for reimplementar o models.Avaliacao)
	// avaliacoes, _ := models.BuscarAvaliacoesDoProjeto(id)

	dados := struct {
		Projeto structs.Projeto
		// Avaliacoes []structs.Avaliacao
	}{
		Projeto: projeto,
	}

	temp.ExecuteTemplate(w, "DetalheProjeto", dados)
}

// ==========================================
// 3. VITRINE DE TALENTOS (ALUNOS)
// ==========================================

func TalentosHandler(w http.ResponseWriter, r *http.Request) {
	// Busca todos os alunos cadastrados pelos Admins
	alunos, _ := models.ListarAlunos()

	dados := struct {
		Alunos []structs.Aluno
	}{
		Alunos: alunos,
	}

	temp.ExecuteTemplate(w, "TalentosPublicos", dados)
}

func DetalheTalentoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	if id == 0 {
		http.Redirect(w, r, "/talentos", http.StatusSeeOther)
		return
	}

	// 1. Busca os dados públicos do Aluno
	aluno, err := models.BuscarAlunoPorID(id)
	if err != nil {
		http.Redirect(w, r, "/talentos", http.StatusSeeOther)
		return
	}

	// 2. Busca o portfólio de projetos que o aluno participou
	projetos, _ := models.BuscarProjetosDoAluno(id)

	dados := struct {
		Aluno    structs.Aluno
		Projetos []structs.Projeto
	}{
		Aluno:    aluno,
		Projetos: projetos,
	}

	temp.ExecuteTemplate(w, "DetalheTalento", dados)
}

// ==========================================
// 4. INTERAÇÃO PÚBLICA (COMENTÁRIOS SEM LOGIN)
// ==========================================

func SalvarAvaliacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		idProjeto, _ := strconv.Atoi(r.FormValue("id_projeto"))
		nome := r.FormValue("nome")
		email := r.FormValue("email")
		comentario := r.FormValue("comentario")
		// nota, _ := strconv.Atoi(r.FormValue("nota"))

		// Como o visitante não tem login, ele preenche nome e e-mail no form.
		// Futuramente: implementar o envio de um token pro e-mail para validar antes de exibir.
		if nome != "" && email != "" && comentario != "" {
			// models.SalvarAvaliacao(idProjeto, nome, email, nota, comentario)
		}

		http.Redirect(w, r, "/projeto?id="+strconv.Itoa(idProjeto)+"&sucesso=comentario", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

