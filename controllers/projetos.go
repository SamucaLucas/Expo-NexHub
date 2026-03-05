package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"nexhub/config"
	"nexhub/models"
	"nexhub/structs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func NovoProjeto(w http.ResponseWriter, r *http.Request) {
	// 1. Autenticação
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == "GET" {
		// 1. Busca as opções no banco
		opcoes, err := models.BuscarTodasTecnologias()
		if err != nil {
			log.Println("Erro ao buscar tecnologias:", err)
		}

		// 2. Manda para o HTML junto com o Usuário
		dados := struct {
			Usuario    structs.Usuario
			TodasTechs []structs.Tecnologia // Lista de opções
		}{
			Usuario:    user,
			TodasTechs: opcoes,
		}

		temp.ExecuteTemplate(w, "NovoProjeto", dados)
		return
	}

	// --- POST: Salvar com Upload ---
	if r.Method == "POST" {
		// Aumenta o limite de memória para upload (10MB)
		r.ParseMultipartForm(10 << 20)

		// 1. Processar Upload da Imagem (MANTIDO IGUAL)
		var caminhoImagem string
		file, handler, err := r.FormFile("imagem_capa")

		if err == nil {
			defer file.Close()
			os.MkdirAll("./static/uploads", os.ModePerm)
			nomeArquivo := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
			caminhoDestino := filepath.Join("static/uploads", nomeArquivo)

			dst, err := os.Create(caminhoDestino)
			if err != nil {
				http.Error(w, "Erro ao criar arquivo", http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "Erro ao salvar arquivo", http.StatusInternalServerError)
				return
			}
			caminhoImagem = "/static/uploads/" + nomeArquivo
		} else {
			caminhoImagem = ""
		}

		// 2. Pegar outros dados do formulário
		titulo := r.FormValue("titulo")
		descricao := r.FormValue("descricao")
		status := r.FormValue("status")
		cidade := r.FormValue("cidade")
		categoria := r.FormValue("categoria")
		outraCategoria := r.FormValue("outra_categoria")
		repo := r.FormValue("link_repositorio")

		if categoria == "Outros" && outraCategoria != "" {
			categoria = outraCategoria
		}
		techsSelecionadas := r.Form["techs"] // Ex: ["Go", "React", "Docker"]

		// 3. Montar Struct
		novoProjeto := structs.Projeto{
			Titulo:      titulo,
			Descricao:   descricao,
			Status:      status,
			Cidade:      cidade,
			Categoria:   categoria,
			LinkRepo:    repo,
			ImagemCapa:  caminhoImagem,
			IdLider:     user.Id,
			Tecnologias: techsSelecionadas,
		}

		_, err = models.CriarProjeto(novoProjeto)
		if err != nil {
			log.Println("Erro ao criar projeto:", err)
			http.Error(w, "Erro ao salvar no banco: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Sucesso!
		http.Redirect(w, r, "/dev/meus-projetos", http.StatusSeeOther)
	}
}

// AtualizarProjetoHandler
func AtualizarProjetoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseMultipartForm(10 << 20)

		idStr := r.FormValue("id")
		id, _ := strconv.Atoi(idStr)
		titulo := r.FormValue("titulo")
		descricao := r.FormValue("descricao")
		status := r.FormValue("status")

		// NOVOS CAMPOS
		cidade := r.FormValue("cidade")
		categoria := r.FormValue("categoria")
		outraCategoria := r.FormValue("outra_categoria")
		repo := r.FormValue("link_repositorio")

		techsSelecionadas := r.Form["techs"] // Ex: ["Go", "React", "Docker"]

		if categoria == "Outros" && outraCategoria != "" {
			categoria = outraCategoria
		}

		// Lógica da Imagem (Mantida igual ao anterior)
		projetoAtual, _ := models.BuscarProjetoPorID(id)
		caminhoImagemFinal := projetoAtual.ImagemCapa

		file, handler, err := r.FormFile("imagem")
		if err == nil {
			defer file.Close()
			nomeArquivo := fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
			caminhoNoDisco := "static/uploads/" + nomeArquivo
			caminhoNoBanco := "/static/uploads/" + nomeArquivo
			os.MkdirAll("static/uploads", os.ModePerm)
			dst, _ := os.Create(caminhoNoDisco)
			defer dst.Close()
			io.Copy(dst, file)
			caminhoImagemFinal = caminhoNoBanco
		}

		// Chama o model com TODOS os argumentos
		err = models.AtualizarProjeto(id, titulo, descricao, status, cidade, categoria, caminhoImagemFinal, repo, techsSelecionadas)
		if err != nil {
			log.Println("Erro ao atualizar:", err)
			http.Error(w, "Erro ao atualizar projeto", http.StatusInternalServerError)
			return
		}
		// 2. NOVO: Processar GALERIA (Vários arquivos)
		// Pega todos os arquivos do input name="galeria"
		files := r.MultipartForm.File["galeria"]
		for _, fileHeader := range files {
			// Abre o arquivo
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}
			defer file.Close()

			// Cria nome único
			nomeArquivo := fmt.Sprintf("galeria_%d_%s", time.Now().UnixNano(), fileHeader.Filename)
			caminhoDestino := filepath.Join("static/uploads", nomeArquivo)

			// Salva no disco
			dst, err := os.Create(caminhoDestino)
			if err == nil {
				io.Copy(dst, file)
				dst.Close()

				// SALVA NO BANCO (Tabela projeto_imagens)
				caminhoBanco := "/static/uploads/" + nomeArquivo
				models.AdicionarImagemGaleria(id, caminhoBanco)
			}
		}

		http.Redirect(w, r, "/dev/meus-projetos", http.StatusSeeOther)
	}
}

