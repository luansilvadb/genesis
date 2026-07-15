package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
)

func TenantRequired(membroRepo repository.MembroRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "X-Tenant-ID header é obrigatório"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "usuário não autenticado"})
			return
		}

		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "usuário inválido"})
			return
		}

		membro, err := membroRepo.GetByUserID(c.Request.Context(), tenantID, userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "erro interno ao verificar acesso"})
			return
		}
		if membro == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "acesso negado a este núcleo"})
			return
		}

		if !membro.Ativo {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "membro desativado neste núcleo"})
			return
		}

		c.Set("tenantID", tenantID)
		c.Set("userRole", string(membro.Role))
		c.Next()
	}
}
