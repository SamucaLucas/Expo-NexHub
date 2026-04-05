package structs

import (
	"strings"
	"time"
)

// ==========================================
// 1. GESTÃO DO SISTEMA (QUEM ACESSA O PAINEL)
// ==========================================

// Usuario representa os Analistas de ADS (Admins)
type Usuario struct {
	IdUsuario       int       `json:"id_usuario" db:"id_usuario"`
	NomeCompleto    string    `json:"nome_completo" db:"nome_completo"`
	Email           string    `json:"email" db:"email"`
	SenhaHash       string    `json:"-" db:"senha_hash"`
	IdCursoAnalista *int      `json:"id_curso_analista" db:"id_curso_analista"`
	FotoPerfil      string    `json:"foto_perfil" db:"foto_perfil"`
	DataCadastro    time.Time `json:"data_cadastro" db:"data_cadastro"`
}

func (u Usuario) PrimeiroNome() string {
	if u.NomeCompleto == "" {
		return ""
	}
	return strings.Split(u.NomeCompleto, " ")[0]
}

type RecuperacaoSenha struct {
	Id        int       `db:"id"`
	Email     string    `db:"email"`
	Codigo    string    `db:"codigo"`
	Expiracao time.Time `db:"expiracao"`
	Usado     bool      `db:"usado"`
	CriadoEm  time.Time `db:"criado_em"`
}

// ==========================================
// 2. DOMÍNIOS
// ==========================================

type Curso struct {
	IdCurso   int    `json:"id_curso" db:"id_curso"`
	NomeCurso string `json:"nome_curso" db:"nome_curso"`
	AreaCurso string `json:"area_curso" db:"area_curso"`
}

type Area struct {
	IdArea   int    `json:"id_area" db:"id_area"`
	NomeArea string `json:"nome_area" db:"nome_area"`
}

// Habilidade substitui a antiga struct 'Tecnologia'
type Habilidade struct {
	IdHabilidade int    `json:"id_habilidade" db:"id_habilidade"`
	NomeHab      string `json:"nome_hab" db:"nome_hab"`
	TipoHab      string `json:"tipo_hab" db:"tipo_hab"` // 'TECNICA' ou 'COMPORTAMENTAL'
}

// ==========================================
// 3. VITRINE DE TALENTOS (ALUNOS)
// ==========================================

type Aluno struct {
	IdAluno       int       `json:"id_aluno" db:"id_aluno"`
	NomeCompleto  string    `json:"nome_completo" db:"nome_completo"`
	IdCurso       *int      `json:"id_curso" db:"id_curso"`
	SemestreAtual *int      `json:"semestre_atual" db:"semestre_atual"`
	Biografia     string    `json:"biografia" db:"biografia"`
	FotoPerfil    string    `json:"foto_perfil" db:"foto_perfil"`
	EmailContato  string    `json:"email_contato" db:"email_contato"`
	LinkedinLink  string    `json:"linkedin_link" db:"linkedin_link"`
	GithubLink    string    `json:"github_link" db:"github_link"`
	PortfolioLink string    `json:"portfolio_link" db:"portfolio_link"`
	CadastradoPor *int      `json:"cadastrado_por" db:"cadastrado_por"` // Qual Admin cadastrou
	DataCadastro  time.Time `json:"data_cadastro" db:"data_cadastro"`

	// Relacionamentos para facilitar no Front-end
	Curso       Curso        `json:"curso"`
	Habilidades []Habilidade `json:"habilidades"`
}

func (a Aluno) PrimeiroNome() string {
	if a.NomeCompleto == "" {
		return ""
	}
	return strings.Split(a.NomeCompleto, " ")[0]
}

// ==========================================
// 4. VITRINE DE PROJETOS MULTIDISCIPLINARES
// ==========================================

type Projeto struct {
	IdProjeto           int       `json:"id_projeto" db:"id_projeto"`
	Titulo              string    `json:"titulo" db:"titulo"`
	Descricao           string    `json:"descricao" db:"descricao"`
	IdCurso             *int      `json:"id_curso" db:"id_curso"`
	IdArea              *int      `json:"id_area" db:"id_area"`
	SemestreLetivo      string    `json:"semestre_letivo" db:"semestre_letivo"`
	ProfessorOrientador string    `json:"professor_orientador" db:"professor_orientador"`
	StatusProjeto       string    `json:"status_projeto" db:"status_projeto"`
	ImagemCapa          string    `json:"imagem_capa" db:"imagem_capa"`
	LinkRepositorio     string    `json:"link_repositorio" db:"link_repositorio"`
	CadastradoPor       *int      `json:"cadastrado_por" db:"cadastrado_por"`
	DataCriacao         time.Time `json:"data_criacao" db:"data_criacao"`
	DataAtualizacao     time.Time `json:"data_atualizacao" db:"data_atualizacao"`

	// Relacionamentos populados nas queries
	Area       Area             `json:"area"`
	Curso      Curso            `json:"curso"`
	Equipe     []Aluno          `json:"equipe"` // Usado com JOIN na tabela projeto_alunos
	Arquivos   []ProjetoArquivo `json:"arquivos"`
	Links      []ProjetoLink    `json:"links"`
	Imagens    []ProjetoImagem  `json:"imagens"`
	Avaliacoes []Avaliacao      `json:"avaliacoes"`

	// Campos Auxiliares para o Front-end
	MediaEstrelas   float64 `json:"media_estrelas"`
	TotalAvaliacoes int     `json:"total_avaliacoes"`
}

type ProjetoArquivo struct {
	IdArquivo      int       `json:"id_arquivo" db:"id_arquivo"`
	IdProjeto      int       `json:"id_projeto" db:"id_projeto"`
	NomeOriginal   string    `json:"nome_original" db:"nome_original"`
	CaminhoArquivo string    `json:"caminho_arquivo" db:"caminho_arquivo"`
	DataUpload     time.Time `json:"data_upload" db:"data_upload"`
}

type ProjetoLink struct {
	IdLink    int    `json:"id_link" db:"id_link"`
	IdProjeto int    `json:"id_projeto" db:"id_projeto"`
	TipoLink  string `json:"tipo_link" db:"tipo_link"` // 'YOUTUBE', 'FORMS', 'PUBLICACAO', 'OUTRO'
	Url       string `json:"url" db:"url"`
	Descricao string `json:"descricao" db:"descricao"`
}

type ProjetoImagem struct {
	IdImagem      int       `json:"id_imagem" db:"id_imagem"`
	IdProjeto     int       `json:"id_projeto" db:"id_projeto"`
	CaminhoImagem string    `json:"caminho_imagem" db:"caminho_imagem"`
	Ordem         int       `json:"ordem" db:"ordem"`
	DataUpload    time.Time `json:"data_upload" db:"data_upload"`
}

// ==========================================
// 5. INTERAÇÃO PÚBLICA
// ==========================================

type Avaliacao struct {
	IdAvaliacao    int       `json:"id_avaliacao" db:"id_avaliacao"`
	IdProjeto      int       `json:"id_projeto" db:"id_projeto"`
	NomeAvaliador  string    `json:"nome_avaliador" db:"nome_avaliador"`
	EmailAvaliador string    `json:"email_avaliador" db:"email_avaliador"`
	Nota           int       `json:"nota" db:"nota"`
	Comentario     string    `json:"comentario" db:"comentario"`
	DataAvaliacao  time.Time `json:"data_avaliacao" db:"data_avaliacao"`
}
