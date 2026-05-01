package structs

import (
	"strings"
	"time"
)

// ==========================================
// 1. GESTÃO DO SISTEMA (QUEM ACESSA O PAINEL)
// ==========================================

type Usuario struct {
	IdUsuario       int       `json:"id_usuario" db:"id_usuario" gorm:"primaryKey;column:id_usuario;autoIncrement"`
	NomeCompleto    string    `json:"nome_completo" db:"nome_completo" gorm:"column:nome_completo;type:varchar(100);not null"`
	Email           string    `json:"email" db:"email" gorm:"column:email;type:varchar(100);unique;not null"`
	SenhaHash       string    `json:"-" db:"senha_hash" gorm:"column:senha_hash;type:varchar(255);not null"`
	IdCursoAnalista *int      `json:"id_curso_analista" db:"id_curso_analista" gorm:"column:id_curso_analista;type:int4"`
	FotoPerfil      string    `json:"foto_perfil" db:"foto_perfil" gorm:"column:foto_perfil;type:varchar(255)"`
	DataCadastro    time.Time `json:"data_cadastro" db:"data_cadastro" gorm:"column:data_cadastro;type:timestamp;default:CURRENT_TIMESTAMP"`

	// Ignorado pelo GORM para evitar conflitos de chaves estrangeiras reversas
	Curso *Curso `json:"-" db:"-" gorm:"-"`
}

func (u Usuario) PrimeiroNome() string {
	if u.NomeCompleto == "" {
		return ""
	}
	return strings.Split(u.NomeCompleto, " ")[0]
}

func (Usuario) TableName() string { return "usuarios" }

