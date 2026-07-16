package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/service"
)

type FinanceiroHandler struct {
	svc *service.FinanceiroService
}

func NewFinanceiroHandler(svc *service.FinanceiroService) *FinanceiroHandler {
	return &FinanceiroHandler{svc: svc}
}

// @Summary Create membro with account
// @Description Adiciona um membro ao núcleo familiar com conta de usuário vinculada
// @Tags Membros
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateMembroWithAccountRequest true "Dados do membro com credenciais"
// @Success 201 {object} dto.MembroResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/membros/with-account [post]
func (h *FinanceiroHandler) CreateMembroWithAccount(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.CreateMembroWithAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	membro, err := h.svc.CreateMembroWithAccount(c.Request.Context(), tenantID, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, membro)
}

// @Summary Create membro
// @Description Adiciona um membro ao núcleo familiar
// @Tags Membros
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateMembroRequest true "Dados do membro"
// @Success 201 {object} dto.MembroResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/membros [post]
func (h *FinanceiroHandler) CreateMembro(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.CreateMembroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	membro, err := h.svc.CreateMembro(c.Request.Context(), tenantID, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, membro)
}

// @Summary Update membro
// @Description Atualiza os dados de um membro do nÃºcleo familiar
// @Tags Membros
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do membro"
// @Param request body dto.UpdateMembroRequest true "Dados para atualizaÃ§Ã£o"
// @Success 200 {object} dto.MembroResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/membros/{id} [put]
func (h *FinanceiroHandler) UpdateMembro(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")

	var req dto.UpdateMembroRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	membro, err := h.svc.UpdateMembro(c.Request.Context(), tenantID, id, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, membro)
}

// @Summary Update gasto
// @Description Atualiza um gasto existente
// @Tags Gastos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID do gasto"
// @Param request body dto.UpdateGastoRequest true "Dados para atualizaÃ§Ã£o"
// @Success 200 {object} dto.GastoResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/gastos/{id} [put]
func (h *FinanceiroHandler) UpdateGasto(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")

	var req dto.UpdateGastoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	gasto, err := h.svc.UpdateGasto(c.Request.Context(), tenantID, id, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gasto)
}

// @Summary List membros
// @Description Lista todos os membros do núcleo familiar
// @Tags Membros
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.MembroResponse
// @Failure 500 {object} map[string]string
// @Router /api/membros [get]
func (h *FinanceiroHandler) ListMembros(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	ctx := c.Request.Context()
	respondListOrPaginated(c,
		func() ([]dto.MembroResponse, error) { return h.svc.ListMembros(ctx, tenantID) },
		func(offset, limit int) ([]dto.MembroResponse, int64, error) {
			return h.svc.ListMembrosPaginated(ctx, tenantID, offset, limit)
		},
	)
}

// @Summary Create cartão
// @Description Adiciona um cartão de crédito ao núcleo
// @Tags Cartões
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCartaoRequest true "Dados do cartão"
// @Success 201 {object} model.Cartao
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/cartoes [post]
func (h *FinanceiroHandler) CreateCartao(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.CreateCartaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	cartao, err := h.svc.CreateCartao(c.Request.Context(), tenantID, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.CartaoToResponse(cartao))
}

// @Summary Create gasto
// @Description Registra uma despesa no núcleo
// @Tags Gastos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateGastoRequest true "Dados do gasto"
// @Success 201 {object} dto.GastoResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/gastos [post]
func (h *FinanceiroHandler) CreateGasto(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.CreateGastoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	gasto, err := h.svc.CreateGasto(c.Request.Context(), tenantID, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gasto)
}

// @Summary List gastos
// @Description Lista despesas do núcleo familiar
// @Tags Gastos
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.GastoResponse
// @Failure 500 {object} map[string]string
// @Router /api/gastos [get]
func (h *FinanceiroHandler) ListGastos(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	ctx := c.Request.Context()
	respondListOrPaginated(c,
		func() ([]dto.GastoResponse, error) { return h.svc.ListGastos(ctx, tenantID) },
		func(offset, limit int) ([]dto.GastoResponse, int64, error) {
			return h.svc.ListGastosPaginated(ctx, tenantID, offset, limit)
		},
	)
}

