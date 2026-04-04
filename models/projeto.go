package models

import (
	"log"
	"nexhub/db"
	"nexhub/structs"
)

// ==========================================
// 1. CRUD BÁSICO DE PROJETOS (ADMIN)
// ==========================================

// CriarProjeto cadastra a base do projeto (executado pelo Admin de ADS)
func CriarProjeto(p structs.Projeto) (int, error) {
	query := `
		INSERT INTO projetos (
			titulo, descricao, id_curso, id_area, semestre_letivo, 
			professor_orientador, status_projeto, imagem_capa, link_repositorio, cadastrado_por
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id_projeto`

	var idGerado int
	err := db.DB.QueryRow(query,
		p.Titulo, p.Descricao, p.IdCurso, p.IdArea, p.SemestreLetivo,
		p.ProfessorOrientador, p.StatusProjeto, p.ImagemCapa, p.LinkRepositorio, p.CadastradoPor,
	).Scan(&idGerado)

	return idGerado, err
}

// AtualizarProjeto atualiza as informações gerais do projeto
func AtualizarProjeto(p structs.Projeto) error {
	query := `
		UPDATE projetos 
		SET titulo=$1, descricao=$2, id_curso=$3, id_area=$4, semestre_letivo=$5, 
			professor_orientador=$6, status_projeto=$7, link_repositorio=$8, data_atualizacao=CURRENT_TIMESTAMP
		WHERE id_projeto=$9`

	_, err := db.DB.Exec(query,
		p.Titulo, p.Descricao, p.IdCurso, p.IdArea, p.SemestreLetivo,
		p.ProfessorOrientador, p.StatusProjeto, p.LinkRepositorio, p.IdProjeto,
	)
	return err
}

// AtualizarCapaProjeto é chamado separadamente caso o admin faça upload de nova capa
func AtualizarCapaProjeto(idProjeto int, imagemCapa string) error {
	query := `UPDATE projetos SET imagem_capa=$1 WHERE id_projeto=$2`
	_, err := db.DB.Exec(query, imagemCapa, idProjeto)
	return err
}

func DeletarProjeto(id int) error {
	query := `DELETE FROM projetos WHERE id_projeto = $1`
	_, err := db.DB.Exec(query, id)
	if err != nil {
		log.Println("Erro ao deletar projeto:", err)
	}
	return err
}

// ==========================================
// 2. BUSCAS E LISTAGENS
// ==========================================

// BuscarDetalhesProjeto traz todas as informações e agregações do projeto para a Vitrine Pública
func BuscarDetalhesProjeto(id int) (structs.Projeto, error) {
	var p structs.Projeto

	// 1. Dados Principais
	query := `
		SELECT 
			p.id_projeto, p.titulo, p.descricao, p.id_curso, p.id_area, 
			COALESCE(p.semestre_letivo, ''), COALESCE(p.professor_orientador, ''), 
			p.status_projeto, COALESCE(p.imagem_capa, ''), COALESCE(p.link_repositorio, ''), 
			p.data_criacao, COALESCE(c.nome_curso, ''), COALESCE(a.nome_area, '')
		FROM projetos p
		LEFT JOIN cursos c ON p.id_curso = c.id_curso
		LEFT JOIN areas a ON p.id_area = a.id_area
		WHERE p.id_projeto = $1`

	var nomeCurso, nomeArea string
	err := db.DB.QueryRow(query, id).Scan(
		&p.IdProjeto, &p.Titulo, &p.Descricao, &p.IdCurso, &p.IdArea,
		&p.SemestreLetivo, &p.ProfessorOrientador, &p.StatusProjeto,
		&p.ImagemCapa, &p.LinkRepositorio, &p.DataCriacao,
		&nomeCurso, &nomeArea,
	)
	if err != nil {
		return p, err
	}
	p.Curso = structs.Curso{NomeCurso: nomeCurso}
	p.Area = structs.Area{NomeArea: nomeArea}

	// 2. Buscar Membros da Equipe (Alunos)
	p.Equipe, _ = BuscarEquipeDoProjeto(id)

	// 3. Buscar Arquivos (PDFs) e Links
	p.Arquivos, _ = BuscarArquivosDoProjeto(id)
	p.Links, _ = BuscarLinksDoProjeto(id)

	// 4. Buscar Galeria de Imagens
	p.Imagens, _ = BuscarGaleriaDoProjeto(id)

	return p, nil
}

