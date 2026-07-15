package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
)

type mockMembroRepoForFinanceiro struct {
	repository.MembroRepository
	membros map[string]*model.MembroCasa
}

func (m *mockMembroRepoForFinanceiro) Create(ctx context.Context, mb *model.MembroCasa) error {
	if mb.ID == "" {
		mb.ID = uuid.New().String()
	}
	m.membros[mb.ID] = mb
	return nil
}

func (m *mockMembroRepoForFinanceiro) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.UserID != nil && *mb.UserID == userID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoForFinanceiro) ListByTenant(ctx context.Context, tenantID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.TenantID == tenantID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoForFinanceiro) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.TenantID == tenantID && mb.UserID != nil && *mb.UserID == userID {
			return mb, nil
		}
	}
	return nil, nil
}

type mockCartaoRepoForFinanceiro struct {
	repository.CartaoRepository
	cartoes map[string]*model.Cartao
}

func (m *mockCartaoRepoForFinanceiro) Create(ctx context.Context, c *model.Cartao) error {
	m.cartoes[c.ID] = c
	return nil
}

type mockGastoRepoForFinanceiro struct {
	repository.GastoRepository
	gastos map[string]*model.Gasto
}

func (m *mockGastoRepoForFinanceiro) Create(ctx context.Context, g *model.Gasto) error {
	m.gastos[g.ID] = g
	return nil
}

func (m *mockGastoRepoForFinanceiro) DeleteDivisoes(ctx context.Context, gastoID, tenantID string) error {
	return nil
}

func (m *mockGastoRepoForFinanceiro) ListByTenant(ctx context.Context, tenantID string) ([]model.Gasto, error) {
	var result []model.Gasto
	for _, g := range m.gastos {
		if g.TenantID == tenantID {
			result = append(result, *g)
		}
	}
	return result, nil
}

type mockContaFixaRepoForFinanceiro struct {
	repository.ContaFixaRepository
	contas map[string]*model.ContaFixa
}

func (m *mockContaFixaRepoForFinanceiro) Create(ctx context.Context, c *model.ContaFixa) error {
	m.contas[c.ID] = c
	return nil
}

func (m *mockContaFixaRepoForFinanceiro) GetByID(ctx context.Context, id, tenantID string) (*model.ContaFixa, error) {
	c, ok := m.contas[id]
	if !ok || c.TenantID != tenantID {
		return nil, nil
	}
	return c, nil
}

type mockAuditRepoForFinanceiro struct {
	repository.AuditLogRepository
	logs []*model.AuditLog
}

func (m *mockAuditRepoForFinanceiro) Create(ctx context.Context, l *model.AuditLog) error {
	m.logs = append(m.logs, l)
	return nil
}

func (m *mockAuditRepoForFinanceiro) ListByTenant(ctx context.Context, tenantID string) ([]model.AuditLog, error) {
	var result []model.AuditLog
	for _, l := range m.logs {
		if l.TenantID == tenantID {
			result = append(result, *l)
		}
	}
	return result, nil
}

type mockValidationRepoForFinanceiro struct {
	repository.ProductValidationRepository
	events map[string]*model.ProductValidationEvent
}

func (m *mockValidationRepoForFinanceiro) Create(ctx context.Context, e *model.ProductValidationEvent) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	m.events[e.ID] = e
	return nil
}

func (m *mockValidationRepoForFinanceiro) ExistsByDedupeKey(ctx context.Context, tenantID, eventType, dedupeKey string) (bool, error) {
	for _, e := range m.events {
		if e.TenantID == tenantID && string(e.Type) == eventType && e.DedupeKey == dedupeKey {
			return true, nil
		}
	}
	return false, nil
}

type mockWSHub struct {
	broadcasts []dto.WSMessage
}

func (h *mockWSHub) Broadcast(tenantID string, msg dto.WSMessage) {
	h.broadcasts = append(h.broadcasts, msg)
}

func (h *mockWSHub) BroadcastAll(msg dto.WSMessage) {
	h.broadcasts = append(h.broadcasts, msg)
}

