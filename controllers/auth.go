package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nexhub/config"
	"nexhub/models"
	"nexhub/structs"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Cadastro gerencia a página de criar conta
func Cadastro(w http.ResponseWriter, r *http.Request) {

	// --- CENÁRIO 1: Acessar a Página (GET) ---
	if r.Method == "GET" {
		err := temp.ExecuteTemplate(w, "Cadastro", nil)
		if err != nil {
			log.Println("Erro ao renderizar tela de cadastro:", err)
		}
		return
	}

	// --- CENÁRIO 2: Enviar o Formulário (POST) ---
	if r.Method == "POST" {
		// 1. Coleta os dados do HTML
		nome := r.FormValue("nome_completo")
		email := r.FormValue("email")
		senha := r.FormValue("senha")
		cidade := r.FormValue("cidade")
		tipo := r.FormValue("tipo_usuario")
		nomeFantasia := r.FormValue("nome_fantasia")

		// 2. Validação simples (Backend)
		if nome == "" || email == "" || senha == "" || tipo == "" {
			dados := struct{ Erro string }{Erro: "Por favor, preencha todos os campos obrigatórios."}
			temp.ExecuteTemplate(w, "Cadastro", dados)
			return
		}

		strings.ToLower(strings.TrimSpace(email))

		// 3. Validação Rigorosa de E-mail (MX, SMTP, etc)
		// Declaração inicial de 'err'
		_, err := models.ValidarEmailRigoroso(email)
		if err != nil {
			dados := struct{ Erro string }{Erro: err.Error()}
			temp.ExecuteTemplate(w, "Cadastro", dados)
			return
		}

		// 4. Verifica se o E-mail já existe no Banco
		// (Ajustei o nome da função para o padrão que criamos: EmailJaCadastrado)
		if models.EmailJaCadastrado(email) {
			dados := struct{ Erro string }{Erro: "Este email já está cadastrado. Tente fazer login."}
			temp.ExecuteTemplate(w, "Cadastro", dados)
			return
		}

		// Define nível
		nivel_profissional := ""
		if tipo == "DEV" {
			nivel_profissional = "Junior"
		} else {
			nivel_profissional = "Empresa"
		}

		// 5. Monta a estrutura de dados
		// (Ajustei de 'structs.Usuario' para 'models.Usuario', verifique onde está sua struct)
		novoUsuario := structs.Usuario{
			NomeCompleto: nome, // Verifique se no model é 'Nome' ou 'NomeCompleto'
			Email:        email,
			SenhaHash:    senha, // A senha será criptografada dentro de CriarUsuario? Se não, faça o hash aqui.
			Cidade:       cidade,
			TipoUsuario:  tipo,
			StatusConta:  "ATIVO", // Se seu model tiver esses campos, descomente
			Nivel:        nivel_profissional,
			NomeFantasia: nomeFantasia,
		}

		// 6. Salva no Banco de Dados
		// CORREÇÃO AQUI: Usamos '=' em vez de ':=' porque 'err' já existe
		err = models.CriarUsuario(novoUsuario)
		if err != nil {
			log.Println("Erro ao cadastrar usuário:", err)
			// Retorna o erro visualmente no template em vez de uma tela branca de erro
			dados := struct{ Erro string }{Erro: "Erro interno ao criar conta. Tente novamente."}
			temp.ExecuteTemplate(w, "Cadastro", dados)
			return
		}

		// 7. Sucesso! Redireciona para o Login
		http.Redirect(w, r, "/login?sucesso=ContaCriada", http.StatusSeeOther)
	}
}

func VerificarEmailDisponivel(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	// Prepara a resposta padrão (Otimista: tudo certo)
	resposta := map[string]interface{}{
		"valido":     true,
		"disponivel": true,
		"mensagem":   "E-mail válido e disponível!",
	}

	// 1. VERIFICAÇÃO DE BANCO (Já existe usuário com esse email?)
	if models.EmailJaCadastrado(email) {
		resposta["disponivel"] = false
		resposta["mensagem"] = "Este e-mail já está cadastrado no sistema."

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resposta)
		return
	}

	// 2. VERIFICAÇÃO RIGOROSA (Usa sua função do Model)
	// A função ValidarEmailRigoroso retorna erro se algo estiver errado (MX, SMTP, Syntax)
	sugestao, err := models.ValidarEmailRigoroso(email)

	if err != nil {
		resposta["valido"] = false
		// O erro.Error() vai conter exatamente as frases que você definiu no model:
		// "formato inválido", "domínio não existe", "conta cheia", etc.
		resposta["mensagem"] = err.Error()

		// Opcional: Se houver sugestão (ex: gmail.com em vez de gamil.com), mandamos também
		if sugestao != "" {
			resposta["sugestao"] = sugestao
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resposta)
}

