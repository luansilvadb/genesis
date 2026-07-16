package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/validator"
)

var ErrEventAlreadyRegistered = errors.New("evento já registrado")
var ErrContaFixaNotFound = errors.New("conta fixa nao encontrada")

// ── Generic pointer helpers ────────────────────────────────────────────────

// setIfNotNil copies *src to *dest when src is non-nil.
// Useful for applying partial-update DTOs onto model structs.
func setIfNotNil[T any](dest *T, src *T) {
	if src != nil {
		*dest = *src
	}
}

// setNullStringIfNotNil unpacks a dto.NullString wrapper into a *string dest,
// preserving the nil-as-unset semantics (src.Value may itself be nil).
func setNullStringIfNotNil(dest **string, src *dto.NullString) {
	if src != nil {
		*dest = src.Value
	}
}

// setPtrIfNotNil assigns *src to *dest when src is non-nil.
// Use when dest is already a pointer field (e.g. **bool ← *bool).
func setPtrIfNotNil[T any](dest **T, src *T) {
	if src != nil {
		*dest = src
	}
}

// ── JSON helpers ────────────────────────────────────────────────────────────

// rawMessageToString converts a *json.RawMessage to *string for model storage
// (jsonb columns in the DB store JSON as strings).
func rawMessageToString(raw *json.RawMessage) *string {
	if raw == nil {
		return nil
	}
	s := string(*raw)
	return &s
}

// stringToRawMessage converts a *string to *json.RawMessage for DTO responses.
func stringToRawMessage(s *string) *json.RawMessage {
	if s == nil {
		return nil
	}
	r := json.RawMessage(*s)
	return &r
}

// ── Service struct ──────────────────────────────────────────────────────────

type FinanceiroService struct {
	db             *gorm.DB
	membroRepo     repository.MembroRepository
	cartaoRepo     repository.CartaoRepository
	faturaRepo     repository.FaturaRepository
	gastoRepo      repository.GastoRepository
	contaFixaRepo  repository.ContaFixaRepository
	auditRepo      repository.AuditLogRepository
	validationRepo repository.ProductValidationRepository
	tenantRepo     repository.TenantRepository
	wsHub          WSHub

	permissionsMu sync.RWMutex
	// permissions is tenant-scoped: map[tenantID]map[role]RolePermissions.
	// The inner map uses the same structure as the previous global map so existing
	// lookup code in GetPermissions continues to work when scoped to a tenant.
	permissions map[string]map[string]dto.RolePermissions
}

