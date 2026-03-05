package models

import (
	"fmt"
	"nexhub/db"
	"nexhub/structs"
	"strings"
)

func BuscarStatsUsuario(idUsuario int) (structs.DashboardStats, error) {
	// Essa query faz duas coisas ao mesmo tempo:
	// 1. Conta quantos projetos você tem (COUNT)
	// 2. Soma as visualizações de todos eles (SUM)
	query := `
        SELECT 
            COUNT(*) as total_projetos,
            COALESCE(SUM(visualizacoes), 0) as total_views
        FROM projetos 
        WHERE id_lider = $1
    `

	var stats structs.DashboardStats

	// O Scan preenche as variáveis. Mensagens deixamos fixo em 0 por enquanto.
	err := db.DB.QueryRow(query, idUsuario).Scan(
		&stats.TotalProjetos,
		&stats.TotalVisualizacoes,
	)

	stats.MensagensNaoLidas = 0 // Placeholder até criarmos o Chat

	return stats, err
}

// Agora aceita 'idUsuarioLogado' para checar o favorito
func BuscarTodosProjetos(termo, categoria, cidade string, idUsuarioLogado int) ([]structs.Projeto, error) {
	query := `
        SELECT 
        p.id_projeto, 
        p.titulo, 
        p.descricao, 
        COALESCE(array_to_string(p.tecnologias, ','), ''), 
        p.status_projeto,   -- Nome correto da coluna
        COALESCE(p.imagem_capa, ''), 
        COALESCE(p.categoria, 'Geral'), 
        COALESCE(p.cidade_projeto, ''), 
        p.visualizacoes,
        u.nome_completo, 
        COALESCE(u.foto_perfil, ''), 
        u.id_usuario,
        EXISTS(
            SELECT 1 FROM favoritos f 
            WHERE f.id_item_salvo = p.id_projeto 
            AND f.id_usuario_quem_salvou = $1 
            AND f.tipo_item = 'PROJETO'
        ) as favoritado,
		COALESCE(p.media_estrelas, 0) as media,  
    COALESCE(p.total_avaliacoes, 0) as total
    FROM projetos p
    JOIN usuarios u ON p.id_lider = u.id_usuario
    WHERE p.status_projeto != 'Oculto'
    AND u.is_banned = FALSE`

	// Argumentos iniciais
	args := []interface{}{idUsuarioLogado}
	contador := 2 // Começa em 2 porque $1 é o ID do usuário

	// 1. Filtro de Texto
	if termo != "" {
		query += fmt.Sprintf(" AND (p.titulo ILIKE $%d OR p.descricao ILIKE $%d)", contador, contador)
		args = append(args, "%"+termo+"%")
		contador++
	}

	// 2. Filtro de Categoria (Area)
	// Só aplica se não estiver vazio E não for "Todas"
	if categoria != "" && categoria != "Todas" {
		query += fmt.Sprintf(" AND p.categoria = $%d", contador)
		args = append(args, categoria)
		contador++
	}

	// 3. Filtro de Cidade
	// Só aplica se não estiver vazio E não for "Todas"
	if cidade != "" && cidade != "Todas" {
		query += fmt.Sprintf(" AND p.cidade_projeto = $%d", contador)
		args = append(args, cidade)
		contador++
	}

	query += " ORDER BY p.id_projeto DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projetos []structs.Projeto
	for rows.Next() {
		var p structs.Projeto
		err = rows.Scan(&p.Id, &p.Titulo, &p.Descricao, &p.Tags, &p.Status, &p.ImagemCapa, &p.Categoria, &p.Cidade, &p.Visualizacoes, &p.NomeLider, &p.FotoLider, &p.IdLider, &p.EstaSalvo, &p.MediaEstrelas, &p.TotalAvaliacoes,)
		if err != nil {
			continue
		}
		p.Tecnologias = p.TagsComoLista()
		projetos = append(projetos, p)
	}
	return projetos, nil
}

