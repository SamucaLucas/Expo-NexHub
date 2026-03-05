package models

import (
	"nexhub/db"
	"nexhub/structs"
	"nexhub/utils"
)

// 1. Enviar Mensagem (Adaptado para tabela 'chat')
func EnviarMensagem(remetente, destinatario int, conteudo string) error {

	// 1. CRIPTOGRAFA O CONTEÚDO AQUI ANTES DE SALVAR
	conteudoCriptografado, err := utils.CriptografarMensagem(conteudo)
	if err != nil {
		return err // Retorna erro se falhar a criptografia
	}
	// Query ajustada para as suas colunas
	query := `INSERT INTO chat (id_remetente, id_destinatario, mensagem) VALUES ($1, $2, $3)`
	_, err = db.DB.Exec(query, remetente, destinatario, conteudoCriptografado)
	return err
}

func ContarTotalNaoLidas(usuarioID int) (int, error) {
	var total int
	// Conta mensagens onde EU sou o destinatário e lida é FALSE
	query := `SELECT COUNT(*) FROM chat WHERE id_destinatario = $1 AND lida = FALSE`
	err := db.DB.QueryRow(query, usuarioID).Scan(&total)
	return total, err
}

func MarcarComoLidas(euId, outroId int) error {
	query := `UPDATE chat SET lida = TRUE WHERE id_destinatario = $1 AND id_remetente = $2`
	_, err := db.DB.Exec(query, euId, outroId)
	return err
}