func NewFinanceiroService(
	db *gorm.DB,
	membroRepo repository.MembroRepository,
	cartaoRepo repository.CartaoRepository,
	faturaRepo repository.FaturaRepository,
	gastoRepo repository.GastoRepository,
	contaFixaRepo repository.ContaFixaRepository,
	auditRepo repository.AuditLogRepository,
	validationRepo repository.ProductValidationRepository,
	tenantRepo repository.TenantRepository,
	wsHub WSHub,
) *FinanceiroService {
	return &FinanceiroService{
		db:             db,
		membroRepo:     membroRepo,
		cartaoRepo:     cartaoRepo,
		faturaRepo:     faturaRepo,
		gastoRepo:      gastoRepo,
		contaFixaRepo:  contaFixaRepo,
		auditRepo:      auditRepo,
		validationRepo: validationRepo,
		tenantRepo:     tenantRepo,
		wsHub:          wsHub,
		permissions:    make(map[string]map[string]dto.RolePermissions),
	}
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func boolPtr(b bool) *bool { return &b }

// createAuditLog is a best-effort audit log write. A failed audit log does not
// roll back the primary operation, but the error is now observable via logging.
func (s *FinanceiroService) createAuditLog(ctx context.Context, tenantID, membroID, acao, detalhes string) {
	if err := s.auditRepo.Create(ctx, &model.AuditLog{
		TenantID: tenantID,
		MembroID: membroID,
		Acao:     acao,
		Detalhes: detalhes,
	}); err != nil {
		log.Printf("ERROR: failed to write audit log [%s] for tenant=%s membro=%s: %v", acao, tenantID, membroID, err)
	}
}

// ── Permissions ─────────────────────────────────────────────────────────────

func defaultPermissions() map[string]dto.RolePermissions {
	return map[string]dto.RolePermissions{
		string(model.RoleAdmin): {
			AllowLancarGasto:          boolPtr(true),
			AllowGerenciarCartoes:     boolPtr(true),
			AllowGerenciarContasFixas: boolPtr(true),
			AllowRegistrarNetting:     boolPtr(true),
			AllowVerAuditLogs:         boolPtr(true),
			AllowFecharPeriodo:        boolPtr(true),
			AllowAlterarRenda:         boolPtr(true),
			AllowAlterarNome:          boolPtr(true),
		},
		string(model.RoleMorador): {
			AllowLancarGasto:          boolPtr(true),
			AllowGerenciarCartoes:     boolPtr(false),
			AllowGerenciarContasFixas: boolPtr(true),
			AllowRegistrarNetting:     boolPtr(true),
			AllowVerAuditLogs:         boolPtr(false),
			AllowFecharPeriodo:        boolPtr(false),
			AllowAlterarRenda:         boolPtr(false),
			AllowAlterarNome:          boolPtr(true),
		},
		string(model.RoleVisualizador): {
			AllowLancarGasto:          boolPtr(false),
			AllowGerenciarCartoes:     boolPtr(false),
			AllowGerenciarContasFixas: boolPtr(false),
			AllowRegistrarNetting:     boolPtr(false),
			AllowVerAuditLogs:         boolPtr(true),
			AllowFecharPeriodo:        boolPtr(false),
			AllowAlterarRenda:         boolPtr(false),
			AllowAlterarNome:          boolPtr(false),
		},
	}
}

// mergePermissionFields applies non-nil partial permission fields onto existing.
func mergePermissionFields(existing *dto.RolePermissions, partial dto.RolePermissions) {
	setPtrIfNotNil(&existing.AllowLancarGasto, partial.AllowLancarGasto)
	setPtrIfNotNil(&existing.AllowGerenciarCartoes, partial.AllowGerenciarCartoes)
	setPtrIfNotNil(&existing.AllowGerenciarContasFixas, partial.AllowGerenciarContasFixas)
	setPtrIfNotNil(&existing.AllowRegistrarNetting, partial.AllowRegistrarNetting)
	setPtrIfNotNil(&existing.AllowVerAuditLogs, partial.AllowVerAuditLogs)
	setPtrIfNotNil(&existing.AllowFecharPeriodo, partial.AllowFecharPeriodo)
	setPtrIfNotNil(&existing.AllowAlterarRenda, partial.AllowAlterarRenda)
	setPtrIfNotNil(&existing.AllowAlterarNome, partial.AllowAlterarNome)
}

// clonePermissionsMap returns a shallow copy of a defaults map merged with tenant overrides.
func clonePermissionsMap(defaults, overrides map[string]dto.RolePermissions) map[string]dto.RolePermissions {
	result := make(map[string]dto.RolePermissions, len(defaults))
	for k, v := range defaults {
		result[k] = v
	}
	for k, v := range overrides {
		result[k] = v
	}
	return result
}

func (s *FinanceiroService) LoadPermissions(ctx context.Context) error {
	tenants, err := s.tenantRepo.ListAll(ctx)
	if err != nil {
		return err
	}
	s.permissionsMu.Lock()
	defer s.permissionsMu.Unlock()
	for _, t := range tenants {
		if t.PermissionsJSON == nil || *t.PermissionsJSON == "" {
			continue
		}
		var overrides map[string]dto.RolePermissions
		if err := json.Unmarshal([]byte(*t.PermissionsJSON), &overrides); err != nil {
			log.Printf("ERROR: failed to unmarshal permissions for tenant %s: %v", t.ID, err)
			continue
		}
		s.permissions[t.ID] = overrides
	}
	return nil
}

func (s *FinanceiroService) GetPermissions(ctx context.Context, tenantID string) map[string]dto.RolePermissions {
	s.permissionsMu.RLock()
	defer s.permissionsMu.RUnlock()

	defaults := defaultPermissions()
	if tenantPerms, ok := s.permissions[tenantID]; ok {
		return clonePermissionsMap(defaults, tenantPerms)
	}

	result := make(map[string]dto.RolePermissions, len(defaults))
	for k, v := range defaults {
		result[k] = v
	}
	return result
}

func (s *FinanceiroService) UpdatePermissions(ctx context.Context, tenantID string, role string, partial dto.RolePermissions) map[string]dto.RolePermissions {
	s.permissionsMu.Lock()
	defer s.permissionsMu.Unlock()

	defaults := defaultPermissions()

	if _, ok := s.permissions[tenantID]; !ok {
		s.permissions[tenantID] = make(map[string]dto.RolePermissions)
	}
	tenantPerms := s.permissions[tenantID]

	if _, ok := tenantPerms[role]; !ok {
		if def, exists := defaults[role]; exists {
			tenantPerms[role] = def
		}
	}

	if existing, ok := tenantPerms[role]; ok {
		mergePermissionFields(&existing, partial)
		tenantPerms[role] = existing
	}

	// Build snapshot inside the write lock to avoid deadlock.
	result := clonePermissionsMap(defaults, tenantPerms)
	rolePerms := result[role]

	s.persistPermissionsJSON(ctx, tenantID, tenantPerms)

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypePermissionsUpdate,
		Payload: map[string]interface{}{"role": role, "permissions": rolePerms},
	})

	return result
}

