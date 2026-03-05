package models

import (
	"fmt"
	"log"
	"nexhub/db"
	"nexhub/structs"

	"github.com/lib/pq"
)

// Incrementa +1 no contador de visualizações do projeto
func IncrementarVisualizacao(idProjeto int) error {
	query := `UPDATE projetos SET visualizacoes = visualizacoes + 1 WHERE id_projeto = $1`
	_, err := db.DB.Exec(query, idProjeto)
	return err
}

// BuscarProjetosDoDev retorna todos os projetos de um usuário específico
func BuscarProjetosDoDev(idUsuario int) ([]structs.Projeto, error) {
	// Query: Seleciona projetos onde o id_lider é igual ao id passado
	sqlStatement := `
		SELECT id_projeto, titulo, descricao, status_projeto, cidade_projeto, imagem_capa, visualizacoes 
		FROM projetos 
		WHERE id_lider = $1 
		ORDER BY id_projeto DESC` // Mostra os mais novos primeiro

	rows, err := db.DB.Query(sqlStatement, idUsuario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projetos []structs.Projeto

	for rows.Next() {
		var p structs.Projeto
		// Precisamos ler na mesma ordem do SELECT acima
		err = rows.Scan(&p.Id, &p.Titulo, &p.Descricao, &p.Status, &p.Cidade, &p.ImagemCapa, &p.Visualizacoes)
		if err != nil {
			return nil, err
		}
		projetos = append(projetos, p)
	}

	return projetos, nil
}

// 1. CRIAR PROJETO (Salva o Array)
func CriarProjeto(p structs.Projeto) (int, error) {
	// Note o pq.Array(p.Tecnologias)
	query := `
        INSERT INTO projetos (titulo, descricao, status_projeto, cidade_projeto, categoria, link_repositorio, imagem_capa, id_lider, tecnologias) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        RETURNING id_projeto`

	var idGerado int
	err := db.DB.QueryRow(query,
		p.Titulo, p.Descricao, p.Status, p.Cidade, p.Categoria, p.LinkRepo, p.ImagemCapa, p.IdLider,
		pq.Array(p.Tecnologias), // Converte []string para TEXT[] do banco
	).Scan(&idGerado)

	return idGerado, err
}

// 2. ATUALIZADO: Buscar projeto para edição (Corrigido: category -> categoria)
func BuscarProjetoPorID(id int) (structs.Projeto, error) {
	// Query principal (Projeto)
	// CORREÇÃO AQUI: Troquei 'category' por 'categoria'
	query := `
        SELECT id_projeto, titulo, descricao, status_projeto, cidade_projeto, categoria, imagem_capa, link_repositorio, tecnologias, visualizacoes
        FROM projetos WHERE id_projeto = $1`

	var p structs.Projeto
	// O Scan do pq.Array converte de volta para []string
	err := db.DB.QueryRow(query, id).Scan(
		&p.Id, &p.Titulo, &p.Descricao, &p.Status, &p.Cidade,
		&p.Categoria, &p.ImagemCapa, &p.LinkRepo,
		pq.Array(&p.Tecnologias),
		&p.Visualizacoes,
	)
	if err != nil {
		return p, err
	}

	// Query secundária (Galeria com IDs) - MANTIDA IGUAL
	queryGaleria := `SELECT id_imagem, caminho_imagem FROM projeto_imagens WHERE id_projeto = $1 ORDER BY id_imagem ASC`
	rows, err := db.DB.Query(queryGaleria, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var img structs.ImagemGaleria
			if err := rows.Scan(&img.Id, &img.Caminho); err == nil {
				p.ImagensGaleria = append(p.ImagensGaleria, img)
			}
		}
	}
	return p, nil
}

// 2. ATUALIZAR PROJETO (Atualiza o Array)
func AtualizarProjeto(id int, titulo, descricao, status, cidade, categoria, imagemCapa, repo string, tecnologias []string) error {
	query := `
        UPDATE projetos 
        SET titulo=$1, descricao=$2, status_projeto=$3, cidade_projeto=$4, categoria=$5, link_repositorio=$6, tecnologias=$7
        WHERE id_projeto=$8`

	// Se tiver imagem nova, atualiza ela tmb (lógica simplificada aqui)
	if imagemCapa != "" {
		query = `UPDATE projetos SET titulo=$1, descricao=$2, status_projeto=$3, cidade_projeto=$4, categoria=$5, link_repositorio=$6, tecnologias=$7, imagem_capa='` + imagemCapa + `' WHERE id_projeto=$8`
	}

	_, err := db.DB.Exec(query, titulo, descricao, status, cidade, categoria, repo, pq.Array(tecnologias), id)
	return err
}

