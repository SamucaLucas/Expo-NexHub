package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
)

// getChaveSecreta busca a chave dinamicamente.
// Como ela é chamada apenas na hora de enviar/receber a mensagem,
// garante que o .env já foi carregado pelo main().
func getChaveSecreta() []byte {
	chave := os.Getenv("SECRET_KEY")

	// AES-256 EXIGE exatos 32 caracteres.
	// Se a chave no .env estiver ausente ou com tamanho errado, usamos uma de backup.
	if len(chave) != 32 {
		return []byte("NexHub-Chave-Secreta-32-Caractrs") // Exatos 32 caracteres
	}

	return []byte(chave)
}

// CriptografarMensagem transforma texto plano em um código ilegível (Hex)
func CriptografarMensagem(texto string) (string, error) {
	if texto == "" {
		return "", nil
	}

	// Pega a chave na hora certa!
	chaveSecreta := getChaveSecreta()

	bloco, err := aes.NewCipher(chaveSecreta)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(bloco)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(texto), nil)

	// Retornamos em formato Hexadecimal para salvar fácil no Banco de Dados (como string)
	return hex.EncodeToString(ciphertext), nil
}

// DescriptografarMensagem reverte o código Hex para o texto original
func DescriptografarMensagem(textoHex string) string {
	if textoHex == "" {
		return ""
	}

	// Tenta converter de Hex para Bytes
	dados, err := hex.DecodeString(textoHex)
	if err != nil {
		// TRUQUE DE MESTRE: Se der erro, significa que é uma mensagem antiga
		// que não estava criptografada. Então retornamos ela mesma!
		return textoHex
	}

	// Pega a chave na hora certa!
	chaveSecreta := getChaveSecreta()

	bloco, err := aes.NewCipher(chaveSecreta)
	if err != nil {
		return textoHex
	}

	gcm, err := cipher.NewGCM(bloco)
	if err != nil {
		return textoHex
	}

	nonceSize := gcm.NonceSize()
	if len(dados) < nonceSize {
		return textoHex
	}

	nonce, ciphertext := dados[:nonceSize], dados[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// Se a senha estiver errada, devolve o texto cifrado para não quebrar a tela
		return textoHex
	}

	return string(plaintext)
}