// persistPermissionsJSON serializes the tenant permissions map and persists it
// to the database. Errors are logged but not returned — permission persistence
// is best-effort (in-memory state takes precedence).
func (s *FinanceiroService) persistPermissionsJSON(ctx context.Context, tenantID string, tenantPerms map[string]dto.RolePermissions) {
	data, err := json.Marshal(tenantPerms)
	if err != nil {
		log.Printf("ERROR: failed to marshal permissions for tenant %s: %v", tenantID, err)
		return
	}
	tenant, tenantErr := s.tenantRepo.GetByID(ctx, tenantID)
	if tenantErr != nil || tenant == nil {
		return
	}
	jsonStr := string(data)
	tenant.PermissionsJSON = &jsonStr
	if saveErr := s.tenantRepo.Update(ctx, tenant); saveErr != nil {
		log.Printf("ERROR: failed to persist permissions for tenant %s: %v", tenantID, saveErr)
	}
}

// ── Gasto defaults & model building ─────────────────────────────────────────

// gastoDefaults holds parsed default/normalized values for gasto creation.
type gastoDefaults struct {
	installments      int
	totalInstallments int
	method            string
	splitMode         model.SplitMode
}

// parseGastoDefaults extracts and normalizes default values from a request.
func parseGastoDefaults(req *dto.CreateGastoRequest) gastoDefaults {
	d := gastoDefaults{
		installments:      1,
		totalInstallments: 1,
		method:            "pix",
		splitMode:         model.SplitModeCustom,
	}
	if req.Installments != nil {
		d.installments = *req.Installments
	}
	if req.TotalInstallments != nil {
		d.totalInstallments = *req.TotalInstallments
	}
	if req.Method != "" {
		d.method = req.Method
	}
	if req.SplitMode != "" {
		d.splitMode = model.SplitMode(req.SplitMode)
	}
	return d
}

// validateDivisoes checks that CUSTOM split mode has at least one divisao and
// that the sum of divisoes matches the total value. The optional label is
// included in error messages for batch operations to identify the offending row.
func validateDivisoes(divisoes []dto.SplitItem, valorTotal int64, label string) error {
	if len(divisoes) == 0 {
		return fmt.Errorf("split mode custom requer ao menos uma divisao")
	}
	var sum int64
	for _, d := range divisoes {
		sum += d.ValorCentavos
	}
	if sum != valorTotal {
		if label != "" {
			return fmt.Errorf("soma das divisoes (%d) nao confere com o valor total (%d) no gasto %q", sum, valorTotal, label)
		}
		return fmt.Errorf("soma das divisoes (%d) nao confere com o valor total (%d)", sum, valorTotal)
	}
	return nil
}

// buildGastoModel constructs a model.Gasto from a request and parsed defaults.
func buildGastoModel(req *dto.CreateGastoRequest, tenantID string, d gastoDefaults) *model.Gasto {
	gasto := &model.Gasto{
		ID:                 uuid.New().String(),
		TenantID:           tenantID,
		FaturaID:           req.FaturaID,
		Descricao:          req.Descricao,
		ValorTotalCentavos: req.ValorTotalCentavos,
		CompradorID:        req.CompradorID,
		Installments:       d.installments,
		TotalInstallments:  d.totalInstallments,
		IsLoan:             req.IsLoan,
		BorrowerID:         req.BorrowerID,
		Method:             d.method,
		CardOwnerID:        req.CardOwnerID,
		IsPrivate:          req.IsPrivate,
		IsSettlement:       req.IsSettlement,
		SettlementDetails:  nil,
		GrupoParcelasID:    req.GrupoParcelasID,
		RecurringBillID:    req.RecurringBillID,
		SplitMode:          d.splitMode,
	}

	if req.SettlementDetails != nil {
		gasto.SettlementDetails = rawMessageToString(req.SettlementDetails)
	}

	for _, div := range req.Divisoes {
		gasto.Divisoes = append(gasto.Divisoes, model.DivisaoGasto{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			MembroID:      div.MembroID,
			ValorCentavos: div.ValorCentavos,
		})
	}

	return gasto
}

// ── Membro ──────────────────────────────────────────────────────────────────

func (s *FinanceiroService) CreateMembro(ctx context.Context, tenantID string, req *dto.CreateMembroRequest) (*dto.MembroResponse, error) {
	membro := &model.MembroCasa{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		Nome:     req.Nome,
		Avatar:   req.Avatar,
		Role:     model.RoleMorador,
	}

	if err := s.membroRepo.Create(ctx, membro); err != nil {
		return nil, err
	}

	s.createAuditLog(ctx, tenantID, membro.ID, "CRIAR_MEMBRO", "Membro "+req.Nome+" criado")

	resp := membroToResponse(membro)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeMemberCreated,
		Payload: resp,
	})
	return resp, nil
}

