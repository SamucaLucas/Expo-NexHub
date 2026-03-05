package models

import (
	"fmt"
	"log"
	"nexhub/db"
	"nexhub/structs"

	"golang.org/x/crypto/bcrypt"
)

// Busca totais e dados para o gráfico
func BuscarEstatisticasAdmin() (int, int, int, int, []string, []int, error) {
	var devs, empresas, projetos, banidos int

	queryTotais := `
        SELECT 
            (SELECT COUNT(*) FROM usuarios WHERE tipo_usuario = 'DEV' AND is_banned = FALSE),
            (SELECT COUNT(*) FROM usuarios WHERE tipo_usuario = 'EMPRESA' AND is_banned = FALSE),
            (SELECT COUNT(*) FROM projetos),
            (SELECT COUNT(*) FROM usuarios WHERE is_banned = TRUE)
    `

	// Se seu banco não suportar subselect no select assim (algumas versões antigas de SQL),
	// mantenha suas queries separadas. Mas no Postgres moderno isso funciona bem.
	err := db.DB.QueryRow(queryTotais).Scan(&devs, &empresas, &projetos, &banidos)
	if err != nil {
		log.Println("Erro ao buscar totais:", err)
		return 0, 0, 0, 0, nil, nil, err
	}

	// 2. DADOS DO GRÁFICO (Crescimento últimos 6 meses)
	// Esta query agrupa por mês e conta quantos usuários foram criados
	queryGrafico := `
		SELECT TO_CHAR(data_cadastro, 'Mon') as mes, COUNT(*) 
		FROM usuarios 
		WHERE data_cadastro > NOW() - INTERVAL '6 months'
		GROUP BY TO_CHAR(data_cadastro, 'Mon'), DATE_TRUNC('month', data_cadastro)
		ORDER BY DATE_TRUNC('month', data_cadastro)
	`

	rows, err := db.DB.Query(queryGrafico)
	if err != nil {
		log.Println("Erro no gráfico:", err)
		// Retorna os totais mesmo se o gráfico falhar
		return devs, empresas, projetos, banidos, []string{}, []int{}, nil
	}
	defer rows.Close()

	var meses []string
	var novosUsers []int

	for rows.Next() {
		var mes string
		var qtd int
		if err := rows.Scan(&mes, &qtd); err == nil {
			meses = append(meses, mes)
			novosUsers = append(novosUsers, qtd)
		}
	}

	return devs, empresas, projetos, banidos, meses, novosUsers, nil
}

// ListarTodosUsuarios busca todos para a tabela do admin
func ListarTodosUsuarios(filtro string) ([]structs.Usuario, error) {
	query := `
        SELECT id_usuario, nome_completo, email, tipo_usuario, 
               COALESCE(cidade, '-'), 
               COALESCE(foto_perfil, ''),
               is_banned  -- <--- Trazendo do banco
        FROM usuarios
        WHERE 1=1
    `
	var args []interface{}

	if filtro != "" {
		query += " AND (nome_completo ILIKE $1 OR email ILIKE $1)"
		args = append(args, "%"+filtro+"%")
	}

	query += " ORDER BY id_usuario DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []structs.Usuario
	for rows.Next() {
		var u structs.Usuario
		// Adicionei &u.IsBanned no final do Scan
		if err := rows.Scan(&u.Id, &u.NomeCompleto, &u.Email, &u.TipoUsuario, &u.Cidade, &u.FotoPerfil, &u.IsBanned); err != nil {
			continue
		}
		usuarios = append(usuarios, u)
	}
	return usuarios, nil
}

func TornarUsuarioAdmin(idUsuario int) error {
	// Adicionei "AND tipo_usuario = 'DEV'" para travar empresas
	_, err := db.DB.Exec("UPDATE usuarios SET tipo_usuario = 'ADMIN' WHERE id_usuario = $1 AND tipo_usuario = 'DEV'", idUsuario)
	return err
}