func TestFinanceiroService_CreateMembro(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	resp, err := svc.CreateMembro(context.Background(), "tenant-1", &dto.CreateMembroRequest{
		Nome:   "João",
		Avatar: "default",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Nome != "João" {
		t.Fatalf("expected nome 'João', got %s", resp.Nome)
	}
	if resp.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	if len(auditRepo.logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(auditRepo.logs))
	}
}

func TestFinanceiroService_ListMembros(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, _ = svc.CreateMembro(context.Background(), "tenant-1", &dto.CreateMembroRequest{Nome: "João", Avatar: "a"})
	_, _ = svc.CreateMembro(context.Background(), "tenant-1", &dto.CreateMembroRequest{Nome: "Maria", Avatar: "b"})

	membros, err := svc.ListMembros(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(membros) != 2 {
		t.Fatalf("expected 2 membros, got %d", len(membros))
	}
}

func TestFinanceiroService_CreateCartao(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	cartao, err := svc.CreateCartao(context.Background(), "tenant-1", &dto.CreateCartaoRequest{
		Nome:                "Nubank",
		DiaFechamento:       15,
		ResponsavelPadraoID: "membro-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cartao.Nome != "Nubank" {
		t.Fatalf("expected 'Nubank', got %s", cartao.Nome)
	}
	if cartao.DiaFechamento != 15 {
		t.Fatalf("expected dia_fechamento 15, got %d", cartao.DiaFechamento)
	}
}

func TestFinanceiroService_CreateGasto_DefaultValues(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	resp, err := svc.CreateGasto(context.Background(), "tenant-1", &dto.CreateGastoRequest{
		Descricao:          "Supermercado",
		ValorTotalCentavos: 5000,
		CompradorID:        "membro-1",
		Divisoes: []dto.SplitItem{
			{MembroID: "membro-1", ValorCentavos: 2500},
			{MembroID: "membro-2", ValorCentavos: 2500},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Descricao != "Supermercado" {
		t.Fatalf("expected 'Supermercado', got %s", resp.Descricao)
	}
	if resp.Installments != 1 {
		t.Fatalf("expected 1 installment, got %d", resp.Installments)
	}
	if resp.SplitMode != "CUSTOM" {
		t.Fatalf("expected CUSTOM split mode, got %s", resp.SplitMode)
	}
	if len(resp.Divisoes) != 2 {
		t.Fatalf("expected 2 divisoes, got %d", len(resp.Divisoes))
	}

	if len(wsHub.broadcasts) != 1 {
		t.Fatalf("expected 1 ws broadcast, got %d", len(wsHub.broadcasts))
	}
	if wsHub.broadcasts[0].Type != dto.WSTypeExpenseCreated {
		t.Fatalf("expected EXPENSE_CREATED broadcast, got %s", wsHub.broadcasts[0].Type)
	}

	if len(auditRepo.logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(auditRepo.logs))
	}
}

func TestFinanceiroService_CreateGasto_WithSetup(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	installments := 3
	totalInstallments := 3
	resp, err := svc.CreateGasto(context.Background(), "tenant-1", &dto.CreateGastoRequest{
		Descricao:          "Curso Online",
		ValorTotalCentavos: 30000,
		CompradorID:        "membro-1",
		Installments:       &installments,
		TotalInstallments:  &totalInstallments,
		Method:             "credit_card",
		SplitMode:          "EQUAL",
		Divisoes: []dto.SplitItem{
			{MembroID: "membro-1", ValorCentavos: 15000},
			{MembroID: "membro-2", ValorCentavos: 15000},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Installments != 3 {
		t.Fatalf("expected 3 installments, got %d", resp.Installments)
	}
	if resp.SplitMode != "EQUAL" {
		t.Fatalf("expected EQUAL split mode, got %s", resp.SplitMode)
	}
}

func TestFinanceiroService_ListGastos(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, _ = svc.CreateGasto(context.Background(), "tenant-1", &dto.CreateGastoRequest{
		Descricao:          "Gasto 1",
		ValorTotalCentavos: 1000,
		CompradorID:        "membro-1",
		Divisoes: []dto.SplitItem{
			{MembroID: "membro-1", ValorCentavos: 1000},
		},
	})

	gastos, err := svc.ListGastos(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(gastos) != 1 {
		t.Fatalf("expected 1 gasto, got %d", len(gastos))
	}
}

func TestFinanceiroService_CreateContaFixa(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	valor := int64(50000)
	conta, err := svc.CreateContaFixa(context.Background(), "tenant-1", &dto.CreateContaFixaRequest{
		Name:               "Internet",
		Icon:               "wifi",
		FixedValueCentavos: &valor,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if conta.Name != "Internet" {
		t.Fatalf("expected 'Internet', got %s", conta.Name)
	}
}

func TestFinanceiroService_RecordValidationEvent_Success(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	err := svc.RecordValidationEvent(context.Background(), "tenant-1", &dto.ValidateEventRequest{
		Type:      "TENANT_CREATED",
		DedupeKey: "tenant-1",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestFinanceiroService_RecordValidationEvent_Duplicate(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_ = svc.RecordValidationEvent(context.Background(), "tenant-1", &dto.ValidateEventRequest{
		Type:      "FIRST_EXPENSE_CREATED",
		DedupeKey: "expense-1",
	})

	err := svc.RecordValidationEvent(context.Background(), "tenant-1", &dto.ValidateEventRequest{
		Type:      "FIRST_EXPENSE_CREATED",
		DedupeKey: "expense-1",
	})
	if err == nil {
		t.Fatal("expected error for duplicate event")
	}
}

func TestFinanceiroService_GetAuditLogs(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	_, _ = svc.CreateMembro(context.Background(), "tenant-1", &dto.CreateMembroRequest{Nome: "João", Avatar: "a"})

	logs, err := svc.GetAuditLogs(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(logs) < 1 {
		t.Fatal("expected at least 1 audit log")
	}
	if logs[0].Acao != "CRIAR_MEMBRO" {
		t.Fatalf("expected CRIAR_MEMBRO, got %s", logs[0].Acao)
	}
}

func TestFinanceiroService_CreateContaFixa_WithDefaultSplit(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	valor := int64(10000)
	conta, err := svc.CreateContaFixa(context.Background(), "tenant-1", &dto.CreateContaFixaRequest{
		Name:               "Internet",
		Icon:               "wifi",
		FixedValueCentavos: &valor,
		DefaultSplit: []dto.SplitItem{
			{MembroID: "m1", ValorCentavos: 5000},
			{MembroID: "m2", ValorCentavos: 5000},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(conta.DefaultSplit) != 2 {
		t.Fatalf("expected 2 split items, got %d", len(conta.DefaultSplit))
	}
	if conta.DefaultSplit[0].MembroID != "m1" {
		t.Fatalf("expected m1, got %s", conta.DefaultSplit[0].MembroID)
	}
	if conta.DefaultSplit[1].ValorCentavos != 5000 {
		t.Fatalf("expected 5000, got %d", conta.DefaultSplit[1].ValorCentavos)
	}
}

func TestFinanceiroService_CreateContaFixa_EmptyDefaultSplit(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	conta, err := svc.CreateContaFixa(context.Background(), "tenant-1", &dto.CreateContaFixaRequest{
		Name: "Agua",
		Icon: "opacity",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(conta.DefaultSplit) != 0 {
		t.Fatalf("expected empty defaultSplit, got %d items", len(conta.DefaultSplit))
	}
}

func TestFinanceiroService_MembroToResponse_IncludesCreatedAt(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	membro, err := svc.CreateMembro(context.Background(), "tenant-1", &dto.CreateMembroRequest{
		Nome:   "Maria",
		Avatar: "🐱",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if membro.CreatedAt == "" {
		t.Fatal("expected createdAt to be populated")
	}
	if membro.Avatar != "🐱" {
		t.Fatalf("expected avatar 🐱, got %s", membro.Avatar)
	}
}

func TestFinanceiroService_GastoToResponse_IncludesRecurringFields(t *testing.T) {
	membroRepo := &mockMembroRepoForFinanceiro{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoForFinanceiro{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepo{}
	gastoRepo := &mockGastoRepoForFinanceiro{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoForFinanceiro{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoForFinanceiro{}
	validationRepo := &mockValidationRepoForFinanceiro{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHub{}

	svc := NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)

	gasto, err := svc.CreateGasto(context.Background(), "tenant-1", &dto.CreateGastoRequest{
		Descricao:          "Compra teste",
		ValorTotalCentavos: 1000,
		CompradorID:        "m1",
		Divisoes: []dto.SplitItem{
			{MembroID: "m1", ValorCentavos: 1000},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if gasto.CardOwnerID != nil {
		t.Fatal("expected cardOwnerId to be nil for pix gasto")
	}

	gastos, err := svc.ListGastos(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(gastos) != 1 {
		t.Fatalf("expected 1 gasto, got %d", len(gastos))
	}
}
