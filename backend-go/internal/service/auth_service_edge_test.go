package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/config"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
)

type mockUsuarioRepoErr struct {
	repository.UsuarioRepository
	users       map[string]*model.Usuario
	failCreate  bool
	failGetByID bool
	failUpdate  bool
}

func (m *mockUsuarioRepoErr) GetByEmail(ctx context.Context, email string) (*model.Usuario, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockUsuarioRepoErr) Create(ctx context.Context, u *model.Usuario) error {
	if m.failCreate {
		return errors.New("database error")
	}
	if u.ID == "" {
		u.ID = "test-user-id"
	}
	m.users[u.ID] = u
	return nil
}

func (m *mockUsuarioRepoErr) GetByID(ctx context.Context, id string) (*model.Usuario, error) {
	if m.failGetByID {
		return nil, errors.New("database error")
	}
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}

func (m *mockUsuarioRepoErr) GetByGoogleID(ctx context.Context, googleID string) (*model.Usuario, error) {
	for _, u := range m.users {
		if u.GoogleID != nil && *u.GoogleID == googleID {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockUsuarioRepoErr) Update(ctx context.Context, u *model.Usuario) error {
	if m.failUpdate {
		return errors.New("database error")
	}
	m.users[u.ID] = u
	return nil
}

type mockTenantRepoErr struct {
	repository.TenantRepository
	tenants    map[string]*model.Tenant
	failCreate bool
}

func (m *mockTenantRepoErr) Create(ctx context.Context, t *model.Tenant) error {
	if m.failCreate {
		return errors.New("database error")
	}
	if t.ID == "" {
		t.ID = "test-tenant-id"
	}
	m.tenants[t.ID] = t
	return nil
}

func (m *mockTenantRepoErr) GetByInviteCode(ctx context.Context, code string) (*model.Tenant, error) {
	for _, t := range m.tenants {
		if t.InviteCode == code {
			return t, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockTenantRepoErr) GetByID(ctx context.Context, id string) (*model.Tenant, error) {
	if t, ok := m.tenants[id]; ok {
		return t, nil
	}
	return nil, ErrUserNotFound
}

type mockMembroRepoErr struct {
	repository.MembroRepository
	membros    map[string]*model.MembroCasa
	failCreate bool
}

func (m *mockMembroRepoErr) Create(ctx context.Context, mb *model.MembroCasa) error {
	if m.failCreate {
		return errors.New("database error")
	}
	m.membros[mb.ID] = mb
	return nil
}

func (m *mockMembroRepoErr) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	var result []model.MembroCasa
	for _, mb := range m.membros {
		if mb.UserID != nil && *mb.UserID == userID {
			result = append(result, *mb)
		}
	}
	return result, nil
}

func (m *mockMembroRepoErr) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	for _, mb := range m.membros {
		if mb.TenantID == tenantID && mb.UserID != nil && *mb.UserID == userID {
			return mb, nil
		}
	}
	return nil, nil
}

type mockResetRepoErr struct {
	repository.PasswordResetTokenRepository
	tokens     map[string]*model.PasswordResetToken
	failCreate bool
	failDelete bool
}

func (m *mockResetRepoErr) Create(ctx context.Context, t *model.PasswordResetToken) error {
	if m.failCreate {
		return errors.New("database error")
	}
	if t.ID == "" {
		t.ID = "test-reset-id"
	}
	m.tokens[t.ID] = t
	return nil
}

func (m *mockResetRepoErr) GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	for _, t := range m.tokens {
		if t.Token == token {
			return t, nil
		}
	}
	return nil, ErrInvalidToken
}

func (m *mockResetRepoErr) Delete(ctx context.Context, id string) error {
	if m.failDelete {
		return errors.New("database error")
	}
	delete(m.tokens, id)
	return nil
}

type testGoogleTokenVerifier struct {
	clientID string
}

func (v *testGoogleTokenVerifier) Verify(credential string, _ string) (*GoogleTokenPayload, error) {
	token, err := jwt.Parse(credential, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.clientID), nil
	})
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	return &GoogleTokenPayload{
		Subject: claims["sub"].(string),
		Email:   claims["email"].(string),
		Name:    claims["name"].(string),
	}, nil
}

