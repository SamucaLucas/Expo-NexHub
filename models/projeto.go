package models

import (
	"log"
	"nexhub/db"
	"nexhub/structs"
	"strconv"
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

	// 1. Dados Principais (AGORA COM AS ESTRELAS E TOTAL DE AVALIAÇÕES)
	query := `
		SELECT 
			p.id_projeto, p.titulo, p.descricao, p.id_curso, p.id_area, 
			COALESCE(p.semestre_letivo, ''), COALESCE(p.professor_orientador, ''), 
			p.status_projeto, COALESCE(p.imagem_capa, ''), COALESCE(p.link_repositorio, ''), 
			p.data_criacao, COALESCE(c.nome_curso, ''), COALESCE(a.nome_area, ''),
			COALESCE(p.media_estrelas, 0), COALESCE(p.total_avaliacoes, 0)
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
		&p.MediaEstrelas, &p.TotalAvaliacoes, // <--- MAPEANDO OS NOVOS CAMPOS
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

	// 5. BUSCA AS AVALIAÇÕES/COMENTÁRIOS DO PROJETO PARA A TELA PÚBLICA
	queryAvaliacoes := `
		SELECT nome_avaliador, nota, comentario, TO_CHAR(data_avaliacao, 'DD/MM/YYYY') 
		FROM avaliacoes 
		WHERE id_projeto = $1 
		ORDER BY data_avaliacao DESC`

	rowsAv, errAv := db.DB.Query(queryAvaliacoes, id)
	if errAv == nil { // Só faz o loop se não der erro no banco
		defer rowsAv.Close()
		for rowsAv.Next() {
			var av structs.Avaliacao
			// Scan tem que bater exatamente com a ordem do SELECT acima
			rowsAv.Scan(&av.NomeAvaliador, &av.Nota, &av.Comentario, &av.DataFormatada)
			p.Avaliacoes = append(p.Avaliacoes, av)
		}
	}

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
			p.id_projeto, p.titulo, p.status_projeto, p.imagem_capa, 
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
		if err := rows.Scan(&p.IdProjeto, &p.Titulo, &p.StatusProjeto, &p.ImagemCapa, &nomeCurso); err == nil {
			p.Curso = structs.Curso{NomeCurso: nomeCurso}
			projetos = append(projetos, p)
		}
	}
	return projetos, nil
}

func ListarProjetosAdmin(busca, cursoId string) ([]structs.Projeto, error) {
	query := `
		SELECT p.id_projeto, p.titulo, COALESCE(p.imagem_capa, ''), p.status_projeto, 
		       COALESCE(c.id_curso, 0), COALESCE(c.nome_curso, 'Multidisciplinar')
		FROM projetos p
		LEFT JOIN cursos c ON p.id_curso = c.id_curso
		WHERE 1=1
	`
	var args []interface{}
	argId := 1

	// Filtro por termo de busca
	if busca != "" {
		query += ` AND p.titulo ILIKE $` + strconv.Itoa(argId)
		args = append(args, "%"+busca+"%")
		argId++
	}

	// Filtro pelo curso selecionado (ou o curso padrão do usuário)
	if cursoId != "" && cursoId != "0" {
		query += ` AND p.id_curso = $` + strconv.Itoa(argId)
		args = append(args, cursoId)
		argId++
	}

	query += ` ORDER BY p.id_projeto DESC`

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Projeto
	for rows.Next() {
		var p structs.Projeto
		var c structs.Curso

		// Adapte os campos do Scan de acordo com a sua Struct de Projeto
		rows.Scan(&p.IdProjeto, &p.Titulo, &p.ImagemCapa, &p.StatusProjeto, &c.IdCurso, &c.NomeCurso)
		p.Curso = c

		lista = append(lista, p)
	}
	return lista, nil
}

// ListarTodasAreas busca a lista de categorias para o formulário de projetos
func ListarTodasAreas() ([]structs.Area, error) {
	query := `SELECT id_area, nome_area FROM areas ORDER BY nome_area ASC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []structs.Area
	for rows.Next() {
		var a structs.Area
		if err := rows.Scan(&a.IdArea, &a.NomeArea); err == nil {
			areas = append(areas, a)
		}
	}
	return areas, nil
}

