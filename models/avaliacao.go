package models

import (
	"errors"
	"nexhub/db"
	"nexhub/structs"

	// Importando a biblioteca do AfterShip
	verifier "github.com/AfterShip/email-verifier"
)

// Inicializa o verificador globalmente
var (
	emailVerifier = verifier.NewVerifier().
		EnableSMTPCheck().    // Verifica se o usuário existe no servidor (SMTP)
		EnableDomainSuggest() // Sugere correções (ex: gmil.com -> gmail.com)
)

// NOVA VALIDAÇÃO RIGOROSA (USANDO AFTERSHIP)
func ValidarEmailRigoroso(email string) (string, error) {
	// A biblioteca faz todo o trabalho pesado
	ret, err := emailVerifier.Verify(email)
	if err != nil {
		return "", err
	}

	// 1. Sintaxe (ex: a@b)
	if !ret.Syntax.Valid {
		return "", errors.New("formato do e-mail é inválido")
	}

	// 2. DNS / MX (ex: teste@naoexiste123.com)
	if !ret.HasMxRecords {
		return "", errors.New("domínio do e-mail não existe")
	}

	// 3. SMTP (ex: usuario_fake@gmail.com)
	// Nota: Alguns provedores bloqueiam checagem SMTP via localhost,
	// mas funciona bem em servidores de produção.
	if !ret.SMTP.Deliverable {
		return "", errors.New("esta conta de e-mail não existe ou está cheia")
	}

	// 4. Se for descartável (opcional, bom para evitar spam)
	if ret.Disposable {
		return "", errors.New("e-mails temporários não são permitidos")
	}

	// Retorna sugestão se houver (ex: did you mean gmail.com?)
	return ret.Suggestion, nil
}

func BuscarAvaliacoesDoProjeto(idProjeto int) ([]structs.Avaliacao, error) {
	// Query: Busca e-mail, nota e comentário
	// Ajuste "avaliacoes" para o nome real da sua tabela se for diferente
	query := `
		SELECT email_avaliador, nota, comentario, nome_avaliador 
		FROM avaliacoes 
		WHERE id_projeto = $1 
		ORDER BY id_avaliacao DESC
	`
	rows, err := db.DB.Query(query, idProjeto)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lista []structs.Avaliacao
	for rows.Next() {
		var a structs.Avaliacao
		// O Scan deve bater com a struct e a ordem do SELECT
		if err := rows.Scan(&a.Email, &a.Nota, &a.Comentario, &a.NomeAvaliador); err != nil {
			continue
		}
		lista = append(lista, a)
	}
	return lista, nil
}

// 2. FUNÇÃO SALVAR (Mantenha simples para o controller tratar o erro)
func SalvarAvaliacao(a structs.Avaliacao) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	queryInsert := `
		INSERT INTO avaliacoes (id_projeto, email_avaliador, nota, comentario, nome_avaliador)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(queryInsert, a.IdProjeto, a.Email, a.Nota, a.Comentario, a.NomeAvaliador)
	if err != nil {
		tx.Rollback()
		return err // Retorna o erro ORIGINAL do banco (ex: duplicate key)
	}

	queryUpdate := `
		UPDATE projetos 
		SET 
			media_estrelas = (SELECT COALESCE(AVG(nota), 0) FROM avaliacoes WHERE id_projeto = $1),
			total_avaliacoes = (SELECT COUNT(*) FROM avaliacoes WHERE id_projeto = $1)
		WHERE id_projeto = $1
	`
	_, err = tx.Exec(queryUpdate, a.IdProjeto)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
