package dto

import (
	"encoding/json"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
)

// NullString is a string that tracks whether it was explicitly set in JSON.
// nil pointer = field not present in JSON.
// pointer to empty string = field explicitly set to null (clear it).
type NullString struct {
	Value *string
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ns.Value = &s
	return nil
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Value == nil {
		return []byte("null"), nil
	}
	return json.Marshal(*ns.Value)
}

type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateMembroRequest struct {
	Nome   string `json:"nome" binding:"required"`
	Avatar string `json:"avatar" binding:"required"`
}

type CreateMembroWithAccountRequest struct {
	Nome     string `json:"nome" binding:"required"`
	Avatar   string `json:"avatar" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type CreateCartaoRequest struct {
	Nome                string `json:"nome" binding:"required"`
	DiaFechamento       int    `json:"diaFechamento" binding:"required,min=1,max=31"`
	ResponsavelPadraoID string `json:"responsavelPadraoId" binding:"required"`
}

type CreateGastoRequest struct {
	Descricao          string      `json:"descricao" binding:"required"`
	ValorTotalCentavos int64       `json:"valorTotalCentavos" binding:"required,min=1"`
	CompradorID        string      `json:"compradorId" binding:"required"`
	FaturaID           *string     `json:"faturaId,omitempty"`
	Installments       *int        `json:"installments,omitempty"`
	TotalInstallments  *int        `json:"totalInstallments,omitempty"`
	IsLoan             bool        `json:"isLoan,omitempty"`
	BorrowerID         *string     `json:"borrowerId,omitempty"`
	Method             string      `json:"method,omitempty" binding:"omitempty,oneof=pix card cash"`
	CardOwnerID        *string     `json:"cardOwnerId,omitempty"`
	IsPrivate          bool        `json:"isPrivate,omitempty"`
	IsSettlement       bool        `json:"isSettlement,omitempty"`
	SettlementDetails  *json.RawMessage `json:"settlementDetails,omitempty"`
	GrupoParcelasID    *string          `json:"grupoParcelasId,omitempty"`
	RecurringBillID    *string          `json:"recurringBillId,omitempty"`
	SplitMode          string           `json:"splitMode,omitempty"`
	Divisoes           []SplitItem      `json:"divisoes,omitempty"`
}

type SplitItem struct {
	MembroID      string `json:"membroId" binding:"required"`
	ValorCentavos int64  `json:"valorCentavos" binding:"required"`
}

type CartaoResponse struct {
	ID                  string `json:"id"`
	Nome                string `json:"nome"`
	DiaFechamento       int    `json:"diaFechamento"`
	ResponsavelPadraoID string `json:"responsavelPadraoId"`
}

func CartaoToResponse(c *model.Cartao) *CartaoResponse {
	return &CartaoResponse{
		ID:                  c.ID,
		Nome:                c.Nome,
		DiaFechamento:       c.DiaFechamento,
		ResponsavelPadraoID: c.ResponsavelPadraoID,
	}
}

func CartoesToResponse(cartoes []model.Cartao) []CartaoResponse {
	resp := make([]CartaoResponse, len(cartoes))
	for i, c := range cartoes {
		resp[i] = *CartaoToResponse(&c)
	}
	return resp
}

type FaturaResponse struct {
	ID                 string  `json:"id"`
	CartaoID           string  `json:"cartaoId"`
	Mes                int     `json:"mes"`
	Ano                int     `json:"ano"`
	ResponsavelID      string  `json:"responsavelId"`
	Status             string  `json:"status"`
	DataPagamentoBanco *string `json:"dataPagamentoBanco,omitempty"`
}

func FaturaToResponse(f *model.Fatura) *FaturaResponse {
	var dataPgto *string
	if f.DataPagamentoBanco != nil {
		s := f.DataPagamentoBanco.Format("2006-01-02T15:04:05Z07:00")
		dataPgto = &s
	}
	return &FaturaResponse{
		ID:                 f.ID,
		CartaoID:           f.CartaoID,
		Mes:                f.Mes,
		Ano:                f.Ano,
		ResponsavelID:      f.ResponsavelID,
		Status:             f.Status,
		DataPagamentoBanco: dataPgto,
	}
}

func FaturasToResponse(faturas []model.Fatura) []FaturaResponse {
	resp := make([]FaturaResponse, len(faturas))
	for i, f := range faturas {
		resp[i] = *FaturaToResponse(&f)
	}
	return resp
}

type CreateContaFixaRequest struct {
	Name               string      `json:"name" binding:"required"`
	Icon               string      `json:"icon" binding:"required"`
	FixedValueCentavos *int64      `json:"fixedValueCentavos,omitempty"`
	DefaultSplit       []SplitItem `json:"defaultSplit,omitempty"`
}

type UpdateContaFixaRequest struct {
	Name               *string     `json:"name,omitempty"`
	Icon               *string     `json:"icon,omitempty"`
	FixedValueCentavos *int64      `json:"fixedValueCentavos,omitempty"`
	DefaultSplit       []SplitItem `json:"defaultSplit,omitempty"`
}

type ContaFixaResponse struct {
	ID                 string      `json:"id"`
	Name               string      `json:"name"`
	Icon               string      `json:"icon"`
	FixedValueCentavos *int64      `json:"fixedValueCentavos,omitempty"`
	DefaultSplit       []SplitItem `json:"defaultSplit"`
	CreatedAt          string      `json:"createdAt"`
}

type JoinTenantRequest struct {
	InviteCode string `json:"inviteCode" binding:"required"`
}

type CreateFaturaRequest struct {
	CartaoID           string  `json:"cartaoId" binding:"required"`
	Mes                int     `json:"mes" binding:"required,min=1,max=12"`
	Ano                int     `json:"ano" binding:"required"`
	ResponsavelID      string  `json:"responsavelId" binding:"required"`
	Status             string  `json:"status" binding:"required"`
	DataPagamentoBanco *string `json:"dataPagamentoBanco,omitempty"`
}

type UpdateMembroRequest struct {
	Nome          *string `json:"nome,omitempty"`
	Avatar        *string `json:"avatar,omitempty"`
	Ativo         *bool   `json:"ativo,omitempty"`
	Role          *string `json:"role,omitempty"`
	RendaCentavos *int64  `json:"rendaCentavos,omitempty"`
}

type UpdateGastoRequest struct {
	Descricao          *string     `json:"descricao,omitempty"`
	ValorTotalCentavos *int64      `json:"valorTotalCentavos,omitempty"`
	CompradorID        *string     `json:"compradorId,omitempty"`
	FaturaID           *NullString `json:"faturaId,omitempty"`
	Installments       *int        `json:"installments,omitempty"`
	TotalInstallments  *int        `json:"totalInstallments,omitempty"`
	IsLoan             *bool       `json:"isLoan,omitempty"`
	BorrowerID         *NullString `json:"borrowerId,omitempty"`
	Method             *string     `json:"method,omitempty" binding:"omitempty,oneof=pix card cash"`
	CardOwnerID        *NullString `json:"cardOwnerId,omitempty"`
	IsPrivate          *bool       `json:"isPrivate,omitempty"`
	IsSettlement       *bool            `json:"isSettlement,omitempty"`
	SettlementDetails  *json.RawMessage `json:"settlementDetails,omitempty"`
	GrupoParcelasID    *NullString      `json:"grupoParcelasId,omitempty"`
	RecurringBillID    *NullString      `json:"recurringBillId,omitempty"`
	SplitMode          *string          `json:"splitMode,omitempty"`
	Divisoes           []SplitItem      `json:"divisoes,omitempty"`
}

type DeleteGastoBatchRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

type RolePermissions struct {
	AllowLancarGasto          *bool `json:"ALLOW_LANCAR_GASTO,omitempty"`
	AllowGerenciarCartoes     *bool `json:"ALLOW_GERENCIAR_CARTOES,omitempty"`
	AllowGerenciarContasFixas *bool `json:"ALLOW_GERENCIAR_CONTAS_FIXAS,omitempty"`
	AllowRegistrarNetting     *bool `json:"ALLOW_REGISTRAR_NETTING,omitempty"`
	AllowVerAuditLogs         *bool `json:"ALLOW_VER_AUDIT_LOGS,omitempty"`
	AllowFecharPeriodo        *bool `json:"ALLOW_FECHAR_PERIODO,omitempty"`
	AllowAlterarRenda         *bool `json:"ALLOW_ALTERAR_RENDA,omitempty"`
	AllowAlterarNome          *bool `json:"ALLOW_ALTERAR_NOME,omitempty"`
}

type ValidateEventRequest struct {
	Type      string `json:"type" binding:"required"`
	DedupeKey string `json:"dedupeKey" binding:"required"`
	PeriodKey string `json:"periodKey,omitempty"`
}

type TenantResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"inviteCode"`
	CreatedAt  string `json:"createdAt"`
}