// BuscarProjetoPorId retorna os detalhes de um projeto específico
func BuscarProjetoPorId(idProjeto int) (structs.Projeto, error) {
	var p structs.Projeto

	// Usamos COALESCE para evitar que o Go trave se algum campo opcional estiver NULL no banco
	query := `
		SELECT id_projeto, titulo, descricao, id_curso, id_area,
			   COALESCE(semestre_letivo, ''), COALESCE(professor_orientador, ''),
			   status_projeto, COALESCE(imagem_capa, ''), COALESCE(link_repositorio, ''), media_estrelas, total_avaliacoes
		FROM projetos
		WHERE id_projeto = $1`

	row := db.DB.QueryRow(query, idProjeto)
	err := row.Scan(
		&p.IdProjeto, &p.Titulo, &p.Descricao, &p.IdCurso, &p.IdArea,
		&p.SemestreLetivo, &p.ProfessorOrientador,
		&p.StatusProjeto, &p.ImagemCapa, &p.LinkRepositorio, &p.MediaEstrelas, &p.TotalAvaliacoes,
	)

	if err != nil {
		return p, err
	}

	// 4. Busca as Avaliações/Comentários do Projeto
	// Usamos TO_CHAR para o PostgreSQL já devolver a data bonitinha (ex: 05/04/2026)
	queryAvaliacoes := `
		SELECT nome_avaliador, nota, comentario, TO_CHAR(data_avaliacao, 'DD/MM/YYYY') 
		FROM avaliacoes 
		WHERE id_projeto = $1 
		ORDER BY data_avaliacao DESC`

	rowsAv, _ := db.DB.Query(queryAvaliacoes, idProjeto)
	for rowsAv.Next() {
		var av structs.Avaliacao // (ou models.Avaliacao)
		rowsAv.Scan(&av.NomeAvaliador, &av.Nota, &av.Comentario, &av.DataFormatada)
		p.Avaliacoes = append(p.Avaliacoes, av)
	}
	rowsAv.Close()

	return p, nil
}

// --- BUSCA O PROJETO COM TUDO DENTRO (Equipe, Links e Arquivos) ---
func BuscarProjetoCompletoPorId(idProjeto int) (structs.Projeto, error) {
	projeto, err := BuscarProjetoPorId(idProjeto) // Usa a função que já criamos!
	if err != nil {
		return projeto, err
	}

	// 1. Busca Equipe
	queryEquipe := `
		SELECT a.id_aluno, a.nome_completo, COALESCE(a.foto_perfil, ''), pa.funcao_no_projeto 
		FROM projeto_alunos pa
		JOIN alunos a ON pa.id_aluno = a.id_aluno
		WHERE pa.id_projeto = $1`
	rowsEq, _ := db.DB.Query(queryEquipe, idProjeto)
	for rowsEq.Next() {
		var membro structs.Aluno
		var funcao string
		rowsEq.Scan(&membro.IdAluno, &membro.NomeCompleto, &membro.FotoPerfil, &funcao)
		membro.Biografia = funcao // Usando o campo Biografia provisoriamente para exibir a função na tela
		projeto.Equipe = append(projeto.Equipe, membro)
	}
	rowsEq.Close()

	// 2. Busca Links Externos
	queryLinks := `SELECT id_link, tipo_link, url, COALESCE(descricao, '') FROM projeto_links WHERE id_projeto = $1`
	rowsLk, _ := db.DB.Query(queryLinks, idProjeto)
	for rowsLk.Next() {
		var link structs.ProjetoLink
		rowsLk.Scan(&link.IdLink, &link.TipoLink, &link.Url, &link.Descricao)
		projeto.Links = append(projeto.Links, link)
	}
	rowsLk.Close()

	// 3. Busca Arquivos PDF
	queryArquivos := `SELECT id_arquivo, nome_original, caminho_arquivo FROM projeto_arquivos WHERE id_projeto = $1`
	rowsArq, _ := db.DB.Query(queryArquivos, idProjeto)
	for rowsArq.Next() {
		var arq structs.ProjetoArquivo
		rowsArq.Scan(&arq.IdArquivo, &arq.NomeOriginal, &arq.CaminhoArquivo)
		projeto.Arquivos = append(projeto.Arquivos, arq)
	}
	rowsArq.Close()

	return projeto, nil
}