func (s *FinanceiroService) CreateMembroWithAccount(ctx context.Context, tenantID string, req *dto.CreateMembroWithAccountRequest) (*dto.MembroResponse, error) {
	if err := validator.ValidatePassword(req.Password); err != nil {
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

	membro := &model.MembroCasa{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		Nome:     req.Nome,
		Avatar:   req.Avatar,
		Role:     model.RoleMorador,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if createErr := tx.Create(user).Error; createErr != nil {
			return createErr
		}
		membro.UserID = &user.ID
		if createErr := tx.Create(membro).Error; createErr != nil {
			return createErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.createAuditLog(ctx, tenantID, membro.ID, "CRIAR_MEMBRO_COM_CONTA", "Membro "+req.Nome+" criado com conta vinculada")

	resp := membroToResponse(membro)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeMemberCreated,
		Payload: resp,
	})
	return resp, nil
}

func (s *FinanceiroService) UpdateMembro(ctx context.Context, tenantID string, membroID string, req *dto.UpdateMembroRequest) (*dto.MembroResponse, error) {
	membro, err := s.membroRepo.GetByID(ctx, membroID, tenantID)
	if err != nil {
		return nil, err
	}
	if membro == nil {
		return nil, fmt.Errorf("membro não encontrado")
	}

	applyMembroFieldUpdates(membro, req)

	if err := s.membroRepo.Update(ctx, membro); err != nil {
		return nil, err
	}

	s.createAuditLog(ctx, tenantID, membro.ID, "ATUALIZAR_MEMBRO", "Membro "+membro.Nome+" atualizado")

	resp := membroToResponse(membro)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeMemberUpdated,
		Payload: resp,
	})
	return resp, nil
}

// applyMembroFieldUpdates applies non-nil partial-update fields from the
// request onto the membro model, preserving unset fields.
func applyMembroFieldUpdates(m *model.MembroCasa, req *dto.UpdateMembroRequest) {
	setIfNotNil(&m.Nome, req.Nome)
	setIfNotNil(&m.Avatar, req.Avatar)
	setIfNotNil(&m.Ativo, req.Ativo)
	if req.Role != nil {
		m.Role = model.Role(*req.Role)
	}
	setPtrIfNotNil(&m.RendaCentavos, req.RendaCentavos)
}

func (s *FinanceiroService) ListMembros(ctx context.Context, tenantID string) ([]dto.MembroResponse, error) {
	membros, err := s.membroRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.MembroResponse, len(membros))
	for i, m := range membros {
		resp[i] = *membroToResponse(&m)
	}
	return resp, nil
}

func (s *FinanceiroService) ListMembrosPaginated(ctx context.Context, tenantID string, offset, limit int) ([]dto.MembroResponse, int64, error) {
	membros, total, err := s.membroRepo.ListByTenantPaginated(ctx, tenantID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.MembroResponse, len(membros))
	for i, m := range membros {
		resp[i] = *membroToResponse(&m)
	}
	return resp, total, nil
}

// ── Cartão ──────────────────────────────────────────────────────────────────

func (s *FinanceiroService) CreateCartao(ctx context.Context, tenantID string, req *dto.CreateCartaoRequest) (*model.Cartao, error) {
	cartao := &model.Cartao{
		ID:                  uuid.New().String(),
		TenantID:            tenantID,
		Nome:                req.Nome,
		DiaFechamento:       req.DiaFechamento,
		ResponsavelPadraoID: req.ResponsavelPadraoID,
	}

	if err := s.cartaoRepo.Create(ctx, cartao); err != nil {
		return nil, err
	}

	s.createAuditLog(ctx, tenantID, req.ResponsavelPadraoID, "CRIAR_CARTAO", "Cartão "+req.Nome+" criado")

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeCardCreated,
		Payload: dto.CartaoToResponse(cartao),
	})
	return cartao, nil
}

func (s *FinanceiroService) ListCartoes(ctx context.Context, tenantID string) ([]model.Cartao, error) {
	return s.cartaoRepo.ListByTenant(ctx, tenantID)
}

func (s *FinanceiroService) ListCartoesPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Cartao, int64, error) {
	return s.cartaoRepo.ListByTenantPaginated(ctx, tenantID, offset, limit)
}

func (s *FinanceiroService) DeleteCartao(ctx context.Context, tenantID, id string) error {
	if err := s.cartaoRepo.Delete(ctx, id, tenantID); err != nil {
		return err
	}

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeCardDeleted,
		Payload: map[string]string{"id": id},
	})
	return nil
}

