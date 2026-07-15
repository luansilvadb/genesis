package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
)

type mockErrorRepo struct {
	failCreate bool
	failList   bool
	failExists bool
}

type mockMembroRepoEdge struct {
	repository.MembroRepository
	membros map[string]*model.MembroCasa
	err     mockErrorRepo
}

func (m *mockMembroRepoEdge) Create(ctx context.Context, mb *model.MembroCasa) error {
	if m.err.failCreate {
		return errors.New("database error")
	}
	if mb.ID == "" {
		mb.ID = uuid.New().String()
	}
	m.membros[mb.ID] = mb
	return nil
}

func (m *mockMembroRepoEdge) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.UserID != nil && *mb.UserID == userID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoEdge) ListByTenant(ctx context.Context, tenantID string) ([]model.MembroCasa, error) {
	if m.err.failList {
		return nil, errors.New("database error")
	}
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.TenantID == tenantID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoEdge) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.TenantID == tenantID && mb.UserID != nil && *mb.UserID == userID {
			return mb, nil
		}
	}
	return nil, nil
}

type mockCartaoRepoEdge struct {
	repository.CartaoRepository
	cartoes map[string]*model.Cartao
	err     mockErrorRepo
}

func (m *mockCartaoRepoEdge) Create(ctx context.Context, c *model.Cartao) error {
	if m.err.failCreate {
		return errors.New("database error")
	}
	m.cartoes[c.ID] = c
	return nil
}

type mockGastoRepoEdge struct {
	repository.GastoRepository
	gastos map[string]*model.Gasto
	err    mockErrorRepo
}

func (m *mockGastoRepoEdge) Create(ctx context.Context, g *model.Gasto) error {
	if m.err.failCreate {
		return errors.New("database error")
	}
	m.gastos[g.ID] = g
	return nil
}

func (m *mockGastoRepoEdge) ListByTenant(ctx context.Context, tenantID string) ([]model.Gasto, error) {
	if m.err.failList {
		return nil, errors.New("database error")
	}
	var result []model.Gasto
	for _, g := range m.gastos {
		if g.TenantID == tenantID {
			result = append(result, *g)
		}
	}
	return result, nil
}

func (m *mockGastoRepoEdge) DeleteDivisoes(ctx context.Context, gastoID, tenantID string) error {
	if m.err.failCreate {
		return errors.New("database error")
	}
	return nil
}

type mockContaFixaRepoEdge struct {
	repository.ContaFixaRepository
	contas map[string]*model.ContaFixa
	err    mockErrorRepo
}

func (m *mockContaFixaRepoEdge) Create(ctx context.Context, c *model.ContaFixa) error {
	if m.err.failCreate {
		return errors.New("database error")
	}
	m.contas[c.ID] = c
	return nil
}

func (m *mockContaFixaRepoEdge) GetByID(ctx context.Context, id, tenantID string) (*model.ContaFixa, error) {
	if m.err.failCreate {
		return nil, errors.New("database error")
	}
	c, ok := m.contas[id]
	if !ok || c.TenantID != tenantID {
		return nil, nil
	}
	return c, nil
}

type mockValidationRepoEdge struct {
	repository.ProductValidationRepository
	events map[string]*model.ProductValidationEvent
	err    mockErrorRepo
}

func (m *mockValidationRepoEdge) Create(ctx context.Context, e *model.ProductValidationEvent) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	m.events[e.ID] = e
	return nil
}

func (m *mockValidationRepoEdge) ExistsByDedupeKey(ctx context.Context, tenantID, eventType, dedupeKey string) (bool, error) {
	if m.err.failExists {
		return false, errors.New("database error")
	}
	for _, e := range m.events {
		if e.TenantID == tenantID && string(e.Type) == eventType && e.DedupeKey == dedupeKey {
			return true, nil
		}
	}
	return false, nil
}