// 3. DELETAR PROJETO (Corrigido)
func DeletarProjeto(id int) {
	// Ajustado para 'id_projeto'
	query := `DELETE FROM projetos WHERE id_projeto = $1`

	_, err := db.DB.Exec(query, id)
	if err != nil {
		log.Println("Erro ao deletar projeto:", err)
	}
}

// 1. Função Nova: SALVAR UMA IMAGEM NA GALERIA
func AdicionarImagemGaleria(idProjeto int, caminhoImagem string) error {
	query := `INSERT INTO projeto_imagens (id_projeto, caminho_imagem) VALUES ($1, $2)`
	_, err := db.DB.Exec(query, idProjeto, caminhoImagem)
	return err
}

// 2. Função Atualizada: BUSCAR DETALHES (Agora busca as fotos da galeria tmb)
// models/projeto.go
func BuscarDetalhesProjeto(id int) (structs.Projeto, string, string, error) {
	// CORREÇÃO: Adicionei 'p.id_lider' no SELECT
	query := `
        SELECT 
            p.id_projeto, p.titulo, p.descricao, p.status_projeto, p.cidade_projeto, 
            COALESCE(p.categoria, 'Tecnologia'), COALESCE(p.imagem_capa, ''), COALESCE(p.link_repositorio, ''), 
            p.tecnologias, 
            p.id_lider,  -- <--- FALTAVA ISSO AQUI
            u.nome_completo, COALESCE(u.foto_perfil, ''),
			COALESCE(p.media_estrelas, 0) as media,  
    COALESCE(p.total_avaliacoes, 0) as total
        FROM projetos p
        JOIN usuarios u ON p.id_lider = u.id_usuario
        WHERE p.id_projeto = $1`

	var p structs.Projeto
	var autorNome, autorAvatar string

	// CORREÇÃO: Adicionei '&p.IdLider' no Scan
	err := db.DB.QueryRow(query, id).Scan(
		&p.Id, &p.Titulo, &p.Descricao, &p.Status, &p.Cidade,
		&p.Categoria, &p.ImagemCapa, &p.LinkRepo,
		pq.Array(&p.Tecnologias),
		&p.IdLider, // <--- PREENCHENDO O ID DO DONO
		&autorNome, &autorAvatar, &p.MediaEstrelas, &p.TotalAvaliacoes,
	)
	if err != nil {
		return p, "", "", err
	}

	// (Parte da Galeria continua igual)
	queryGaleria := `SELECT id_imagem, caminho_imagem FROM projeto_imagens WHERE id_projeto = $1 ORDER BY id_imagem ASC`
	rows, err := db.DB.Query(queryGaleria, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var img structs.ImagemGaleria
			if err := rows.Scan(&img.Id, &img.Caminho); err == nil {
				p.ImagensGaleria = append(p.ImagensGaleria, img)
			}
		}
	}

	return p, autorNome, autorAvatar, err
}

func DeletarImagemGaleria(idImagem int) {
	// OBS: Idealmente, deletaríamos o arquivo do disco também,
	// mas para simplificar, vamos deletar só do banco por enquanto.
	query := `DELETE FROM projeto_imagens WHERE id_imagem = $1`
	_, err := db.DB.Exec(query, idImagem)
	if err != nil {
		log.Println("Erro ao deletar imagem da galeria:", err)
	}
}

//equipes

// 1. Buscar membros de um projeto específico
func BuscarMembrosDoProjeto(idProjeto int) ([]structs.MembroEquipe, error) {
	query := `
        SELECT u.id_usuario, u.nome_completo, COALESCE(u.foto_perfil, ''), 
        e.funcao_no_projeto, e.data_entrada
        FROM equipe_projeto e
        JOIN usuarios u ON e.id_usuario = u.id_usuario
        WHERE e.id_projeto = $1 AND is_banned = false
        ORDER BY e.data_entrada ASC
    `
	rows, err := db.DB.Query(query, idProjeto)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipe []structs.MembroEquipe
	for rows.Next() {
		var m structs.MembroEquipe
		// Scan deve seguir a ordem do SELECT
		rows.Scan(&m.IdUsuario, &m.Nome, &m.Foto, &m.Funcao, &m.DataEntrada)
		equipe = append(equipe, m)
	}
	return equipe, nil
}

// 2. Adicionar um membro na tabela de ligação

