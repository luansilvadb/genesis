package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ── Enums ────────────────────────────────────────────────────────────────────

type Role string

const (
	RoleAdmin        Role = "ADMIN"
	RoleMorador      Role = "MORADOR"
	RoleVisualizador Role = "VISUALIZADOR"
)

type SplitMode string

const (
	SplitModeEqual  SplitMode = "EQUAL"
	SplitModeIncome SplitMode = "INCOME"
	SplitModeCustom SplitMode = "CUSTOM"
)

type ValidationEventType string

// ── Models ───────────────────────────────────────────────────────────────────

type Tenant struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name            string    `gorm:"not null" json:"name"`
	InviteCode      string    `gorm:"uniqueIndex;not null" json:"inviteCode"`
	PermissionsJSON *string   `gorm:"column:permissions_json;type:jsonb" json:"permissionsJson,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (Tenant) TableName() string { return "tenants" }

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return nil
}

type Usuario struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Nome         string    `gorm:"not null" json:"nome"`
	PasswordHash *string   `gorm:"column:password_hash" json:"-"`
	GoogleID     *string   `gorm:"uniqueIndex;column:google_id" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Usuario) TableName() string { return "usuarios" }

func (u *Usuario) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

type MembroCasa struct {
	ID            string    `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;type:uuid;primaryKey" json:"tenantId"`
	Nome          string    `gorm:"not null" json:"nome"`
	Avatar        string    `gorm:"not null" json:"avatar"`
	Ativo         bool      `gorm:"default:true" json:"ativo"`
	Role          Role      `gorm:"default:MORADOR" json:"role"`
	UserID        *string   `gorm:"column:user_id;type:uuid" json:"userId,omitempty"`
	RendaCentavos *int64    `gorm:"column:renda_centavos" json:"rendaCentavos,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"createdAt"`

	Tenant  *Tenant  `gorm:"foreignKey:TenantID" json:"-"`
	Usuario *Usuario `gorm:"foreignKey:UserID" json:"-"`
}

func (MembroCasa) TableName() string { return "membros_casa" }

type Cartao struct {
	ID                  string    `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID            string    `gorm:"column:tenant_id;type:uuid;primaryKey" json:"tenantId"`
	Nome                string    `gorm:"not null" json:"nome"`
	DiaFechamento       int       `gorm:"column:dia_fechamento;not null" json:"diaFechamento"`
	ResponsavelPadraoID string    `gorm:"column:responsavel_padrao_id;not null" json:"responsavelPadraoId"`
	CreatedAt           time.Time `gorm:"autoCreateTime" json:"createdAt"`

	Tenant            *Tenant     `gorm:"foreignKey:TenantID" json:"-"`
	ResponsavelPadrao *MembroCasa `gorm:"foreignKey:ResponsavelPadraoID,TenantID;references:ID,TenantID" json:"-"`
}

func (Cartao) TableName() string { return "cartoes" }

type Fatura struct {
	ID                 string     `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID           string     `gorm:"column:tenant_id;type:uuid;primaryKey" json:"tenantId"`
	CartaoID           string     `gorm:"column:cartao_id;not null" json:"cartaoId"`
	Mes                int        `gorm:"not null" json:"mes"`
	Ano                int        `gorm:"not null" json:"ano"`
	ResponsavelID      string     `gorm:"column:responsavel_id;not null" json:"responsavelId"`
	Status             string     `gorm:"not null" json:"status"`
	DataPagamentoBanco *time.Time `gorm:"column:data_pagamento_banco" json:"dataPagamentoBanco,omitempty"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"createdAt"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"-"`
}

func (Fatura) TableName() string { return "faturas" }

