package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/config"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/validator"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("email ou senha inválidos")
	ErrEmailAlreadyExists = errors.New("email já cadastrado")
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrInvalidToken       = errors.New("token inválido ou expirado")
)

type GoogleTokenPayload struct {
	Subject string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
}

type GoogleTokenVerifier interface {
	Verify(credential string, audience string) (*GoogleTokenPayload, error)
}

type AuthService struct {
	cfg           *config.Config
	db            *gorm.DB
	usuarioRepo   repository.UsuarioRepository
	tenantRepo    repository.TenantRepository
	membroRepo    repository.MembroRepository
	resetRepo     repository.PasswordResetTokenRepository
	emailSvc      *EmailService
	googleVerif   GoogleTokenVerifier
	googleVerifMu sync.RWMutex
	wsHub         WSHub
}

func NewAuthService(
	cfg *config.Config,
	db *gorm.DB,
	usuarioRepo repository.UsuarioRepository,
	tenantRepo repository.TenantRepository,
	membroRepo repository.MembroRepository,
	resetRepo repository.PasswordResetTokenRepository,
	emailSvc *EmailService,
	wsHub WSHub,
) *AuthService {
	return &AuthService{
		cfg:         cfg,
		db:          db,
		usuarioRepo: usuarioRepo,
		tenantRepo:  tenantRepo,
		membroRepo:  membroRepo,
		resetRepo:   resetRepo,
		emailSvc:    emailSvc,
		googleVerif: newDefaultGoogleTokenVerifier(),
		wsHub:       wsHub,
	}
}

// checkEmailAvailable returns nil when the email is free, ErrEmailAlreadyExists
// when already taken, or a wrapped error on unexpected database failures.
func (s *AuthService) checkEmailAvailable(ctx context.Context, email string) error {
	existing, err := s.usuarioRepo.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, repository.ErrRecordNotFound) && !errors.Is(err, ErrUserNotFound) {
		return fmt.Errorf("erro ao verificar email: %w", err)
	}
	if existing != nil {
		return ErrEmailAlreadyExists
	}
	return nil
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	if err := validator.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	if err := s.checkEmailAvailable(ctx, req.Email); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	hashStr := string(hash)

	user := &model.Usuario{
		Email:        req.Email,
		Nome:         req.Nome,
		PasswordHash: &hashStr,
	}

	if err = s.usuarioRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	s.maybeJoinByInvite(ctx, user, req.InviteCode)

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			ID:    user.ID,
			Email: user.Email,
			Nome:  user.Nome,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.usuarioRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}

	if err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		// Log failed login to allow monitoring for brute-force attempts.
		// A distributed rate limiter (e.g. Redis) should be added for
		// production deployments with multiple instances.
		log.Printf("failed login attempt for user %s (email: %s)", user.ID, user.Email)
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			ID:    user.ID,
			Email: user.Email,
			Nome:  user.Nome,
		},
	}, nil
}

func (s *AuthService) GoogleLogin(ctx context.Context, req *dto.GoogleLoginRequest) (*dto.AuthResponse, error) {
	payload, err := s.verifyGoogleToken(req.Credential)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.resolveGoogleUser(ctx, payload)
	if err != nil {
		return nil, err
	}

	s.maybeJoinByInvite(ctx, user, req.InviteCode)

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			ID:    user.ID,
			Email: user.Email,
			Nome:  user.Nome,
		},
	}, nil
}

