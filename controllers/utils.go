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
	Dados         interface{}     // Lista de Projetos ou Talentos

	// Filtros para manter o formulário preenchido
	FiltroTermo      string
	FiltroCat        string
	FiltroCidade     string
	FiltroDisponivel string
	FiltroNivel string
}

// GetUserFromSession verifica se existe um usuário logado na sessão.
// Se não houver, retorna um usuário vazio (Id=0), permitindo acesso anônimo.
func GetUserFromSession(r *http.Request) structs.Usuario {
	// 1. Pega a sessão
	session, _ := config.Store.Get(r, "nexhub-session")

	// 2. Verifica se tem ID salvo
	if userId, ok := session.Values["userId"].(int); ok {
		// 3. Busca os dados atualizados no banco
		usuario, err := models.BuscarUsuarioPorID(userId)
		if err == nil {
			return usuario
		}
	}

	// Retorna struct vazia se não estiver logado ou der erro
	return structs.Usuario{}
}