// @Summary Create conta fixa
// @Description Registra uma conta fixa/recorrente
// @Tags Contas Fixas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateContaFixaRequest true "Dados da conta fixa"
// @Success 201 {object} dto.ContaFixaResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/contas-fixas [post]
func (h *FinanceiroHandler) CreateContaFixa(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.CreateContaFixaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	conta, err := h.svc.CreateContaFixa(c.Request.Context(), tenantID, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, conta)
}

// @Summary Update conta fixa
// @Description Atualiza uma conta fixa existente
// @Tags Contas Fixas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID da conta fixa"
// @Param request body dto.UpdateContaFixaRequest true "Dados atualizados"
// @Success 200 {object} dto.ContaFixaResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/contas-fixas/:id [put]
func (h *FinanceiroHandler) UpdateContaFixa(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")

	var req dto.UpdateContaFixaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	conta, err := h.svc.UpdateContaFixa(c.Request.Context(), tenantID, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrContaFixaNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "conta fixa nao encontrada"})
			return
		}
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, conta)
}

// @Summary List cartoes
// @Description Lista todos os cartões do núcleo familiar
// @Tags Cartões
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Cartao
// @Failure 500 {object} map[string]string
// @Router /api/cartoes [get]
func (h *FinanceiroHandler) ListCartoes(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var params dto.PaginationParams
	if err := c.ShouldBindQuery(&params); err == nil && params.Paginated {
		params.Normalize()
		cartoes, total, err := h.svc.ListCartoesPaginated(c.Request.Context(), tenantID, params.Offset(), params.PageSize)
		if err != nil {
			respondInternalError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.NewPaginatedResponse(dto.CartoesToResponse(cartoes), total, params))
		return
	}

	cartoes, err := h.svc.ListCartoes(c.Request.Context(), tenantID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	if cartoes == nil {
		cartoes = []model.Cartao{}
	}
	c.JSON(http.StatusOK, dto.CartoesToResponse(cartoes))
}

// @Summary Delete cartao
// @Description Exclui um cartão do núcleo
// @Tags Cartões
// @Security BearerAuth
// @Param id path string true "ID do cartão"
// @Success 204 {object} nil
// @Failure 500 {object} map[string]string
// @Router /api/cartoes/{id} [delete]
func (h *FinanceiroHandler) DeleteCartao(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")
	if err := h.svc.DeleteCartao(c.Request.Context(), tenantID, id); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary List contas fixas
// @Description Lista todas as contas fixas do núcleo
// @Tags Contas Fixas
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.ContaFixaResponse
// @Failure 500 {object} map[string]string
// @Router /api/contas-fixas [get]
func (h *FinanceiroHandler) ListContasFixas(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	ctx := c.Request.Context()
	respondListOrPaginated(c,
		func() ([]dto.ContaFixaResponse, error) { return h.svc.ListContasFixas(ctx, tenantID) },
		func(offset, limit int) ([]dto.ContaFixaResponse, int64, error) {
			return h.svc.ListContasFixasPaginated(ctx, tenantID, offset, limit)
		},
	)
}

// @Summary Delete conta fixa
// @Description Exclui uma conta fixa do núcleo
// @Tags Contas Fixas
// @Security BearerAuth
// @Param id path string true "ID da conta fixa"
// @Success 204 {object} nil
// @Failure 500 {object} map[string]string
// @Router /api/contas-fixas/{id} [delete]
func (h *FinanceiroHandler) DeleteContaFixa(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")
	if err := h.svc.DeleteContaFixa(c.Request.Context(), tenantID, id); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary Create fatura
// @Description Registra uma fatura de cartão de crédito
// @Tags Faturas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateFaturaRequest true "Dados da fatura"
// @Success 201 {object} model.Fatura
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/faturas [post]
func (h *FinanceiroHandler) CreateFatura(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.CreateFaturaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	fatura, err := h.svc.CreateFatura(c.Request.Context(), tenantID, &req)
	if err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.FaturaToResponse(fatura))
}

// @Summary Create fatura batch
// @Description Registra múltiplas faturas de uma vez
// @Tags Faturas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body []dto.CreateFaturaRequest true "Lista de faturas"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/faturas/batch [post]
func (h *FinanceiroHandler) CreateFaturaBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var reqs []dto.CreateFaturaRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.svc.CreateFaturaBatch(c.Request.Context(), tenantID, reqs); err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "faturas registradas"})
}

