package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
)

type mockMembroRepoTenant struct {
	repository.MembroRepository
}

func (m *mockMembroRepoTenant) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	return &model.MembroCasa{Role: model.RoleMorador, Ativo: true}, nil
}

func (m *mockMembroRepoTenant) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	return nil, nil
}

func TestTenantRequired_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/protected", TenantRequired(&mockMembroRepoTenant{}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["message"] != "X-Tenant-ID header é obrigatório" {
		t.Fatalf("expected 'X-Tenant-ID header é obrigatório', got '%s'", body["message"])
	}
}

func TestTenantRequired_ValidHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/protected", func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Next()
	}, TenantRequired(&mockMembroRepoTenant{}), func(c *gin.Context) {
		tenantID := c.GetString("tenantID")
		if tenantID != "tenant-123" {
			t.Fatalf("expected tenantID 'tenant-123', got '%s'", tenantID)
		}
		role := c.GetString("userRole")
		if role != "MORADOR" {
			t.Fatalf("expected role 'MORADOR', got '%s'", role)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Tenant-ID", "tenant-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestTenantRequired_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/protected", TenantRequired(&mockMembroRepoTenant{}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Tenant-ID", "tenant-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

type mockMembroNotFound struct {
	repository.MembroRepository
}

func (m *mockMembroNotFound) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	return nil, nil
}

func (m *mockMembroNotFound) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	return nil, nil
}

func TestTenantRequired_UserNotInTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/protected", func(c *gin.Context) {
		c.Set("userID", "user-999")
		c.Next()
	}, TenantRequired(&mockMembroNotFound{}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-Tenant-ID", "tenant-999")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["message"] != "acesso negado a este núcleo" {
		t.Fatalf("expected 'acesso negado a este núcleo', got '%s'", body["message"])
	}
}
