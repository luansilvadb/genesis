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

func createUpgrader(origins []string) gorillaWS.Upgrader {
	return gorillaWS.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return false
			}
			for _, allowed := range origins {
				if allowed == "*" || allowed == origin {
					return true
				}
				// Match localhost variants: http://localhost:PORT is equivalent to
				// http://127.0.0.1:PORT and http://[::1]:PORT in development.
				if isLocalhostVariant(allowed, origin) {
					return true
				}
			}
			log.Printf("websocket: origin not allowed: %s (allowed: %v)", origin, origins)
			return false
		},
	}
}

// isLocalhostVariant checks whether two origins differ only in the localhost
// representation (localhost vs 127.0.0.1 vs [::1]) while sharing the same
// scheme and port. This prevents spurious CheckOrigin failures when the
// frontend is accessed via a different localhost alias.
func isLocalhostVariant(a, b string) bool {
	// Fast path: exact match is handled by the caller.
	if a == b {
		return true
	}
	// Replace known localhost hostnames with a canonical form for comparison.
	canonical := func(origin string) string {
		// Remove scheme prefix to extract host:port.
		var scheme, rest string
		switch {
		case strings.HasPrefix(origin, "http://"):
			scheme, rest = "http", origin[len("http://"):]
		case strings.HasPrefix(origin, "https://"):
			scheme, rest = "https", origin[len("https://"):]
		default:
			return origin
		}
		// Split host:port
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

	tenantRepo := repository.NewGormTenantRepo(db)
	usuarioRepo := repository.NewGormUsuarioRepo(db)
	membroRepo := repository.NewGormMembroRepo(db)
	cartaoRepo := repository.NewGormCartaoRepo(db)
	faturaRepo := repository.NewGormFaturaRepo(db)
	gastoRepo := repository.NewGormGastoRepo(db)
	contaFixaRepo := repository.NewGormContaFixaRepo(db)
	auditRepo := repository.NewGormAuditLogRepo(db)
	validationRepo := repository.NewGormProductValidationRepo(db)
	resetRepo := repository.NewGormPasswordResetTokenRepo(db)

	emailSvc := service.NewEmailService(cfg)

	wsHub := ws.NewHub()

	upgrader := createUpgrader(cfg.CORSOrigins)

	authSvc := service.NewAuthService(cfg, db, usuarioRepo, tenantRepo, membroRepo, resetRepo, emailSvc)
	financeiroSvc := service.NewFinanceiroService(
		db,
		membroRepo, cartaoRepo, faturaRepo, gastoRepo,
		contaFixaRepo, auditRepo, validationRepo, tenantRepo, wsHub,
	)

	if err := financeiroSvc.LoadPermissions(context.Background()); err != nil {
		log.Printf("warning: failed to load tenant permissions: %v", err)
	}

	authHandler := handler.NewAuthHandler(authSvc)
	financeiroHandler := handler.NewFinanceiroHandler(financeiroSvc)

	validator.RegisterGinValidators()

	r := gin.Default()

	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORSMiddleware(cfg.CORSOrigins))

	r.GET("/health", handler.HealthCheck)
	r.GET("/health/deep", handler.DeepHealthCheck(db))
	// Swagger documentation is disabled by default in production.
	// Set SWAGGER_ENABLED=true in .env to enable it.
	if cfg.EnableSwagger {
		swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler)
		r.GET("/swagger/*any", func(c *gin.Context) {
			if c.Param("any") == "/" {
				c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
				return
			}
			swaggerHandler(c)
		})
	}

	inviteRateLimit := middleware.RateLimit(10, time.Minute)
	r.GET("/api/tenants/invite/:code", inviteRateLimit, authHandler.InvitePreview)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			authRateLimit := middleware.RateLimit(5, time.Minute)
			auth.POST("/register", authRateLimit, authHandler.Register)
			auth.POST("/login", authRateLimit, authHandler.Login)
			auth.POST("/google", authRateLimit, authHandler.GoogleLogin)
			auth.POST("/forgot-password", authRateLimit, authHandler.ForgotPassword)
			auth.POST("/reset-password", authRateLimit, authHandler.ResetPassword)
		}

		protected := api.Group("")
		protected.Use(middleware.JWTAuth(cfg.JWTSecret))
		protected.Use(middleware.CSRFToken())
		{
			protected.GET("/auth/me", authHandler.Session)
			protected.POST("/tenants", authHandler.CreateTenant)
			protected.POST("/tenants/join", authHandler.JoinTenant)

			tenant := protected.Group("")
			tenant.Use(middleware.TenantRequired(membroRepo))
			{
				tenant.GET("/membros", financeiroHandler.ListMembros)

				tenant.GET("/cartoes", financeiroHandler.ListCartoes)

				tenant.GET("/contas-fixas", financeiroHandler.ListContasFixas)

				tenant.GET("/faturas", financeiroHandler.ListFaturas)

				tenant.GET("/gastos", financeiroHandler.ListGastos)

				tenant.GET("/audit-logs", financeiroHandler.GetAuditLogs)
				tenant.GET("/tenants/permissions", financeiroHandler.GetPermissions)

				write := tenant.Group("")
				write.Use(middleware.RoleRequired(model.RoleAdmin, model.RoleMorador))
				write.Use(middleware.RateLimit(60, time.Minute))
				{
					write.POST("/membros/with-account", financeiroHandler.CreateMembroWithAccount)

					write.POST("/membros", financeiroHandler.CreateMembro)
					write.PUT("/membros/:id", financeiroHandler.UpdateMembro)

					write.POST("/cartoes", financeiroHandler.CreateCartao)
					write.DELETE("/cartoes/:id", financeiroHandler.DeleteCartao)

					write.POST("/contas-fixas", financeiroHandler.CreateContaFixa)
					write.PUT("/contas-fixas/:id", financeiroHandler.UpdateContaFixa)
					write.DELETE("/contas-fixas/:id", financeiroHandler.DeleteContaFixa)

					write.POST("/faturas", financeiroHandler.CreateFatura)
					write.POST("/faturas/batch", financeiroHandler.CreateFaturaBatch)

					write.POST("/gastos", financeiroHandler.CreateGasto)
					write.PUT("/gastos/:id", financeiroHandler.UpdateGasto)
					write.POST("/gastos/batch", financeiroHandler.CreateGastoBatch)
					write.DELETE("/gastos/:id", financeiroHandler.DeleteGasto)
					write.POST("/gastos/delete-batch", financeiroHandler.DeleteGastoBatch)

					write.POST("/validation-events", financeiroHandler.RecordValidationEvent)
				}

				// Permission management is restricted to ADMIN only —
				// MORADOR should not be able to escalate privileges.
				admin := tenant.Group("")
				admin.Use(middleware.RoleRequired(model.RoleAdmin))
				admin.Use(middleware.RateLimit(60, time.Minute))
				{
					admin.PATCH("/tenants/permissions/:role", financeiroHandler.UpdatePermissions)
				}
			}
		}
	}

	r.GET("/ws", func(c *gin.Context) {
		// Token extraction priority (most secure first):
		// 1. Authorization: Bearer <token> header (programmatic clients)
		// 2. Sec-WebSocket-Protocol: divi.<token> subprotocol (browser clients)
		// 3. ?token= query param (legacy fallback, logged by proxies)
		tokenSource := "none"
		tokenStr := c.Query("token")
		if tokenStr != "" {
			tokenSource = "query"
		}
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
				tokenSource = "header"
			}
		}
		if tokenStr == "" {
			if proto := c.GetHeader("Sec-WebSocket-Protocol"); strings.HasPrefix(proto, "divi.") {
				tokenStr = strings.TrimPrefix(proto, "divi.")
				tokenSource = "subprotocol"
			}
		}
		if tokenStr == "" {
			log.Printf("websocket: missing token (origin=%s, has_upgrade=%v)",
				c.GetHeader("Origin"),
				c.GetHeader("Upgrade") != "")
			c.JSON(http.StatusBadRequest, gin.H{"message": "token query param, Authorization header, or WebSocket subprotocol required"})
			return
		}

		tenantID := c.Query("tenant_id")
		if tenantID == "" {
			log.Printf("websocket: missing tenant_id")
			c.JSON(http.StatusBadRequest, gin.H{"message": "tenant_id query param required"})
			return
		}

		userID, err := middleware.ValidateToken(cfg.JWTSecret, tokenStr)
		if err != nil {
			log.Printf("websocket: invalid token (source=%s, tenant=%s): %v", tokenSource, tenantID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "token inválido"})
			return
		}

		membro, err := membroRepo.GetByUserID(c.Request.Context(), tenantID, userID)
		if err != nil {
			log.Printf("websocket: membership lookup error (user=%s, tenant=%s, source=%s): %v",
				userID, tenantID, tokenSource, err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "erro interno ao verificar acesso"})
			return
		}
		if membro == nil {
			log.Printf("websocket: membership denied (user=%s, tenant=%s, source=%s): not a member",
				userID, tenantID, tokenSource)
			c.JSON(http.StatusForbidden, gin.H{"message": "acesso negado a este núcleo"})
			return
		}

		log.Printf("websocket: auth OK (user=%s, tenant=%s, source=%s, origin=%s)",
			userID, tenantID, tokenSource, c.GetHeader("Origin"))

		// Se não for uma requisição de handshake de WebSocket (como a chamada preflight do frontend),
		// respondemos com 200 OK para evitar um erro 400 Bad Request estético no console do navegador.
		if !strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			c.Status(http.StatusOK)
			return
		}

		log.Printf("websocket: upgrading connection")

		var upgradeHeader http.Header
		if tokenSource == "subprotocol" {
			upgradeHeader = make(http.Header)
			upgradeHeader.Set("Sec-WebSocket-Protocol", c.GetHeader("Sec-WebSocket-Protocol"))
		}

		conn, upErr := upgrader.Upgrade(c.Writer, c.Request, upgradeHeader)
		if upErr != nil {
			log.Printf("websocket: upgrade failed (user=%s, tenant=%s, origin=%s): %v",
				userID, tenantID, c.GetHeader("Origin"), upErr)
			return
		}

		log.Printf("websocket: connected (user=%s, tenant=%s)", userID, tenantID)
		ws.HandleClient(wsHub, conn, tenantID)
	})

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