func TestAuthService_Register_RepoError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{
		users:      make(map[string]*model.Usuario),
		failCreate: true,
	}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Email:    "test@example.com",
		Nome:     "Test User",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestAuthService_GoogleLogin_Success_NewUser(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		GoogleOAuthID: "test-google-client-id",
	}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)
	svc.SetGoogleTokenVerifier(&testGoogleTokenVerifier{clientID: cfg.GoogleOAuthID})

	claims := jwt.MapClaims{
		"sub":   "google-sub-123",
		"email": "google@example.com",
		"name":  "Google User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-google-client-id"))

	resp, err := svc.GoogleLogin(context.Background(), &dto.GoogleLoginRequest{Credential: tokenStr})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.User.Email != "google@example.com" {
		t.Fatalf("expected email google@example.com, got %s", resp.User.Email)
	}

	if len(usuarioRepo.users) != 1 {
		t.Fatalf("expected 1 user created, got %d", len(usuarioRepo.users))
	}
}

func TestAuthService_GoogleLogin_Success_ExistingByGoogleID(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		GoogleOAuthID: "test-google-client-id",
	}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	googleID := "existing-google-sub"
	usuarioRepo.users["existing-id"] = &model.Usuario{
		ID:       "existing-id",
		Email:    "existing@example.com",
		Nome:     "Existing User",
		GoogleID: &googleID,
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)
	svc.SetGoogleTokenVerifier(&testGoogleTokenVerifier{clientID: cfg.GoogleOAuthID})

	claims := jwt.MapClaims{
		"sub":   googleID,
		"email": "existing@example.com",
		"name":  "Existing User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-google-client-id"))

	resp, err := svc.GoogleLogin(context.Background(), &dto.GoogleLoginRequest{Credential: tokenStr})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.User.Email != "existing@example.com" {
		t.Fatalf("expected email existing@example.com, got %s", resp.User.Email)
	}

	if len(usuarioRepo.users) != 1 {
		t.Fatalf("expected 1 user (unchanged), got %d", len(usuarioRepo.users))
	}
}

func TestAuthService_GoogleLogin_Success_ExistingByEmail(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		GoogleOAuthID: "test-google-client-id",
	}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["existing-id"] = &model.Usuario{
		ID:    "existing-id",
		Email: "existing@example.com",
		Nome:  "Existing User",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)
	svc.SetGoogleTokenVerifier(&testGoogleTokenVerifier{clientID: cfg.GoogleOAuthID})

	claims := jwt.MapClaims{
		"sub":   "new-google-sub",
		"email": "existing@example.com",
		"name":  "Existing User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-google-client-id"))

	resp, err := svc.GoogleLogin(context.Background(), &dto.GoogleLoginRequest{Credential: tokenStr})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.User.Email != "existing@example.com" {
		t.Fatalf("expected email existing@example.com, got %s", resp.User.Email)
	}

	updatedUser := usuarioRepo.users["existing-id"]
	if updatedUser.GoogleID == nil || *updatedUser.GoogleID != "new-google-sub" {
		t.Fatal("expected GoogleID to be updated on existing user")
	}
}

func TestAuthService_ForgotPassword_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", FrontendURL: "http://localhost:5173"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}
	emailSvc := NewEmailService(cfg)

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@example.com",
		Nome:  "User",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, emailSvc, nil)

	err := svc.ForgotPassword(context.Background(), "user@example.com")
	if err == nil {
		t.Fatal("expected error since SMTP is not available")
	}

	if len(resetRepo.tokens) != 1 {
		t.Fatalf("expected 1 reset token created despite email failure, got %d", len(resetRepo.tokens))
	}
}

func TestAuthService_ForgotPassword_CreateTokenError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{
		tokens:     make(map[string]*model.PasswordResetToken),
		failCreate: true,
	}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@example.com",
		Nome:  "User",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	err := svc.ForgotPassword(context.Background(), "user@example.com")
	if err == nil {
		t.Fatal("expected error when reset token creation fails, got nil")
	}
}