// 1. TELA DE EDIÇÃO (GET)
func EditarProjetoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Pega o ID da URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println("ID inválido:", err)
		http.Redirect(w, r, "/dev/meus-projetos", http.StatusSeeOther)
		return
	}

	// 2. Busca o projeto (que já traz o array de techs salvas)
	projeto, err := models.BuscarProjetoPorID(id)
	if err != nil {
		log.Println("Projeto não encontrado:", err)
		http.Redirect(w, r, "/dev/meus-projetos", http.StatusSeeOther)
		return
	}

	equipe, _ := models.BuscarMembrosDoProjeto(projeto.Id)
	projeto.Equipe = equipe

	// 3. Busca o "cardápio" completo de opções (Tabela tecnologias)
	opcoes, err := models.BuscarTodasTecnologias()
	if err != nil {
		log.Println("Erro ao buscar opções de tecnologias:", err)
		// Não precisa travar, pode renderizar sem opções se der erro
	}

	saves, _ := models.ContarSavesEmpresa(id)

	// 4. Monta o pacote de dados para o HTML
	dados := struct {
		Projeto    structs.Projeto
		TodasTechs []structs.Tecnologia // Importante: O nome no HTML é .TodasTechs
		Saves      int
		Usuario    structs.Usuario
	}{
		Projeto:    projeto,
		TodasTechs: opcoes,
		Saves:      saves,
	}

	// 5. Renderiza
	temp.ExecuteTemplate(w, "EditarProjeto", dados)
}

// 3. EXCLUIR PROJETO (GET/POST)
func DeletarProjetoHandler(w http.ResponseWriter, r *http.Request) {
	// Pega o ID da URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)

	if err == nil {
		// Chama o MODEL corrigido
		models.DeletarProjeto(id)
	}

	// Volta pra lista
	http.Redirect(w, r, "/dev/meus-projetos", http.StatusSeeOther)
}

// controllers/projetos.go

func DetalhesProjetoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Tenta pegar o ID da URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || idStr == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 2. Busca os dados do projeto
	projeto, autorNome, autorFoto, err := models.BuscarDetalhesProjeto(id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// 3. Busca a Equipe
	equipe, _ := models.BuscarMembrosDoProjeto(id)

	// 4. BUSCA AS AVALIAÇÕES (Isso traz os dados do banco)
	avaliacoes, err := models.BuscarAvaliacoesDoProjeto(id)
	if err != nil {
		fmt.Println("Erro ao buscar avaliações:", err)
	}

	// 5. Variáveis de Controle e Sessão
	ehDono := false
	ehMembro := false
	estaSalvo := false
	role := "Visitante"

	session, _ := config.Store.Get(r, "nexhub-session")
	if usuarioID, ok := session.Values["userId"].(int); ok {
		usuarioLogado, _ := models.BuscarUsuarioPorID(usuarioID)

		if usuarioLogado.TipoUsuario == "EMPRESA" {
			role = "Empresa"
			estaSalvo = models.ChecarSeFavoritou(usuarioID, projeto.Id, "PROJETO")
		} else if usuarioLogado.TipoUsuario == "DEV" {
			role = "Dev"
		} else if usuarioLogado.TipoUsuario == "ADMIN" {
			role = "Admin"
		}

		if usuarioID == projeto.IdLider {
			ehDono = true
		}
		for _, membro := range equipe {
			if membro.IdUsuario == usuarioID {
				ehMembro = true
				break
			}
		}
	}
	usuarioID, ok := session.Values["userId"].(int)
	if !ok {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

	usuarioLogado, err := models.BuscarUsuarioPorID(usuarioID)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !ehDono && !ehMembro {
		go models.IncrementarVisualizacao(id)
	}

	// 6. Lógica de Alertas via URL
	erroURL := r.URL.Query().Get("erro")
	sucessoURL := r.URL.Query().Get("sucesso")
	sugestao := r.URL.Query().Get("sugestao")

	var msgAlerta, tipoAlerta string

	if sucessoURL == "true" {
		msgAlerta = "✨ Avaliação enviada e confirmada com sucesso!"
		tipoAlerta = "sucesso"
	} else if erroURL != "" {
		tipoAlerta = "erro"
		switch erroURL {
		case "EmailInexistente":
			msgAlerta = "🚫 E-mail não encontrado. Verifique se digitou corretamente."
		case "EmailTemporario":
			msgAlerta = "⚠️ E-mails temporários não são permitidos."
		case "EmailInvalido":
			msgAlerta = "❌ Formato de e-mail inválido."
		case "JaAvaliou":
			msgAlerta = "✋ Você já avaliou este projeto com este e-mail."
			tipoAlerta = "aviso"
		case "Sugestao":
			msgAlerta = fmt.Sprintf("🤔 Você quis dizer %s? Corrija e tente novamente.", sugestao)
			tipoAlerta = "aviso"
		default:
			msgAlerta = "❌ Ocorreu um erro interno. Tente novamente."
		}
	}

	// 7. MONTA A ESTRUTURA PARA O HTML
	// ATENÇÃO AQUI: Você precisa DECLARAR o campo e DEPOIS preencher.
	dados := struct {
		Projeto     structs.Projeto
		AutorNome   string
		AutorAvatar string
		Role        string
		Equipe      []structs.MembroEquipe
		EhMembro    bool
		Salvou      bool
		Usuario     structs.Usuario

		// --- O HTML SÓ VÊ O QUE ESTÁ DECLARADO AQUI EMBAIXO ---
		Avaliacoes     []models.Avaliacao // <------ OBRIGATÓRIO ESTAR AQUI
		MensagemAlerta string             // <------ OBRIGATÓRIO ESTAR AQUI
		TipoAlerta     string             // <------ OBRIGATÓRIO ESTAR AQUI
	}{
		Projeto:     projeto,
		AutorNome:   autorNome,
		AutorAvatar: autorFoto,
		Role:        role,
		Equipe:      equipe,
		EhMembro:    ehMembro,
		Salvou:      estaSalvo,
		Usuario:     usuarioLogado,

		// --- AQUI PREENCHEMOS OS DADOS ---
		Avaliacoes:     avaliacoes, // <------ AQUI ENTRA A LISTA QUE BUSCAMOS NO PASSO 4
		MensagemAlerta: msgAlerta,
		TipoAlerta:     tipoAlerta,
	}

	temp.ExecuteTemplate(w, "DetalheProjeto", dados)
}

func DeletarImagemHandler(w http.ResponseWriter, r *http.Request) {
	// Pega o ID da imagem e o ID do projeto da URL
	idImgStr := r.URL.Query().Get("id_img")
	idProjetoStr := r.URL.Query().Get("id_projeto")

	idImg, _ := strconv.Atoi(idImgStr)

	// Chama o model para deletar
	models.DeletarImagemGaleria(idImg)

	// Redireciona de volta para a tela de edição do projeto
	http.Redirect(w, r, "/dev/projeto/editar?id="+idProjetoStr, http.StatusSeeOther)
}

func AdicionarMembroHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		idProjeto, _ := strconv.Atoi(r.FormValue("id_projeto"))
		idUsuario, _ := strconv.Atoi(r.FormValue("id_usuario")) // ID do dev selecionado
		funcao := r.FormValue("funcao_membro")

		models.AdicionarMembroEquipe(idProjeto, idUsuario, funcao)

		// Recarrega a página de edição
		http.Redirect(w, r, "/dev/projeto/editar?id="+strconv.Itoa(idProjeto)+"#equipe-section", http.StatusSeeOther)
	}
}