// --- FUNÇÕES DE EQUIPE ---
func AdicionarMembroEquipe(idProjeto, idAluno int, funcao string) error {
	_, err := db.DB.Exec(`INSERT INTO projeto_alunos (id_projeto, id_aluno, funcao_no_projeto) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, idProjeto, idAluno, funcao)
	return err
}

func RemoverMembroEquipe(idProjeto, idAluno int) error {
	_, err := db.DB.Exec(`DELETE FROM projeto_alunos WHERE id_projeto = $1 AND id_aluno = $2`, idProjeto, idAluno)
	return err
}

// --- FUNÇÕES DE ARQUIVOS (PDF) ---
func SalvarArquivoProjeto(idProjeto int, nomeOriginal, caminho string) error {
	_, err := db.DB.Exec(`INSERT INTO projeto_arquivos (id_projeto, nome_original, caminho_arquivo) VALUES ($1, $2, $3)`, idProjeto, nomeOriginal, caminho)
	return err
}

func RemoverArquivoProjeto(idArquivo int) error {
	_, err := db.DB.Exec(`DELETE FROM projeto_arquivos WHERE id_arquivo = $1`, idArquivo)
	return err
}

func RemoverLinkProjeto(idLink int) error {
	_, err := db.DB.Exec("DELETE FROM projeto_links WHERE id_link = $1", idLink)
	return err
}

// ListarProjetosPublicos traz os projetos para a vitrine aplicando os filtros de busca
func ListarProjetosPublicos(busca, curso, status string) ([]structs.Projeto, error) {
	query := `
		SELECT 
			p.id_projeto, p.titulo, p.descricao, p.status_projeto, 
			COALESCE(p.imagem_capa, ''), COALESCE(c.nome_curso, '')
		FROM projetos p
		LEFT JOIN cursos c ON p.id_curso = c.id_curso
		WHERE p.status_projeto != 'OCULTO' 
	`

	var args []interface{}
	argId := 1

	if busca != "" {
		query += ` AND (p.titulo ILIKE $` + strconv.Itoa(argId) + ` OR p.descricao ILIKE $` + strconv.Itoa(argId) + `)`
		args = append(args, "%"+busca+"%", "%"+busca+"%")
		argId += 2
	}
	if curso != "" {
		query += ` AND c.nome_curso ILIKE $` + strconv.Itoa(argId)
		args = append(args, "%"+curso+"%")
		argId++
	}
	if status != "" {
		query += ` AND p.status_projeto ILIKE $` + strconv.Itoa(argId)
		args = append(args, "%"+status+"%")
		argId++
	}

	query += ` ORDER BY p.id_projeto DESC`

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Projeto
	for rows.Next() {
		var p structs.Projeto
		var nomeCurso string
		// Usamos Scan para ler do banco
		rows.Scan(&p.IdProjeto, &p.Titulo, &p.Descricao, &p.StatusProjeto, &p.ImagemCapa, &nomeCurso)
		p.Curso = structs.Curso{NomeCurso: nomeCurso}
		lista = append(lista, p)
	}
	return lista, nil
}

// Conta o total de cursos cadastrados
func ObterTotalCursos() int {
	var total int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM cursos").Scan(&total)
	if err != nil {
		return 0
	}
	return total
}

// Conta o total de alunos na vitrine
func ObterTotalAlunos() int {
	var total int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM alunos").Scan(&total)
	if err != nil {
		return 0
	}
	return total
}

// Conta o total de projetos (que não estejam ocultos)
func ObterTotalProjetos() int {
	var total int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM projetos WHERE status_projeto != 'OCULTO'").Scan(&total)
	if err != nil {
		return 0
	}
	return total
}
