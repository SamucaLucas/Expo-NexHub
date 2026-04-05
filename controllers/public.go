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
