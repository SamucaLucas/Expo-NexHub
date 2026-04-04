package controllers

import (
	"log"
	"net/http"
	"nexhub/config"
	"nexhub/models"
	"nexhub/structs"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// ==========================================
// 1. LOGIN
// ==========================================

func Cadastro(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		cursos, _ := models.ListarTodosCursos()

		dados := struct {
			Cursos []structs.Curso
			Erro   string
		}{
			Cursos: cursos,
			Erro:   r.URL.Query().Get("erro"),
		}

		temp.ExecuteTemplate(w, "Cadastro", dados)
		return
	}

	if r.Method == "POST" {
		nome := r.FormValue("nome_completo")
		email := r.FormValue("email")
		senha := r.FormValue("senha")
		cursoStr := r.FormValue("id_curso_analista") // Pega a escolha do select

		if nome == "" || email == "" || senha == "" {
			http.Redirect(w, r, "/cadastro?erro=CamposVazios", http.StatusSeeOther)
			return
		}

		if models.VerificarEmailExiste(email) {
			http.Redirect(w, r, "/cadastro?erro=EmailEmUso", http.StatusSeeOther)
			return
		}

		// Lógica do Curso Analista
		var idCursoAnalista *int
		if cursoStr != "" && cursoStr != "geral" {
			id, err := strconv.Atoi(cursoStr)
			if err == nil {
				idCursoAnalista = &id
			}
		}

		usuario := structs.Usuario{
			NomeCompleto:    nome,
			Email:           email,
			SenhaHash:       senha,
			IdCursoAnalista: idCursoAnalista,
		}

		err := models.CriarUsuario(usuario)
		if err != nil {
			log.Println("Erro ao criar admin:", err)
			http.Redirect(w, r, "/cadastro?erro=Banco", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/login?sucesso=CadastroRealizado", http.StatusSeeOther)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		temp.ExecuteTemplate(w, "Login", nil)
		return
	}

	if r.Method == "POST" {
		email := r.FormValue("email")
		senha := r.FormValue("senha")

		usuario, err := models.LoginUsuario(email, senha)
		if err != nil {
			http.Redirect(w, r, "/login?erro=CredenciaisInvalidas", http.StatusSeeOther)
			return
		}

		session, _ := config.Store.Get(r, "nexhub-session")
		session.Values["user_id"] = usuario.IdUsuario
		session.Values["user_tipo"] = "ADMIN"

		// Guarda o curso na sessão (0 = Geral)
		if usuario.IdCursoAnalista != nil {
			session.Values["curso_analista"] = *usuario.IdCursoAnalista
		} else {
			session.Values["curso_analista"] = 0
		}

		session.Save(r, w)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	}
}

// ==========================================
// 3. LOGOUT
// ==========================================

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := config.Store.Get(r, "nexhub-session")
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ==========================================
// 4. RECUPERAÇÃO DE SENHA
// ==========================================

func EsqueciSenhaPage(w http.ResponseWriter, r *http.Request) {
	temp.ExecuteTemplate(w, "EsqueciSenha", nil)
}

func SolicitarResetHandler(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))

	if email == "" {
		http.Redirect(w, r, "/esqueci-senha?erro=EmailVazio", http.StatusSeeOther)
		return
	}

	if !models.VerificarEmailExiste(email) {
		http.Redirect(w, r, "/recuperar/codigo?email="+email, http.StatusSeeOther)
		return
	}

	err := models.CriarSolicitacaoRecuperacao(email)
	if err != nil {
		log.Println("Erro ao criar solicitação de recuperação:", err)
		http.Redirect(w, r, "/esqueci-senha?erro=ErroInterno", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/recuperar/codigo?email="+email, http.StatusSeeOther)
}

func ValidarCodigoPage(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	dados := struct{ Email string }{Email: email}
	temp.ExecuteTemplate(w, "ValidarCodigo", dados)
}

func VerificarCodigoHandler(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	codigo := strings.TrimSpace(r.FormValue("codigo"))

	valido := models.ValidarCodigoRecuperacao(email, codigo)

	if !valido {
		http.Redirect(w, r, "/recuperar/codigo?email="+email+"&erro=CodigoInvalido", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/recuperar/nova-senha?email="+email+"&code="+codigo, http.StatusSeeOther)
}

func NovaSenhaPage(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	codigo := r.URL.Query().Get("code")

	if !models.ValidarCodigoRecuperacao(email, codigo) {
		http.Redirect(w, r, "/login?erro=AcessoNegado", http.StatusSeeOther)
		return
	}

	dados := struct {
		Email  string
		Codigo string
	}{Email: email, Codigo: codigo}

	temp.ExecuteTemplate(w, "NovaSenha", dados)
}

func SalvarNovaSenhaHandler(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	codigo := strings.TrimSpace(r.FormValue("codigo"))
	novaSenha := r.FormValue("nova_senha")
	confirmarSenha := r.FormValue("confirmar_senha")

	if novaSenha != confirmarSenha {
		http.Redirect(w, r, "/recuperar/nova-senha?email="+email+"&code="+codigo+"&erro=SenhasNaoConferem", http.StatusSeeOther)
		return
	}

	if !models.ValidarCodigoRecuperacao(email, codigo) {
		http.Redirect(w, r, "/login?erro=Expirou", http.StatusSeeOther)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erro ao criptografar senha", 500)
		return
	}

	err = models.AtualizarSenhaPeloEmail(email, string(hash))
	if err != nil {
		log.Println("Erro ao salvar nova senha:", err)
		http.Redirect(w, r, "/login?erro=ErroBanco", http.StatusSeeOther)
		return
	}

	models.MarcarCodigoComoUsado(email, codigo)
	http.Redirect(w, r, "/login?sucesso=SenhaAlterada", http.StatusSeeOther)
}