func TestFinanceiroService_CreateMembro_RepoError(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{
		membros: make(map[string]*model.MembroCasa),
		err:     mockErrorRepo{failCreate: true},
	}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, err := svc.CreateMembro(context.Background(), "tenant-1", &dto.CreateMembroRequest{
		Nome:   "João",
		Avatar: "default",
	})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestFinanceiroService_CreateCartao_RepoError(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{
		cartoes: make(map[string]*model.Cartao),
		err:     mockErrorRepo{failCreate: true},
	}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, err := svc.CreateCartao(context.Background(), "tenant-1", &dto.CreateCartaoRequest{
		Nome:                "Nubank",
		DiaFechamento:       15,
		ResponsavelPadraoID: "membro-1",
	})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestFinanceiroService_CreateGasto_RepoError(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{
		gastos: make(map[string]*model.Gasto),
		err:    mockErrorRepo{failCreate: true},
	}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, err := svc.CreateGasto(context.Background(), "tenant-1", &dto.CreateGastoRequest{
		Descricao:          "Gasto",
		ValorTotalCentavos: 5000,
		CompradorID:        "membro-1",
	})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestFinanceiroService_CreateGasto_LoanAndPrivate(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	borrowerID := "membro-2"
	resp, err := svc.CreateGasto(context.Background(), "tenant-1", &dto.CreateGastoRequest{
		Descricao:          "Empréstimo",
		ValorTotalCentavos: 10000,
		CompradorID:        "membro-1",
		IsLoan:             true,
		BorrowerID:         &borrowerID,
		IsPrivate:          true,
		Method:             "money",
		SplitMode:          "EQUAL",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !resp.IsLoan {
		t.Fatal("expected IsLoan to be true")
	}
}

func TestFinanceiroService_CreateContaFixa_NilFixedValue(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	conta, err := svc.CreateContaFixa(context.Background(), "tenant-1", &dto.CreateContaFixaRequest{
		Name: "Internet",
		Icon: "wifi",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if conta.Name != "Internet" {
		t.Fatalf("expected 'Internet', got %s", conta.Name)
	}
	if conta.FixedValueCentavos != nil {
		t.Fatal("expected FixedValueCentavos to be nil")
	}
}

func TestFinanceiroService_CreateContaFixa_RepoError(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{
		contas: make(map[string]*model.ContaFixa),
		err:    mockErrorRepo{failCreate: true},
	}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	valor := int64(50000)
	_, err := svc.CreateContaFixa(context.Background(), "tenant-1", &dto.CreateContaFixaRequest{
		Name:               "Internet",
		Icon:               "wifi",
		FixedValueCentavos: &valor,
	})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestFinanceiroService_ListMembros_RepoError(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{
		membros: make(map[string]*model.MembroCasa),
		err:     mockErrorRepo{failList: true},
	}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, err := svc.ListMembros(context.Background(), "tenant-1")
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestFinanceiroService_ListGastos_RepoError(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{
		gastos: make(map[string]*model.Gasto),
		err:    mockErrorRepo{failList: true},
	}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, err := svc.ListGastos(context.Background(), "tenant-1")
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestFinanceiroService_RecordValidationEvent_RepoError(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{
		events: make(map[string]*model.ProductValidationEvent),
		err:    mockErrorRepo{failExists: true},
	}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	err := svc.RecordValidationEvent(context.Background(), "tenant-1", &dto.ValidateEventRequest{
		Type:      "TENANT_CREATED",
		DedupeKey: "tenant-1",
	})
	if err == nil {
		t.Fatal("expected error when exists check fails, got nil")
	}
}

func TestFinanceiroService_MembroToResponse_NilFields(t *testing.T) {
	membro := &model.MembroCasa{
		ID:     "membro-1",
		Nome:   "Test",
		Avatar: "default",
		Role:   model.RoleMorador,
	}

	resp := membroToResponse(membro)
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.RendaCentavos != nil {
		t.Fatal("expected Renda to be nil when RendaCentavos is nil")
	}
	if resp.UserID != "" {
		t.Fatal("expected empty UserID when UserID is nil")
	}
}

func TestFinanceiroService_MembroToResponse_WithFields(t *testing.T) {
	renda := int64(500000)
	userID := "user-123"
	membro := &model.MembroCasa{
		ID:            "membro-1",
		Nome:          "Test",
		Avatar:        "avatar-url",
		Ativo:         true,
		Role:          model.RoleAdmin,
		RendaCentavos: &renda,
		UserID:        &userID,
	}

	resp := membroToResponse(membro)
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.RendaCentavos == nil || *resp.RendaCentavos != 500000 {
		t.Fatalf("expected Renda 500000, got %v", resp.RendaCentavos)
	}
	if resp.UserID != "user-123" {
		t.Fatalf("expected UserID 'user-123', got '%s'", resp.UserID)
	}
	if resp.Role != "ADMIN" {
		t.Fatalf("expected role 'ADMIN', got '%s'", resp.Role)
	}
}

func TestFinanceiroService_RecordValidationEvent_WithPeriodKey(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	err := svc.RecordValidationEvent(context.Background(), "tenant-1", &dto.ValidateEventRequest{
		Type:      "MONTHLY_CHECK",
		DedupeKey: "2024-01",
		PeriodKey: "2024-01",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, e := range validationRepo.events {
		if e.PeriodKey == nil || *e.PeriodKey != "2024-01" {
			t.Fatal("expected PeriodKey to be set")
		}
	}
}

func TestFinanceiroService_ListMembros_Empty(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	membros, err := svc.ListMembros(context.Background(), "empty-tenant")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(membros) != 0 {
		t.Fatalf("expected 0 membros, got %d", len(membros))
	}
}

func TestFinanceiroService_GetAuditLogs_Empty(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	logs, err := svc.GetAuditLogs(context.Background(), "tenant-empty")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(logs) != 0 {
		t.Fatalf("expected 0 logs, got %d", len(logs))
	}
}

func TestFinanceiroService_ListGastos_Empty(t *testing.T) {
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoEdge{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoEdge{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoEdge{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoEdge{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	gastos, err := svc.ListGastos(context.Background(), "empty-tenant")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(gastos) != 0 {
		t.Fatalf("expected 0 gastos, got %d", len(gastos))
	}
}
