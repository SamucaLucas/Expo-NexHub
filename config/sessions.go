package config

import (
	"log"

	"github.com/gorilla/sessions"
)

var Store = sessions.NewCookieStore([]byte("27840835082784083508278408350801"))

// Esta função init() é executada automaticamente quando o pacote 'config' é usado pela primeira vez.
func init() {
	// Configurações importantes para os cookies da sessão
	Store.Options = &sessions.Options{
		Path:     "/",       // O cookie será válido para todo o site
		MaxAge:   86400 * 7, // Expira em 7 dias (em segundos)
		HttpOnly: true,      // O cookie não pode ser acessado via JavaScript (mais seguro)
		// SameSite: sessions.SameSiteLaxMode, // Descomente se tiver problemas com redirecionamentos de outros sites
	}
	log.Println("Configuração da Store de Sessão inicializada.") // Log para sabermos que isso executou
}