// ── Gasto ───────────────────────────────────────────────────────────────────

func (s *FinanceiroService) CreateGasto(ctx context.Context, tenantID string, req *dto.CreateGastoRequest) (*dto.GastoResponse, error) {
	defaults := parseGastoDefaults(req)

	if defaults.splitMode == model.SplitModeCustom {
		if err := validateDivisoes(req.Divisoes, req.ValorTotalCentavos, ""); err != nil {
			return nil, err
		}
	}

	gasto := buildGastoModel(req, tenantID, defaults)

	if err := s.gastoRepo.Create(ctx, gasto); err != nil {
		return nil, err
	}

	s.createAuditLog(ctx, tenantID, req.CompradorID, "CRIAR_GASTO", "Gasto "+req.Descricao+" criado")

	resp := gastoToResponse(gasto)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeExpenseCreated,
		Payload: resp,
	})
	return resp, nil
}

func (s *FinanceiroService) UpdateGasto(ctx context.Context, tenantID string, gastoID string, req *dto.UpdateGastoRequest) (*dto.GastoResponse, error) {
	gasto, err := s.gastoRepo.GetByID(ctx, gastoID, tenantID)
	if err != nil {
		return nil, err
	}
	if gasto == nil {
		return nil, fmt.Errorf("gasto não encontrado")
	}

	applyGastoFieldUpdates(gasto, req)

	// Revalidate split sums when relevant fields change (total, split mode, or
	// explicit new divisoes). This prevents the sum invariant from being silently
	// broken when only ValorTotalCentavos or SplitMode is updated.
	needsDivisaoUpdate := req.ValorTotalCentavos != nil || req.SplitMode != nil || len(req.Divisoes) > 0
	if needsDivisaoUpdate {
		if err := s.updateGastoDivisoes(ctx, gasto, req, tenantID, gastoID); err != nil {
			return nil, err
		}
	} else {
		if err := s.gastoRepo.Update(ctx, gasto); err != nil {
			return nil, err
		}
	}

	s.createAuditLog(ctx, tenantID, gasto.CompradorID, "ATUALIZAR_GASTO", "Gasto "+gasto.Descricao+" atualizado")

	resp := gastoToResponse(gasto)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeExpenseUpdated,
		Payload: resp,
	})
	return resp, nil
}

// applyGastoFieldUpdates applies all scalar field updates from the request to the gasto model.
func applyGastoFieldUpdates(gasto *model.Gasto, req *dto.UpdateGastoRequest) {
	// Scalar fields — direct pointer dereference
	setIfNotNil(&gasto.Descricao, req.Descricao)
	setIfNotNil(&gasto.ValorTotalCentavos, req.ValorTotalCentavos)
	setIfNotNil(&gasto.CompradorID, req.CompradorID)
	setIfNotNil(&gasto.Installments, req.Installments)
	setIfNotNil(&gasto.TotalInstallments, req.TotalInstallments)
	setIfNotNil(&gasto.IsLoan, req.IsLoan)
	setIfNotNil(&gasto.Method, req.Method)
	setIfNotNil(&gasto.IsPrivate, req.IsPrivate)
	setIfNotNil(&gasto.IsSettlement, req.IsSettlement)

	// Optional foreign-key IDs — unpack NullString wrapper
	setNullStringIfNotNil(&gasto.FaturaID, req.FaturaID)
	setNullStringIfNotNil(&gasto.BorrowerID, req.BorrowerID)
	setNullStringIfNotNil(&gasto.CardOwnerID, req.CardOwnerID)
	setNullStringIfNotNil(&gasto.GrupoParcelasID, req.GrupoParcelasID)
	setNullStringIfNotNil(&gasto.RecurringBillID, req.RecurringBillID)

	// Fields needing type conversion
	if req.SettlementDetails != nil {
		gasto.SettlementDetails = rawMessageToString(req.SettlementDetails)
	}
	if req.SplitMode != nil && *req.SplitMode != "" {
		gasto.SplitMode = model.SplitMode(*req.SplitMode)
	}
}

// updateGastoDivisoes validates, replaces, and persists divisoes for a gasto update.
func (s *FinanceiroService) updateGastoDivisoes(ctx context.Context, gasto *model.Gasto, req *dto.UpdateGastoRequest, tenantID, gastoID string) error {
	divisoes := gasto.Divisoes
	if len(req.Divisoes) > 0 {
		divisoes = make([]model.DivisaoGasto, len(req.Divisoes))
		for i, d := range req.Divisoes {
			divisoes[i] = model.DivisaoGasto{
				ID:            uuid.New().String(),
				TenantID:      tenantID,
				GastoID:       gastoID,
				MembroID:      d.MembroID,
				ValorCentavos: d.ValorCentavos,
			}
		}
	}

	var sum int64
	for _, d := range divisoes {
		sum += d.ValorCentavos
	}
	if sum != gasto.ValorTotalCentavos {
		return fmt.Errorf("soma das divisoes (%d) nao confere com o valor total (%d)", sum, gasto.ValorTotalCentavos)
	}

	if len(req.Divisoes) > 0 {
		gasto.Divisoes = divisoes
		return s.persistDivisaoUpdate(ctx, gasto, gastoID, tenantID)
	}
	return s.gastoRepo.Update(ctx, gasto)
}

