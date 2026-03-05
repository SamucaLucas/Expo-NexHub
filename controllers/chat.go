package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nexhub/config"
	"nexhub/models"
	"nexhub/structs"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// --- CONFIGURAÇÃO DO WEBSOCKET ---
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Em produção, valide a origem. Para dev, permitimos tudo.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Mapa seguro para guardar usuários conectados (UserID -> Conexão)
var (
	conexoes = make(map[int]*websocket.Conn)
	mu       sync.Mutex // Mutex para evitar conflitos de escrita no mapa
)

// 1. ROTA DE CONEXÃO (Handshake)
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Pegar ID da sessão
	session, _ := config.Store.Get(r, "nexhub-session")
	usuarioID, ok := session.Values["userId"].(int)
	if !ok {
		return // Não conecta se não estiver logado
	}

	// Atualiza a conexão HTTP para WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Erro no upgrade WS:", err)
		return
	}

	// Registra o usuário no mapa
	mu.Lock()
	conexoes[usuarioID] = ws
	mu.Unlock()

	// Loop eterno para manter a conexão viva e ler mensagens (se houver)
	for {
		// O chat atual só recebe via POST, mas precisamos ler aqui para detectar desconexão
		_, _, err := ws.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(conexoes, usuarioID) // Remove do mapa ao desconectar
			mu.Unlock()
			break
		}
	}
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Validar Sessão
	session, _ := config.Store.Get(r, "nexhub-session")
	usuarioID, ok := session.Values["userId"].(int)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	usuarioLogado, _ := models.BuscarUsuarioPorID(usuarioID)

	var conversa []structs.Mensagem
	var destinatario structs.Usuario
	var projetoAtual structs.Projeto
	chatAberto := false
	destinatarioBanido := false
	isGrupo := false

	// 2. Lógica de Seleção de Chat
	destIdStr := r.URL.Query().Get("id")
	projetoIdStr := r.URL.Query().Get("projeto")

	// SE CLICOU NUM GRUPO
	if projetoIdStr != "" {
		projId, _ := strconv.Atoi(projetoIdStr)
		if projId > 0 {
			chatAberto = true
			isGrupo = true
			projetoAtual, _ = models.BuscarProjetoPorID(projId)
			conversa, _ = models.BuscarHistoricoGrupo(projId, usuarioID)
		}
	} else if destIdStr != "" {
		// SE CLICOU NUMA PESSOA (Lógica 1x1 antiga mantida intacta)
		destId, _ := strconv.Atoi(destIdStr)
		if destId > 0 && destId != usuarioID {
			chatAberto = true
			destinatario, _ = models.BuscarUsuarioPorID(destId)

			if destinatario.IsBanned {
				destinatarioBanido = true
			}

			models.MarcarComoLidas(usuarioID, destId)
			conversa, _ = models.BuscarHistoricoConversa(usuarioID, destId)
		}
	}

	// 3. Busca lista de contatos e grupos para a barra lateral
	contatos, _ := models.BuscarContatosRecentes(usuarioID)
	grupos, _ := models.BuscarMeusProjetosChat(usuarioID)

	// Marca o ativo visualmente na sidebar
	if chatAberto {
		if isGrupo {
			for i := range grupos {
				if grupos[i].ProjetoId == projetoAtual.Id {
					grupos[i].NaoLidas = 0 // Simula lido visualmente
				}
			}
		} else {
			for i := range contatos {
				if contatos[i].UsuarioId == destinatario.Id {
					contatos[i].Ativo = true
					contatos[i].NaoLidas = 0
				}
			}
		}
	}

	// Empacota tudo para mandar para o HTML
	dados := structs.DadosChat{
		UsuarioLogado:      usuarioLogado,
		Contatos:           contatos,
		Grupos:             grupos,
		ConversaAtual:      conversa,
		Destinatario:       destinatario,
		ChatAberto:         chatAberto,
		DestinatarioBanido: destinatarioBanido,
		IsGrupo:            isGrupo,      // NOVO
		ProjetoAtual:       projetoAtual, // NOVO
	}

	err := temp.ExecuteTemplate(w, "ChatFull", dados)
	if err != nil {
		log.Println("❌ Erro ao renderizar Chat:", err)
	}
}

