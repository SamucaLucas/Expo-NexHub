package models

import (
	"nexhub/db"
	"nexhub/structs"
)

// CriarAluno insere um novo perfil na vitrine (ação executada pelo Admin)
func CriarAluno(a structs.Aluno) (int, error) {
	query := `
		INSERT INTO alunos (
			nome_completo, id_curso, semestre_atual, biografia, foto_perfil, 
			email_contato, linkedin_link, github_link, portfolio_link, cadastrado_por
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id_aluno`

	var idGerado int
	err := db.DB.QueryRow(query,
		a.NomeCompleto, a.IdCurso, a.SemestreAtual, a.Biografia, a.FotoPerfil,
		a.EmailContato, a.LinkedinLink, a.GithubLink, a.PortfolioLink, a.CadastradoPor,
	).Scan(&idGerado)

	return idGerado, err
}

// BuscarAlunoPorId traz os detalhes completos de um aluno específico para edição e perfil público
func BuscarAlunoPorID(id int) (structs.Aluno, error) {
	query := `
		SELECT 
			a.id_aluno, a.nome_completo, a.id_curso, a.semestre_atual, 
			COALESCE(a.biografia, ''), COALESCE(a.foto_perfil, ''), 
			COALESCE(a.email_contato, ''), COALESCE(a.linkedin_link, ''), 
			COALESCE(a.github_link, ''), COALESCE(a.portfolio_link, ''),
			COALESCE(c.nome_curso, '')
		FROM alunos a
		LEFT JOIN cursos c ON a.id_curso = c.id_curso
		WHERE a.id_aluno = $1`

	var a structs.Aluno
	var nomeCurso string

	err := db.DB.QueryRow(query, id).Scan(
		&a.IdAluno, &a.NomeCompleto, &a.IdCurso, &a.SemestreAtual,
		&a.Biografia, &a.FotoPerfil, &a.EmailContato, &a.LinkedinLink,
		&a.GithubLink, &a.PortfolioLink, &nomeCurso,
	)

	// Se não deu erro, preenche a struct aninhada do curso
	if err == nil {
		a.Curso = structs.Curso{NomeCurso: nomeCurso}
	}

	return a, err
}

// ListarAlunos traz um resumo de todos os alunos para a página "Talentos" e tabelas administrativas
func ListarAlunos() ([]structs.Aluno, error) {
	query := `
		SELECT 
			a.id_aluno, a.nome_completo, a.id_curso, a.semestre_atual, 
			COALESCE(a.foto_perfil, ''), COALESCE(a.biografia, ''), COALESCE(c.nome_curso, '')
		FROM alunos a
		LEFT JOIN cursos c ON a.id_curso = c.id_curso
		ORDER BY a.id_aluno DESC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alunos []structs.Aluno
	for rows.Next() {
		var a structs.Aluno
		var nomeCurso string

		err = rows.Scan(
			&a.IdAluno, &a.NomeCompleto, &a.IdCurso, &a.SemestreAtual,
			&a.FotoPerfil, &a.Biografia, &nomeCurso,
		)
		if err != nil {
			continue
		}
		a.Curso = structs.Curso{NomeCurso: nomeCurso}
		alunos = append(alunos, a)
	}

	return alunos, nil
}

// AtualizarAluno modifica APENAS OS DADOS TEXTUAIS do aluno
func AtualizarAluno(a structs.Aluno) error {
	query := `
		UPDATE alunos 
		SET nome_completo=$1, id_curso=$2, semestre_atual=$3, biografia=$4, 
		    email_contato=$5, linkedin_link=$6, github_link=$7, portfolio_link=$8
		WHERE id_aluno=$9`

	_, err := db.DB.Exec(query,
		a.NomeCompleto, a.IdCurso, a.SemestreAtual, a.Biografia,
		a.EmailContato, a.LinkedinLink, a.GithubLink, a.PortfolioLink,
		a.IdAluno,
	)
	return err
}

// AtualizarFotoAluno salva o caminho da nova imagem separadamente
func AtualizarFotoAluno(idAluno int, caminhoFoto string) error {
	query := `UPDATE alunos SET foto_perfil=$1 WHERE id_aluno=$2`
	_, err := db.DB.Exec(query, caminhoFoto, idAluno)
	return err
}

// DeletarAluno exclui o registro de um aluno da plataforma
func DeletarAluno(id int) error {
	query := `DELETE FROM alunos WHERE id_aluno = $1`
	_, err := db.DB.Exec(query, id)
	return err
}