type MembroResponse struct {
	ID            string `json:"id"`
	Nome          string `json:"nome"`
	Avatar        string `json:"avatar"`
	Ativo         bool   `json:"ativo"`
	Role          string `json:"role"`
	RendaCentavos *int64 `json:"rendaCentavos,omitempty"`
	UserID        string `json:"userId,omitempty"`
	CreatedAt     string `json:"createdAt"`
}

type GastoResponse struct {
	ID                 string      `json:"id"`
	Descricao          string      `json:"descricao"`
	ValorTotalCentavos int64       `json:"valorTotalCentavos"`
	CompradorID        string      `json:"compradorId"`
	FaturaID           *string     `json:"faturaId,omitempty"`
	Installments       int         `json:"installments"`
	TotalInstallments  int         `json:"totalInstallments"`
	Method             string      `json:"method"`
	IsLoan             bool        `json:"isLoan"`
	IsPrivate          bool        `json:"isPrivate"`
	BorrowerID         *string     `json:"borrowerId,omitempty"`
	CardOwnerID        *string     `json:"cardOwnerId,omitempty"`
	RecurringBillID    *string     `json:"recurringBillId,omitempty"`
	GrupoParcelasID    *string     `json:"grupoParcelasId,omitempty"`
	IsSettlement       bool             `json:"isSettlement"`
	SettlementDetails  *json.RawMessage `json:"settlementDetails,omitempty"`
	CreatedAt          string           `json:"createdAt"`
	SplitMode          string      `json:"splitMode"`
	Divisoes           []SplitItem `json:"divisoes,omitempty"`
}

type PaginationParams struct {
	Page      int  `form:"page"`
	PageSize  int  `form:"page_size"`
	Paginated bool `form:"paginated"`
}

func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int64 `json:"totalPages"`
}

func NewPaginatedResponse[T any](data []T, total int64, params PaginationParams) PaginatedResponse[T] {
	totalPages := (total + int64(params.PageSize) - 1) / int64(params.PageSize)
	if totalPages < 1 {
		totalPages = 1
	}
	return PaginatedResponse[T]{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}
}
