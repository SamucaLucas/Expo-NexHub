package controllers

import (
	"encoding/json"
	"fmt"
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Busca as contagens reais no banco de dados usando as funções que acabamos de criar
	totalCursos := models.ObterTotalCursos()
	totalAlunos := models.ObterTotalAlunos()
	totalProjetos := models.ObterTotalProjetos()

	// Empacota os dados para enviar para o HTML
	dados := struct {
		TotalCursos   int
		TotalAlunos   int
		TotalProjetos int
	}{
		TotalCursos:   totalCursos,
		TotalAlunos:   totalAlunos,
		TotalProjetos: totalProjetos,
	}

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
	// Captura os filtros que vêm da URL quando o usuário pesquisa
	busca := r.URL.Query().Get("q")
	curso := r.URL.Query().Get("curso")
	status := r.URL.Query().Get("status")

	// Puxa do Model
	projetos, _ := models.ListarProjetosPublicos(busca, curso, status)

	// 🌟 NOVO: Busca a lista de cursos dinamicamente
	cursos, _ := models.ListarTodosCursos()

	// Monta a caixa de dados para o HTML
	dados := struct {
		Projetos    []structs.Projeto
		Cursos      []structs.Curso // Passamos a lista de cursos
		FiltroCurso string          // Guarda qual curso ele pesquisou
		FiltroBusca string          // Guarda o termo de busca
	}{
		Projetos:    projetos,
		Cursos:      cursos,
		FiltroCurso: curso,
		FiltroBusca: busca,
	}

	temp.ExecuteTemplate(w, "ProjetosPublicos", dados)
}

func DetalheProjetoHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	if id == 0 {
		http.Redirect(w, r, "/projetos", http.StatusSeeOther)
		return
	}

	// 1. Busca todos os dados do projeto (arquivos, links, equipe, galeria E AVALIAÇÕES)
	projeto, err := models.BuscarDetalhesProjeto(id)
	if err != nil {
		http.Redirect(w, r, "/projetos", http.StatusSeeOther)
		return
	}

	// 2. Monta a caixa de dados (Só precisamos passar o Projeto, pois as avaliações já estão dentro dele!)
	dados := struct {
		Projeto structs.Projeto
	}{
		Projeto: projeto,
	}

	temp.ExecuteTemplate(w, "DetalheProjeto", dados)
}

// ==========================================
// 3. VITRINE DE TALENTOS (ALUNOS)
// ==========================================

func TalentosHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Captura os filtros da URL
	busca := r.URL.Query().Get("q")
	curso := r.URL.Query().Get("curso")

	// 2. Busca os alunos filtrados
	alunos, _ := models.ListarTalentosPublicos(busca, curso)

	// 3. Busca a lista de cursos dinamicamente (reutilizando a função que criamos)
	cursos, _ := models.ListarTodosCursos()

	dados := struct {
		Alunos      []structs.Aluno
		Cursos      []structs.Curso
		FiltroCurso string
		FiltroBusca string
	}{
		Alunos:      alunos,
		Cursos:      cursos,
		FiltroCurso: curso,
		FiltroBusca: busca,
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
		nota, _ := strconv.Atoi(r.FormValue("nota"))

		nome := r.FormValue("nome_avaliador")
		email := r.FormValue("email_avaliador")
		comentario := r.FormValue("comentario")

		if nome != "" && email != "" && comentario != "" {

			// 🛡️ A NOSSA VALIDAÇÃO RIGOROSA ENTRA AQUI
			_, errEmail := models.ValidarEmailRigoroso(email)
			if errEmail != nil {
				fmt.Println("❌ BARRADO NA VALIDAÇÃO DE E-MAIL:", errEmail)
				// Redireciona de volta avisando que o e-mail falhou
				http.Redirect(w, r, "/projeto?id="+strconv.Itoa(idProjeto)+"&erro=email_invalido", http.StatusSeeOther)
				return
			}

			novaAvaliacao := structs.Avaliacao{ // ou models.Avaliacao
				IdProjeto:     idProjeto,
				NomeAvaliador: nome,
				Email:         email,
				Nota:          nota,
				Comentario:    comentario,
			}

			err := models.SalvarAvaliacao(novaAvaliacao)
			if err != nil {
				fmt.Println("❌ ERRO NO BANCO DE DADOS:", err)
			}
		}

		// Se deu tudo certo, redireciona com sucesso!
		http.Redirect(w, r, "/projeto?id="+strconv.Itoa(idProjeto)+"&sucesso=comentario", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func ValidarEmailAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	email := r.URL.Query().Get("email")

	if email == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"valido": false, "mensagem": "E-mail não pode estar vazio."})
		return
	}

	// Chama a SUA função rigorosa que está lá no models
	_, err := models.ValidarEmailRigoroso(email)
	if err != nil {
		// Se deu erro (e-mail falso, domínio inexistente), devolve a mensagem de erro!
		json.NewEncoder(w).Encode(map[string]interface{}{"valido": false, "mensagem": err.Error()})
		return
	}

	// Se passou direto, o e-mail é real!
	json.NewEncoder(w).Encode(map[string]interface{}{"valido": true})
}
