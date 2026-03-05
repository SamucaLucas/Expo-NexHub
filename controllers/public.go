package controllers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"nexhub/config"
	"nexhub/models"
	"nexhub/structs"
	"strconv"
)

// Carrega TODOS os templates das subpastas (Public, Admin, Dev, etc)
// O template.Must garante que se houver erro no HTML, o sistema avisa na hora de abrir
var temp = template.Must(template.ParseGlob("templates/**/*.html"))

func DetalhesTalentoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Pega o ID da URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 2. Busca os dados do Talento (Dono do Perfil)
	talento, err := models.BuscarUsuarioPorID(id)
	if err != nil {
		http.Error(w, "Talento não encontrado", 404)
		return
	}

	// 3. Busca o Portfólio
	projetos, _ := models.BuscarPortfolioCompleto(id)

	// 4. Lógica de Quem está Visitando (Usuario Logado)
	var usuarioLogado structs.Usuario // Começa vazio (para visitantes não logados)
	role := "Visitante"
	estaSalvo := false

	session, _ := config.Store.Get(r, "nexhub-session")
	if userId, ok := session.Values["userId"].(int); ok {
		// Busca quem está logado
		usuarioEncontrado, err := models.BuscarUsuarioPorID(userId)
		if err == nil {
			usuarioLogado = usuarioEncontrado

			// Define o papel (Role) baseado em quem está logado
			if usuarioLogado.TipoUsuario == "EMPRESA" {
				role = "Empresa"
				estaSalvo = models.ChecarSeFavoritou(usuarioLogado.Id, talento.Id, "DEV")
			} else if usuarioLogado.TipoUsuario == "DEV" {
				role = "Dev"
			} else if usuarioLogado.TipoUsuario == "ADMIN" {
				role = "Admin"
			}
		}
	}

	// 5. Monta o pacote de dados SEPARANDO Visitante de Talento
	dados := struct {
		Usuario  structs.Usuario // Quem vê a página (para Navbar e Notificações)
		Talento  structs.Usuario // Quem é dono do perfil (Conteúdo principal)
		Projetos []structs.Projeto
		Role     string
		Salvou   bool
	}{
		Usuario:  usuarioLogado, // Passa o usuário logado aqui!
		Talento:  talento,       // Passa o dono do perfil aqui!
		Projetos: projetos,
		Role:     role,
		Salvou:   estaSalvo,
	}

	// 6. Renderiza
	err = temp.ExecuteTemplate(w, "DetalheTalento", dados)
	if err != nil {
		// Loga o erro no terminal para você saber o motivo da tela branca
		fmt.Println("❌ Erro ao renderizar DetalheTalento:", err)
	}
}

// 1. INDEX (HOME)
func Index(w http.ResponseWriter, r *http.Request) {
	// Verifica sessão (sem travar se for anônimo)
	user := GetUserFromSession(r)

	dados := PageData2{
		Usuario: user,
	}

	err := temp.ExecuteTemplate(w, "Index", dados)
	if err != nil {
		log.Println("Erro ao renderizar a Home:", err)
	}
}

// 2. PÁGINA SOBRE
func Sobre(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromSession(r)

	dados := PageData2{
		Usuario: user,
	}

	err := temp.ExecuteTemplate(w, "Sobre", dados)
	if err != nil {
		log.Println("Erro ao renderizar Sobre:", err)
	}
}

// 3. VITRINE DE PROJETOS (Com Filtros e Sessão)
func ProjetosVitrineHandler(w http.ResponseWriter, r *http.Request) {
	// Captura Filtros
	termo := r.URL.Query().Get("q")
	categoria := r.URL.Query().Get("categoria")
	cidade := r.URL.Query().Get("cidade")

	// Busca Dados
	lista, err := models.BuscarProjetosComFiltro(termo, categoria, cidade)
	if err != nil {
		log.Println("Erro ao buscar projetos filtrados:", err)
	}

	// Pega Usuário Logado
	user := GetUserFromSession(r)

	// Monta PageData
	dados := PageData2{
		Usuario:      user,
		Dados:        lista, // A lista de projetos vai aqui
		FiltroTermo:  termo,
		FiltroCat:    categoria,
		FiltroCidade: cidade,
	}

	temp.ExecuteTemplate(w, "Projetos", dados)
}

// 4. VITRINE DE TALENTOS (Com Filtros e Sessão)
func TalentosVitrineHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Pega Usuário Logado PRIMEIRO
	// (Precisamos do user.Id AGORA para a query saber o que está favoritado)
	user := GetUserFromSession(r)

	// 2. Captura Filtros
	termo := r.URL.Query().Get("q")
	disponivel := r.URL.Query().Get("disponivel")
	nivel := r.URL.Query().Get("nivel") // <--- Captura o nível da URL

	// 3. Busca Dados (Passando todos os 4 argumentos na ordem correta do Model)
	lista, err := models.BuscarTalentosComFiltro(termo, nivel, disponivel, user.Id)
	if err != nil {
		log.Println("Erro ao buscar talentos:", err)
	}

	// 4. Monta PageData
	// Certifique-se que sua struct PageData2 tem o campo FiltroNivel
	dados := struct {
		Usuario          structs.Usuario
		Dados            []structs.Usuario
		FiltroTermo      string
		FiltroDisponivel string
		FiltroNivel      string // <--- Adicionado para manter o select marcado
	}{
		Usuario:          user,
		Dados:            lista,
		FiltroTermo:      termo,
		FiltroDisponivel: disponivel,
		FiltroNivel:      nivel,
	}

	temp.ExecuteTemplate(w, "Talentos", dados)
}

// Estruturas de Dados Mock (Futuramente virão do Banco de Dados)
type ProjetoData struct {
	Titulo      string
	Descricao   string
	Capa        string
	Categoria   string
	Status      string
	AutorNome   string
	AutorAvatar string
	Techs       []string
	Imagens     []string
}

type TalentoData struct {
	Nome   string
	Titulo string
	Bio    string
	Cidade string
	Avatar string
	Skills []string
}
