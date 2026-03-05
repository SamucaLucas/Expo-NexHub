package models

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"nexhub/db"
	"nexhub/services"
	"time"
)

// 1. Gera um código de 6 dígitos criptograficamente seguro
func GerarCodigoSeguro() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000)) // 0 a 999999
	if err != nil {
		return "", err
	}
	// Formata com zeros a esquerda (ex: 004512)
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// 2. Cria a solicitação no banco
func CriarSolicitacaoRecuperacao(email string) error {
	codigo, err := GerarCodigoSeguro()
	if err != nil {
		return err
	}

	// Expira em 15 minutos
	expiracao := time.Now().Add(15 * time.Minute)

	query := `INSERT INTO recuperacao_senha (email, codigo, expiracao) VALUES ($1, $2, $3)`
	_, err = db.DB.Exec(query, email, codigo, expiracao)

	if err == nil {
		// Envia o e-mail em uma goroutine para não travar o site
		go services.EnviarEmailRecuperacao(email, codigo)
	}

	return err
}

// 3. Valida se o código é válido, não expirou e não foi usado
func ValidarCodigoRecuperacao(email, codigo string) bool {
	var id int
	query := `
		SELECT id FROM recuperacao_senha 
		WHERE email = $1 AND codigo = $2 
		AND usado = FALSE 
		AND expiracao > NOW()
		ORDER BY id DESC LIMIT 1
	`
	err := db.DB.QueryRow(query, email, codigo).Scan(&id)
	return err == nil // Se achou ID, é válido
}

// 4. Marca o código como usado (após trocar a senha)
func MarcarCodigoComoUsado(email, codigo string) {
	db.DB.Exec("UPDATE recuperacao_senha SET usado = TRUE WHERE email = $1 AND codigo = $2", email, codigo)
}

// 5. Atualiza a senha do usuário
// 5. Atualiza a senha do usuário
func AtualizarSenhaPeloEmail(email, novaSenhaHash string) error {
    // Lembre-se de manter o nome da coluna correto (senha ou senha_hash)
    query := "UPDATE usuarios SET senha_hash = $1 WHERE email = $2"
    
    result, err := db.DB.Exec(query, novaSenhaHash, email)
    if err != nil {
        return err
    }

    // Opcional: Verifica se realmente achou alguém para atualizar
    linhas, _ := result.RowsAffected()
    if linhas == 0 {
        return fmt.Errorf("usuário não encontrado")
    }

    return nil
}
