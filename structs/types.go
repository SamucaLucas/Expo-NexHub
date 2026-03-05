package structs

import (
	"strings"
	"time"
)

// Usuario representa a tabela 'usuarios' do banco de dados
type Usuario struct {
	Id           int
	NomeCompleto string
	Email        string
	SenhaHash    string
	Nivel        string
	TipoUsuario  string // 'ADMIN', 'DEV', 'EMPRESA'
	StatusConta  string // 'ATIVO', 'BANIDO'

	// Campos Gerais
	Cidade       string
	Biografia    string
	FotoPerfil   string
	DataCadastro time.Time

	// Campos Específicos de Desenvolvedor (DEV)
	TituloProfissional    string
	DisponivelParaEquipes bool
	GithubLink            string
	LinkedinLink          string
	PortfolioLink         string
	Skills                string

	// Campos Específicos de Empresa (EMPRESA)
	NomeFantasia string
	SiteEmpresa  string
	RamoAtuacao  string
	EstaSalvo    bool
	IsBanned     bool
}

// structs/types.go

// PrimeiroNome retorna apenas a primeira palavra do NomeCompleto
func (u Usuario) PrimeiroNome() string {
	if u.NomeCompleto == "" {
		return ""
	}
	// Separa por espaço e pega o primeiro item
	return strings.Split(u.NomeCompleto, " ")[0]
}

func (u Usuario) SkillsComoLista() []string {
	if u.Skills == "" {
		return []string{}
	}
	// Separa por vírgula
	listaBruta := strings.Split(u.Skills, ",")

	var listaLimpa []string
	for _, item := range listaBruta {
		// Remove espaços extras (ex: " Java" vira "Java")
		s := strings.TrimSpace(item)
		if s != "" {
			listaLimpa = append(listaLimpa, s)
		}
	}
	return listaLimpa
}

// Projeto representa a tabela 'projetos'
type Projeto struct {
	Id            int
	Titulo        string
	Descricao     string
	Status        string // 'Planejado', 'Em Andamento', 'Concluido'
	Cidade        string // Mapeia para 'cidade_projeto'
	Categoria     string // Mapeia para 'categoria' (NOVO)
	ImagemCapa    string
	LinkRepo      string
	Tags          string
	Visualizacoes int
	IdLider       int

	ImagensGaleria []ImagemGaleria
	Tecnologias    []string

	NomeLider string
	FotoLider string
	Equipe    []MembroEquipe
	EstaSalvo bool

	MediaEstrelas   float64
	TotalAvaliacoes int
}

// MembroEquipe representa uma linha da tabela equipe_projeto + dados do usuario
type MembroEquipe struct {
	IdUsuario   int
	Nome        string
	Foto        string
	Funcao      string // O cargo (ex: "Frontend Dev")
	DataEntrada time.Time
}

// Helper para tags (já existente)
func (p Projeto) TagsComoLista() []string {
	if p.Tags == "" {
		return []string{}
	}
	listaBruta := strings.Split(p.Tags, ",")
	var listaLimpa []string
	for _, item := range listaBruta {
		s := strings.TrimSpace(item)
		if s != "" {
			listaLimpa = append(listaLimpa, s)
		}
	}
	return listaLimpa
}

type DashboardStats struct {
	TotalProjetos      int
	TotalVisualizacoes int
	MensagensNaoLidas  int // Vamos deixar 0 por enquanto
}

type ImagemGaleria struct {
	Id      int    // ID na tabela projeto_imagens
	Caminho string // O caminho /static/uploads/...
}

type Tecnologia struct {
	Id   int
	Nome string
}

type AdminDashboardData struct {
	Usuario Usuario
	TotalDevs     int
	TotalEmpresas int
	TotalProjetos int
	TotalBanidos  int

	// Para os Gráficos
	ChartMeses      []string // Ex: ["Jan", "Fev"]
	ChartNovosUsers []int    // Ex: [10, 20]
	NaoLidas        int
}

type ProjetoAdmin struct {
	Id        int
	Titulo    string
	DonoNome  string // Nome do usuário dono
	Categoria string
	Status    string // "Concluido", "Em Andamento", "Oculto"
}

type Mensagem struct {
	Id             int       // No banco: id_mensagem
	RemetenteId    int       // No banco: id_remetente
	DestinatarioId *int      // No banco: id_destinatario
	ProjetoId      *int       
	Conteudo       string    // No banco: mensagem
	DataEnvio      time.Time // No banco: data_envio
	Lido           bool      // No banco: lida

	// Campos auxiliares para o Front-end
	EhMinha       bool
	HoraFormatada string
	NomeRemetente  string
	FotoRemetente  string
}

type ContatoChat struct {
	UsuarioId      int
	Nome           string
	Avatar         string
	UltimaMensagem string
	DataUltima     time.Time
	Ativo          bool
	NaoLidas       int
	TipoUsuario    string
}

// Representa um chat de grupo (Projeto) na barra lateral
type GrupoChat struct {
    ProjetoId      int
    NomeProjeto    string
    CapaProjeto    string // Imagem do projeto
    UltimaMensagem string
    NaoLidas       int
}

type DadosChat struct {
	UsuarioLogado      Usuario
	Contatos           []ContatoChat
	Grupos             []GrupoChat
	ConversaAtual      []Mensagem
	Destinatario       Usuario
	ChatAberto         bool // Se tem alguém selecionado
	DestinatarioBanido bool

	IsGrupo            bool
	ProjetoAtual       Projeto
}

// Estrutura para enviar dados via Socket
type NotificacaoMensagem struct {
	Tipo          string `json:"tipo"` // ex: "nova_mensagem"
	Conteudo      string `json:"conteudo"`
	RemetenteID   int    `json:"remetente_id"`
	HoraFormatada string `json:"hora"`
}
