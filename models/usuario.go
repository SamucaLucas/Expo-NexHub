package models

import (
	"nexhub/db"
	"nexhub/structs"

	"golang.org/x/crypto/bcrypt"
)

// CriarUsuario cadastra um novo Analista/Admin do sistema (apenas acesso)
func CriarUsuario(u structs.Usuario) error {
	senhaCriptografada, err := bcrypt.GenerateFromPassword([]byte(u.SenhaHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Atualizado com id_curso_analista
	sqlStatement := `
		INSERT INTO usuarios (nome_completo, email, senha_hash, id_curso_analista)
		VALUES ($1, $2, $3, $4)`

	_, err = db.DB.Exec(sqlStatement, u.NomeCompleto, u.Email, string(senhaCriptografada), u.IdCursoAnalista)
	return err
}

func LoginUsuario(email, senha string) (structs.Usuario, error) {
	var u structs.Usuario
	// Atualizado com id_curso_analista
	query := `SELECT id_usuario, nome_completo, email, senha_hash, id_curso_analista 
			  FROM usuarios WHERE email = $1`

	row := db.DB.QueryRow(query, email)
	err := row.Scan(&u.IdUsuario, &u.NomeCompleto, &u.Email, &u.SenhaHash, &u.IdCursoAnalista)

	if err != nil {
		return u, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(senha))
	if err != nil {
		return u, err
	}

	u.SenhaHash = ""
	return u, nil
}

// VerificarEmailExiste checa rapidamente se o email está em uso (usado no cadastro)
func VerificarEmailExiste(email string) bool {
	var existe bool
	query := `SELECT EXISTS(SELECT 1 FROM usuarios WHERE email = $1)`

	err := db.DB.QueryRow(query, email).Scan(&existe)
	if err != nil {
		return false
	}
	return existe
}

// BuscarUsuarioPorID retorna os dados do Admin logado
func BuscarUsuarioPorID(id int) (structs.Usuario, error) {
	var u structs.Usuario

	sqlStatement := `
        SELECT id_usuario, nome_completo, email 
        FROM usuarios WHERE id_usuario = $1`

	row := db.DB.QueryRow(sqlStatement, id)
	err := row.Scan(&u.IdUsuario, &u.NomeCompleto, &u.Email)

	return u, err
}

// AtualizarPerfil atualiza os dados básicos do Admin
func AtualizarPerfil(u structs.Usuario) error {
	query := `
        UPDATE usuarios 
        SET nome_completo=$1, email=$2, senha_hash=$3
        WHERE id_usuario=$4
    `
	_, err := db.DB.Exec(query,
		u.NomeCompleto,
		u.Email,
		u.SenhaHash, // Assume que o controller já criptografou se foi alterada
		u.IdUsuario,
	)
	return err
}

// BuscarUsuarioPorEmail busca para redefinição de senha
func BuscarUsuarioPorEmail(email string) (structs.Usuario, error) {
	query := `SELECT id_usuario, nome_completo, email, senha_hash 
              FROM usuarios WHERE email = $1`

	var u structs.Usuario
	err := db.DB.QueryRow(query, email).Scan(&u.IdUsuario, &u.NomeCompleto, &u.Email, &u.SenhaHash)

	return u, err
}

// EmailJaCadastrado verifica duplicidade
func EmailJaCadastrado(email string) bool {
	var id int
	query := `SELECT id_usuario FROM usuarios WHERE email = $1 LIMIT 1`
	err := db.DB.QueryRow(query, email).Scan(&id)

	// Se err for nil, achou o id
	return err == nil
}

// ListarTodosCursos busca a lista para popular o formulário de cadastro
func ListarTodosCursos() ([]structs.Curso, error) {
	query := `SELECT id_curso, nome_curso FROM cursos ORDER BY nome_curso ASC`
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cursos []structs.Curso
	for rows.Next() {
		var c structs.Curso
		if err := rows.Scan(&c.IdCurso, &c.NomeCurso); err == nil {
			cursos = append(cursos, c)
		}
	}
	return cursos, nil
}

// --- ESTRUTURA AUXILIAR PARA A PÁGINA SOBRE ---
type AnalistaCard struct {
	NomeCompleto     string
	Email            string
	CursoResponsavel string
}

// ListarAnalistasParaSobre busca os admins e o curso que eles atendem
func ListarAnalistasParaSobre() ([]AnalistaCard, error) {
	query := `
		SELECT u.nome_completo, u.email, COALESCE(c.nome_curso, 'Geral (Todos os Cursos)') as curso_responsavel
		FROM usuarios u
		LEFT JOIN cursos c ON u.id_curso_analista = c.id_curso
		ORDER BY c.nome_curso ASC, u.nome_completo ASC`
		
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []AnalistaCard
	for rows.Next() {
		var a AnalistaCard
		if err := rows.Scan(&a.NomeCompleto, &a.Email, &a.CursoResponsavel); err == nil {
			lista = append(lista, a)
		}
	}
	return lista, nil
}