// persistDivisaoUpdate deletes old divisoes and saves new ones in a transaction
// when db is available, or falls back to non-transactional for tests.
func (s *FinanceiroService) persistDivisaoUpdate(ctx context.Context, gasto *model.Gasto, gastoID, tenantID string) error {
	if s.db != nil {
		return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txRepo := repository.NewGormGastoRepo(tx)
			if err := txRepo.DeleteDivisoes(ctx, gastoID, tenantID); err != nil {
				return err
			}
			return txRepo.Update(ctx, gasto)
		})
	}
	// Fallback for tests without a *gorm.DB: non-transactional.
	if err := s.gastoRepo.DeleteDivisoes(ctx, gastoID, tenantID); err != nil {
		return err
	}
	return s.gastoRepo.Update(ctx, gasto)
}

func (s *FinanceiroService) ListGastos(ctx context.Context, tenantID string) ([]dto.GastoResponse, error) {
	gastos, err := s.gastoRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.GastoResponse, len(gastos))
	for i, g := range gastos {
		resp[i] = *gastoToResponse(&g)
	}
	return resp, nil
}

func (s *FinanceiroService) ListGastosPaginated(ctx context.Context, tenantID string, offset, limit int) ([]dto.GastoResponse, int64, error) {
	gastos, total, err := s.gastoRepo.ListByTenantPaginated(ctx, tenantID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.GastoResponse, len(gastos))
	for i, g := range gastos {
		resp[i] = *gastoToResponse(&g)
	}
	return resp, total, nil
}

func (s *FinanceiroService) CreateGastoBatch(ctx context.Context, tenantID string, gastos []dto.CreateGastoRequest) error {
	type gastoInfo struct {
		ID          string
		Descricao   string
		CompradorID string
	}

	var created []gastoInfo

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range gastos {
			req := &gastos[i]

			defaults := parseGastoDefaults(req)

			if defaults.splitMode == model.SplitModeCustom {
				if err := validateDivisoes(req.Divisoes, req.ValorTotalCentavos, req.Descricao); err != nil {
					return err
				}
			}

			gasto := buildGastoModel(req, tenantID, defaults)

			if err := tx.Create(gasto).Error; err != nil {
				return err
			}

			created = append(created, gastoInfo{
				ID:          gasto.ID,
				Descricao:   req.Descricao,
				CompradorID: req.CompradorID,
			})
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, g := range created {
		s.createAuditLog(ctx, tenantID, g.CompradorID, "CRIAR_GASTO", "Gasto "+g.Descricao+" criado em lote")
	}

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeExpenseCreated,
		Payload: map[string]interface{}{"action": "batch-created"},
	})
	return nil
}

func (s *FinanceiroService) DeleteGasto(ctx context.Context, tenantID, id string) error {
	if err := s.gastoRepo.Delete(ctx, id, tenantID); err != nil {
		return err
	}

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeExpenseDeleted,
		Payload: map[string]string{"id": id},
	})
	return nil
}

func (s *FinanceiroService) DeleteGastoBatch(ctx context.Context, tenantID string, ids []string) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range ids {
			if err := tx.Where("gasto_id = ? AND tenant_id = ?", id, tenantID).Delete(&model.DivisaoGasto{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&model.Gasto{}, "id = ? AND tenant_id = ?", id, tenantID).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeExpenseDeleted,
		Payload: map[string]interface{}{"ids": ids, "action": "batch-deleted"},
	})
	return nil
}

// ── Conta Fixa ──────────────────────────────────────────────────────────────

func (s *FinanceiroService) CreateContaFixa(ctx context.Context, tenantID string, req *dto.CreateContaFixaRequest) (*dto.ContaFixaResponse, error) {
	conta := &model.ContaFixa{
		ID:                 uuid.New().String(),
		TenantID:           tenantID,
		Name:               req.Name,
		Icon:               req.Icon,
		FixedValueCentavos: req.FixedValueCentavos,
	}

	if len(req.DefaultSplit) > 0 {
		data, err := json.Marshal(req.DefaultSplit)
		if err != nil {
			return nil, err
		}
		conta.DefaultSplit = string(data)
	}

	if err := s.contaFixaRepo.Create(ctx, conta); err != nil {
		return nil, err
	}

	resp := contaFixaToResponse(conta)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeFixedBillCreated,
		Payload: resp,
	})
	return resp, nil
}

