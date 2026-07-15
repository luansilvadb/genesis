package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/service"
)

type mockMembroRepoFHEdge struct {
	repository.MembroRepository
	membros map[string]*model.MembroCasa
}

func (m *mockMembroRepoFHEdge) Create(ctx context.Context, mb *model.MembroCasa) error {
	m.membros[mb.ID] = mb
	return nil
}

func (m *mockMembroRepoFHEdge) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.UserID != nil && *mb.UserID == userID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoFHEdge) ListByTenant(ctx context.Context, tenantID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.TenantID == tenantID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoFHEdge) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.MembroCasa, int64, error) {
	all, err := m.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

func (m *mockMembroRepoFHEdge) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.TenantID == tenantID && mb.UserID != nil && *mb.UserID == userID {
			return mb, nil
		}
	}
	return nil, nil
}

type mockCartaoRepoFHEdge struct {
	repository.CartaoRepository
	cartoes map[string]*model.Cartao
}

func (m *mockCartaoRepoFHEdge) Create(ctx context.Context, c *model.Cartao) error {
	m.cartoes[c.ID] = c
	return nil
}

func (m *mockCartaoRepoFHEdge) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Cartao, int64, error) {
	var all []model.Cartao
	for _, c := range m.cartoes {
		if c.TenantID == tenantID {
			all = append(all, *c)
		}
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

type mockGastoRepoFHEdge struct {
	repository.GastoRepository
	gastos map[string]*model.Gasto
}

func (m *mockGastoRepoFHEdge) Create(ctx context.Context, g *model.Gasto) error {
	m.gastos[g.ID] = g
	return nil
}

func (m *mockGastoRepoFHEdge) ListByTenant(ctx context.Context, tenantID string) ([]model.Gasto, error) {
	var result []model.Gasto
	for _, g := range m.gastos {
		if g.TenantID == tenantID {
			result = append(result, *g)
		}
	}
	return result, nil
}

func (m *mockGastoRepoFHEdge) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Gasto, int64, error) {
	all, err := m.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

func (m *mockGastoRepoFHEdge) DeleteDivisoes(ctx context.Context, gastoID, tenantID string) error {
	return nil
}

type mockContaFixaRepoFHEdge struct {
	repository.ContaFixaRepository
	contas map[string]*model.ContaFixa
}

func (m *mockContaFixaRepoFHEdge) Create(ctx context.Context, c *model.ContaFixa) error {
	m.contas[c.ID] = c
	return nil
}

func (m *mockContaFixaRepoFHEdge) GetByID(ctx context.Context, id, tenantID string) (*model.ContaFixa, error) {
	c, ok := m.contas[id]
	if !ok || c.TenantID != tenantID {
		return nil, nil
	}
	return c, nil
}

func (m *mockContaFixaRepoFHEdge) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.ContaFixa, int64, error) {
	var all []model.ContaFixa
	for _, c := range m.contas {
		if c.TenantID == tenantID {
			all = append(all, *c)
		}
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

type mockAuditRepoFHEdge struct {
	repository.AuditLogRepository
	logs []*model.AuditLog
}

func (m *mockAuditRepoFHEdge) Create(ctx context.Context, l *model.AuditLog) error {
	m.logs = append(m.logs, l)
	return nil
}

func (m *mockAuditRepoFHEdge) ListByTenant(ctx context.Context, tenantID string) ([]model.AuditLog, error) {
	var result []model.AuditLog
	for _, l := range m.logs {
		if l.TenantID == tenantID {
			result = append(result, *l)
		}
	}
	return result, nil
}

func (m *mockAuditRepoFHEdge) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.AuditLog, int64, error) {
	all, err := m.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

type mockValidationRepoFHEdge struct {
	repository.ProductValidationRepository
	events map[string]*model.ProductValidationEvent
}

func (m *mockValidationRepoFHEdge) Create(ctx context.Context, e *model.ProductValidationEvent) error {
	m.events[e.ID] = e
	return nil
}

func (m *mockValidationRepoFHEdge) ExistsByDedupeKey(ctx context.Context, tenantID, eventType, dedupeKey string) (bool, error) {
	for _, e := range m.events {
		if e.TenantID == tenantID && string(e.Type) == eventType && e.DedupeKey == dedupeKey {
			return true, nil
		}
	}
	return false, nil
}

type mockWSHubFHEdge struct{}

func (h *mockWSHubFHEdge) Broadcast(tenantID string, msg dto.WSMessage) {}

func (h *mockWSHubFHEdge) BroadcastAll(msg dto.WSMessage) {}

func TestListMembrosHandler_NoTenantReturnsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setup := func() *FinanceiroHandler {
		membroRepo := &mockMembroRepoFHEdge{membros: make(map[string]*model.MembroCasa)}
		cartaoRepo := &mockCartaoRepoFHEdge{cartoes: make(map[string]*model.Cartao)}
		faturaRepo := &mockFaturaRepoFH{}
		gastoRepo := &mockGastoRepoFHEdge{gastos: make(map[string]*model.Gasto)}
		contaFixaRepo := &mockContaFixaRepoFHEdge{contas: make(map[string]*model.ContaFixa)}
		auditRepo := &mockAuditRepoFHEdge{}
		validationRepo := &mockValidationRepoFHEdge{events: make(map[string]*model.ProductValidationEvent)}
		wsHub := &mockWSHubFHEdge{}
		svc := service.NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)
		return NewFinanceiroHandler(svc)
	}
	handler := setup()

	r := gin.New()
	r.GET("/membros", handler.ListMembros)

	req := httptest.NewRequest(http.MethodGet, "/membros", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with empty list when tenantID missing, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateMembroHandler_NoTenantStillSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setup := func() *FinanceiroHandler {
		membroRepo := &mockMembroRepoFHEdge{membros: make(map[string]*model.MembroCasa)}
		cartaoRepo := &mockCartaoRepoFHEdge{cartoes: make(map[string]*model.Cartao)}
		faturaRepo := &mockFaturaRepoFH{}
		gastoRepo := &mockGastoRepoFHEdge{gastos: make(map[string]*model.Gasto)}
		contaFixaRepo := &mockContaFixaRepoFHEdge{contas: make(map[string]*model.ContaFixa)}
		auditRepo := &mockAuditRepoFHEdge{}
		validationRepo := &mockValidationRepoFHEdge{events: make(map[string]*model.ProductValidationEvent)}
		wsHub := &mockWSHubFHEdge{}
		svc := service.NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)
		return NewFinanceiroHandler(svc)
	}
	handler := setup()

	r := gin.New()
	r.POST("/membros", handler.CreateMembro)

	body, _ := json.Marshal(dto.CreateMembroRequest{Nome: "João", Avatar: "default"})
	req := httptest.NewRequest(http.MethodPost, "/membros", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateGastoHandler_NoTenantStillSucceeds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	setup := func() *FinanceiroHandler {
		membroRepo := &mockMembroRepoFHEdge{membros: make(map[string]*model.MembroCasa)}
		cartaoRepo := &mockCartaoRepoFHEdge{cartoes: make(map[string]*model.Cartao)}
		faturaRepo := &mockFaturaRepoFH{}
		gastoRepo := &mockGastoRepoFHEdge{gastos: make(map[string]*model.Gasto)}
		contaFixaRepo := &mockContaFixaRepoFHEdge{contas: make(map[string]*model.ContaFixa)}
		auditRepo := &mockAuditRepoFHEdge{}
		validationRepo := &mockValidationRepoFHEdge{events: make(map[string]*model.ProductValidationEvent)}
		wsHub := &mockWSHubFHEdge{}
		svc := service.NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)
		return NewFinanceiroHandler(svc)
	}
	handler := setup()

	r := gin.New()
	r.POST("/gastos", handler.CreateGasto)

	body, _ := json.Marshal(dto.CreateGastoRequest{
		Descricao:          "Teste",
		ValorTotalCentavos: 1000,
		CompradorID:        "membro-1",
		Divisoes: []dto.SplitItem{
			{MembroID: "membro-1", ValorCentavos: 1000},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/gastos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateCartaoHandler_InvalidBody_EmptyJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/cartoes", handler.CreateCartao)

	req := httptest.NewRequest(http.MethodPost, "/cartoes", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestCreateContaFixaHandler_InvalidBody_EmptyJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/contas-fixas", handler.CreateContaFixa)

	req := httptest.NewRequest(http.MethodPost, "/contas-fixas", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestRecordValidationEventHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/validation-events", handler.RecordValidationEvent)

	req := httptest.NewRequest(http.MethodPost, "/validation-events", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d", w.Code)
	}
}

type mockMembroRepoFail struct {
	repository.MembroRepository
}

func (m *mockMembroRepoFail) ListByTenant(ctx context.Context, tenantID string) ([]model.MembroCasa, error) {
	return nil, errors.New("database error")
}

func (m *mockMembroRepoFail) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.MembroCasa, int64, error) {
	all, err := m.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

func (m *mockMembroRepoFail) Create(ctx context.Context, mb *model.MembroCasa) error {
	return errors.New("database error")
}

func (m *mockMembroRepoFail) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	return nil, nil
}

func (m *mockMembroRepoFail) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	return nil, errors.New("database error")
}

type mockCartaoRepoFail struct {
	repository.CartaoRepository
}

func (m *mockCartaoRepoFail) Create(ctx context.Context, c *model.Cartao) error {
	return errors.New("database error")
}

func (m *mockCartaoRepoFail) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Cartao, int64, error) {
	return nil, 0, errors.New("database error")
}

type mockGastoRepoFail struct {
	repository.GastoRepository
}

func (m *mockGastoRepoFail) Create(ctx context.Context, g *model.Gasto) error {
	return errors.New("database error")
}

func (m *mockGastoRepoFail) ListByTenant(ctx context.Context, tenantID string) ([]model.Gasto, error) {
	return nil, errors.New("database error")
}

func (m *mockGastoRepoFail) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Gasto, int64, error) {
	all, err := m.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

func (m *mockGastoRepoFail) DeleteDivisoes(ctx context.Context, gastoID, tenantID string) error {
	return nil
}

type mockContaFixaRepoFail struct {
	repository.ContaFixaRepository
}

func (m *mockContaFixaRepoFail) Create(ctx context.Context, c *model.ContaFixa) error {
	return errors.New("database error")
}

func (m *mockContaFixaRepoFail) GetByID(ctx context.Context, id, tenantID string) (*model.ContaFixa, error) {
	return nil, errors.New("database error")
}

func (m *mockContaFixaRepoFail) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.ContaFixa, int64, error) {
	return nil, 0, errors.New("database error")
}

type mockAuditRepoFail struct {
	repository.AuditLogRepository
}

func (m *mockAuditRepoFail) ListByTenant(ctx context.Context, tenantID string) ([]model.AuditLog, error) {
	return nil, errors.New("database error")
}

func (m *mockAuditRepoFail) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.AuditLog, int64, error) {
	all, err := m.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(all))
	start := offset
	if start > int(total) {
		start = int(total)
	}
	end := start + limit
	if end > int(total) {
		end = int(total)
	}
	return all[start:end], total, nil
}

