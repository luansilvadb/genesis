package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	gorillaWS "github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/config"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/database"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/handler"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/middleware"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/service"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/validator"
	ws "github.com/luansilvadb/financeiro-divi/backend-go/internal/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/luansilvadb/financeiro-divi/backend-go/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Divi API
// @version 1.0
// @description API de controle financeiro do Divi
// @host localhost:3000
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// repos holds all repository instances initialized at startup.
type repos struct {
	tenantRepo     repository.TenantRepository
	usuarioRepo    repository.UsuarioRepository
	membroRepo     repository.MembroRepository
	cartaoRepo     repository.CartaoRepository
	faturaRepo     repository.FaturaRepository
	gastoRepo      repository.GastoRepository
	contaFixaRepo  repository.ContaFixaRepository
	auditRepo      repository.AuditLogRepository
	validationRepo repository.ProductValidationRepository
	resetRepo      repository.PasswordResetTokenRepository
}

func createUpgrader(origins []string) gorillaWS.Upgrader {
	return gorillaWS.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return false
			}
			for _, allowed := range origins {
				if allowed == "*" || allowed == origin || isLocalhostVariant(allowed, origin) {
					return true
				}
			}
			log.Printf("websocket: origin not allowed: %s (allowed: %v)", origin, origins)
			return false
		},
	}
}

func isLocalhostVariant(a, b string) bool {
	if a == b {
		return true
	}
	canonical := func(origin string) string {
		var scheme, rest string
		switch {
		case strings.HasPrefix(origin, "http://"):
			scheme, rest = "http", origin[len("http://"):]
		case strings.HasPrefix(origin, "https://"):
			scheme, rest = "https", origin[len("https://"):]
		default:
			return origin
		}
		hostPart, port, hasPort := strings.Cut(rest, ":")
		if !hasPort {
			return origin
		}
		switch hostPart {
		case "localhost", "127.0.0.1", "[::1]":
			hostPart = "localhost"
		}
		return scheme + "://" + hostPart + ":" + port
	}
	return canonical(a) == canonical(b)
}

func initRepositories(db *gorm.DB) repos {
	return repos{
		tenantRepo:     repository.NewGormTenantRepo(db),
		usuarioRepo:    repository.NewGormUsuarioRepo(db),
		membroRepo:     repository.NewGormMembroRepo(db),
		cartaoRepo:     repository.NewGormCartaoRepo(db),
		faturaRepo:     repository.NewGormFaturaRepo(db),
		gastoRepo:      repository.NewGormGastoRepo(db),
		contaFixaRepo:  repository.NewGormContaFixaRepo(db),
		auditRepo:      repository.NewGormAuditLogRepo(db),
		validationRepo: repository.NewGormProductValidationRepo(db),
		resetRepo:      repository.NewGormPasswordResetTokenRepo(db),
	}
}