func (s *FinanceiroService) UpdateContaFixa(ctx context.Context, tenantID, contaID string, req *dto.UpdateContaFixaRequest) (*dto.ContaFixaResponse, error) {
	conta, err := s.contaFixaRepo.GetByID(ctx, contaID, tenantID)
	if err != nil {
		return nil, err
	}
	if conta == nil {
		return nil, ErrContaFixaNotFound
	}

	if req.Name != nil {
		conta.Name = *req.Name
	}
	if req.Icon != nil {
		conta.Icon = *req.Icon
	}
	if req.FixedValueCentavos != nil {
		v := *req.FixedValueCentavos
		conta.FixedValueCentavos = &v
	}
	if req.DefaultSplit != nil {
		data, err := json.Marshal(req.DefaultSplit)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar default split: %w", err)
		}
		conta.DefaultSplit = string(data)
	}

	if err := s.contaFixaRepo.Update(ctx, conta); err != nil {
		return nil, err
	}

	s.createAuditLog(ctx, tenantID, "", "ATUALIZAR_CONTA_FIXA", "Conta fixa "+conta.Name+" atualizada")

	resp := contaFixaToResponse(conta)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeFixedBillUpdated,
		Payload: resp,
	})
	return resp, nil
}

func (s *FinanceiroService) ListContasFixas(ctx context.Context, tenantID string) ([]dto.ContaFixaResponse, error) {
	contas, err := s.contaFixaRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.ContaFixaResponse, len(contas))
	for i, c := range contas {
		resp[i] = *contaFixaToResponse(&c)
	}
	return resp, nil
}

func (s *FinanceiroService) ListContasFixasPaginated(ctx context.Context, tenantID string, offset, limit int) ([]dto.ContaFixaResponse, int64, error) {
	contas, total, err := s.contaFixaRepo.ListByTenantPaginated(ctx, tenantID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.ContaFixaResponse, len(contas))
	for i, c := range contas {
		resp[i] = *contaFixaToResponse(&c)
	}
	return resp, total, nil
}

func (s *FinanceiroService) DeleteContaFixa(ctx context.Context, tenantID, id string) error {
	if err := s.contaFixaRepo.Delete(ctx, id, tenantID); err != nil {
		return err
	}

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeFixedBillDeleted,
		Payload: map[string]string{"id": id},
	})
	return nil
}

// ── Fatura ──────────────────────────────────────────────────────────────────

// parseDataPagamento converte uma string RFC3339Nano opcional para *time.Time.
func parseDataPagamento(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339Nano, *raw)
	if err != nil {
		return nil, fmt.Errorf("data_pagamento_banco inválida %q: %w", *raw, err)
	}
	return &t, nil
}

func buildFaturaModel(tenantID string, req *dto.CreateFaturaRequest, dataPgto *time.Time) *model.Fatura {
	return &model.Fatura{
		ID:                 uuid.New().String(),
		TenantID:           tenantID,
		CartaoID:           req.CartaoID,
		Mes:                req.Mes,
		Ano:                req.Ano,
		ResponsavelID:      req.ResponsavelID,
		Status:             req.Status,
		DataPagamentoBanco: dataPgto,
	}
}

func (s *FinanceiroService) CreateFatura(ctx context.Context, tenantID string, req *dto.CreateFaturaRequest) (*model.Fatura, error) {
	dataPgto, err := parseDataPagamento(req.DataPagamentoBanco)
	if err != nil {
		return nil, err
	}

	fatura := buildFaturaModel(tenantID, req, dataPgto)

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if createErr := s.faturaRepo.CreateOrUpdate(ctx, tx, fatura); createErr != nil {
			return createErr
		}
		// Re-query the fatura after upsert so the in-memory ID matches the DB row.
		// fatura.ID must be cleared before First() because GORM adds the primary
		// key as an implicit filter when the struct has a non-zero PK. If the
		// upsert performed an UPDATE (row already existed), the DB row's ID
		// differs from the newly generated UUID, and the implicit AND id = ?
		// would produce "record not found".
		fatura.ID = ""
		return tx.Where("cartao_id = ? AND mes = ? AND ano = ? AND tenant_id = ?",
			req.CartaoID, req.Mes, req.Ano, tenantID).First(fatura).Error
	})
	if err != nil {
		return nil, err
	}

	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeInvoiceUpdated,
		Payload: dto.FaturaToResponse(fatura),
	})
	return fatura, nil
}