type mockValidationRepoFail struct {
	repository.ProductValidationRepository
}

func (m *mockValidationRepoFail) ExistsByDedupeKey(ctx context.Context, tenantID, eventType, dedupeKey string) (bool, error) {
	return false, errors.New("database error")
}

func setupFailingFinanceiroHandler() *FinanceiroHandler {
	membroRepo := &mockMembroRepoFail{}
	cartaoRepo := &mockCartaoRepoFail{}
	faturaRepo := &mockFaturaRepoFH{}
	gastoRepo := &mockGastoRepoFail{}
	contaFixaRepo := &mockContaFixaRepoFail{}
	auditRepo := &mockAuditRepoFail{}
	validationRepo := &mockValidationRepoFail{}
	wsHub := &mockWSHubFHEdge{}

	svc := service.NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)
	return NewFinanceiroHandler(svc)
}

func TestCreateMembroHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.POST("/membros", handler.CreateMembro)

	body, _ := json.Marshal(dto.CreateMembroRequest{Nome: "João", Avatar: "default"})
	req := httptest.NewRequest(http.MethodPost, "/membros", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}

func TestListMembrosHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.GET("/membros", handler.ListMembros)
	req := httptest.NewRequest(http.MethodGet, "/membros", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}

func TestCreateCartaoHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.POST("/cartoes", handler.CreateCartao)

	body, _ := json.Marshal(dto.CreateCartaoRequest{
		Nome:                "Nubank",
		DiaFechamento:       15,
		ResponsavelPadraoID: "membro-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/cartoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}

func TestCreateGastoHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.POST("/gastos", handler.CreateGasto)

	body, _ := json.Marshal(dto.CreateGastoRequest{
		Descricao:          "Teste",
		ValorTotalCentavos: 1000,
		CompradorID:        "membro-1",
		Divisoes: []dto.SplitItem{
			{MembroID: "membro-1", ValorCentavos: 1000},
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/gastos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}

func TestListGastosHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.GET("/gastos", handler.ListGastos)
	req := httptest.NewRequest(http.MethodGet, "/gastos", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}

func TestCreateContaFixaHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.POST("/contas-fixas", handler.CreateContaFixa)

	body, _ := json.Marshal(dto.CreateContaFixaRequest{Name: "Internet", Icon: "wifi"})
	req := httptest.NewRequest(http.MethodPost, "/contas-fixas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}

func TestRecordValidationEventHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.POST("/validation-events", handler.RecordValidationEvent)

	body, _ := json.Marshal(dto.ValidateEventRequest{Type: "TENANT_CREATED", DedupeKey: "tenant-1"})
	req := httptest.NewRequest(http.MethodPost, "/validation-events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}

func TestGetAuditLogsHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFailingFinanceiroHandler()

	r := gin.New()
	r.GET("/audit-logs", handler.GetAuditLogs)
	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}