// 2. Buscar Histórico (Adaptado para tabela 'chat')
func BuscarHistoricoConversa(euId, outroId int) ([]structs.Mensagem, error) {
	// Query ajustada: id_mensagem, id_remetente, id_destinatario, mensagem...
	query := `
		SELECT id_mensagem, id_remetente, id_destinatario, mensagem, data_envio, lida
		FROM chat
		WHERE (id_remetente = $1 AND id_destinatario = $2)
		   OR (id_remetente = $2 AND id_destinatario = $1)
		ORDER BY data_envio ASC
	`
	rows, err := db.DB.Query(query, euId, outroId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []structs.Mensagem
	for rows.Next() {
		var m structs.Mensagem
		// Scan na ordem exata do SELECT acima
		err = rows.Scan(&m.Id, &m.RemetenteId, &m.DestinatarioId, &m.Conteudo, &m.DataEnvio, &m.Lido)
		if err != nil {
			continue
		}

		// ---> DESCRIPTOGRAFA A MENSAGEM AQUI <---
		m.Conteudo = utils.DescriptografarMensagem(m.Conteudo)

		m.EhMinha = (m.RemetenteId == euId)
		m.HoraFormatada = m.DataEnvio.Format("15:04")
		msgs = append(msgs, m)
	}
	return msgs, nil
}

// 3. Buscar Contatos (Adaptado para tabela 'chat' e 'usuarios(id_usuario)')
func BuscarContatosRecentes(euId int) ([]structs.ContatoChat, error) {
	// Adicionei a linha do COUNT para preencher o NaoLidas
	query := `
		SELECT DISTINCT ON (u.id_usuario)
			u.id_usuario, u.nome_completo, u.foto_perfil,
			c.mensagem, c.data_envio,
			(SELECT COUNT(*) FROM chat WHERE id_remetente = u.id_usuario AND id_destinatario = $1 AND lida = FALSE) AS nao_lidas, u.tipo_usuario
		FROM usuarios u
		JOIN chat c ON (c.id_remetente = u.id_usuario OR c.id_destinatario = u.id_usuario)
		WHERE (c.id_remetente = $1 OR c.id_destinatario = $1)
		AND u.id_usuario != $1
		ORDER BY u.id_usuario, c.data_envio DESC
	`

	rows, err := db.DB.Query(query, euId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contatos []structs.ContatoChat
	for rows.Next() {
		var c structs.ContatoChat
		var fotoNull *string

		// Adicionei &c.NaoLidas no final do Scan
		err = rows.Scan(&c.UsuarioId, &c.Nome, &fotoNull, &c.UltimaMensagem, &c.DataUltima, &c.NaoLidas, &c.TipoUsuario)
		if err != nil {
			continue
		}
		if fotoNull != nil {
			c.Avatar = *fotoNull
		}

		// ---> DESCRIPTOGRAFA A ÚLTIMA MENSAGEM (PRÉVIA) AQUI <---
		c.UltimaMensagem = utils.DescriptografarMensagem(c.UltimaMensagem)

		contatos = append(contatos, c)
	}

	return contatos, nil
}

func BuscarMeusProjetosChat(usuarioId int) ([]structs.GrupoChat, error) {
	// SQL Ajustado com base nas structs:
	// - Tabela projetos: 'id', 'titulo', 'imagem_capa', 'id_lider'
	// - Tabela equipe_projeto: 'id_projeto', 'id_usuario'
	query := `
		SELECT 
			p.id_projeto, 
			p.titulo, 
			p.imagem_capa,
			(SELECT mensagem FROM chat WHERE id_projeto = p.id_projeto ORDER BY data_envio DESC LIMIT 1) AS ultima_mensagem
		FROM projetos p
		WHERE p.id_lider = $1 
		   OR p.id_projeto IN (SELECT id_projeto FROM equipe_projeto WHERE id_usuario = $1)
		ORDER BY p.id_projeto DESC
	`

	rows, err := db.DB.Query(query, usuarioId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grupos []structs.GrupoChat
	for rows.Next() {
		var g structs.GrupoChat
		var ultimaMsgNull *string
		var capaNull *string

		// Lemos os dados exatamente na ordem do SELECT
		err = rows.Scan(&g.ProjetoId, &g.NomeProjeto, &capaNull, &ultimaMsgNull)
		if err != nil {
			continue
		}

		// Tratamento de campos que podem ser nulos no banco
		if capaNull != nil {
			g.CapaProjeto = *capaNull
		}

		if ultimaMsgNull != nil {
			// DESCRIPTOGRAFA a última mensagem do grupo para a prévia!
			g.UltimaMensagem = utils.DescriptografarMensagem(*ultimaMsgNull)
		} else {
			g.UltimaMensagem = "Nenhuma mensagem no grupo ainda."
		}

		grupos = append(grupos, g)
	}

	return grupos, nil
}

// BuscarHistoricoGrupo procura as mensagens de um projeto específico
func BuscarHistoricoGrupo(projetoId, euId int) ([]structs.Mensagem, error) {
	query := `
		SELECT c.id_mensagem, c.id_remetente, c.id_projeto, c.mensagem, c.data_envio, u.nome_completo, u.foto_perfil
		FROM chat c
		JOIN usuarios u ON c.id_remetente = u.id_usuario
		WHERE c.id_projeto = $1
		ORDER BY c.data_envio ASC
	`

	rows, err := db.DB.Query(query, projetoId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []structs.Mensagem
	for rows.Next() {
		var m structs.Mensagem
		var fotoNull *string

		err = rows.Scan(&m.Id, &m.RemetenteId, &m.ProjetoId, &m.Conteudo, &m.DataEnvio, &m.NomeRemetente, &fotoNull)
		if err != nil {
			continue
		}

		if fotoNull != nil {
			m.FotoRemetente = *fotoNull
		}

		// Descriptografa a mensagem
		m.Conteudo = utils.DescriptografarMensagem(m.Conteudo)
		m.EhMinha = (m.RemetenteId == euId)
		m.HoraFormatada = m.DataEnvio.Format("15:04")
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func EnviarMensagemGrupo(remetenteId, projetoId int, conteudo string) error {
	conteudoCriptografado, err := utils.CriptografarMensagem(conteudo)
	if err != nil {
		return err
	}
	// O id_destinatario fica NULL porque é um grupo
	query := `INSERT INTO chat (id_remetente, id_projeto, mensagem) VALUES ($1, $2, $3)`
	_, err = db.DB.Exec(query, remetenteId, projetoId, conteudoCriptografado)
	return err
}

// BuscarIDsMembrosProjeto pega o ID do Líder + IDs da Equipe
func BuscarIDsMembrosProjeto(projetoId int) ([]int, error) {
	query := `
		SELECT id_usuario FROM equipe_projeto WHERE id_projeto = $1
		UNION
		SELECT id_lider FROM projetos WHERE id = $1
	`
	rows, err := db.DB.Query(query, projetoId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
