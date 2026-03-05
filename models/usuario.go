package models

import (
	"nexhub/db"
	"nexhub/structs"

	"golang.org/x/crypto/bcrypt"
)

// CriarUsuario recebe um usuário preenchido, gera o Hash da senha e salva no banco
func CriarUsuario(u structs.Usuario) error {
	// 1. Criptografar a senha (Nunca salvar senha pura!)
	senhaCriptografada, err := bcrypt.GenerateFromPassword([]byte(u.SenhaHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 2. Preparar o SQL de Inserção
	// Inserimos os dados básicos do cadastro inicial
	sqlStatement := `
		INSERT INTO usuarios (nome_completo, email, senha_hash, tipo_usuario, cidade, status_conta, nivel_profissional, nome_fantasia)
		VALUES ($1, $2, $3, $4, $5, 'ATIVO', $6, $7)`

	// 3. Executar no Banco
	_, err = db.DB.Exec(sqlStatement, u.NomeCompleto, u.Email, string(senhaCriptografada), u.TipoUsuario, u.Cidade, u.Nivel, u.NomeFantasia)
	if err != nil {
		return err
	}

	return nil
}

// LoginUsuario busca um usuário pelo email e verifica se a senha bate
func LoginUsuario(email, senha string) (structs.Usuario, error) {
	var u structs.Usuario

	// 1. Buscar o usuário pelo E-mail
	query := `SELECT id_usuario, nome_completo, email, senha_hash, tipo_usuario, is_banned 
			FROM usuarios WHERE email = $1`

	row := db.DB.QueryRow(query, email)
	err := row.Scan(&u.Id, &u.NomeCompleto, &u.Email, &u.SenhaHash, &u.TipoUsuario, &u.IsBanned)

	if err != nil {
		return u, err // Retorna erro se não achar o email
	}

	// 2. Verificar a Senha (Comparar o Hash do banco com a senha digitada)
	err = bcrypt.CompareHashAndPassword([]byte(u.SenhaHash), []byte(senha))
	if err != nil {
		return u, err // Retorna erro se a senha estiver errada
	}

	// 3. Sucesso! Retorna os dados do usuário (sem a senha, por segurança)
	u.SenhaHash = ""
	return u, nil
}

func VerificarEmailExiste(email string) bool {
	var existe bool
	// SELECT EXISTS retorna true ou false direto do banco, é muito rápido
	query := `SELECT EXISTS(SELECT 1 FROM usuarios WHERE email = $1)`

	err := db.DB.QueryRow(query, email).Scan(&existe)
	if err != nil {
		return false // Se der erro no banco, assumimos que não existe para não travar
	}
	return existe
}

func BuscarUsuarioPorID(id int) (structs.Usuario, error) {
	var u structs.Usuario

	// ADICIONEI: foto_perfil
	// ADICIONEI: COALESCE em cidade e tipo_usuario (segurança contra NULL)
	sqlStatement := `
        SELECT id_usuario, nome_completo, email, 
               COALESCE(cidade, ''), COALESCE(tipo_usuario, ''), 
               COALESCE(biografia, ''), COALESCE(titulo_profissional, ''), 
               COALESCE(github_link, ''), COALESCE(linkedin_link, ''), 
               disponivel_para_trabalho, COALESCE(skills, ''),
               COALESCE(foto_perfil, ''), senha_hash, 
			   COALESCE(nome_fantasia, ''), 
               COALESCE(site_empresa, ''), 
               COALESCE(ramo_atuacao, ''),
			   COALESCE(nivel_profissional, ''), 
			   is_banned
        FROM usuarios WHERE id_usuario = $1`

	row := db.DB.QueryRow(sqlStatement, id)

	err := row.Scan(
		&u.Id,
		&u.NomeCompleto,
		&u.Email,
		&u.Cidade,
		&u.TipoUsuario,
		&u.Biografia,
		&u.TituloProfissional,
		&u.GithubLink,
		&u.LinkedinLink,
		&u.DisponivelParaEquipes,
		&u.Skills,
		&u.FotoPerfil,
		&u.SenhaHash,
		&u.NomeFantasia,
		&u.SiteEmpresa,
		&u.RamoAtuacao,
		&u.Nivel,
		&u.IsBanned,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func AtualizarPerfil(u structs.Usuario) error {
	// ADICIONEI: senha_hash=$11 e mudei o WHERE para $12
	query := `
        UPDATE usuarios 
        SET nome_completo=$1, titulo_profissional=$2, cidade=$3, 
            disponivel_para_trabalho=$4, biografia=$5, skills=$6, 
            github_link=$7, linkedin_link=$8, email=$9, foto_perfil=$10,
            senha_hash=$11, nivel_profissional = $12
        WHERE id_usuario=$13
    `
	_, err := db.DB.Exec(query,
		u.NomeCompleto, u.TituloProfissional, u.Cidade,
		u.DisponivelParaEquipes, u.Biografia, u.Skills,
		u.GithubLink, u.LinkedinLink, u.Email, u.FotoPerfil,
		u.SenhaHash, u.Nivel, // Passamos a senha (nova ou velha)
		u.Id,
	)
	return err
}

// Busca usuário apenas pelo email para validar senha no controller
func BuscarUsuarioPorEmail(email string) (structs.Usuario, error) {
	// ATENÇÃO: O SQL precisa selecionar o campo senha_hash
	query := `SELECT id_usuario, nome_completo, email, senha_hash, tipo_usuario, is_banned 
              FROM usuarios WHERE email = $1`

	var u structs.Usuario
	// ATENÇÃO: A ordem do Scan deve bater com o SELECT acima
	// Verifique se u.SenhaHash está recebendo o valor do banco
	err := db.DB.QueryRow(query, email).Scan(&u.Id, &u.NomeCompleto, &u.Email, &u.SenhaHash, &u.TipoUsuario, &u.IsBanned)

	if err != nil {
		return u, err
	}
	return u, nil
}

// Verifica se o email já existe no banco
func EmailJaCadastrado(email string) bool {
	var id int
	query := `SELECT id_usuario FROM usuarios WHERE email = $1 LIMIT 1`
	err := db.DB.QueryRow(query, email).Scan(&id)
	
	// Se err for nil, significa que ACHOU um id (então existe)
	// Se der erro (sql.ErrNoRows), não existe
	return err == nil 
}