func (s *FinanceiroService) CreateFaturaBatch(ctx context.Context, tenantID string, reqs []dto.CreateFaturaRequest) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range reqs {
			req := &reqs[i]

			dataPgto, err := parseDataPagamento(req.DataPagamentoBanco)
			if err != nil {
				return err
			}

			fatura := buildFaturaModel(tenantID, req, dataPgto)
			if err := s.faturaRepo.CreateOrUpdate(ctx, tx, fatura); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Notifica clients sobre faturas alteradas (batch)
	s.wsHub.Broadcast(tenantID, dto.WSMessage{
		Type:    dto.WSTypeInvoiceUpdated,
		Payload: map[string]interface{}{"action": "batch-updated"},
	})
	return nil
}

func (s *FinanceiroService) ListFaturas(ctx context.Context, tenantID string) ([]model.Fatura, error) {
	return s.faturaRepo.ListByTenant(ctx, tenantID)
}

func (s *FinanceiroService) ListFaturasPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Fatura, int64, error) {
	return s.faturaRepo.ListByTenantPaginated(ctx, tenantID, offset, limit)
}

// ── Audit & Validation ──────────────────────────────────────────────────────

func (s *FinanceiroService) RecordValidationEvent(ctx context.Context, tenantID string, req *dto.ValidateEventRequest) error {
	exists, err := s.validationRepo.ExistsByDedupeKey(ctx, tenantID, req.Type, req.DedupeKey)
	if err != nil {
		return err
	}
	if exists {
		return ErrEventAlreadyRegistered
	}

	event := &model.ProductValidationEvent{
		TenantID:  tenantID,
		Type:      model.ValidationEventType(req.Type),
		DedupeKey: req.DedupeKey,
	}
	if req.PeriodKey != "" {
		event.PeriodKey = &req.PeriodKey
	}

	return s.validationRepo.Create(ctx, event)
}

func (s *FinanceiroService) GetAuditLogs(ctx context.Context, tenantID string) ([]model.AuditLog, error) {
	return s.auditRepo.ListByTenant(ctx, tenantID)
}

func (s *FinanceiroService) GetAuditLogsPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.AuditLog, int64, error) {
	return s.auditRepo.ListByTenantPaginated(ctx, tenantID, offset, limit)
}

// ── Response mappers ────────────────────────────────────────────────────────

func membroToResponse(m *model.MembroCasa) *dto.MembroResponse {
	r := &dto.MembroResponse{
		ID:        m.ID,
		Nome:      m.Nome,
		Avatar:    m.Avatar,
		Ativo:     m.Ativo,
		Role:      string(m.Role),
		CreatedAt: m.CreatedAt.Format(time.RFC3339Nano),
	}

	if m.RendaCentavos != nil {
		r.RendaCentavos = m.RendaCentavos
	}
	if m.UserID != nil {
		r.UserID = *m.UserID
	}

	return r
}

func gastoToResponse(g *model.Gasto) *dto.GastoResponse {
	r := &dto.GastoResponse{
		ID:                 g.ID,
		Descricao:          g.Descricao,
		ValorTotalCentavos: g.ValorTotalCentavos,
		CompradorID:        g.CompradorID,
		FaturaID:           g.FaturaID,
		Installments:       g.Installments,
		TotalInstallments:  g.TotalInstallments,
		Method:             g.Method,
		IsLoan:             g.IsLoan,
		IsPrivate:          g.IsPrivate,
		BorrowerID:         g.BorrowerID,
		CardOwnerID:        g.CardOwnerID,
		RecurringBillID:    g.RecurringBillID,
		GrupoParcelasID:    g.GrupoParcelasID,
		IsSettlement:       g.IsSettlement,
		SettlementDetails:  stringToRawMessage(g.SettlementDetails),
		SplitMode:          string(g.SplitMode),
		CreatedAt:          g.CreatedAt.Format(time.RFC3339Nano),
	}

	for _, d := range g.Divisoes {
		r.Divisoes = append(r.Divisoes, dto.SplitItem{
			MembroID:      d.MembroID,
			ValorCentavos: d.ValorCentavos,
		})
	}

	return r
}

func contaFixaToResponse(c *model.ContaFixa) *dto.ContaFixaResponse {
	r := &dto.ContaFixaResponse{
		ID:                 c.ID,
		Name:               c.Name,
		Icon:               c.Icon,
		FixedValueCentavos: c.FixedValueCentavos,
		CreatedAt:          c.CreatedAt.Format(time.RFC3339Nano),
	}

	if c.DefaultSplit != "" && c.DefaultSplit != "[]" {
		var split []dto.SplitItem
		if err := json.Unmarshal([]byte(c.DefaultSplit), &split); err == nil {
			r.DefaultSplit = split
		}
	}

	if r.DefaultSplit == nil {
		r.DefaultSplit = []dto.SplitItem{}
	}

	return r
}
