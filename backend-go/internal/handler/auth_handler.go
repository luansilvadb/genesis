package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// @Summary Register a new user
// @Description Cria uma nova conta de usuário
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Dados de registro"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := h.authSvc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			// Generic message to prevent email enumeration.
			// HTTP 409 is preserved, but the message does not reveal
			// whether the email already exists or another conflict occurred.
			log.Printf("tentativa de registro com email existente: %s", req.Email)
			c.JSON(http.StatusConflict, gin.H{"message": "Nao foi possivel completar o cadastro. Verifique os dados informados e tente novamente."})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": userFacingError(err)})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// @Summary Login
// @Description Autentica um usuário existente
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Credenciais"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := h.authSvc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": userFacingError(err)})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Google OAuth login
// @Description Autentica ou registra usuário via Google
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.GoogleLoginRequest true "Credencial Google"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	resp, err := h.authSvc.GoogleLogin(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Forgot password
// @Description Solicita link de recuperação de senha
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Email"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.authSvc.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "se o email existir, um link de recuperação foi enviado"})
}

// @Summary Reset password
// @Description Redefine a senha usando token recebido por email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Token e nova senha"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.authSvc.ResetPassword(c.Request.Context(), req.Token, req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "senha redefinida com sucesso"})
}

// @Summary Get current session
// @Description Retorna os dados da sessão do usuário autenticado
// @Tags Auth
// @Security BearerAuth
// @Success 200 {object} dto.SessionResponse
// @Failure 500 {object} map[string]string
// @Router /api/auth/me [get]
func (h *AuthHandler) Session(c *gin.Context) {
	userID := c.GetString("userID")

	resp, err := h.authSvc.Session(c.Request.Context(), userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Create tenant
// @Description Cria um novo núcleo familiar (tenant)
// @Tags Tenant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTenantRequest true "Nome do tenant"
// @Success 201 {object} dto.TenantResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/tenants [post]
func (h *AuthHandler) CreateTenant(c *gin.Context) {
	var req dto.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userID := c.GetString("userID")

	tenant, err := h.authSvc.CreateTenant(c.Request.Context(), req.Name, userID)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.TenantResponse{
		ID:         tenant.ID,
		Name:       tenant.Name,
		InviteCode: tenant.InviteCode,
		CreatedAt:  tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// @Summary Get invite preview
// @Description Obtém informações do núcleo familiar a partir do código de convite
// @Tags Tenant
// @Produce json
// @Param code path string true "Código de convite"
// @Success 200 {object} dto.InvitePreviewResponse
// @Failure 404 {object} map[string]string
// @Router /api/tenants/invite/{code} [get]
func (h *AuthHandler) InvitePreview(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "código de convite é obrigatório"})
		return
	}

	resp, err := h.authSvc.GetInvitePreview(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// @Summary Join tenant
// @Description Entra em um núcleo familiar via código de convite
// @Tags Tenant
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.JoinTenantRequest true "Código de convite"
// @Success 200 {object} dto.TenantResponse
// @Failure 400 {object} map[string]string
// @Router /api/tenants/join [post]
func (h *AuthHandler) JoinTenant(c *gin.Context) {
	var req dto.JoinTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userID := c.GetString("userID")

	tenant, err := h.authSvc.JoinTenant(c.Request.Context(), req.InviteCode, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.TenantResponse{
		ID:         tenant.ID,
		Name:       tenant.Name,
		InviteCode: tenant.InviteCode,
		CreatedAt:  tenant.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