// RemoverAdmin (Mantém como está, pois agora sabemos que todo Admin era um Dev antes)
func RemoverAdmin(idUsuario int) error {
	_, err := db.DB.Exec("UPDATE usuarios SET tipo_usuario = 'DEV' WHERE id_usuario = $1", idUsuario)
	return err
}

// 3. Função para Banir/Desbanir (Toggle)
// Se for TRUE vira FALSE, se for FALSE vira TRUE
func AlternarBanimento(idUsuario int) error {
	_, err := db.DB.Exec("UPDATE usuarios SET is_banned = NOT is_banned WHERE id_usuario = $1", idUsuario)
	return err
}

func ListarTodosProjetos(busca string, status string) ([]structs.ProjetoAdmin, error) {

	query := `
        SELECT 
            p.id_projeto, 
            COALESCE(p.titulo, 'Sem Título'), 
            COALESCE(u.nome_completo, 'Desconhecido'), 
            COALESCE(p.categoria, '-'), 
            COALESCE(p.status_projeto, 'Em Andamento')
        FROM projetos p
        LEFT JOIN usuarios u ON p.id_lider = u.id_usuario
        WHERE 1=1
    `
	var args []interface{}
	counter := 1

	if busca != "" {
		query += fmt.Sprintf(" AND p.titulo ILIKE $%d", counter)
		args = append(args, "%"+busca+"%")
		counter++
	}

	if status != "" {
		query += fmt.Sprintf(" AND p.status_projeto = $%d", counter)
		args = append(args, status)
		counter++
	}

	query += " ORDER BY p.id_projeto DESC"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		log.Println("ERRO SQL (Query Projetos):", err) // <--- Log para ver no terminal
		return nil, err
	}
	defer rows.Close()

	var projetos []structs.ProjetoAdmin
	for rows.Next() {
		var p structs.ProjetoAdmin
		// A ordem aqui deve ser EXATAMENTE a mesma do SELECT
		if err := rows.Scan(&p.Id, &p.Titulo, &p.DonoNome, &p.Categoria, &p.Status); err != nil {
			log.Println("ERRO SCAN (Linha Projeto):", err) // <--- Log se falhar ao ler linha
			continue
		}
		projetos = append(projetos, p)
	}

	return projetos, nil
}

// AlterarStatusProjeto muda o status (ex: para "Oculto" ou "Aprovado")
func AlterarStatusProjeto(id int, novoStatus string) error {
	_, err := db.DB.Exec("UPDATE projetos SET status_projeto = $1 WHERE id_projeto = $2", novoStatus, id)
	return err
}

// ExcluirProjeto remove permanentemente (Cuidado!)
func ExcluirProjeto(id int) error {
	_, err := db.DB.Exec("DELETE FROM projetos WHERE id_projeto = $1", id)
	return err
}

// AtualizarPerfil salva as alterações do usuário
func AtualizarPerfilAdmin(id int, nome, email, senhaPura, fotoPath string) error {

	// CENÁRIO 1: Usuário NÃO trocou a senha (senhaPura vazia)
	if senhaPura == "" {
		query := `
            UPDATE usuarios 
            SET nome_completo = $1, email = $2, foto_perfil = $3 
            WHERE id_usuario = $4
        `
		_, err := db.DB.Exec(query, nome, email, fotoPath, id)
		return err
	}

	// CENÁRIO 2: Usuário TROCOU a senha
	// 1. Gera o Hash aqui no Model
	bytes, err := bcrypt.GenerateFromPassword([]byte(senhaPura), 14)
	if err != nil {
		return err
	}
	novoHash := string(bytes)

	// 2. Atualiza tudo, incluindo a senha
	query := `
        UPDATE usuarios 
        SET nome_completo = $1, email = $2, foto_perfil = $3, senha_hash = $4
        WHERE id_usuario = $5
    `
	_, err = db.DB.Exec(query, nome, email, fotoPath, novoHash, id)
	return err
}
