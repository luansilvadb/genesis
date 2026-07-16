package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/config"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
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
	return nil, ErrUserNotFound
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
	return nil, ErrUserNotFound
}

func (m *mockUsuarioRepo) GetByGoogleID(ctx context.Context, googleID string) (*model.Usuario, error) {
	for _, u := range m.users {
		if u.GoogleID != nil && *u.GoogleID == googleID {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
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
	return nil, ErrUserNotFound
}

func (m *mockTenantRepo) GetByID(ctx context.Context, id string) (*model.Tenant, error) {
	if t, ok := m.tenants[id]; ok {
		return t, nil
	}
	return nil, ErrUserNotFound
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

type mockResetRepo struct {
	repository.PasswordResetTokenRepository
	tokens map[string]*model.PasswordResetToken
}

func (m *mockResetRepo) Create(ctx context.Context, t *model.PasswordResetToken) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	m.tokens[t.ID] = t
	return nil
}

func (m *mockResetRepo) GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	for _, t := range m.tokens {
		if t.Token == token {
			return t, nil
		}
	}
	return nil, ErrInvalidToken
}

func (m *mockResetRepo) Delete(ctx context.Context, id string) error {
	delete(m.tokens, id)
	return nil
}

func TestAuthService_Register_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret-123"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	resp, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Email:    "test@example.com",
		Nome:     "Test User",
		Password: "Password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected token to be non-empty")
	}

	if resp.User.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", resp.User.Email)
	}
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, _ = svc.Register(context.Background(), &dto.RegisterRequest{
		Email:    "dup@example.com",
		Nome:     "User",
		Password: "Password123",
	})

	_, err := svc.Register(context.Background(), &dto.RegisterRequest{
		Email:    "dup@example.com",
		Nome:     "User 2",
		Password: "Password456",
	})

	if err != ErrEmailAlreadyExists {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, _ = svc.Register(context.Background(), &dto.RegisterRequest{
		Email:    "login@test.com",
		Nome:     "Login User",
		Password: "SecurePass1",
	})

	resp, err := svc.Login(context.Background(), &dto.LoginRequest{
		Email:    "login@test.com",
		Password: "SecurePass1",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected token to be non-empty")
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "AnyPass1",
	})

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, _ = svc.Register(context.Background(), &dto.RegisterRequest{
		Email:    "wrongpass@test.com",
		Nome:     "User",
		Password: "CorrectPass1",
	})

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Email:    "wrongpass@test.com",
		Password: "WrongPass1",
	})

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_NilPasswordHash(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	user := &model.Usuario{
		Email: "google-only@test.com",
		Nome:  "Google User",
	}
	_ = usuarioRepo.Create(context.Background(), user)

	_, err := svc.Login(context.Background(), &dto.LoginRequest{
		Email:    "google-only@test.com",
		Password: "anypass",
	})

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for user with nil password hash, got %v", err)
	}
}

func TestAuthService_GoogleLogin_InvalidCredential(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", GoogleOAuthID: "test-client-id"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_, err := svc.GoogleLogin(context.Background(), &dto.GoogleLoginRequest{Credential: "invalid-credential"})

	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for invalid credential, got %v", err)
	}
}

func TestAuthService_ForgotPassword_NonExistentUser(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	err := svc.ForgotPassword(context.Background(), "nonexistent@test.com")

	if err != nil {
		t.Fatalf("expected nil for non-existent user, got %v", err)
	}
}

func TestAuthService_ResetPassword_InvalidToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	err := svc.ResetPassword(context.Background(), "invalid-token", "NewPass123")

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_ResetPassword_ExpiredToken(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_ = usuarioRepo.Create(context.Background(), &model.Usuario{
		Email: "user@test.com",
		Nome:  "User",
	})

	_ = resetRepo.Create(context.Background(), &model.PasswordResetToken{
		Token:     "expired-token",
		UserID:    "some-user-id",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	})

	err := svc.ResetPassword(context.Background(), "expired-token", "NewPass123")

	if err != ErrInvalidToken {
		t.Fatalf("expected ErrInvalidToken for expired token, got %v", err)
	}
}

func TestAuthService_CreateTenant_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	user := &model.Usuario{Email: "admin@test.com", Nome: "Admin"}
	_ = usuarioRepo.Create(context.Background(), user)

	tenant, err := svc.CreateTenant(context.Background(), "My House", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tenant.Name != "My House" {
		t.Fatalf("expected tenant name 'My House', got %s", tenant.Name)
	}
	if tenant.InviteCode == "" {
		t.Fatal("expected invite code to be non-empty")
	}

	if len(membroRepo.membros) != 1 {
		t.Fatalf("expected 1 member, got %d", len(membroRepo.membros))
	}

	for _, m := range membroRepo.membros {
		if m.Role != model.RoleAdmin {
			t.Fatalf("expected admin role, got %s", m.Role)
		}
		if m.TenantID != tenant.ID {
			t.Fatalf("expected tenant id %s, got %s", tenant.ID, m.TenantID)
		}
	}
}

func TestAuthService_JoinTenant_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	_ = tenantRepo.Create(context.Background(), &model.Tenant{
		Name:       "Existing House",
		InviteCode: "INVITE123",
	})

	user := &model.Usuario{Email: "member@test.com", Nome: "Member"}
	_ = usuarioRepo.Create(context.Background(), user)

	tenant, err := svc.JoinTenant(context.Background(), "INVITE123", user.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tenant.Name != "Existing House" {
		t.Fatalf("expected 'Existing House', got %s", tenant.Name)
	}
}

func TestAuthService_JoinTenant_InvalidCode(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret"}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	user := &model.Usuario{Email: "user@test.com", Nome: "User"}
	_ = usuarioRepo.Create(context.Background(), user)

	_, err := svc.JoinTenant(context.Background(), "INVALIDCODE", user.ID)
	if err == nil {
		t.Fatal("expected error for invalid invite code")
	}
}

func TestAuthService_GenerateToken_Valid(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExpiry: 24 * time.Hour,
	}
	usuarioRepo := &mockUsuarioRepo{users: make(map[string]*model.Usuario)}
	tenantRepo := &mockTenantRepo{tenants: make(map[string]*model.Tenant)}
	membroRepo := &mockMembroRepo{membros: make(map[string]*model.MembroCasa)}
	resetRepo := &mockResetRepo{tokens: make(map[string]*model.PasswordResetToken)}

	svc := NewAuthService(cfg, nil, usuarioRepo, tenantRepo, membroRepo, resetRepo, nil, nil)

	tokenStr, err := svc.generateToken("user-id", "user@test.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["sub"] != "user-id" {
		t.Fatalf("expected sub 'user-id', got %v", claims["sub"])
	}
	if claims["email"] != "user@test.com" {
		t.Fatalf("expected email 'user@test.com', got %v", claims["email"])
	}
}
