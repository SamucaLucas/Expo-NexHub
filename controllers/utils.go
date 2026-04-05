package controllers

import (
	"net/http"
	"nexhub/config"
	"nexhub/models"
	"nexhub/structs"
)

// PageData é a estrutura "coringa" para passar dados para os templates públicos
type PageData2 struct {
	Usuario structs.Usuario // Dados do usuário (se logado)
	Dados   interface{}     // Lista de Projetos ou Talentos

	// Filtros para manter o formulário preenchido
	FiltroTermo      string
	FiltroCat        string
	FiltroCidade     string
	FiltroDisponivel string
	FiltroNivel      string
}

// GetUserFromSession verifica se existe um usuário logado na sessão.
// Se não houver, retorna um usuário vazio (Id=0), permitindo acesso anônimo.
func GetUserFromSession(r *http.Request) structs.Usuario {
	session, _ := config.Store.Get(r, "nexhub-session")

	// Verifica se existe um ID de usuário na sessão
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		return structs.Usuario{} // Retorna vazio, o que causa o redirecionamento pro index
	}

	// Verifica se o tipo da conta logada é ADMIN
	userTipo, ok := session.Values["user_tipo"].(string)
	if !ok || userTipo != "ADMIN" {
		return structs.Usuario{}
	}

	// Busca os dados reais do banco
	usuario, err := models.BuscarUsuarioPorID(userID)
	if err != nil {
		return structs.Usuario{}
	}

	return usuario
}