// Login gerencia a entrada no sistema
func Login(w http.ResponseWriter, r *http.Request) {

	// CENÁRIO 1: Acessar a página (GET)
	if r.Method != "POST" {
		temp.ExecuteTemplate(w, "Login", nil)
		return
	}

	// CENÁRIO 2: Enviar dados (POST)
	email := r.FormValue("email")
	senha := r.FormValue("senha")

	// --- CORREÇÃO AQUI ---
	// Remove espaços em branco nas pontas e força tudo para minúsculo
	email = strings.ToLower(strings.TrimSpace(email))
	// ---------------------

	// 1. Busca APENAS os dados do usuário pelo email (agora normalizado)
	usuario, err := models.BuscarUsuarioPorEmail(email)

	// 2. VALIDAÇÃO DE SENHA E EXISTÊNCIA
	if err != nil || !CheckPasswordHash(senha, usuario.SenhaHash) {
		dados := struct{ Erro string }{Erro: "Email ou senha incorretos."}
		temp.ExecuteTemplate(w, "Login", dados)
		return
	}

	// 3. VALIDAÇÃO DE BANIMENTO
	if usuario.IsBanned {
		dados := struct{ Erro string }{Erro: "Esta conta foi suspensa por violar as diretrizes da plataforma."}
		temp.ExecuteTemplate(w, "Login", dados)
		return
	}

	// 4. CRIA A SESSÃO COM GORILLA
	session, _ := config.Store.Get(r, "nexhub-session")
	session.Values["userId"] = usuario.Id
	err = session.Save(r, w)

	if err != nil {
		http.Error(w, "Erro de sessão", 500)
		return
	}

	// 5. REDIRECIONAMENTO
	if usuario.TipoUsuario == "DEV" {
		http.Redirect(w, r, "/dev/dashboard", http.StatusSeeOther)
	} else if usuario.TipoUsuario == "EMPRESA" {
		http.Redirect(w, r, "/empresa/dashboard", http.StatusSeeOther)
	} else if usuario.TipoUsuario == "ADMIN" {
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := config.Store.Get(r, "nexhub-session")

	// Define o tempo de vida para -1 (Isso deleta o cookie imediatamente)
	session.Options.MaxAge = -1
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func autenticarEBuscarUsuario(r *http.Request) (structs.Usuario, error) {
	// 1. Pega a sessão (o Gorilla cuida de descriptografar)
	session, _ := config.Store.Get(r, "nexhub-session")

	// 2. Verifica se existe um ID de usuário salvo na sessão
	// O 'ok' verifica se a conversão para int funcionou
	id, ok := session.Values["userId"].(int)
	if !ok {
		return structs.Usuario{}, fmt.Errorf("usuário não logado")
	}

	// 3. Busca os dados completos no banco pelo ID
	// (Precisamos dessa função no Model, veja o Passo 4)
	usuario, err := models.BuscarUsuarioPorID(id)
	if err != nil {
		return structs.Usuario{}, err
	}

	return usuario, nil
}

func AtualizarPerfilHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Busca o usuário atual (já com a senha antiga carregada na struct)
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		r.ParseMultipartForm(10 << 20)

		// --- ATUALIZAÇÃO DOS DADOS NORMAIS ---
		user.NomeCompleto = r.FormValue("nome")
		user.TituloProfissional = r.FormValue("titulo")
		user.Cidade = r.FormValue("cidade")
		user.Biografia = r.FormValue("bio")
		user.Skills = r.FormValue("skills")
		user.GithubLink = r.FormValue("github")
		user.LinkedinLink = r.FormValue("linkedin")
		user.Email = r.FormValue("email")
		user.Nivel = r.FormValue("nivel")

		if r.FormValue("disponivel") == "true" {
			user.DisponivelParaEquipes = true
		} else {
			user.DisponivelParaEquipes = false
		}

		// --- LÓGICA DA SENHA (NOVA) ---
		novaSenha := r.FormValue("nova_senha")
		confirmarSenha := r.FormValue("confirmar_senha")

		// Só tenta mudar a senha se o usuário digitou algo
		if novaSenha != "" {
			// 1. Validação: Senhas conferem?
			if novaSenha != confirmarSenha {
				// Se não batem, redireciona com erro (idealmente mostraria msg na tela)
				log.Println("Erro: As senhas não conferem!")
				http.Redirect(w, r, "/dev/perfil?erro=senhas_nao_conferem", http.StatusSeeOther)
				return
			}

			// 2. Hash da nova senha
			hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Erro ao criptografar senha", 500)
				return
			}

			// 3. Atualiza a struct com o NOVO hash
			user.SenhaHash = string(hash)
			log.Println("Senha alterada com sucesso!")
		}
		// SE novaSenha == "", a gente NÃO mexe no user.SenhaHash.
		// Como 'user' veio do banco lá em cima, ele mantém o hash antigo.

		// --- UPLOAD DA FOTO (MANTIDO) ---
		file, handler, err := r.FormFile("foto_perfil")
		if err == nil {
			defer file.Close()
			nomeArquivo := fmt.Sprintf("avatar_%d_%s", user.Id, handler.Filename)
			caminho := "static/uploads/" + nomeArquivo
			os.MkdirAll("static/uploads", os.ModePerm)
			dst, _ := os.Create(caminho)
			defer dst.Close()
			io.Copy(dst, file)
			user.FotoPerfil = "/" + caminho
		}

		// SALVA TUDO NO BANCO
		err = models.AtualizarPerfil(user)
		if err != nil {
			log.Println("Erro ao atualizar perfil:", err)
		}

		http.Redirect(w, r, "/dev/perfil?sucesso=true", http.StatusSeeOther)
	}
}

