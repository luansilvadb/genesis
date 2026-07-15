package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/service"
)

type mockMembroRepoFH struct {
	repository.MembroRepository
	membros map[string]*model.MembroCasa
}

func (m *mockMembroRepoFH) Create(ctx context.Context, mb *model.MembroCasa) error {
	if mb.ID == "" {
		mb.ID = uuid.New().String()
	}
	m.membros[mb.ID] = mb
	return nil
}

func (m *mockMembroRepoFH) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.UserID != nil && *mb.UserID == userID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoFH) ListByTenant(ctx context.Context, tenantID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.TenantID == tenantID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoFH) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.MembroCasa, int64, error) {
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

func (m *mockMembroRepoFH) GetByID(ctx context.Context, id, tenantID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.ID == id && mb.TenantID == tenantID {
			return mb, nil
		}
	}
	return nil, nil
}

func (m *mockMembroRepoFH) Update(ctx context.Context, mb *model.MembroCasa) error {
	m.membros[mb.ID] = mb
	return nil
}

func (m *mockMembroRepoFH) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.TenantID == tenantID && mb.UserID != nil && *mb.UserID == userID {
			return mb, nil
		}
	}
	return nil, nil
}

type mockCartaoRepoFH struct {
	repository.CartaoRepository
	cartoes map[string]*model.Cartao
}

func (m *mockCartaoRepoFH) Create(ctx context.Context, c *model.Cartao) error {
	m.cartoes[c.ID] = c
	return nil
}

func (m *mockCartaoRepoFH) ListByTenant(ctx context.Context, tenantID string) ([]model.Cartao, error) {
	var all []model.Cartao
	for _, c := range m.cartoes {
		if c.TenantID == tenantID {
			all = append(all, *c)
		}
	}
	return all, nil
}

func (m *mockCartaoRepoFH) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Cartao, int64, error) {
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

type mockGastoRepoFH struct {
	repository.GastoRepository
	gastos map[string]*model.Gasto
}

func (m *mockGastoRepoFH) Create(ctx context.Context, g *model.Gasto) error {
	m.gastos[g.ID] = g
	return nil
}

func (m *mockGastoRepoFH) GetByID(ctx context.Context, id, tenantID string) (*model.Gasto, error) {
	for _, g := range m.gastos {
		if g.ID == id && g.TenantID == tenantID {
			return g, nil
		}
	}
	return nil, nil
}

func (m *mockGastoRepoFH) Update(ctx context.Context, g *model.Gasto) error {
	m.gastos[g.ID] = g
	return nil
}

func (m *mockGastoRepoFH) DeleteDivisoes(ctx context.Context, gastoID, tenantID string) error {
	return nil
}

func (m *mockGastoRepoFH) ListByTenant(ctx context.Context, tenantID string) ([]model.Gasto, error) {
	var result []model.Gasto
	for _, g := range m.gastos {
		if g.TenantID == tenantID {
			result = append(result, *g)
		}
	}
	return result, nil
}

func (m *mockGastoRepoFH) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Gasto, int64, error) {
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

type mockContaFixaRepoFH struct {
	repository.ContaFixaRepository
	contas map[string]*model.ContaFixa
}

func (m *mockContaFixaRepoFH) Create(ctx context.Context, c *model.ContaFixa) error {
	m.contas[c.ID] = c
	return nil
}

func (m *mockContaFixaRepoFH) GetByID(ctx context.Context, id, tenantID string) (*model.ContaFixa, error) {
	c, ok := m.contas[id]
	if !ok || c.TenantID != tenantID {
		return nil, nil
	}
	return c, nil
}

func (m *mockContaFixaRepoFH) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.ContaFixa, int64, error) {
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

type mockAuditRepoFH struct {
	repository.AuditLogRepository
	logs []*model.AuditLog
}

func (m *mockAuditRepoFH) Create(ctx context.Context, l *model.AuditLog) error {
	m.logs = append(m.logs, l)
	return nil
}

func (m *mockAuditRepoFH) ListByTenant(ctx context.Context, tenantID string) ([]model.AuditLog, error) {
	var result []model.AuditLog
	for _, l := range m.logs {
		if l.TenantID == tenantID {
			result = append(result, *l)
		}
	}
	return result, nil
}

func (m *mockAuditRepoFH) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.AuditLog, int64, error) {
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

type mockValidationRepoFH struct {
	repository.ProductValidationRepository
	events map[string]*model.ProductValidationEvent
}

func (m *mockValidationRepoFH) Create(ctx context.Context, e *model.ProductValidationEvent) error {
	m.events[e.ID] = e
	return nil
}

func (m *mockValidationRepoFH) ExistsByDedupeKey(ctx context.Context, tenantID, eventType, dedupeKey string) (bool, error) {
	for _, e := range m.events {
		if e.TenantID == tenantID && string(e.Type) == eventType && e.DedupeKey == dedupeKey {
			return true, nil
		}
	}
	return false, nil
}

type mockWSHubFH struct{}

func (h *mockWSHubFH) Broadcast(tenantID string, msg dto.WSMessage) {}

func (h *mockWSHubFH) BroadcastAll(msg dto.WSMessage) {}

func setupFinanceiroHandler() *FinanceiroHandler {
	membroRepo := &mockMembroRepoFH{membros: make(map[string]*model.MembroCasa)}
	cartaoRepo := &mockCartaoRepoFH{cartoes: make(map[string]*model.Cartao)}
	faturaRepo := &mockFaturaRepoFH{}
	gastoRepo := &mockGastoRepoFH{gastos: make(map[string]*model.Gasto)}
	contaFixaRepo := &mockContaFixaRepoFH{contas: make(map[string]*model.ContaFixa)}
	auditRepo := &mockAuditRepoFH{}
	validationRepo := &mockValidationRepoFH{events: make(map[string]*model.ProductValidationEvent)}
	wsHub := &mockWSHubFH{}

	svc := service.NewFinanceiroService(nil, membroRepo, cartaoRepo, faturaRepo, gastoRepo, contaFixaRepo, auditRepo, validationRepo, nil, wsHub)
	return NewFinanceiroHandler(svc)
}

type mockFaturaRepoFH struct {
	repository.FaturaRepository
}

func TestCreateMembroHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/membros", handler.CreateMembro)

	body, _ := json.Marshal(dto.CreateMembroRequest{
		Nome:   "João",
		Avatar: "default",
	})

	req := httptest.NewRequest(http.MethodPost, "/membros", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.MembroResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid response, got error: %v", err)
	}
	if resp.Nome != "João" {
		t.Fatalf("expected 'João', got %s", resp.Nome)
	}
}

func TestCreateMembroHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/membros", handler.CreateMembro)

	req := httptest.NewRequest(http.MethodPost, "/membros", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListMembrosHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.GET("/membros", handler.ListMembros)

	req := httptest.NewRequest(http.MethodGet, "/membros", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateCartaoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/cartoes", handler.CreateCartao)

	body, _ := json.Marshal(dto.CreateCartaoRequest{
		Nome:                "Nubank",
		DiaFechamento:       15,
		ResponsavelPadraoID: "membro-1",
	})

	req := httptest.NewRequest(http.MethodPost, "/cartoes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.Cartao
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("expected valid response, got error: %v", err)
	}
	if resp.Nome != "Nubank" {
		t.Fatalf("expected 'Nubank', got %s", resp.Nome)
	}
}

func TestCreateCartaoHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/cartoes", handler.CreateCartao)

	req := httptest.NewRequest(http.MethodPost, "/cartoes", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateGastoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/gastos", handler.CreateGasto)
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-1")
		c.Next()
	})

	body, _ := json.Marshal(dto.CreateGastoRequest{
		Descricao:          "Supermercado",
		ValorTotalCentavos: 10000,
		CompradorID:        "membro-1",
		Divisoes: []dto.SplitItem{
			{MembroID: "membro-1", ValorCentavos: 5000},
			{MembroID: "membro-2", ValorCentavos: 5000},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/gastos", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateGastoHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/gastos", handler.CreateGasto)

	req := httptest.NewRequest(http.MethodPost, "/gastos", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListGastosHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.GET("/gastos", handler.ListGastos)

	req := httptest.NewRequest(http.MethodGet, "/gastos", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateContaFixaHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/contas-fixas", handler.CreateContaFixa)

	body, _ := json.Marshal(dto.CreateContaFixaRequest{
		Name: "Internet",
		Icon: "wifi",
	})

	req := httptest.NewRequest(http.MethodPost, "/contas-fixas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateContaFixaHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/contas-fixas", handler.CreateContaFixa)

	req := httptest.NewRequest(http.MethodPost, "/contas-fixas", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestRecordValidationEventHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/validation-events", handler.RecordValidationEvent)

	body, _ := json.Marshal(dto.ValidateEventRequest{
		Type:      "TENANT_CREATED",
		DedupeKey: "tenant-1",
	})

	req := httptest.NewRequest(http.MethodPost, "/validation-events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRecordValidationEventHandler_Duplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.POST("/validation-events", handler.RecordValidationEvent)

	body, _ := json.Marshal(dto.ValidateEventRequest{
		Type:      "FIRST_EXPENSE_CREATED",
		DedupeKey: "expense-1",
	})

	req := httptest.NewRequest(http.MethodPost, "/validation-events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req2 := httptest.NewRequest(http.MethodPost, "/validation-events", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate, got %d", w2.Code)
	}
}

func TestGetAuditLogsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupFinanceiroHandler()

	r := gin.New()
	r.GET("/audit-logs", handler.GetAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