func setupDatabase(cfg *config.Config) *gorm.DB {
	if err := database.EnsureDatabaseExists(cfg.DatabaseURL); err != nil {
		log.Fatalf("failed to ensure database exists: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Tenant{},
		&model.Usuario{},
		&model.MembroCasa{},
		&model.Cartao{},
		&model.Fatura{},
		&model.Gasto{},
		&model.DivisaoGasto{},
		&model.ContaFixa{},
		&model.AuditLog{},
		&model.ProductValidationEvent{},
		&model.PasswordResetToken{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	ensureFaturaConstraint(db)
	dropLegacyIndex(db)
	runMigrations(db)

	return db
}

func ensureFaturaConstraint(db *gorm.DB) {
	if err := db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'faturas_tenant_cartao_mes_ano_key'
				AND conrelid = 'faturas'::regclass
			) THEN
				ALTER TABLE faturas ADD CONSTRAINT faturas_tenant_cartao_mes_ano_key
				UNIQUE (tenant_id, cartao_id, mes, ano);
			END IF;
		END $$;
	`).Error; err != nil {
		log.Fatalf("failed to ensure fatura unique constraint: %v", err)
	}
}

func dropLegacyIndex(db *gorm.DB) {
	if err := db.Exec(`DROP INDEX IF EXISTS idx_fatura_unica`).Error; err != nil {
		log.Printf("warning: failed to drop legacy index idx_fatura_unica: %v", err)
	}
}

func runMigrations(db *gorm.DB) {
	if err := database.RunSQLMigrations(db, "migrations"); err != nil {
		log.Printf("warning: raw SQL migrations failed: %v", err)
	}
}

func setupServices(cfg *config.Config, db *gorm.DB, r repos, wsHub *ws.Hub) (*service.AuthService, *service.FinanceiroService) {
	emailSvc := service.NewEmailService(cfg)
	authSvc := service.NewAuthService(cfg, db, r.usuarioRepo, r.tenantRepo, r.membroRepo, r.resetRepo, emailSvc, wsHub)
	financeiroSvc := service.NewFinanceiroService(
		db,
		r.membroRepo, r.cartaoRepo, r.faturaRepo, r.gastoRepo,
		r.contaFixaRepo, r.auditRepo, r.validationRepo, r.tenantRepo, wsHub,
	)
	return authSvc, financeiroSvc
}

func setupRoutes(
	r *gin.Engine,
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	financeiroHandler *handler.FinanceiroHandler,
	membroRepo repository.MembroRepository,
	upgrader gorillaWS.Upgrader,
	wsHub *ws.Hub,
) {
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

	r.GET("/health", handler.HealthCheck)

	if cfg.EnableSwagger {
		setupSwagger(r)
	}

	inviteRateLimit := middleware.RateLimit(10, time.Minute)
	r.GET("/api/tenants/invite/:code", inviteRateLimit, authHandler.InvitePreview)

	api := r.Group("/api")
	{
		setupAuthRoutes(api, authHandler)
		setupProtectedRoutes(api, cfg, authHandler, financeiroHandler, membroRepo)
	}

	setupWebSocket(r, cfg, membroRepo, upgrader, wsHub)
}

func setupSwagger(r *gin.Engine) {
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
	r.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "/" {
			c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
			return
		}
		swaggerHandler(c)
	})
}

func setupAuthRoutes(api *gin.RouterGroup, authHandler *handler.AuthHandler) {
	auth := api.Group("/auth")
	authRateLimit := middleware.RateLimit(5, time.Minute)
	auth.POST("/register", authRateLimit, authHandler.Register)
	auth.POST("/login", authRateLimit, authHandler.Login)
	auth.POST("/google", authRateLimit, authHandler.GoogleLogin)
	auth.POST("/forgot-password", authRateLimit, authHandler.ForgotPassword)
	auth.POST("/reset-password", authRateLimit, authHandler.ResetPassword)
}

func setupProtectedRoutes(
	api *gin.RouterGroup,
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	financeiroHandler *handler.FinanceiroHandler,
	membroRepo repository.MembroRepository,
) {
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	protected.Use(middleware.CSRFToken())

	protected.GET("/auth/me", authHandler.Session)
	protected.POST("/tenants", authHandler.CreateTenant)
	protected.POST("/tenants/join", authHandler.JoinTenant)

	tenant := protected.Group("")
	tenant.Use(middleware.TenantRequired(membroRepo))

	setupTenantReadRoutes(tenant, financeiroHandler)
	setupTenantWriteRoutes(tenant, financeiroHandler)
	setupTenantAdminRoutes(tenant, financeiroHandler)
}

func setupTenantReadRoutes(tenant *gin.RouterGroup, h *handler.FinanceiroHandler) {
	tenant.GET("/membros", h.ListMembros)
	tenant.GET("/cartoes", h.ListCartoes)
	tenant.GET("/contas-fixas", h.ListContasFixas)
	tenant.GET("/faturas", h.ListFaturas)
	tenant.GET("/gastos", h.ListGastos)
	tenant.GET("/audit-logs", h.GetAuditLogs)
	tenant.GET("/tenants/permissions", h.GetPermissions)
}

func setupTenantWriteRoutes(tenant *gin.RouterGroup, h *handler.FinanceiroHandler) {
	write := tenant.Group("")
	write.Use(middleware.RoleRequired(model.RoleAdmin, model.RoleMorador))
	write.Use(middleware.RateLimit(60, time.Minute))

	write.POST("/membros/with-account", h.CreateMembroWithAccount)
	write.POST("/membros", h.CreateMembro)
	write.PUT("/membros/:id", h.UpdateMembro)

	write.POST("/cartoes", h.CreateCartao)
	write.DELETE("/cartoes/:id", h.DeleteCartao)

	write.POST("/contas-fixas", h.CreateContaFixa)
	write.PUT("/contas-fixas/:id", h.UpdateContaFixa)
	write.DELETE("/contas-fixas/:id", h.DeleteContaFixa)

	write.POST("/faturas", h.CreateFatura)
	write.POST("/faturas/batch", h.CreateFaturaBatch)

	write.POST("/gastos", h.CreateGasto)
	write.PUT("/gastos/:id", h.UpdateGasto)
	write.POST("/gastos/batch", h.CreateGastoBatch)
	write.DELETE("/gastos/:id", h.DeleteGasto)
	write.POST("/gastos/delete-batch", h.DeleteGastoBatch)

	write.POST("/validation-events", h.RecordValidationEvent)
}

func setupTenantAdminRoutes(tenant *gin.RouterGroup, h *handler.FinanceiroHandler) {
	admin := tenant.Group("")
	admin.Use(middleware.RoleRequired(model.RoleAdmin))
	admin.Use(middleware.RateLimit(60, time.Minute))
	admin.PATCH("/tenants/permissions/:role", h.UpdatePermissions)
}

// wsAuth holds the authenticated WebSocket connection context.
type wsAuth struct {
	userID      string
	tenantID    string
	tokenSource string
}

// validateWSAuth extracts and validates the token, then checks tenant membership.
// Returns the auth context on success; writes an error JSON response and returns nil on failure.
func validateWSAuth(c *gin.Context, cfg *config.Config, membroRepo repository.MembroRepository) *wsAuth {
	tokenStr, tokenSource := extractWSToken(c)
	if tokenStr == "" {
		log.Printf("websocket: missing token (origin=%s, has_upgrade=%v)",
			c.GetHeader("Origin"), c.GetHeader("Upgrade") != "")
		c.JSON(http.StatusBadRequest, gin.H{"message": "token query param, Authorization header, or WebSocket subprotocol required"})
		return nil
	}

	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		log.Printf("websocket: missing tenant_id")
		c.JSON(http.StatusBadRequest, gin.H{"message": "tenant_id query param required"})
		return nil
	}

	userID, err := middleware.ValidateToken(cfg.JWTSecret, tokenStr)
	if err != nil {
		log.Printf("websocket: invalid token (source=%s, tenant=%s): %v", tokenSource, tenantID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "token inválido"})
		return nil
	}

	membro, err := membroRepo.GetByUserID(c.Request.Context(), tenantID, userID)
	if err != nil {
		log.Printf("websocket: membership lookup error (user=%s, tenant=%s, source=%s): %v",
			userID, tenantID, tokenSource, err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "erro interno ao verificar acesso"})
		return nil
	}
	if membro == nil {
		log.Printf("websocket: membership denied (user=%s, tenant=%s, source=%s): not a member",
			userID, tenantID, tokenSource)
		c.JSON(http.StatusForbidden, gin.H{"message": "acesso negado a este núcleo"})
		return nil
	}

	return &wsAuth{userID: userID, tenantID: tenantID, tokenSource: tokenSource}
}

func setupWebSocket(
	r *gin.Engine,
	cfg *config.Config,
	membroRepo repository.MembroRepository,
	upgrader gorillaWS.Upgrader,
	wsHub *ws.Hub,
) {
	r.GET("/ws", func(c *gin.Context) {
		auth := validateWSAuth(c, cfg, membroRepo)
		if auth == nil {
			return
		}

		log.Printf("websocket: auth OK (user=%s, tenant=%s, source=%s, origin=%s)",
			auth.userID, auth.tenantID, auth.tokenSource, c.GetHeader("Origin"))

		if !strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			c.Status(http.StatusOK)
			return
		}

		log.Printf("websocket: upgrading connection")

		var upgradeHeader http.Header
		if auth.tokenSource == "subprotocol" {
			upgradeHeader = make(http.Header)
			upgradeHeader.Set("Sec-WebSocket-Protocol", c.GetHeader("Sec-WebSocket-Protocol"))
		}

		conn, upErr := upgrader.Upgrade(c.Writer, c.Request, upgradeHeader)
		if upErr != nil {
			log.Printf("websocket: upgrade failed (user=%s, tenant=%s, origin=%s): %v",
				auth.userID, auth.tenantID, c.GetHeader("Origin"), upErr)
			return
		}

		log.Printf("websocket: connected (user=%s, tenant=%s)", auth.userID, auth.tenantID)
		ws.HandleClient(wsHub, conn, auth.tenantID)
	})
}

func extractWSToken(c *gin.Context) (tokenStr, source string) {
	source = "none"

	if tokenStr = c.Query("token"); tokenStr != "" {
		source = "query"
	}

	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			source = "header"
		}
	}

	if tokenStr == "" {
		if proto := c.GetHeader("Sec-WebSocket-Protocol"); strings.HasPrefix(proto, "divi.") {
			tokenStr = strings.TrimPrefix(proto, "divi.")
			source = "subprotocol"
		}
	}

	return
}

func startServer(r *gin.Engine, cfg *config.Config) {
	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
			log.Printf("Server running on %s (TLS)", addr)
			if err := srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server failed: %v", err)
			}
		} else {
			log.Printf("Server running on %s", addr)
			log.Printf("WARNING: TLS not configured. Set TLS_CERT_FILE and TLS_KEY_FILE for HTTPS.")
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("server failed: %v", err)
			}
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("warning: .env file not found: %v", err)
	}
	cfg := config.Load()

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is not defined. Application cannot start.")
	}

	if len(cfg.CORSOrigins) == 1 && cfg.CORSOrigins[0] == "*" {
		log.Println("warning: CORS_ORIGINS is set to '*', accepting all origins")
	}

	db := setupDatabase(cfg)
	r := initRepositories(db)
	upgrader := createUpgrader(cfg.CORSOrigins)
	wsHub := ws.NewHub()

	authSvc, financeiroSvc := setupServices(cfg, db, r, wsHub)

	if err := financeiroSvc.LoadPermissions(context.Background()); err != nil {
		log.Printf("warning: failed to load tenant permissions: %v", err)
	}

	authHandler := handler.NewAuthHandler(authSvc)
	financeiroHandler := handler.NewFinanceiroHandler(financeiroSvc)

	validator.RegisterGinValidators()

	router := gin.Default()
	setupRoutes(router, cfg, authHandler, financeiroHandler, r.membroRepo, upgrader, wsHub)

	// Inject db for deep health check (replaces the nil passed earlier).
	router.GET("/health/deep", handler.DeepHealthCheck(db))

	startServer(router, cfg)
}
