package handler

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/config"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/service"
)

type mockUsuarioRepoEdge struct {
	repository.UsuarioRepository
	users map[string]*model.Usuario
}

func (m *mockUsuarioRepoEdge) GetByEmail(ctx context.Context, email string) (*model.Usuario, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (m *mockUsuarioRepoEdge) Create(ctx context.Context, u *model.Usuario) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUsuarioRepoEdge) GetByID(ctx context.Context, id string) (*model.Usuario, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, service.ErrUserNotFound
}

func (m *mockUsuarioRepoEdge) GetByGoogleID(ctx context.Context, googleID string) (*model.Usuario, error) {
	for _, u := range m.users {
		if u.GoogleID != nil && *u.GoogleID == googleID {
			return u, nil
		}
	}
	return nil, service.ErrUserNotFound
}

func (m *mockUsuarioRepoEdge) Update(ctx context.Context, u *model.Usuario) error {
	m.users[u.ID] = u
	return nil
}

type mockTenantRepoEdge struct {
	repository.TenantRepository
	tenants map[string]*model.Tenant
}

func (m *mockTenantRepoEdge) Create(ctx context.Context, t *model.Tenant) error {
	m.tenants[t.ID] = t
	return nil
}

func (m *mockTenantRepoEdge) GetByInviteCode(ctx context.Context, code string) (*model.Tenant, error) {
	for _, t := range m.tenants {
		if t.InviteCode == code {
			return t, nil
		}
	}
	return nil, service.ErrUserNotFound
}

type mockMembroRepoEdge struct {
	repository.MembroRepository
	membros map[string]*model.MembroCasa
}

func (m *mockMembroRepoEdge) Create(ctx context.Context, mb *model.MembroCasa) error {
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

func (m *mockMembroRepoEdge) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.TenantID == tenantID && mb.UserID != nil && *mb.UserID == userID {
			return mb, nil
		}
	}
	return nil, nil
}

func (m *mockMembroRepoEdge) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.MembroCasa, int64, error) {
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

type mockResetRepoEdge struct {
	repository.PasswordResetTokenRepository
	tokens map[string]*model.PasswordResetToken
}

func (m *mockResetRepoEdge) Create(ctx context.Context, t *model.PasswordResetToken) error {
	m.tokens[t.ID] = t
	return nil
}

func (m *mockResetRepoEdge) GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	for _, t := range m.tokens {
		if t.Token == token {
			return t, nil
		}
	}
	return nil, service.ErrInvalidToken
}

func (m *mockResetRepoEdge) Delete(ctx context.Context, id string) error {
	delete(m.tokens, id)
	return nil
}

//nolint:unused // test helper kept for symmetry with setupAuthHandler, may be used in future edge-case tests
func setupAuthHandlerWithConfig(cfg *config.Config) *AuthHandler {
	usuarioRepo := &mockUsuarioRepoEdge{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoEdge{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoEdge{tokens: make(map[string]*model.PasswordResetToken)}

	authSvc := service.NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil)
	return NewAuthHandler(authSvc)
}

func setupGoogleAuthHandler(cfg *config.Config) *AuthHandler {
	usuarioRepo := &mockUsuarioRepoEdge{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoEdge{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoEdge{tokens: make(map[string]*model.PasswordResetToken)}

	authSvc := service.NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil)
	authSvc.SetGoogleTokenVerifier(&testGoogleTokenVerifierEdge{clientID: cfg.GoogleOAuthID})
	return NewAuthHandler(authSvc)
}

type testGoogleTokenVerifierEdge struct {
	clientID string
}

func (v *testGoogleTokenVerifierEdge) Verify(credential string, _ string) (*service.GoogleTokenPayload, error) {
	token, err := jwt.Parse(credential, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.clientID), nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &service.GoogleTokenPayload{
		Subject: claims["sub"].(string),
		Email:   claims["email"].(string),
		Name:    claims["name"].(string),
	}, nil
}

func TestGoogleLoginHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret:     "test-secret",
		GoogleOAuthID: "test-google-client-id",
	}
	handler := setupGoogleAuthHandler(cfg)

	r := gin.New()
	r.POST("/google", handler.GoogleLogin)

	claims := jwt.MapClaims{
		"sub":   "google-sub",
		"email": "google-user@test.com",
		"name":  "Google User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-google-client-id"))

	body, _ := json.Marshal(dto.GoogleLoginRequest{Credential: tokenStr})
	req := httptest.NewRequest(http.MethodPost, "/google", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp dto.AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal("expected valid response body")
	}
	if resp.User.Email != "google-user@test.com" {
		t.Fatalf("expected email 'google-user@test.com', got '%s'", resp.User.Email)
	}
}