type RecuperacaoSenha struct {
	Id        int       `db:"id" gorm:"primaryKey;column:id;autoIncrement"`
	Email     string    `db:"email" gorm:"column:email;type:varchar(255);not null"`
	Codigo    string    `db:"codigo" gorm:"column:codigo;type:varchar(6);not null"`
	Expiracao time.Time `db:"expiracao" gorm:"column:expiracao;type:timestamp;not null"`
	Usado     bool      `db:"usado" gorm:"column:usado;default:false"`
	CriadoEm  time.Time `db:"criado_em" gorm:"column:criado_em;type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (RecuperacaoSenha) TableName() string { return "recuperacao_senha" }

// ==========================================
// 2. DOMÍNIOS
// ==========================================

type Curso struct {
	IdCurso   int    `json:"id_curso" db:"id_curso" gorm:"primaryKey;column:id_curso;autoIncrement"`
	NomeCurso string `json:"nome_curso" db:"nome_curso" gorm:"column:nome_curso;type:varchar(100);unique;not null"`
	AreaCurso string `json:"area_curso" db:"area_curso" gorm:"column:area_curso;type:varchar(100)"`
}

func (Curso) TableName() string { return "cursos" }

type Area struct {
	IdArea   int    `json:"id_area" db:"id_area" gorm:"primaryKey;column:id_area;autoIncrement"`
	NomeArea string `json:"nome_area" db:"nome_area" gorm:"column:nome_area;type:varchar(50);unique;not null"`
}

func (Area) TableName() string { return "areas" }

type Habilidade struct {
	IdHabilidade int    `json:"id_habilidade" db:"id_habilidade" gorm:"primaryKey;column:id_habilidade;autoIncrement"`
	NomeHab      string `json:"nome_hab" db:"nome_hab" gorm:"column:nome_hab;type:varchar(100);unique;not null"`
	TipoHab      string `json:"tipo_hab" db:"tipo_hab" gorm:"column:tipo_hab;type:varchar(30);default:'TECNICA';check:tipo_hab IN ('TECNICA', 'COMPORTAMENTAL')"`
}

func (Habilidade) TableName() string { return "habilidades" }

// ==========================================
// 3. VITRINE DE TALENTOS (ALUNOS)
// ==========================================

type Aluno struct {
	IdAluno       int       `json:"id_aluno" db:"id_aluno" gorm:"primaryKey;column:id_aluno;autoIncrement"`
	NomeCompleto  string    `json:"nome_completo" db:"nome_completo" gorm:"column:nome_completo;type:varchar(100);not null"`
	IdCurso       *int      `json:"id_curso" db:"id_curso" gorm:"column:id_curso;type:int4"`
	SemestreAtual *int      `json:"semestre_atual" db:"semestre_atual" gorm:"column:semestre_atual;type:int2"`
	Biografia     string    `json:"biografia" db:"biografia" gorm:"column:biografia;type:text"`
	FotoPerfil    string    `json:"foto_perfil" db:"foto_perfil" gorm:"column:foto_perfil;type:varchar(255)"`
	EmailContato  string    `json:"email_contato" db:"email_contato" gorm:"column:email_contato;type:varchar(150)"`
	LinkedinLink  string    `json:"linkedin_link" db:"linkedin_link" gorm:"column:linkedin_link;type:varchar(255)"`
	GithubLink    string    `json:"github_link" db:"github_link" gorm:"column:github_link;type:varchar(255)"`
	PortfolioLink string    `json:"portfolio_link" db:"portfolio_link" gorm:"column:portfolio_link;type:varchar(255)"`
	CadastradoPor *int      `json:"cadastrado_por" db:"cadastrado_por" gorm:"column:cadastrado_por;type:int4"`
	DataCadastro  time.Time `json:"data_cadastro" db:"data_cadastro" gorm:"column:data_cadastro;type:timestamp;default:CURRENT_TIMESTAMP"`

	// Relacionamentos Complexos (Mantidos via SQL puro, ignorados no AutoMigrate)
	Curso       Curso    `json:"curso" db:"-" gorm:"-"`
	Cadastrador *Usuario `json:"-" db:"-" gorm:"-"`

	// Many2Many Funciona perfeitamente no GORM
	Habilidades []Habilidade `json:"habilidades" db:"-" gorm:"many2many:aluno_habilidades;joinForeignKey:id_aluno;joinReferences:id_habilidade;constraint:OnDelete:CASCADE;"`
}

func (a Aluno) PrimeiroNome() string {
	if a.NomeCompleto == "" {
		return ""
	}
	return strings.Split(a.NomeCompleto, " ")[0]
}

func (Aluno) TableName() string { return "alunos" }

// ==========================================
// 4. VITRINE DE PROJETOS MULTIDISCIPLINARES
// ==========================================

type Projeto struct {
	IdProjeto           int       `json:"id_projeto" db:"id_projeto" gorm:"primaryKey;column:id_projeto;autoIncrement"`
	Titulo              string    `json:"titulo" db:"titulo" gorm:"column:titulo;type:varchar(100);not null"`
	Descricao           string    `json:"descricao" db:"descricao" gorm:"column:descricao;type:text;not null"`
	IdCurso             *int      `json:"id_curso" db:"id_curso" gorm:"column:id_curso;type:int4"`
	IdArea              *int      `json:"id_area" db:"id_area" gorm:"column:id_area;type:int4"`
	SemestreLetivo      string    `json:"semestre_letivo" db:"semestre_letivo" gorm:"column:semestre_letivo;type:varchar(20)"`
	ProfessorOrientador string    `json:"professor_orientador" db:"professor_orientador" gorm:"column:professor_orientador;type:varchar(100)"`
	StatusProjeto       string    `json:"status_projeto" db:"status_projeto" gorm:"column:status_projeto;type:varchar(30);default:'PLANEJADO';check:status_projeto IN ('PLANEJADO', 'EM_ANDAMENTO', 'CONCLUIDO', 'OCULTO')"`
	ImagemCapa          string    `json:"imagem_capa" db:"imagem_capa" gorm:"column:imagem_capa;type:varchar(255)"`
	LinkRepositorio     string    `json:"link_repositorio" db:"link_repositorio" gorm:"column:link_repositorio;type:varchar(255)"`
	CadastradoPor       *int      `json:"cadastrado_por" db:"cadastrado_por" gorm:"column:cadastrado_por;type:int4"`
	DataCriacao         time.Time `json:"data_criacao" db:"data_criacao" gorm:"column:data_criacao;type:timestamp;default:CURRENT_TIMESTAMP"`
	DataAtualizacao     time.Time `json:"data_atualizacao" db:"data_atualizacao" gorm:"column:data_atualizacao;type:timestamp;default:CURRENT_TIMESTAMP"`

	MediaEstrelas   float64 `json:"media_estrelas" db:"media_estrelas" gorm:"column:media_estrelas;type:numeric(3,2);default:0"`
	TotalAvaliacoes int     `json:"total_avaliacoes" db:"total_avaliacoes" gorm:"column:total_avaliacoes;type:int4;default:0"`

	// Ignorados pelo Gorm para evitar o bug de Foreign Key invertida
	Area        Area     `json:"area" db:"-" gorm:"-"`
	Curso       Curso    `json:"curso" db:"-" gorm:"-"`
	Cadastrador *Usuario `json:"-" db:"-" gorm:"-"`

	// Relacionamentos HasMany e Many2Many (Funcionam perfeitamente)
	Equipe     []Aluno          `json:"equipe" db:"-" gorm:"many2many:projeto_alunos;joinForeignKey:id_projeto;joinReferences:id_aluno;constraint:OnDelete:CASCADE;"`
	Arquivos   []ProjetoArquivo `json:"arquivos" db:"-" gorm:"foreignKey:IdProjeto;constraint:OnDelete:CASCADE;"`
	Links      []ProjetoLink    `json:"links" db:"-" gorm:"foreignKey:IdProjeto;constraint:OnDelete:CASCADE;"`
	Imagens    []ProjetoImagem  `json:"imagens" db:"-" gorm:"foreignKey:IdProjeto;constraint:OnDelete:CASCADE;"`
	Avaliacoes []Avaliacao      `json:"avaliacoes" db:"-" gorm:"foreignKey:IdProjeto;constraint:OnDelete:CASCADE;"`
}

func (Projeto) TableName() string { return "projetos" }

type ProjetoArquivo struct {
	IdArquivo      int       `json:"id_arquivo" db:"id_arquivo" gorm:"primaryKey;column:id_arquivo;autoIncrement"`
	IdProjeto      *int      `json:"id_projeto" db:"id_projeto" gorm:"column:id_projeto;type:int4"`
	NomeOriginal   string    `json:"nome_original" db:"nome_original" gorm:"column:nome_original;type:varchar(255);not null"`
	CaminhoArquivo string    `json:"caminho_arquivo" db:"caminho_arquivo" gorm:"column:caminho_arquivo;type:varchar(255);not null"`
	DataUpload     time.Time `json:"data_upload" db:"data_upload" gorm:"column:data_upload;type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (ProjetoArquivo) TableName() string { return "projeto_arquivos" }

type ProjetoLink struct {
	IdLink    int    `json:"id_link" db:"id_link" gorm:"primaryKey;column:id_link;autoIncrement"`
	IdProjeto *int   `json:"id_projeto" db:"id_projeto" gorm:"column:id_projeto;type:int4"`
	TipoLink  string `json:"tipo_link" db:"tipo_link" gorm:"column:tipo_link;type:varchar(30);default:'OUTRO';check:tipo_link IN ('YOUTUBE', 'FORMS', 'PUBLICACAO', 'OUTRO')"`
	Url       string `json:"url" db:"url" gorm:"column:url;type:varchar(500);not null"`
	Descricao string `json:"descricao" db:"descricao" gorm:"column:descricao;type:varchar(255)"`
}

func (ProjetoLink) TableName() string { return "projeto_links" }

type ProjetoImagem struct {
	IdImagem      int       `json:"id_imagem" db:"id_imagem" gorm:"primaryKey;column:id_imagem;autoIncrement"`
	IdProjeto     *int      `json:"id_projeto" db:"id_projeto" gorm:"column:id_projeto;type:int4"`
	CaminhoImagem string    `json:"caminho_imagem" db:"caminho_imagem" gorm:"column:caminho_imagem;type:varchar(255);not null"`
	Ordem         int       `json:"ordem" db:"ordem" gorm:"column:ordem;type:int4;default:0"`
	DataUpload    time.Time `json:"data_upload" db:"data_upload" gorm:"column:data_upload;type:timestamp;default:CURRENT_TIMESTAMP"`
}

func (ProjetoImagem) TableName() string { return "projeto_imagens" }

// ==========================================
// 5. INTERAÇÃO PÚBLICA
// ==========================================

type Avaliacao struct {
	IdAvaliacao    int       `json:"id_avaliacao" db:"id_avaliacao" gorm:"primaryKey;column:id_avaliacao;autoIncrement"`
	IdProjeto      *int      `json:"id_projeto" db:"id_projeto" gorm:"column:id_projeto;type:int4"`
	NomeAvaliador  string    `json:"nome_avaliador" db:"nome_avaliador" gorm:"column:nome_avaliador;type:varchar(255);not null"`
	EmailAvaliador string    `json:"email_avaliador" db:"email_avaliador" gorm:"column:email_avaliador;type:varchar(150);not null"`
	Nota           int       `json:"nota" db:"nota" gorm:"column:nota;type:int4;check:nota >= 1 AND nota <= 5"`
	Comentario     string    `json:"comentario" db:"comentario" gorm:"column:comentario;type:text"`
	DataAvaliacao  time.Time `json:"data_avaliacao" db:"data_avaliacao" gorm:"column:data_avaliacao;type:timestamp;default:CURRENT_TIMESTAMP"`
	DataFormatada  string    `json:"data_formatada" db:"-" gorm:"-"`
}

func (Avaliacao) TableName() string { return "avaliacoes" }
