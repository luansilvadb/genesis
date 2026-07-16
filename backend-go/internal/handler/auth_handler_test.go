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
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/config"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/service"
)

type mockUsuarioRepo struct {
	repository.UsuarioRepository
	users map[string]*model.Usuario
}

func (m *mockUsuarioRepo) GetByEmail(ctx context.Context, email string) (*model.Usuario, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (m *mockUsuarioRepo) Create(ctx context.Context, u *model.Usuario) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockUsuarioRepo) GetByID(ctx context.Context, id string) (*model.Usuario, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, service.ErrUserNotFound
}

func (m *mockUsuarioRepo) GetByGoogleID(ctx context.Context, googleID string) (*model.Usuario, error) {
	for _, u := range m.users {
		if u.GoogleID != nil && *u.GoogleID == googleID {
			return u, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (m *mockUsuarioRepo) Update(ctx context.Context, u *model.Usuario) error {
	m.users[u.ID] = u
	return nil
}

type mockTenantRepo struct {
	repository.TenantRepository
	tenants map[string]*model.Tenant
}

func (m *mockTenantRepo) Create(ctx context.Context, t *model.Tenant) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	m.tenants[t.ID] = t
	return nil
}

func (m *mockTenantRepo) GetByInviteCode(ctx context.Context, code string) (*model.Tenant, error) {
	for _, t := range m.tenants {
		if t.InviteCode == code {
			return t, nil
		}
	}
	return nil, service.ErrUserNotFound
}

type mockMembroRepo struct {
	repository.MembroRepository
	membros map[string]*model.MembroCasa
}

func (m *mockMembroRepo) Create(ctx context.Context, mb *model.MembroCasa) error {
	m.membros[mb.ID] = mb
	return nil
}

func (m *mockMembroRepo) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.UserID != nil && *mb.UserID == userID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepo) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.TenantID == tenantID && mb.UserID != nil && *mb.UserID == userID {
			return mb, nil
		}
	}
	return nil, nil
}

func (m *mockMembroRepo) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.MembroCasa, int64, error) {
	var all []model.MembroCasa
	for _, mb := range m.membros {
		if mb.TenantID == tenantID {
			all = append(all, *mb)
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

type mockResetRepo struct {
	repository.PasswordResetTokenRepository
	tokens map[string]*model.PasswordResetToken
}

func (m *mockResetRepo) Create(ctx context.Context, t *model.PasswordResetToken) error {
	m.tokens[t.ID] = t
	return nil
}

func (m *mockResetRepo) GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	for _, t := range m.tokens {
		if t.Token == token {
			return t, nil
		}
	}
	return nil, service.ErrInvalidToken
}

func (m *mockResetRepo) Delete(ctx context.Context, id string) error {
	delete(m.tokens, id)
	return nil
}

func setupTestHandler() *AuthHandler {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	authSvc := service.NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)
	return NewAuthHandler(authSvc)
}

func setUserIDFromHeader(c *gin.Context) {
	if userID := c.GetHeader("userID"); userID != "" {
		c.Set("userID", userID)
	}
	c.Next()
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/register", handler.Register)

	body, _ := json.Marshal(dto.RegisterRequest{
		Email:    "new@test.com",
		Nome:     "New User",
		Password: "SecurePass1",
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal("expected valid response body")
	}

	if resp.User.Email != "new@test.com" {
		t.Fatalf("expected email new@test.com, got %s", resp.User.Email)
	}
}

func TestRegisterHandler_Duplicate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/register", handler.Register)

	body, _ := json.Marshal(dto.RegisterRequest{
		Email:    "dup2@test.com",
		Nome:     "User",
		Password: "Pass1234",
	})

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate, got %d", w2.Code)
	}
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/register", handler.Register)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}

func TestLoginHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)

	body, _ := json.Marshal(dto.RegisterRequest{
		Email:    "login-test@test.com",
		Nome:     "Login User",
		Password: "MyPass123",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginBody, _ := json.Marshal(dto.LoginRequest{
		Email:    "login-test@test.com",
		Password: "MyPass123",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp dto.AuthResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatal("expected valid response")
	}
	if resp.Token == "" {
		t.Fatal("expected token to be non-empty")
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/login", handler.Login)

	body, _ := json.Marshal(dto.LoginRequest{
		Email:    "wrong@test.com",
		Password: "wrongpass",
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/login", handler.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}

func TestGoogleLoginHandler_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/google", handler.GoogleLogin)

	req := httptest.NewRequest(http.MethodPost, "/google", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}

func TestForgotPasswordHandler_Request(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/forgot-password", handler.ForgotPassword)

	body, _ := json.Marshal(dto.ForgotPasswordRequest{
		Email: "anyone@test.com",
	})

	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestForgotPasswordHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/forgot-password", handler.ForgotPassword)

	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}

func TestResetPasswordHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/reset-password", handler.ResetPassword)

	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}

func TestResetPasswordHandler_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/reset-password", handler.ResetPassword)

	body, _ := json.Marshal(dto.ResetPasswordRequest{
		Token:    "invalid-token",
		Password: "NewPass123",
	})

	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid token, got %d", w.Code)
	}
}

func TestCreateTenantHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.Use(setUserIDFromHeader)
	router.POST("/register", handler.Register)
	router.POST("/tenants", handler.CreateTenant)

	regBody, _ := json.Marshal(dto.RegisterRequest{
		Email:    "admin@casa.com",
		Nome:     "Admin",
		Password: "Pass1234",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var regResp dto.AuthResponse
	_ = json.Unmarshal(w.Body.Bytes(), &regResp)

	body, _ := json.Marshal(dto.CreateTenantRequest{
		Name: "Minha Casa",
	})

	req2 := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("userID", regResp.User.ID)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w2.Code, w2.Body.String())
	}

	var tenantResp dto.TenantResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &tenantResp); err != nil {
		t.Fatal("expected valid tenant response")
	}
	if tenantResp.Name != "Minha Casa" {
		t.Fatalf("expected 'Minha Casa', got %s", tenantResp.Name)
	}
}

func TestCreateTenantHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/tenants", handler.CreateTenant)

	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}

func TestJoinTenantHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.Use(setUserIDFromHeader)
	router.POST("/register", handler.Register)
	router.POST("/tenants", handler.CreateTenant)
	router.POST("/tenants/join", handler.JoinTenant)

	regBody, _ := json.Marshal(dto.RegisterRequest{
		Email:    "owner@casa.com",
		Nome:     "Owner",
		Password: "Pass1234",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var ownerResp dto.AuthResponse
	_ = json.Unmarshal(w.Body.Bytes(), &ownerResp)

	createBody, _ := json.Marshal(dto.CreateTenantRequest{Name: "Casa"})
	req2 := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(createBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("userID", ownerResp.User.ID)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	var tenantResp dto.TenantResponse
	_ = json.Unmarshal(w2.Body.Bytes(), &tenantResp)

	regBody2, _ := json.Marshal(dto.RegisterRequest{
		Email:    "guest@casa.com",
		Nome:     "Guest",
		Password: "Pass4567",
	})
	req3 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody2))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	var guestResp dto.AuthResponse
	_ = json.Unmarshal(w3.Body.Bytes(), &guestResp)

	joinBody, _ := json.Marshal(dto.JoinTenantRequest{
		InviteCode: tenantResp.InviteCode,
	})
	req4 := httptest.NewRequest(http.MethodPost, "/tenants/join", bytes.NewReader(joinBody))
	req4.Header.Set("Content-Type", "application/json")
	req4.Header.Set("userID", guestResp.User.ID)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w4.Code, w4.Body.String())
	}
}

func TestJoinTenantHandler_InvalidCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	router := gin.New()
	router.Use(setUserIDFromHeader)
	router.POST("/register", handler.Register)
	router.POST("/tenants/join", handler.JoinTenant)

	regBody, _ := json.Marshal(dto.RegisterRequest{
		Email:    "user@test.com",
		Nome:     "User",
		Password: "Pass1234",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var regResp dto.AuthResponse
	_ = json.Unmarshal(w.Body.Bytes(), &regResp)

	body, _ := json.Marshal(dto.JoinTenantRequest{
		InviteCode: "INVALID",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/tenants/join", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("userID", regResp.User.ID)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid code, got %d", w2.Code)
	}
}

func TestJoinTenantHandler_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/tenants/join", handler.JoinTenant)

	req := httptest.NewRequest(http.MethodPost, "/tenants/join", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid body, got %d", w.Code)
	}
}