// @Summary List faturas
// @Description Lista todas as faturas do núcleo
// @Tags Faturas
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.Fatura
// @Failure 500 {object} map[string]string
// @Router /api/faturas [get]
func (h *FinanceiroHandler) ListFaturas(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var params dto.PaginationParams
	if err := c.ShouldBindQuery(&params); err == nil && params.Paginated {
		params.Normalize()
		faturas, total, err := h.svc.ListFaturasPaginated(c.Request.Context(), tenantID, params.Offset(), params.PageSize)
		if err != nil {
			respondInternalError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.NewPaginatedResponse(dto.FaturasToResponse(faturas), total, params))
		return
	}

	faturas, err := h.svc.ListFaturas(c.Request.Context(), tenantID)
	if err != nil {
		respondInternalError(c, err)
		return
	}
	if faturas == nil {
		faturas = []model.Fatura{}
	}
	c.JSON(http.StatusOK, dto.FaturasToResponse(faturas))
}

// @Summary Create gasto batch
// @Description Registra múltiplos gastos de uma vez
// @Tags Gastos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body []dto.CreateGastoRequest true "Lista de gastos"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/gastos/batch [post]
func (h *FinanceiroHandler) CreateGastoBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var reqs []dto.CreateGastoRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.svc.CreateGastoBatch(c.Request.Context(), tenantID, reqs); err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "gastos registrados"})
}

// @Summary Delete gasto
// @Description Exclui um gasto do núcleo
// @Tags Gastos
// @Security BearerAuth
// @Param id path string true "ID do gasto"
// @Success 204 {object} nil
// @Failure 500 {object} map[string]string
// @Router /api/gastos/{id} [delete]
func (h *FinanceiroHandler) DeleteGasto(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	id := c.Param("id")
	if err := h.svc.DeleteGasto(c.Request.Context(), tenantID, id); err != nil {
		respondInternalError(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// @Summary Delete gasto batch
// @Description Exclui múltiplos gastos de uma vez
// @Tags Gastos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.DeleteGastoBatchRequest true "Lista de IDs"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/gastos/delete-batch [post]
func (h *FinanceiroHandler) DeleteGastoBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.DeleteGastoBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.svc.DeleteGastoBatch(c.Request.Context(), tenantID, req.IDs); err != nil {
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "gastos excluídos"})
}

// @Summary Get permissions
// @Description Obtém as permissões por role
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]dto.RolePermissions
// @Router /api/tenants/permissions [get]
func (h *FinanceiroHandler) GetPermissions(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	perms := h.svc.GetPermissions(c.Request.Context(), tenantID)
	c.JSON(http.StatusOK, perms)
}

// @Summary Update permissions
// @Description Atualiza permissões de uma role
// @Tags Permissions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role path string true "Role name"
// @Param request body dto.RolePermissions true "Permissions to update"
// @Success 200 {object} map[string]dto.RolePermissions
// @Router /api/tenants/permissions/{role} [patch]
func (h *FinanceiroHandler) UpdatePermissions(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	role := c.Param("role")
	if role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "role é obrigatória"})
		return
	}

	var partial dto.RolePermissions
	if err := c.ShouldBindJSON(&partial); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result := h.svc.UpdatePermissions(c.Request.Context(), tenantID, role, partial)
	c.JSON(http.StatusOK, result)
}

// @Summary Record validation event
// @Description Registra um evento de validação de produto
// @Tags Validação
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ValidateEventRequest true "Dados do evento"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/validation-events [post]
func (h *FinanceiroHandler) RecordValidationEvent(c *gin.Context) {
	tenantID := c.GetString("tenantID")

	var req dto.ValidateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := h.svc.RecordValidationEvent(c.Request.Context(), tenantID, &req); err != nil {
		if errors.Is(err, service.ErrEventAlreadyRegistered) {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		respondInternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "evento registrado"})
}

// @Summary Get audit logs
// @Description Obtém logs de auditoria do núcleo
// @Tags Auditoria
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.AuditLog
// @Failure 500 {object} map[string]string
// @Router /api/audit-logs [get]
func (h *FinanceiroHandler) GetAuditLogs(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	ctx := c.Request.Context()
	respondListOrPaginated(c,
		func() ([]model.AuditLog, error) { return h.svc.GetAuditLogs(ctx, tenantID) },
		func(offset, limit int) ([]model.AuditLog, int64, error) {
			return h.svc.GetAuditLogsPaginated(ctx, tenantID, offset, limit)
		},
	)
}