// BuscarProjetosComFiltro constrói a query dinamicamente baseada nos inputs
func BuscarProjetosComFiltro(termo, categoria, cidade string) ([]structs.Projeto, error) {
	// 1. Query Base (Idêntica a sua original, sem o WHERE/ORDER BY ainda)
	queryBase := `
        SELECT p.id_projeto, p.titulo, p.descricao, 
               COALESCE(array_to_string(p.tecnologias, ','), ''), 
               p.status_projeto, 
               COALESCE(p.imagem_capa, ''), 
               COALESCE(p.categoria, 'Geral'), 
               COALESCE(p.cidade_projeto, ''), 
               p.visualizacoes,
               u.nome_completo, COALESCE(u.foto_perfil, ''), u.id_usuario
        FROM projetos p
        JOIN usuarios u ON p.id_lider = u.id_usuario
		WHERE p.status_projeto != 'Oculto'
    `

	// 2. Preparação dos Filtros Dinâmicos
	var condicoes []string
	var args []interface{}
	contador := 1 // Controla o número do placeholder ($1, $2, etc)

	// Filtro 1: Termo de Busca (Título ou Descrição)
	if termo != "" {
		// ILIKE faz busca case-insensitive (ignora maiúsculas/minúsculas)
		// Usamos o mesmo índice (contador) duas vezes para passar o argumento apenas uma vez
		condicao := fmt.Sprintf("(p.titulo ILIKE $%d OR p.descricao ILIKE $%d)", contador, contador)
		condicoes = append(condicoes, condicao)
		args = append(args, "%"+termo+"%") // Adiciona % para buscar em qualquer parte do texto
		contador++
	}

	// Filtro 2: Categoria (Se não for vazia e nem "Todas")
	if categoria != "" && categoria != "Todas" {
		condicoes = append(condicoes, fmt.Sprintf("p.categoria = $%d", contador))
		args = append(args, categoria)
		contador++
	}

	// Filtro 3: Cidade (Se não for vazia e nem "Todas")
	if cidade != "" && cidade != "Todas" {
		condicoes = append(condicoes, fmt.Sprintf("p.cidade_projeto = $%d", contador))
		args = append(args, cidade)
		contador++
	}

	// 3. Montagem Final da Query
	if len(condicoes) > 0 {
		queryBase += " AND " + strings.Join(condicoes, " AND ")
	}

	// Adiciona a ordenação no final
	queryBase += " ORDER BY p.id_projeto DESC"

	// 4. Execução (Passando os args dinâmicos com "...")
	rows, err := db.DB.Query(queryBase, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 5. Scan (Exatamente igual ao seu código original)
	var projetos []structs.Projeto
	for rows.Next() {
		var p structs.Projeto
		err = rows.Scan(
			&p.Id, &p.Titulo, &p.Descricao, &p.Tags, &p.Status,
			&p.ImagemCapa, &p.Categoria, &p.Cidade, &p.Visualizacoes,
			&p.NomeLider, &p.FotoLider, &p.IdLider,
		)
		if err != nil {
			continue
		}
		p.Tecnologias = p.TagsComoLista()
		projetos = append(projetos, p)
	}

	return projetos, nil
}

func BuscarTodosTalentos(termo, nivel, cidade string, idUsuarioLogado int) ([]structs.Usuario, error) {
	// Adicionei COALESCE(nivel_profissional, '') no SELECT
	query := `
        SELECT 
        id_usuario, 
        nome_completo, 
        COALESCE(titulo_profissional, 'Dev'), 
        COALESCE(foto_perfil, ''), 
        COALESCE(skills, ''), 
        disponivel_para_trabalho, 
        COALESCE(cidade, ''), 
        COALESCE(nivel_profissional, ''),
        EXISTS(
            SELECT 1 FROM favoritos f 
            WHERE f.id_item_salvo = usuarios.id_usuario 
            AND f.id_usuario_quem_salvou = $1 
            AND f.tipo_item = 'DEV'
        ) as favoritado
    FROM usuarios 
    WHERE tipo_usuario = 'DEV'
    AND is_banned = FALSE`

	args := []interface{}{idUsuarioLogado}
	contador := 2

	// 1. Filtro de Texto
	if termo != "" {
		query += fmt.Sprintf(" AND (nome_completo ILIKE $%d OR skills ILIKE $%d)", contador, contador)
		args = append(args, "%"+termo+"%")
		contador++
	}

	// 2. Filtro de Nível (NOVO)
	if nivel != "" && nivel != "Todos" {
		// Busca exata no campo nivel_profissional
		query += fmt.Sprintf(" AND nivel_profissional = $%d", contador)
		args = append(args, nivel)
		contador++
	}

	// 3. Filtro de Cidade
	if cidade != "" && cidade != "Todas" {
		query += fmt.Sprintf(" AND cidade = $%d", contador)
		args = append(args, cidade)
		contador++
	}

	query += " ORDER BY id_usuario DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var talentos []structs.Usuario
	for rows.Next() {
		var u structs.Usuario
		var skillsStr string
		// Scan atualizado com a nova coluna Nivel
		err = rows.Scan(&u.Id, &u.NomeCompleto, &u.TituloProfissional, &u.FotoPerfil, &skillsStr, &u.DisponivelParaEquipes, &u.Cidade, &u.Nivel, &u.EstaSalvo)
		if err != nil {
			continue
		}
		u.Skills = skillsStr
		talentos = append(talentos, u)
	}
	return talentos, nil
}

// models/usuario.go

// BuscarTalentosComFiltro busca devs com filtros opcionais
func BuscarTalentosComFiltro(termo string, nivel string, disponivel string, idUsuarioLogado int) ([]structs.Usuario, error) {
	// 1. Query Base (Com a verificação de FAVORITO)
	// Note o $1 ali no EXISTS, que será o idUsuarioLogado
	queryBase := `
        SELECT id_usuario, nome_completo, 
               COALESCE(titulo_profissional, ''), 
               COALESCE(foto_perfil, ''), 
               COALESCE(skills, ''), 
               disponivel_para_trabalho,
               COALESCE(nivel_profissional, ''),
               EXISTS(SELECT 1 FROM favoritos f WHERE f.id_item_salvo = usuarios.id_usuario AND f.id_usuario_quem_salvou = $1 AND f.tipo_item = 'DEV') as esta_salvo
        FROM usuarios 
        WHERE tipo_usuario = 'DEV' AND is_banned = FALSE
    `

	// 2. Montagem dos Filtros
	// Começamos o args com o ID do usuário (para o $1 do EXISTS)
	args := []interface{}{idUsuarioLogado}

	// O contador começa em 2 porque o $1 já está ocupado pelo idUsuarioLogado
	contador := 2
	var condicoes []string

	// --- FILTRO POR TEXTO ---
	if termo != "" {
		// Busca no nome, no cargo ou nas skills
		// Usamos contador, contador+1, contador+2
		condicao := fmt.Sprintf("(nome_completo ILIKE $%d OR titulo_profissional ILIKE $%d OR skills ILIKE $%d)", contador, contador+1, contador+2)
		condicoes = append(condicoes, condicao)

		termoBusca := "%" + termo + "%"
		args = append(args, termoBusca, termoBusca, termoBusca) // Adiciona 3 vezes
		contador += 3
	}

	// --- FILTRO POR NÍVEL (Faltava isso) ---
	if nivel != "" && nivel != "Todos" {
		condicoes = append(condicoes, fmt.Sprintf("nivel_profissional = $%d", contador))
		args = append(args, nivel)
		contador++
	}

	// --- FILTRO POR DISPONIBILIDADE ---
	if disponivel == "true" {
		condicoes = append(condicoes, fmt.Sprintf("disponivel_para_trabalho = $%d", contador))
		args = append(args, true)
		contador++
	}

	// Se houver filtros extras, adiciona com AND
	if len(condicoes) > 0 {
		queryBase += " AND " + strings.Join(condicoes, " AND ")
	}

	queryBase += " ORDER BY id_usuario DESC"

	// 3. Execução
	rows, err := db.DB.Query(queryBase, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var talentos []structs.Usuario
	for rows.Next() {
		var u structs.Usuario
		// Lembre-se de scanear o &u.EstaSalvo no final
		err = rows.Scan(
			&u.Id,
			&u.NomeCompleto,
			&u.TituloProfissional,
			&u.FotoPerfil,
			&u.Skills,
			&u.DisponivelParaEquipes,
			&u.Nivel,
			&u.EstaSalvo, // <--- Importante para o botão funcionar
		)
		if err != nil {
			continue
		}
		talentos = append(talentos, u)
	}
	return talentos, nil
}