// 2. API DE ENVIO (Atualizada para disparar o WebSocket com DETALHES DO REMETENTE)
// 2. API DE ENVIO (Atualizada para suportar Chat 1x1 e Chat de Projetos/Grupos)
func EnviarMensagemAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}

	session, _ := config.Store.Get(r, "nexhub-session")
	remetenteID, _ := session.Values["userId"].(int)

	// Pega os dois possíveis IDs do formulário (um deles estará vazio)
	destIdStr := r.FormValue("destinatario_id")
	projetoIdStr := r.FormValue("projeto_id")
	conteudo := strings.TrimSpace(r.FormValue("conteudo"))

	if conteudo == "" {
		return
	}

	// Busca os dados de quem está enviando
	usuarioRemetente, _ := models.BuscarUsuarioPorID(remetenteID)

	if projetoIdStr != "" {
		projetoId, _ := strconv.Atoi(projetoIdStr)

		if projetoId > 0 {
			// A. Salva no Banco como mensagem de Grupo
			err := models.EnviarMensagemGrupo(remetenteID, projetoId, conteudo)
			if err != nil {
				fmt.Println("❌ ERRO AO SALVAR MSG GRUPO NO BANCO:", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"sucesso": false,
					"erro":    err.Error(),
				})
				return
			}

			// B. Busca quem faz parte da equipe
			membros, _ := models.BuscarIDsMembrosProjeto(projetoId)
			projetoAtual, _ := models.BuscarProjetoPorID(projetoId)

			notificacao := map[string]interface{}{
				"tipo":           "nova_mensagem_grupo",
				"conteudo":       conteudo,
				"remetente_id":   remetenteID,
				"projeto_id":     projetoId,
				"hora":           "Agora",
				"remetente_nome": usuarioRemetente.NomeCompleto,
				"remetente_foto": usuarioRemetente.FotoPerfil,
				"projeto_nome":   projetoAtual.Titulo,
			}

			// C. Dispara WebSocket para toda a equipe
			mu.Lock()
			for _, membroID := range membros {
				if membroID == remetenteID {
					continue // Não envia notificação para si mesmo
				}
				if socketDestino, online := conexoes[membroID]; online {
					socketDestino.WriteJSON(notificacao)
				}
			}
			mu.Unlock()
		}

	} else if destIdStr != "" {
		destId, _ := strconv.Atoi(destIdStr)

		if destId > 0 {
			usuarioDestino, _ := models.BuscarUsuarioPorID(destId)

			// Verifica se o destino está banido
			if usuarioDestino.IsBanned {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"sucesso": false,
					"erro":    "Este usuário está banido e não pode receber mensagens.",
				})
				return
			}

			// A. Salva no Banco Privado
			err := models.EnviarMensagem(remetenteID, destId, conteudo)
			if err != nil {
				fmt.Println("❌ ERRO AO SALVAR MENSAGEM PRIVADA NO BANCO:", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"sucesso": false,
					"erro":    err.Error(),
				})
				return
			}

			notificacao := map[string]interface{}{
				"tipo":           "nova_mensagem",
				"conteudo":       conteudo,
				"remetente_id":   remetenteID,
				"hora":           "Agora",
				"remetente_nome": usuarioRemetente.NomeCompleto,
				"remetente_foto": usuarioRemetente.FotoPerfil,
				"remetente_tipo": usuarioRemetente.TipoUsuario,
			}

			// B. Dispara WebSocket para a pessoa
			mu.Lock()
			socketDestino, online := conexoes[destId]
			if online {
				socketDestino.WriteJSON(notificacao)
			}
			mu.Unlock()
		}
	} else {
		// Se não tiver nem ID de projeto nem de pessoa
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sucesso": true,
	})
}