func TestGoogleLoginHandler_InvalidCredential(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		JWTSecret:     "test-secret",
		GoogleOAuthID: "test-google-client-id",
	}
	handler := setupGoogleAuthHandler(cfg)

	r := gin.New()
	r.POST("/google", handler.GoogleLogin)

	body, _ := json.Marshal(dto.GoogleLoginRequest{Credential: "invalid-jwt"})
	req := httptest.NewRequest(http.MethodPost, "/google", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid credential, got %d", w.Code)
	}
}

func TestForgotPasswordHandler_WithUserReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWTSecret: "test-secret", FrontendURL: "http://localhost:5173"}
	usuarioRepo := &mockUsuarioRepoEdge{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoEdge{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoEdge{tokens: make(map[string]*model.PasswordResetToken)}
	emailSvc := service.NewEmailService(cfg)

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@example.com",
		Nome:  "User",
	}

	authSvc := service.NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, emailSvc)
	handler := NewAuthHandler(authSvc)

	r := gin.New()
	r.POST("/forgot-password", handler.ForgotPassword)

	body, _ := json.Marshal(dto.ForgotPasswordRequest{Email: "user@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 since SMTP unavailable, got %d", w.Code)
	}
}

func TestResetPasswordHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoEdge{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoEdge{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoEdge{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@example.com",
		Nome:  "User",
	}

	tokenBytes := sha256.Sum256([]byte("valid-reset-token"))
	hashedToken := hex.EncodeToString(tokenBytes[:])
	resetToken := &model.PasswordResetToken{
		ID:        "reset-1",
		Token:     hashedToken,
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	resetRepo.tokens["reset-1"] = resetToken

	authSvc := service.NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil)
	handler := NewAuthHandler(authSvc)

	r := gin.New()
	r.POST("/reset-password", handler.ResetPassword)

	body, _ := json.Marshal(dto.ResetPasswordRequest{
		Token:    "valid-reset-token",
		Password: "newSecurePass123",
	})
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateTenantHandler_InvalidBody_MissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/tenants", handler.CreateTenant)

	body, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("userID", "user-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

type mockTenantRepoFail struct {
	repository.TenantRepository
}

func (m *mockTenantRepoFail) Create(ctx context.Context, t *model.Tenant) error {
	return errors.New("database error")
}

func (m *mockTenantRepoFail) GetByInviteCode(ctx context.Context, code string) (*model.Tenant, error) {
	return nil, errors.New("database error")
}

func TestCreateTenantHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	usuarioRepo := &mockUsuarioRepoEdge{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoFail{}
	membroRepo := &mockMembroRepoEdge{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoEdge{tokens: make(map[string]*model.PasswordResetToken)}
	cfg := &config.Config{JWTSecret: "test-secret"}

	authSvc := service.NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil)
	handler := NewAuthHandler(authSvc)

	r := gin.New()
	r.POST("/tenants", handler.CreateTenant)

	body, _ := json.Marshal(dto.CreateTenantRequest{Name: "House"})
	req := httptest.NewRequest(http.MethodPost, "/tenants", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("userID", "user-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for service error, got %d", w.Code)
	}
}