// BuscarProjetosDoAluno lista todos os projetos onde o aluno participou (para o portfólio)
func BuscarProjetosDoAluno(idAluno int) ([]structs.Projeto, error) {
	query := `
		SELECT 
			p.id_projeto, p.titulo, p.status_projeto, 
			COALESCE(p.imagem_capa, ''), COALESCE(c.nome_curso, '')
		FROM projetos p
		JOIN projeto_alunos pa ON p.id_projeto = pa.id_projeto
		LEFT JOIN cursos c ON p.id_curso = c.id_curso
		WHERE pa.id_aluno = $1 AND p.status_projeto != 'OCULTO'
		ORDER BY p.id_projeto DESC`

	rows, err := db.DB.Query(query, idAluno)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projetos []structs.Projeto
	for rows.Next() {
		var p structs.Projeto
		var nomeCurso string
		if err := rows.Scan(&p.IdProjeto, &p.Titulo, &p.StatusProjeto, &p.ImagemCapa, &nomeCurso); err == nil {
			p.Curso = structs.Curso{NomeCurso: nomeCurso}
			projetos = append(projetos, p)
		}
	}
	return projetos, nil
}

// ==========================================
// 3. GESTÃO DE EQUIPE (VÍNCULO COM ALUNOS)
// ==========================================

func BuscarEquipeDoProjeto(idProjeto int) ([]structs.Aluno, error) {
	query := `
		SELECT a.id_aluno, a.nome_completo, COALESCE(a.foto_perfil, ''), COALESCE(pa.funcao_no_projeto, '')
		FROM projeto_alunos pa
		JOIN alunos a ON pa.id_aluno = a.id_aluno
		WHERE pa.id_projeto = $1
	`
	rows, err := db.DB.Query(query, idProjeto)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var equipe []structs.Aluno
	for rows.Next() {
		var aluno structs.Aluno
		var funcao string
		if err := rows.Scan(&aluno.IdAluno, &aluno.NomeCompleto, &aluno.FotoPerfil, &funcao); err == nil {
			aluno.Biografia = funcao // Usando o campo Biografia provisoriamente para enviar a "Função" pro template
			equipe = append(equipe, aluno)
		}
	}
	return equipe, nil
}

func AdicionarMembroEquipe(idProjeto, idAluno int, funcao string) error {
	query := `
		INSERT INTO projeto_alunos (id_projeto, id_aluno, funcao_no_projeto)
		VALUES ($1, $2, $3)
		ON CONFLICT (id_projeto, id_aluno) DO UPDATE 
		SET funcao_no_projeto = EXCLUDED.funcao_no_projeto`
	_, err := db.DB.Exec(query, idProjeto, idAluno, funcao)
	return err
}

func RemoverMembroEquipe(idProjeto, idAluno int) error {
	query := `DELETE FROM projeto_alunos WHERE id_projeto = $1 AND id_aluno = $2`
	_, err := db.DB.Exec(query, idProjeto, idAluno)
	return err
}

