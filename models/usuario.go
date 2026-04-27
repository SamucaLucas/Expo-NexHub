package models

import (
	"nexhub/db"
	"nexhub/structs"
	"time"

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
	query := `
		SELECT id_usuario, nome_completo, email, senha_hash, id_curso_analista, foto_perfil 
		FROM usuarios 
		WHERE id_usuario = $1`

	row := db.DB.QueryRow(query, id)

	var fotoPerfil *string
	// Agora pegando o &u.SenhaHash na ordem certa do SELECT
	err := row.Scan(&u.IdUsuario, &u.NomeCompleto, &u.Email, &u.SenhaHash, &u.IdCursoAnalista, &fotoPerfil)

	if fotoPerfil != nil {
		u.FotoPerfil = *fotoPerfil
	}

	return u, err
}

// AtualizarPerfil atualiza os dados básicos do Admin
func AtualizarPerfil(u structs.Usuario) error {
	query := `
        UPDATE usuarios 
        SET nome_completo=$1, email=$2, senha_hash=$3, foto_perfil=$4, id_curso_analista = $5
        WHERE id_usuario=$6
    `
	_, err := db.DB.Exec(query,
		u.NomeCompleto,
		u.Email,
		u.SenhaHash,
		u.FotoPerfil,
		u.IdCursoAnalista,
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
// ListarTodosCursos busca todos os cursos cadastrados para o filtro da vitrine
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
		// Verifica se a struct Curso no seu types.go tem esses exatos campos.
		// Se forem diferentes, ajuste aqui (ex: c.ID, c.Nome)
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
	FotoPerfil       string // DE VOLTA!
}

func ListarAnalistasParaSobre() ([]AnalistaCard, error) {
	query := `
		SELECT u.nome_completo, u.email, COALESCE(u.foto_perfil, ''), COALESCE(c.nome_curso, 'Geral (Todos os Cursos)') as curso_responsavel
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
		if err := rows.Scan(&a.NomeCompleto, &a.Email, &a.FotoPerfil, &a.CursoResponsavel); err == nil {
			lista = append(lista, a)
		}
	}
	return lista, nil
}

type AnalistaAdmin struct {
	IdUsuario        int
	NomeCompleto     string
	Email            string
	FotoPerfil       string
	CursoResponsavel string
	DataCadastro     time.Time
}

// ListarTodosAnalistas busca a lista de quem tem acesso ao painel
func ListarTodosAnalistas() ([]AnalistaAdmin, error) {
	query := `
		SELECT u.id_usuario, u.nome_completo, u.email, COALESCE(u.foto_perfil, ''), COALESCE(c.nome_curso, 'Analista Geral (Admin)'), u.data_cadastro
		FROM usuarios u
		LEFT JOIN cursos c ON u.id_curso_analista = c.id_curso
		ORDER BY u.nome_completo ASC`

	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []AnalistaAdmin
	for rows.Next() {
		var a AnalistaAdmin
		if err := rows.Scan(&a.IdUsuario, &a.NomeCompleto, &a.Email, &a.FotoPerfil, &a.CursoResponsavel, &a.DataCadastro); err == nil {
			lista = append(lista, a)
		}
	}
	return lista, nil
}

// DeletarAnalista exclui um acesso do sistema
func DeletarAnalista(id int) error {
	_, err := db.DB.Exec("DELETE FROM usuarios WHERE id_usuario = $1", id)
	return err
}