func TestAuthService_ResetPassword_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@example.com",
		Nome:  "User",
	}

	tokenBytes := sha256.Sum256([]byte("valid-reset-token"))
	hashedToken := hex.EncodeToString(tokenBytes[:])
	resetRepo.tokens["reset-1"] = &model.PasswordResetToken{
		ID:        "reset-1",
		Token:     hashedToken,
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	err := svc.ResetPassword(context.Background(), "valid-reset-token", "newSecurePass123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updatedUser := usuarioRepo.users["user-1"]
	if updatedUser.PasswordHash == nil {
		t.Fatal("expected password hash to be set")
	}

	if len(resetRepo.tokens) != 0 {
		t.Fatal("expected reset token to be deleted after use")
	}
}

func TestAuthService_ResetPassword_UserNotFound(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	tokenBytes := sha256.Sum256([]byte("valid-token"))
	hashedToken := hex.EncodeToString(tokenBytes[:])
	resetRepo.tokens["reset-1"] = &model.PasswordResetToken{
		ID:        "reset-1",
		Token:     hashedToken,
		UserID:    "nonexistent-user",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	err := svc.ResetPassword(context.Background(), "valid-token", "NewPass123")
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestAuthService_CreateTenant_UserNotFound(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, err := svc.CreateTenant(context.Background(), "House", "nonexistent-user")
	if err == nil {
		t.Fatal("expected error when user not found, got nil")
	}
}

func TestAuthService_CreateTenant_RepoError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{
		tenants:    make(map[string]*model.Tenant),
		failCreate: true,
	}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "admin@test.com",
		Nome:  "Admin",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, err := svc.CreateTenant(context.Background(), "House", "user-1")
	if err == nil {
		t.Fatal("expected error when tenant repo fails, got nil")
	}
}

func TestAuthService_GoogleLogin_UpdateError(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:     "test-secret",
		GoogleOAuthID: "test-google-client-id",
	}
	usuarioRepo := &mockUsuarioRepoErr{
		users:      make(map[string]*model.Usuario),
		failUpdate: true,
	}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["existing-id"] = &model.Usuario{
		ID:    "existing-id",
		Email: "existing@example.com",
		Nome:  "Existing User",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	claims := jwt.MapClaims{
		"sub":   "new-google-sub",
		"email": "existing@example.com",
		"name":  "Existing User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte("test-google-client-id"))

	_, err := svc.GoogleLogin(context.Background(), &dto.GoogleLoginRequest{Credential: tokenStr})
	if err == nil {
		t.Fatal("expected error when Update fails, got nil")
	}
}

func TestAuthService_ResetPassword_RepoError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{
		users:       make(map[string]*model.Usuario),
		failGetByID: true,
	}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@example.com",
		Nome:  "User",
	}

	resetRepo.tokens["reset-1"] = &model.PasswordResetToken{
		ID:        "reset-1",
		Token:     "valid-reset-token",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	err := svc.ResetPassword(context.Background(), "valid-reset-token", "newSecurePass123")
	if err == nil {
		t.Fatal("expected error when Update fails, got nil")
	}
}

func TestAuthService_CreateTenant_MembroRepoError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}

	failMembroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa), failCreate: true}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "admin@test.com",
		Nome:  "Admin",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, failMembroRepo, resetRepo, nil, nil)

	_, err := svc.CreateTenant(context.Background(), "House", "user-1")
	if err == nil {
		t.Fatal("expected error when membro repo fails, got nil")
	}
}

func TestAuthService_JoinTenant_UserNotFound(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	tenantRepo.tenants["tenant-1"] = &model.Tenant{
		ID:         "tenant-1",
		Name:       "House",
		InviteCode: "INVITE123",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, err := svc.JoinTenant(context.Background(), "INVITE123", "nonexistent-user")
	if err != ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestAuthService_JoinTenant_MembroCreateFails(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{
		membros:    make(map[string]*model.MembroCasa),
		failCreate: true,
	}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	tenantRepo.tenants["tenant-1"] = &model.Tenant{
		ID:         "tenant-1",
		Name:       "House",
		InviteCode: "INVITE123",
	}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@test.com",
		Nome:  "User",
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, err := svc.JoinTenant(context.Background(), "INVITE123", "user-1")
	if err == nil {
		t.Fatal("expected error when membro repo fails, got nil")
	}
}

func TestAuthService_ResetPassword_UpdateError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepoErr{
		users:      make(map[string]*model.Usuario),
		failUpdate: true,
	}
	tenantRepo := &mockTenantRepoErr{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepoErr{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepoErr{tokens: make(map[string]*model.PasswordResetToken)}

	usuarioRepo.users["user-1"] = &model.Usuario{
		ID:    "user-1",
		Email: "user@example.com",
		Nome:  "User",
	}

	resetRepo.tokens["reset-1"] = &model.PasswordResetToken{
		ID:        "reset-1",
		Token:     "valid-reset-token",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	err := svc.ResetPassword(context.Background(), "valid-reset-token", "newSecurePass123")
	if err == nil {
		t.Fatal("expected error when Update fails, got nil")
	}
}