// PesquisarAlunosPorNome busca alunos no BD para adicionar à equipe (API)
func PesquisarAlunosPorNome(termo string) ([]structs.Aluno, error) {
	query := `
		SELECT id_aluno, nome_completo, COALESCE(foto_perfil, '')
		FROM alunos 
		WHERE nome_completo ILIKE $1 
		LIMIT 5`

	rows, err := db.DB.Query(query, "%"+termo+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Aluno
	for rows.Next() {
		var a structs.Aluno
		if err := rows.Scan(&a.IdAluno, &a.NomeCompleto, &a.FotoPerfil); err == nil {
			lista = append(lista, a)
		}
	}
	return lista, nil
}

// ==========================================
// 4. ANEXOS, LINKS E GALERIA (Q5)
// ==========================================

func BuscarArquivosDoProjeto(idProjeto int) ([]structs.ProjetoArquivo, error) {
	query := `SELECT id_arquivo, nome_original, caminho_arquivo, data_upload FROM projeto_arquivos WHERE id_projeto = $1`
	rows, err := db.DB.Query(query, idProjeto)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var arquivos []structs.ProjetoArquivo
	for rows.Next() {
		var arq structs.ProjetoArquivo
		if err := rows.Scan(&arq.IdArquivo, &arq.NomeOriginal, &arq.CaminhoArquivo, &arq.DataUpload); err == nil {
			arquivos = append(arquivos, arq)
		}
	}
	return arquivos, nil
}

func AdicionarLinkProjeto(link structs.ProjetoLink) error {
	query := `INSERT INTO projeto_links (id_projeto, tipo_link, url, descricao) VALUES ($1, $2, $3, $4)`
	_, err := db.DB.Exec(query, link.IdProjeto, link.TipoLink, link.Url, link.Descricao)
	return err
}

func BuscarLinksDoProjeto(idProjeto int) ([]structs.ProjetoLink, error) {
	query := `SELECT id_link, tipo_link, url, descricao FROM projeto_links WHERE id_projeto = $1`
	rows, err := db.DB.Query(query, idProjeto)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []structs.ProjetoLink
	for rows.Next() {
		var l structs.ProjetoLink
		if err := rows.Scan(&l.IdLink, &l.TipoLink, &l.Url, &l.Descricao); err == nil {
			links = append(links, l)
		}
	}
	return links, nil
}

// Galeria de Imagens
func AdicionarImagemGaleria(idProjeto int, caminhoImagem string) error {
	query := `INSERT INTO projeto_imagens (id_projeto, caminho_imagem) VALUES ($1, $2)`
	_, err := db.DB.Exec(query, idProjeto, caminhoImagem)
	return err
}

func BuscarGaleriaDoProjeto(idProjeto int) ([]structs.ProjetoImagem, error) {
	query := `SELECT id_imagem, caminho_imagem FROM projeto_imagens WHERE id_projeto = $1 ORDER BY id_imagem ASC`
	rows, err := db.DB.Query(query, idProjeto)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imagens []structs.ProjetoImagem
	for rows.Next() {
		var img structs.ProjetoImagem
		if err := rows.Scan(&img.IdImagem, &img.CaminhoImagem); err == nil {
			imagens = append(imagens, img)
		}
	}
	return imagens, nil
}

func DeletarImagemGaleria(idImagem int) error {
	query := `DELETE FROM projeto_imagens WHERE id_imagem = $1`
	_, err := db.DB.Exec(query, idImagem)
	return err
}

// ListarTodosProjetosAdmin busca a lista para a tabela do painel de controle
func ListarTodosProjetosAdmin() ([]structs.Projeto, error) {
	query := `
		SELECT 
			p.id_projeto, p.titulo, p.status_projeto, 
			COALESCE(c.nome_curso, 'Geral') as nome_curso
		FROM projetos p
		LEFT JOIN cursos c ON p.id_curso = c.id_curso
		ORDER BY p.id_projeto DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projetos []structs.Projeto
	for rows.Next() {
		var p structs.Projeto
		var nomeCurso string
		if err := rows.Scan(&p.IdProjeto, &p.Titulo, &p.StatusProjeto, &nomeCurso); err == nil {
			p.Curso = structs.Curso{NomeCurso: nomeCurso}
			projetos = append(projetos, p)
		}
	}
	return projetos, nil
}
