package models

import (
	"nexhub/db"
	"nexhub/structs"
)

// Atualiza apenas os dados pertinentes à Empresa
func AtualizarPerfilEmpresa(u structs.Usuario) error {
	query := `
        UPDATE usuarios 
        SET nome_fantasia=$1, site_empresa=$2, ramo_atuacao=$3, 
            cidade=$4, biografia=$5, foto_perfil=$6, nome_completo=$7,
            email=$8, senha_hash=$9  -- ADICIONE ISSO
        WHERE id_usuario=$10
    `
	_, err := db.DB.Exec(query,
		u.NomeFantasia, u.SiteEmpresa, u.RamoAtuacao,
		u.Cidade, u.Biografia, u.FotoPerfil, u.NomeCompleto,
		u.Email, u.SenhaHash, // Passa o hash
		u.Id,
	)
	return err
}

// Busca estatísticas para o Dashboard da Empresa
func BuscarStatsEmpresa(idUsuario int) (structs.DashboardStats, error) {
	// Conta quantos favoritos do tipo 'PROJETO' e 'DEV' a empresa tem
	query := `
        SELECT 
            (SELECT COUNT(*) FROM favoritos WHERE id_usuario_quem_salvou = $1 AND tipo_item = 'PROJETO') as total_proj,
            (SELECT COUNT(*) FROM favoritos WHERE id_usuario_quem_salvou = $1 AND tipo_item = 'DEV') as total_devs
    `
	var stats structs.DashboardStats

	// Reutilizando a struct DashboardStats (TotalProjetos vira ProjetosSalvos, etc)
	err := db.DB.QueryRow(query, idUsuario).Scan(&stats.TotalProjetos, &stats.TotalVisualizacoes) // Usando TotalVisualizacoes como placeholder para DevsSalvos

	return stats, err
}

// AlternarFavorito: Se não existe, salva. Se existe, remove. (Like/Dislike)
func AlternarFavorito(idUsuario, idItem int, tipoItem string) (bool, error) {
	// 1. Verifica se já existe
	var existe bool
	queryCheck := `SELECT EXISTS(SELECT 1 FROM favoritos WHERE id_usuario_quem_salvou=$1 AND id_item_salvo=$2 AND tipo_item=$3)`
	err := db.DB.QueryRow(queryCheck, idUsuario, idItem, tipoItem).Scan(&existe)
	if err != nil {
		return false, err
	}

	if existe {
		// REMOVE (Unsave)
		_, err = db.DB.Exec(`DELETE FROM favoritos WHERE id_usuario_quem_salvou=$1 AND id_item_salvo=$2 AND tipo_item=$3`, idUsuario, idItem, tipoItem)
		return false, err // Retorna false indicando que "não está mais salvo"
	} else {
		// ADICIONA (Save)
		_, err = db.DB.Exec(`INSERT INTO favoritos (id_usuario_quem_salvou, id_item_salvo, tipo_item) VALUES ($1, $2, $3)`, idUsuario, idItem, tipoItem)
		return true, err // Retorna true indicando que "está salvo"
	}
}

// Verifica se o usuário já salvou aquele item (para pintar o botão)
func ChecarSeFavoritou(idUsuario, idItem int, tipoItem string) bool {
	var existe bool
	query := `SELECT EXISTS(SELECT 1 FROM favoritos WHERE id_usuario_quem_salvou=$1 AND id_item_salvo=$2 AND tipo_item=$3)`
	db.DB.QueryRow(query, idUsuario, idItem, tipoItem).Scan(&existe)
	return existe
}

func BuscarProjetosSalvos(idEmpresa int) ([]structs.Projeto, error) {
	query := `
		SELECT p.id_projeto, p.titulo, p.descricao, 
		       COALESCE(p.imagem_capa, ''), p.status_projeto, 
		       COALESCE(p.categoria, 'Geral'), p.cidade_projeto
		FROM favoritos f
		JOIN projetos p ON f.id_item_salvo = p.id_projeto
		WHERE f.id_usuario_quem_salvou = $1 AND f.tipo_item = 'PROJETO' AND p.status_projeto != 'Oculto'
		ORDER BY f.data_salvo DESC
	`
	rows, err := db.DB.Query(query, idEmpresa)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Projeto
	for rows.Next() {
		var p structs.Projeto
		if err := rows.Scan(&p.Id, &p.Titulo, &p.Descricao, &p.ImagemCapa, &p.Status, &p.Categoria, &p.Cidade); err == nil {
			lista = append(lista, p)
		}
	}
	return lista, nil
}

// BuscarTalentosSalvos: Retorna lista de devs que a empresa favoritou
func BuscarTalentosSalvos(idEmpresa int) ([]structs.Usuario, error) {
	query := `
		SELECT u.id_usuario, u.nome_completo, COALESCE(u.titulo_profissional, 'Dev'), 
		       COALESCE(u.foto_perfil, ''), u.cidade, u.disponivel_para_trabalho
		FROM favoritos f
		JOIN usuarios u ON f.id_item_salvo = u.id_usuario
		WHERE f.id_usuario_quem_salvou = $1 AND f.tipo_item = 'DEV' AND is_banned = FALSE
		ORDER BY f.data_salvo DESC
	`
	rows, err := db.DB.Query(query, idEmpresa)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Usuario
	for rows.Next() {
		var u structs.Usuario
		if err := rows.Scan(&u.Id, &u.NomeCompleto, &u.TituloProfissional, &u.FotoPerfil, &u.Cidade, &u.DisponivelParaEquipes); err == nil {
			lista = append(lista, u)
		}
	}
	return lista, nil
}
