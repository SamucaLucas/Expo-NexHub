package routers

import (
	"net/http"
	"nexhub/controllers"
)

func CarregarRotas() {
	// 1. Configurar arquivos estáticos (CSS, Imagens, JS)
	// Isso permite que o HTML encontre o /static/css/style.css
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 2. Rotas Públicas
	http.HandleFunc("/", controllers.Index)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/logout", controllers.LogoutHandler)

	http.HandleFunc("/talentos", controllers.TalentosVitrineHandler)
	http.HandleFunc("/projetos", controllers.ProjetosVitrineHandler)
	http.HandleFunc("/cadastro", controllers.Cadastro)
	http.HandleFunc("/sobre", controllers.Sobre)
	http.HandleFunc("/avaliar/salvar", controllers.SalvarAvaliacaoHandler)

	// Chat - Acessível para Dev e Empresa
	http.HandleFunc("/chat", controllers.ChatHandler)
	http.HandleFunc("/chat/enviar", controllers.EnviarMensagemAPI)
	http.HandleFunc("/ws", controllers.HandleWebSocket)

	//Rotas Logadas Dev
	http.HandleFunc("/dev/dashboard", controllers.DashboardDev)
	http.HandleFunc("/dev/meus-projetos", controllers.MeusProjetos)
	http.HandleFunc("/dev/perfil", controllers.PerfilHandler)
	http.HandleFunc("/dev/perfil/salvar", controllers.AtualizarPerfilHandler)
	http.HandleFunc("/dev/novo-projeto", controllers.NovoProjeto)
	http.HandleFunc("/dev/projeto/editar", controllers.EditarProjetoHandler)
	http.HandleFunc("/dev/projeto/atualizar", controllers.AtualizarProjetoHandler) // Rota do POST do form
	http.HandleFunc("/dev/projeto/deletar", controllers.DeletarProjetoHandler)
	http.HandleFunc("/projeto/galeria/deletar", controllers.DeletarImagemHandler)
	// Gestão de Equipe
	http.HandleFunc("/dev/projeto/equipe/adicionar", controllers.AdicionarMembroHandler)
	http.HandleFunc("/dev/projeto/equipe/remover", controllers.RemoverMembroHandler)
	http.HandleFunc("/dev/projeto/sair", controllers.SairDoProjetoHandler)

	//Rotas Logadas Empresa
	http.HandleFunc("/empresa/dashboard", controllers.DashboardEmpresa)
	http.HandleFunc("/empresa/perfil", controllers.PerfilEmpresa)
	http.HandleFunc("/empresa/projetos", controllers.EmpresaProjetos)
	http.HandleFunc("/empresa/talentos", controllers.EmpresaTalentos)
	http.HandleFunc("/favoritar", controllers.ToggleFavoritoHandler)

	// --- ÁREA ADMIN ---
	http.HandleFunc("/admin/dashboard", controllers.AdminDashboardHandler)
	http.HandleFunc("/admin/usuarios", controllers.AdminUsuariosHandler)
	http.HandleFunc("/admin/promover", controllers.AdminPromoverHandler)
	http.HandleFunc("/admin/projetos", controllers.AdminProjetosHandler)
	http.HandleFunc("/admin/remover-admin", controllers.AdminRemoverAdminHandler)
	http.HandleFunc("/admin/banir", controllers.AdminBanirHandler)
	http.HandleFunc("/admin/status-projeto", controllers.AdminAlterarStatusProjetoHandler)
	http.HandleFunc("/admin/excluir-projeto", controllers.AdminExcluirProjetoHandler)
	http.HandleFunc("/admin/perfil", controllers.AdminPerfilHandler)
	http.HandleFunc("/admin/perfil/salvar", controllers.AdminSalvarPerfilHandler)

	http.HandleFunc("/projeto/detalhes", controllers.DetalhesProjetoHandler)
	http.HandleFunc("/talento/detalhes", controllers.DetalhesTalentoHandler)
	http.HandleFunc("/empresa/detalhes", controllers.DetalheEmpresaHandler)

	// API (Pode ser acessada por Devs logados)
	http.HandleFunc("/api/pesquisar-devs", controllers.ApiPesquisarDevs)

	// --- ROTAS DE RECUPERAÇÃO DE SENHA ---

	// 1. Esqueci a Senha (GET exibe página, POST envia email)

	http.HandleFunc("/esqueci-senha", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SolicitarResetHandler(w, r)
		} else {
			controllers.EsqueciSenhaPage(w, r)
		}
	})

	// 2. Validar Código (GET exibe página, POST valida)
	http.HandleFunc("/recuperar/codigo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.VerificarCodigoHandler(w, r)
		} else {
			controllers.ValidarCodigoPage(w, r)
		}
	})

	// 3. Nova Senha (GET exibe formulário)

	http.HandleFunc("/recuperar/nova-senha", controllers.NovaSenhaPage)

	// 4. Salvar Nova Senha (POST salva no banco)
	http.HandleFunc("/recuperar/salvar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			controllers.SalvarNovaSenhaHandler(w, r)
		} else {
			// Se tentarem acessar essa URL via GET, manda pra home
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})

	// API de Verificação de Email
	// Removemos o .Methods("GET") e passamos a função direto
	http.HandleFunc("/api/verificar-email", func(w http.ResponseWriter, r *http.Request) {
		// Opcional: Garante que só aceita GET
		if r.Method != http.MethodGet {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}
		controllers.VerificarEmailDisponivel(w, r)
	})
}