type Gasto struct {
	ID                 string    `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID           string    `gorm:"column:tenant_id;type:uuid;primaryKey" json:"tenantId"`
	FaturaID           *string   `gorm:"column:fatura_id" json:"faturaId,omitempty"`
	Descricao          string    `gorm:"not null" json:"descricao"`
	ValorTotalCentavos int64     `gorm:"column:valor_total_centavos;not null" json:"valorTotalCentavos"`
	CompradorID        string    `gorm:"column:comprador_id;not null" json:"compradorId"`
	Installments       int       `gorm:"default:1" json:"installments"`
	TotalInstallments  int       `gorm:"column:total_installments;default:1" json:"totalInstallments"`
	IsLoan             bool      `gorm:"column:is_loan;default:false" json:"isLoan"`
	BorrowerID         *string   `gorm:"column:borrower_id" json:"borrowerId,omitempty"`
	RecurringBillID    *string   `gorm:"column:recurring_bill_id" json:"recurringBillId,omitempty"`
	IsSettlement       bool      `gorm:"column:is_settlement;default:false" json:"isSettlement"`
	SettlementDetails  *string   `gorm:"column:settlement_details;type:jsonb" json:"settlementDetails,omitempty"`
	Method             string    `gorm:"default:pix" json:"method"`
	CardOwnerID        *string   `gorm:"column:card_owner_id" json:"cardOwnerId,omitempty"`
	GrupoParcelasID    *string   `gorm:"column:grupo_parcelas_id" json:"grupoParcelasId,omitempty"`
	IsPrivate          bool      `gorm:"column:is_private;default:false" json:"isPrivate"`
	SplitMode          SplitMode `gorm:"column:split_mode;default:CUSTOM" json:"splitMode"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"createdAt"`

	Divisoes []DivisaoGasto `gorm:"foreignKey:GastoID,TenantID;references:ID,TenantID;constraint:OnDelete:CASCADE" json:"divisoes,omitempty"`
}

func (Gasto) TableName() string { return "gastos" }

type DivisaoGasto struct {
	ID            string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID      string `gorm:"column:tenant_id;type:uuid;index" json:"tenantId"`
	GastoID       string `gorm:"column:gasto_id" json:"gastoId"`
	MembroID      string `gorm:"column:membro_id" json:"membroId"`
	ValorCentavos int64  `gorm:"column:valor_centavos;not null" json:"valorCentavos"`
}

func (DivisaoGasto) TableName() string { return "divisoes_gasto" }

type ContaFixa struct {
	ID                 string    `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID           string    `gorm:"column:tenant_id;type:uuid;primaryKey" json:"tenantId"`
	Name               string    `gorm:"not null" json:"name"`
	Icon               string    `gorm:"not null" json:"icon"`
	FixedValueCentavos *int64    `gorm:"column:fixed_value_centavos" json:"fixedValueCentavos,omitempty"`
	DefaultSplit       string    `gorm:"type:jsonb;default:'[]'" json:"defaultSplit"`
	CreatedAt          time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (ContaFixa) TableName() string { return "contas_fixas" }

type AuditLog struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;type:uuid" json:"tenantId"`
	MembroID  string    `gorm:"column:membro_id" json:"membroId"`
	Acao      string    `gorm:"not null" json:"acao"`
	Detalhes  string    `gorm:"not null" json:"detalhes"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type ProductValidationEvent struct {
	ID        string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  string              `gorm:"column:tenant_id;type:uuid;uniqueIndex:idx_validation_dedupe" json:"tenant_id"`
	Type      ValidationEventType `gorm:"not null;uniqueIndex:idx_validation_dedupe" json:"type"`
	DedupeKey string              `gorm:"column:dedupe_key;not null;uniqueIndex:idx_validation_dedupe" json:"dedupe_key"`
	PeriodKey *string             `gorm:"column:period_key" json:"period_key,omitempty"`
	CreatedAt time.Time           `gorm:"autoCreateTime" json:"created_at"`
}

func (ProductValidationEvent) TableName() string { return "product_validation_events" }

type PasswordResetToken struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	UserID    string    `gorm:"column:user_id;type:uuid;not null" json:"user_id"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null" json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PasswordResetToken) TableName() string { return "password_reset_tokens" }