// Handler para a rota GET /dev/perfil
func PerfilHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Verifica quem é o usuário logado
	// (Essa função autenticarEBuscarUsuario já busca tudo no banco usando o BuscarUsuarioPorID que atualizamos)
	usuarioLogado, err := autenticarEBuscarUsuario(r)

	// Se não estiver logado ou der erro, manda pro login
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	totalNaoLidas, _ := models.ContarTotalNaoLidas(usuarioLogado.Id)

	// 2. Prepara os dados para o HTML
	dados := struct {
		Usuario  structs.Usuario
		NaoLidas int
	}{
		Usuario:  usuarioLogado,
		NaoLidas: totalNaoLidas,
	}

	// 3. Renderiza o template preenchido
	temp.ExecuteTemplate(w, "PerfilDev", dados)
}

// Função auxiliar: Retorna TRUE se a senha bater, FALSE se falhar
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// --- 1. TELA: DIGITAR O EMAIL ---
func EsqueciSenhaPage(w http.ResponseWriter, r *http.Request) {
	temp.ExecuteTemplate(w, "EsqueciSenha", nil)
}

func SolicitarResetHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")

	// Verifica se o usuário existe (Segurança: Não avise se não existir!)
	_, err := models.BuscarUsuarioPorEmail(email) // Assumindo que você tem essa func
	if err == nil {
		// Se existe, gera o código e envia
		models.CriarSolicitacaoRecuperacao(email)
	}

	// Redireciona para a tela de digitar o código SEMPRE (para não revelar se o email existe)
	http.Redirect(w, r, "/recuperar/codigo?email="+email, http.StatusSeeOther)
}

// --- 2. TELA: DIGITAR O CÓDIGO ---
func ValidarCodigoPage(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	dados := struct{ Email string }{Email: email}
	temp.ExecuteTemplate(w, "ValidarCodigo", dados)
}

func VerificarCodigoHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	codigo := r.FormValue("codigo")

	if models.ValidarCodigoRecuperacao(email, codigo) {
		// Se válido, manda para a tela de nova senha
		// Passamos email e código na URL (ou hidden form) para validar novamente na hora de salvar
		http.Redirect(w, r, fmt.Sprintf("/recuperar/nova-senha?email=%s&code=%s", email, codigo), http.StatusSeeOther)
	} else {
		// Código inválido
		http.Redirect(w, r, "/recuperar/codigo?email="+email+"&erro=CodigoInvalido", http.StatusSeeOther)
	}
}

// --- 3. TELA: NOVA SENHA ---
func NovaSenhaPage(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	code := r.URL.Query().Get("code")

	// Segurança extra: verifica o código de novo para ninguém acessar a URL direto
	if !models.ValidarCodigoRecuperacao(email, code) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	dados := struct{ Email, Codigo string }{Email: email, Codigo: code}
	temp.ExecuteTemplate(w, "NovaSenha", dados)
}

func SalvarNovaSenhaHandler(w http.ResponseWriter, r *http.Request) {
	// Limpa espaços em branco que podem vir do copy-paste
	email := strings.TrimSpace(r.FormValue("email"))
	codigo := strings.TrimSpace(r.FormValue("codigo"))
	novaSenha := r.FormValue("nova_senha")

	// 1. Valida o código
	if !models.ValidarCodigoRecuperacao(email, codigo) {
		http.Redirect(w, r, "/login?erro=Expirou", http.StatusSeeOther)
		return
	}

	// 2. Hash da nova senha
	hash, err := bcrypt.GenerateFromPassword([]byte(novaSenha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Erro ao criptografar senha", 500)
		return
	}

	// 3. Atualiza senha
	err = models.AtualizarSenhaPeloEmail(email, string(hash))
	if err != nil {
		// Se der erro (ex: email não achou ou coluna errada), mostra no terminal e avisa user
		fmt.Println("ERRO FATAL AO SALVAR SENHA:", err)
		http.Redirect(w, r, "/login?erro=ErroInternoAoSalvar", http.StatusSeeOther)
		return
	}

	// 4. Queima o código (só se a senha salvou com sucesso)
	models.MarcarCodigoComoUsado(email, codigo)

	http.Redirect(w, r, "/login?sucesso=SenhaAlterada", http.StatusSeeOther)
}
