package routers

import (
	"net/http"
	"nexhub/controllers"
)

func CarregarRotas() {
	// 1. Configurar arquivos estáticos (CSS, Imagens, JS)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// ==========================================
	// 2. ROTAS PÚBLICAS (Visitantes)
	// ==========================================
	http.HandleFunc("/", controllers.IndexHandler)
	http.HandleFunc("/sobre", controllers.SobreHandler)

	// Vitrine de Talentos (Alunos)
	http.HandleFunc("/talentos", controllers.TalentosHandler)
	http.HandleFunc("/talento", controllers.DetalheTalentoHandler) // ex: /talento?id=5

	// Vitrine de Projetos
	http.HandleFunc("/projetos", controllers.ProjetosHandler)
	http.HandleFunc("/projeto", controllers.DetalheProjetoHandler) // ex: /projeto?id=10

	// Interação do Visitante
	http.HandleFunc("/avaliar/salvar", controllers.SalvarAvaliacaoHandler)

	// ==========================================
	// 3. AUTENTICAÇÃO (Admins / ADS)
	// ==========================================
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/cadastro", controllers.Cadastro)
	http.HandleFunc("/logout", controllers.Logout)

	// ==========================================
	// 4. ROTAS DO ADMIN (Gestão da Plataforma)
	// ==========================================

	// Dashboard e Perfil
	http.HandleFunc("/admin/dashboard", controllers.AdminDashboardHandler)
	http.HandleFunc("/admin/perfil", controllers.AdminPerfilHandler)
	http.HandleFunc("/admin/perfil/salvar", controllers.AdminSalvarPerfilHandler)

	// Gestão de Alunos
	http.HandleFunc("/admin/alunos", controllers.AdminAlunosHandler)
	http.HandleFunc("/admin/alunos/salvar", controllers.AdminSalvarAlunoHandler)
	http.HandleFunc("/admin/alunos/excluir", controllers.AdminExcluirAlunoHandler)

	// Gestão de Projetos
	http.HandleFunc("/admin/projetos", controllers.AdminProjetosHandler)
	http.HandleFunc("/admin/projetos/salvar", controllers.AdminSalvarProjetoHandler)
	http.HandleFunc("/admin/projetos/status", controllers.AdminAlterarStatusProjetoHandler)
	http.HandleFunc("/admin/projetos/excluir", controllers.AdminExcluirProjetoHandler)
	http.HandleFunc("/admin/projetos/editar", controllers.AdminEditarProjetoHandler)

	// Ações da Tela de Edição de Projetos
	http.HandleFunc("/admin/projetos/atualizar", controllers.AdminAtualizarProjetoHandler)
	
	http.HandleFunc("/admin/projetos/equipe/adicionar", controllers.AdminProjetoAdicionarEquipeHandler)
	http.HandleFunc("/admin/projetos/equipe/remover", controllers.AdminProjetoRemoverEquipeHandler)
	
	http.HandleFunc("/admin/projetos/links/adicionar", controllers.AdminProjetoAdicionarLinkHandler)
	http.HandleFunc("/admin/projetos/links/remover", controllers.AdminProjetoRemoverLinkHandler)
	
	http.HandleFunc("/admin/projetos/arquivos/upload", controllers.AdminProjetoUploadArquivoHandler)
	http.HandleFunc("/admin/projetos/arquivos/remover", controllers.AdminProjetoRemoverArquivoHandler)

	// Gestão de Analistas (Exclusivo Admin Geral)
	http.HandleFunc("/admin/analistas", controllers.AdminAnalistasHandler)
	http.HandleFunc("/admin/analistas/excluir", controllers.AdminExcluirAnalistaHandler)

	// ==========================================
	// 5. API (Uso interno via JavaScript)
	// ==========================================

	// Autocomplete para buscar alunos e colocar na equipe do projeto
	http.HandleFunc("/api/alunos/pesquisar", controllers.ApiPesquisarAlunos)

	// ==========================================
	// 6. RECUPERAÇÃO DE SENHA
	// ==========================================
	// Mantido as rotas originais, caso os controllers/recuperacao.go ainda existam.
	http.HandleFunc("/esqueci-senha", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SolicitarResetHandler(w, r)
		} else {
			controllers.EsqueciSenhaPage(w, r)
		}
	})

	http.HandleFunc("/recuperar/codigo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.VerificarCodigoHandler(w, r)
		} else {
			controllers.ValidarCodigoPage(w, r)
		}
	})

	http.HandleFunc("/recuperar/nova-senha", controllers.NovaSenhaPage)

	http.HandleFunc("/recuperar/salvar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SalvarNovaSenhaHandler(w, r)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})
}