func RemoverMembroHandler(w http.ResponseWriter, r *http.Request) {
	idProjetoStr := r.URL.Query().Get("id_projeto")
	idUsuarioStr := r.URL.Query().Get("id_usuario")

	idProjeto, _ := strconv.Atoi(idProjetoStr)
	idUsuario, _ := strconv.Atoi(idUsuarioStr)

	models.RemoverMembroEquipe(idProjeto, idUsuario)

	http.Redirect(w, r, "/dev/projeto/editar?id="+idProjetoStr+"#equipe-section", http.StatusSeeOther)
}

// controllers/projetos.go

// SairDoProjetoHandler permite que o próprio membro saia da equipe
func SairDoProjetoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Identifica quem está logado
	user, err := autenticarEBuscarUsuario(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// 2. Identifica de qual projeto ele quer sair
	idProjetoStr := r.URL.Query().Get("id")
	idProjeto, err := strconv.Atoi(idProjetoStr)
	if err != nil {
		http.Redirect(w, r, "/dev/dashboard", http.StatusSeeOther)
		return
	}

	// 3. Remove do banco (Reutiliza a função do Model existente)
	// Como passamos user.Id (da sessão), é impossível ele remover outra pessoa
	err = models.RemoverMembroEquipe(idProjeto, user.Id)
	if err != nil {
		log.Println("Erro ao sair do projeto:", err)
	}

	// 4. Redireciona para "Meus Projetos" (Pois ele não faz mais parte daquele)
	http.Redirect(w, r, "/dev/meus-projetos", http.StatusSeeOther)
}

func SalvarAvaliacaoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id_projeto")
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	notaStr := r.FormValue("nota")
	comentario := r.FormValue("comentario")
	nomeAvaliador := r.FormValue("nomeAvaliador")

	idProjeto, _ := strconv.Atoi(idStr)
	nota, _ := strconv.Atoi(notaStr)

	// Base URL para redirecionamento
	redirectURL := fmt.Sprintf("/projeto/detalhes?id=%d", idProjeto)

	// 1. Validação Básica
	if idProjeto == 0 || nota < 1 || nota > 5 {
		http.Redirect(w, r, redirectURL+"&erro=DadosInvalidos", http.StatusSeeOther)
		return
	}

	// 2. VALIDAÇÃO RIGOROSA (AfterShip)
	sugestao, err := models.ValidarEmailRigoroso(email)
	if err != nil {
		fmt.Println("Erro Validação Email:", err)

		// Vamos passar o erro específico para a URL
		if strings.Contains(err.Error(), "não existe") {
			http.Redirect(w, r, redirectURL+"&erro=EmailInexistente", http.StatusSeeOther)
		} else if strings.Contains(err.Error(), "temporário") {
			http.Redirect(w, r, redirectURL+"&erro=EmailTemporario", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, redirectURL+"&erro=EmailInvalido", http.StatusSeeOther)
		}
		return
	}

	// Se a biblioteca sugeriu uma correção (ex: gmil -> gmail), avisamos o usuário
	if sugestao != "" {
		// Opcional: Você pode forçar o erro para ele corrigir
		// Ou aceitar e salvar. Aqui vamos forçar ele a corrigir.
		http.Redirect(w, r, redirectURL+"&erro=Sugestao&sugestao="+sugestao, http.StatusSeeOther)
		return
	}

	// 3. Salvar
	av := models.Avaliacao{
		IdProjeto:     idProjeto,
		NomeAvaliador: nomeAvaliador,
		Email:         email,
		Nota:          nota,
		Comentario:    comentario,
	}

	if err := models.SalvarAvaliacao(av); err != nil {
		fmt.Println("Erro DB:", err.Error())

		msgErro := strings.ToLower(err.Error())
		if strings.Contains(msgErro, "duplicate key") ||
			strings.Contains(msgErro, "unique constraint") ||
			strings.Contains(msgErro, "email_projeto_unico") {

			// REDIRECIONA COM O ERRO ESPECÍFICO
			http.Redirect(w, r, redirectURL+"&erro=JaAvaliou", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, redirectURL+"&erro=ErroInterno", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, redirectURL+"&sucesso=true", http.StatusSeeOther)
}