func AdicionarMembroEquipe(idProjeto, idUsuario int, funcao string) error {
	// 1. Primeiro verifica se o usuário está disponível
	var disponivel bool
	err := db.DB.QueryRow("SELECT disponivel_para_trabalho FROM usuarios WHERE id_usuario = $1 AND is_banned = false", idUsuario).Scan(&disponivel)

	if err != nil {
		return err
	}
	if !disponivel {
		return fmt.Errorf("este usuário não está disponível para projetos")
	}

	// 2. Se estiver disponível, prossegue com a inserção
	query := `
        INSERT INTO equipe_projeto (id_projeto, id_usuario, funcao_no_projeto)
        VALUES ($1, $2, $3)
        ON CONFLICT (id_projeto, id_usuario) DO UPDATE 
        SET funcao_no_projeto = EXCLUDED.funcao_no_projeto
    `
	_, err = db.DB.Exec(query, idProjeto, idUsuario, funcao)
	return err
}

// 3. Remover membro
func RemoverMembroEquipe(idProjeto, idUsuario int) error {
	query := `DELETE FROM equipe_projeto WHERE id_projeto = $1 AND id_usuario = $2`
	_, err := db.DB.Exec(query, idProjeto, idUsuario)
	return err
}

// 4. Buscar usuários para o Auto-Complete (API)
// Busca apenas DEVs que NÃO sejam o próprio líder (passamos o liderID para excluir)
// models/projeto.go

// PesquisarDevsPorNome busca devs para o autocomplete (FILTRANDO DISPONIBILIDADE)
func PesquisarDevsPorNome(termo string) ([]structs.Usuario, error) {
	query := `
        SELECT id_usuario, nome_completo, COALESCE(foto_perfil, ''), COALESCE(titulo_profissional, '')
        FROM usuarios 
        WHERE tipo_usuario = 'DEV' 
        AND disponivel_para_trabalho = true 
		AND is_banned = false
        AND nome_completo ILIKE $1 
        LIMIT 5`

	rows, err := db.DB.Query(query, "%"+termo+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Usuario
	for rows.Next() {
		var u structs.Usuario
		var titulo string

		// Usei var titulo string para simplificar, o COALESCE no SQL já garante que não vem NULL
		err = rows.Scan(&u.Id, &u.NomeCompleto, &u.FotoPerfil, &titulo)
		if err != nil {
			continue
		}

		u.TituloProfissional = titulo
		lista = append(lista, u)
	}
	return lista, nil
}

// models/projeto.go

// BuscarPortfolioCompleto traz projetos onde o usuário é LÍDER ou MEMBRO
func BuscarPortfolioCompleto(idUsuario int) ([]structs.Projeto, error) {
	// Usamos DISTINCT para evitar duplicatas caso o líder se adicione como membro
	query := `
        SELECT DISTINCT 
            p.id_projeto, p.titulo, p.descricao, p.status_projeto, 
            p.cidade_projeto, COALESCE(p.imagem_capa, ''), p.visualizacoes, 
            COALESCE(p.categoria, 'Geral'), p.id_lider,
            COALESCE(array_to_string(p.tecnologias, ','), '')
        FROM projetos p
        LEFT JOIN equipe_projeto e ON p.id_projeto = e.id_projeto
        WHERE p.id_lider = $1 OR e.id_usuario = $1
        ORDER BY p.id_projeto DESC
    `

	rows, err := db.DB.Query(query, idUsuario)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projetos []structs.Projeto

	for rows.Next() {
		var p structs.Projeto
		err = rows.Scan(
			&p.Id, &p.Titulo, &p.Descricao, &p.Status,
			&p.Cidade, &p.ImagemCapa, &p.Visualizacoes,
			&p.Categoria, &p.IdLider, &p.Tags,
		)
		if err != nil {
			continue
		}
		p.Tecnologias = p.TagsComoLista()
		projetos = append(projetos, p)
	}

	return projetos, nil
}

// ContarSavesEmpresa conta quantas empresas favoritaram este projeto
func ContarSavesEmpresa(idProjeto int) (int, error) {
	query := `
        SELECT COUNT(*)
        FROM favoritos f
        JOIN usuarios u ON f.id_usuario_quem_salvou = u.id_usuario
        WHERE f.tipo_item = 'PROJETO' 
        AND f.id_item_salvo = $1 
        AND u.tipo_usuario = 'EMPRESA'
    `
	var total int
	err := db.DB.QueryRow(query, idProjeto).Scan(&total)
	return total, err
}