// resolveGoogleUser finds an existing user by Google ID, or links/creates one
// by email when this is the first Google login for that email address.
func (s *AuthService) resolveGoogleUser(ctx context.Context, payload *GoogleTokenPayload) (*model.Usuario, error) {
	user, err := s.usuarioRepo.GetByGoogleID(ctx, payload.Subject)
	if err == nil {
		return user, nil
	}

	// Google ID not found — try matching by email
	userByEmail, lookupErr := s.usuarioRepo.GetByEmail(ctx, payload.Email)
	if lookupErr != nil && !errors.Is(lookupErr, repository.ErrRecordNotFound) && !errors.Is(lookupErr, ErrUserNotFound) {
		return nil, fmt.Errorf("erro ao verificar email existente: %w", lookupErr)
	}

	if userByEmail != nil {
		// Existing user without Google link — attach Google ID
		userByEmail.GoogleID = &payload.Subject
		if err := s.usuarioRepo.Update(ctx, userByEmail); err != nil {
			return nil, err
		}
		return userByEmail, nil
	}

	// Brand-new user via Google
	user = &model.Usuario{
		Email:    payload.Email,
		Nome:     payload.Name,
		GoogleID: &payload.Subject,
	}
	if err := s.usuarioRepo.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.usuarioRepo.GetByEmail(ctx, email)

	// Constant-time response to prevent email enumeration via timing.
	// Always generate a token (even for non-existent users) and discard it
	// when the user doesn't exist.
	tokenBytes := make([]byte, 32)
	if _, randErr := rand.Read(tokenBytes); randErr != nil {
		return randErr
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	if err != nil || user == nil {
		// User not found — still spend CPU time to avoid timing oracle.
		// Hash a dummy value so the bcrypt-like cost is simulated.
		//nolint:errcheck // intentional: timing side-channel mitigation, result is discarded
		bcrypt.GenerateFromPassword([]byte(tokenStr), bcrypt.DefaultCost)
		return nil
	}

	// Hash the token before storage so a DB read doesn't expose active reset tokens.
	tokenHash := sha256.Sum256([]byte(tokenStr))
	hashedToken := hex.EncodeToString(tokenHash[:])

	resetToken := &model.PasswordResetToken{
		Token:     hashedToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}

	if err := s.resetRepo.Create(ctx, resetToken); err != nil {
		return err
	}

	if s.emailSvc != nil {
		return s.emailSvc.SendPasswordReset(email, tokenStr)
	}
	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := validator.ValidatePassword(newPassword); err != nil {
		return err
	}

	tokenHash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(tokenHash[:])
	resetToken, err := s.resetRepo.GetByToken(ctx, hashedToken)
	if err != nil || resetToken == nil {
		return ErrInvalidToken
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return ErrInvalidToken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashStr := string(hash)

	user, err := s.usuarioRepo.GetByID(ctx, resetToken.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	user.PasswordHash = &hashStr
	if err := s.usuarioRepo.Update(ctx, user); err != nil {
		return err
	}

	return s.resetRepo.Delete(ctx, resetToken.ID)
}

func (s *AuthService) generateToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(s.cfg.JWTExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) SetGoogleTokenVerifier(v GoogleTokenVerifier) {
	s.googleVerifMu.Lock()
	defer s.googleVerifMu.Unlock()
	s.googleVerif = v
}

func (s *AuthService) verifyGoogleToken(credential string) (*GoogleTokenPayload, error) {
	s.googleVerifMu.RLock()
	defer s.googleVerifMu.RUnlock()
	if s.googleVerif == nil {
		return nil, ErrInvalidCredentials
	}
	return s.googleVerif.Verify(credential, s.cfg.GoogleOAuthID)
}

type defaultGoogleTokenVerifier struct {
	client *http.Client
}

// newDefaultGoogleTokenVerifier creates a token verifier that uses Google's
// tokeninfo endpoint (POST to oauth2.googleapis.com/tokeninfo).
//
// NOTE: This approach sends the credential to Google's servers for validation
// rather than performing local cryptographic signature verification using the
// official Google Identity Platform client library (golang.org/x/oauth2).
// The tokeninfo endpoint validates the token server-side, but it introduces a
// network dependency and the endpoint is not intended for production credential
// verification at scale. Consider migrating to the official library for local
// JWT verification.
func newDefaultGoogleTokenVerifier() *defaultGoogleTokenVerifier {
	return &defaultGoogleTokenVerifier{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (v *defaultGoogleTokenVerifier) Verify(credential string, audience string) (*GoogleTokenPayload, error) {
	// Use POST to avoid logging the credential in URL query parameters.
	body := fmt.Sprintf("id_token=%s", credential)
	resp, err := v.client.Post("https://oauth2.googleapis.com/tokeninfo", "application/x-www-form-urlencoded", strings.NewReader(body))
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidCredentials
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	var payload GoogleTokenPayload
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Validate email_verified claim to prevent account takeovers.
	var full struct {
		EmailVerified string `json:"email_verified"`
		Audience      string `json:"aud"`
	}
	if err := json.Unmarshal(respBody, &full); err == nil {
		if full.Audience != "" && full.Audience != audience {
			return nil, ErrInvalidCredentials
		}
		if full.EmailVerified == "false" {
			return nil, ErrInvalidCredentials
		}
	}

	return &payload, nil
}

func (s *AuthService) Session(ctx context.Context, userID string) (*dto.SessionResponse, error) {
	membros, err := s.membroRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(membros) == 0 {
		return &dto.SessionResponse{Tenants: []dto.SessionTenant{}}, nil
	}

	// Batch-fetch all tenants in a single query (avoids N+1).
	tenantIDs := make([]string, len(membros))
	for i, m := range membros {
		tenantIDs[i] = m.TenantID
	}

	tenantList, err := s.tenantRepo.GetByIDs(ctx, tenantIDs)
	if err != nil {
		return nil, err
	}

	tenants := make([]dto.SessionTenant, 0, len(tenantList))
	for _, tenant := range tenantList {
		tenants = append(tenants, dto.SessionTenant{
			ID:         tenant.ID,
			Name:       tenant.Name,
			InviteCode: tenant.InviteCode,
			CreatedAt:  tenant.CreatedAt.Format(time.RFC3339),
		})
	}

	return &dto.SessionResponse{Tenants: tenants}, nil
}

func (s *AuthService) CreateTenant(ctx context.Context, name, userID string) (*model.Tenant, error) {
	code := uuid.New().String()[:8]

	user, err := s.usuarioRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if s.db == nil {
		return s.createTenantWithoutTx(ctx, name, code, user)
	}

	var tenant *model.Tenant
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenant = &model.Tenant{
			Name:       name,
			InviteCode: code,
		}
		if createErr := tx.Create(tenant).Error; createErr != nil {
			return createErr
		}

		membro := &model.MembroCasa{
			ID:       uuid.New().String(),
			TenantID: tenant.ID,
			Nome:     user.Nome,
			Avatar:   "default",
			Role:     model.RoleAdmin,
			UserID:   &user.ID,
		}
		if createErr := tx.Create(membro).Error; createErr != nil {
			return createErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *AuthService) createTenantWithoutTx(ctx context.Context, name, code string, user *model.Usuario) (*model.Tenant, error) {
	tenant := &model.Tenant{
		Name:       name,
		InviteCode: code,
	}
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	membro := &model.MembroCasa{
		ID:       uuid.New().String(),
		TenantID: tenant.ID,
		Nome:     user.Nome,
		Avatar:   "default",
		Role:     model.RoleAdmin,
		UserID:   &user.ID,
	}
	if err := s.membroRepo.Create(ctx, membro); err != nil {
		return nil, err
	}

	return tenant, nil
}

func (s *AuthService) GetInvitePreview(ctx context.Context, inviteCode string) (*dto.InvitePreviewResponse, error) {
	tenant, err := s.tenantRepo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, errors.New("código de convite inválido")
	}

	membros, err := s.membroRepo.ListByTenant(ctx, tenant.ID)
	if err != nil {
		return nil, err
	}

	preview := &dto.InvitePreviewResponse{
		ID:                 tenant.ID,
		Name:               tenant.Name,
		MembrosDisponiveis: make([]dto.InvitePreviewMembro, 0, len(membros)),
	}

	for _, m := range membros {
		if m.Ativo {
			preview.MembrosDisponiveis = append(preview.MembrosDisponiveis, dto.InvitePreviewMembro{
				ID:     m.ID,
				Nome:   m.Nome,
				Avatar: m.Avatar,
			})
		}
	}

	return preview, nil
}

// maybeJoinByInvite attempts to join a tenant by invite code when one is
// provided. A nil or empty invite is silently ignored.
func (s *AuthService) maybeJoinByInvite(ctx context.Context, user *model.Usuario, inviteCode *string) {
	if inviteCode != nil && *inviteCode != "" {
		s.joinTenantByInvite(ctx, user, *inviteCode)
	}
}

func (s *AuthService) joinTenantByInvite(ctx context.Context, user *model.Usuario, inviteCode string) {
	tenant, err := s.tenantRepo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		log.Printf("joinTenantByInvite: invalid invite code %q: %v", inviteCode, err)
		return
	}

	existing, err := s.membroRepo.GetByUserID(ctx, tenant.ID, user.ID)
	if err != nil {
		log.Printf("joinTenantByInvite: error checking membership for user %s in tenant %s: %v", user.ID, tenant.ID, err)
		return
	}

	// Member already exists — reactivate if inactive, then done.
	if existing != nil {
		if !existing.Ativo {
			existing.Ativo = true
			if err := s.membroRepo.Update(ctx, existing); err != nil {
				log.Printf("joinTenantByInvite: error reactivating member %s: %v", existing.ID, err)
				return
			}
			s.broadcastMemberChange(tenant.ID, existing)
		}
		return
	}

	// Create new member for this tenant.
	membro := &model.MembroCasa{
		ID:       uuid.New().String(),
		TenantID: tenant.ID,
		Nome:     user.Nome,
		Avatar:   "default",
		Role:     model.RoleMorador,
		UserID:   &user.ID,
	}
	if err := s.membroRepo.Create(ctx, membro); err != nil {
		log.Printf("joinTenantByInvite: error creating member for user %s in tenant %s: %v", user.ID, tenant.ID, err)
		return
	}
	s.broadcastMemberChange(tenant.ID, membro)
}

func (s *AuthService) JoinTenant(ctx context.Context, inviteCode, userID string) (*model.Tenant, error) {
	tenant, err := s.tenantRepo.GetByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, errors.New("código de convite inválido")
	}

	user, err := s.usuarioRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	existing, err := s.membroRepo.GetByUserID(ctx, tenant.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar membresia: %w", err)
	}
	if existing != nil {
		if !existing.Ativo {
			existing.Ativo = true
			if err := s.membroRepo.Update(ctx, existing); err != nil {
				return nil, err
			}
			s.broadcastMemberChange(tenant.ID, existing)
			return tenant, nil
		}
		return nil, errors.New("você já é membro deste núcleo")
	}

	membro := &model.MembroCasa{
		ID:       uuid.New().String(),
		TenantID: tenant.ID,
		Nome:     user.Nome,
		Avatar:   "default",
		Role:     model.RoleMorador,
		UserID:   &user.ID,
	}

	if err := s.membroRepo.Create(ctx, membro); err != nil {
		return nil, err
	}

	s.broadcastMemberChange(tenant.ID, membro)

	return tenant, nil
}

// broadcastMemberChange envia um evento WebSocket para notificar outros membros
// do tenant sobre a entrada ou reativação de um membro.
func (s *AuthService) broadcastMemberChange(tenantID string, m *model.MembroCasa) {
	if s.wsHub == nil {
		return
	}
	resp := dto.MembroResponse{
		ID:     m.ID,
		Nome:   m.Nome,
		Avatar: m.Avatar,
		Role:   string(m.Role),
		Ativo:  m.Ativo,
	}
	if m.UserID != nil {
		resp.UserID = *m.UserID
	}
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeMemberCreated,
		Payload: resp,
	})
}
