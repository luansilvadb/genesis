package repository

import (
	"context"

	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *model.Tenant) error
	GetByID(ctx context.Context, id string) (*model.Tenant, error)
	GetByIDs(ctx context.Context, ids []string) ([]model.Tenant, error)
	GetByInviteCode(ctx context.Context, code string) (*model.Tenant, error)
	Update(ctx context.Context, tenant *model.Tenant) error
	ListAll(ctx context.Context) ([]model.Tenant, error)
}

type UsuarioRepository interface {
	Create(ctx context.Context, user *model.Usuario) error
	GetByID(ctx context.Context, id string) (*model.Usuario, error)
	GetByEmail(ctx context.Context, email string) (*model.Usuario, error)
	GetByGoogleID(ctx context.Context, googleID string) (*model.Usuario, error)
	Update(ctx context.Context, user *model.Usuario) error
}

type MembroRepository interface {
	Create(ctx context.Context, membro *model.MembroCasa) error
	GetByID(ctx context.Context, id, tenantID string) (*model.MembroCasa, error)
	GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error)
	ListByTenant(ctx context.Context, tenantID string) ([]model.MembroCasa, error)
	ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.MembroCasa, int64, error)
	ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error)
	Update(ctx context.Context, membro *model.MembroCasa) error
}

type CartaoRepository interface {
	Create(ctx context.Context, cartao *model.Cartao) error
	ListByTenant(ctx context.Context, tenantID string) ([]model.Cartao, error)
	ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Cartao, int64, error)
	Delete(ctx context.Context, id, tenantID string) error
}

type FaturaRepository interface {
	Create(ctx context.Context, fatura *model.Fatura) error
	CreateOrUpdate(ctx context.Context, tx *gorm.DB, fatura *model.Fatura) error
	GetByID(ctx context.Context, id, tenantID string) (*model.Fatura, error)
	ListByTenant(ctx context.Context, tenantID string) ([]model.Fatura, error)
	ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Fatura, int64, error)
	Update(ctx context.Context, fatura *model.Fatura) error
}

type GastoRepository interface {
	Create(ctx context.Context, gasto *model.Gasto) error
	GetByID(ctx context.Context, id, tenantID string) (*model.Gasto, error)
	ListByTenant(ctx context.Context, tenantID string) ([]model.Gasto, error)
	ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Gasto, int64, error)
	Update(ctx context.Context, gasto *model.Gasto) error
	Delete(ctx context.Context, id, tenantID string) error
	DeleteDivisoes(ctx context.Context, gastoID, tenantID string) error
}

type ContaFixaRepository interface {
	Create(ctx context.Context, conta *model.ContaFixa) error
	GetByID(ctx context.Context, id, tenantID string) (*model.ContaFixa, error)
	ListByTenant(ctx context.Context, tenantID string) ([]model.ContaFixa, error)
	ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.ContaFixa, int64, error)
	Update(ctx context.Context, conta *model.ContaFixa) error
	Delete(ctx context.Context, id, tenantID string) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
	ListByTenant(ctx context.Context, tenantID string) ([]model.AuditLog, error)
	ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.AuditLog, int64, error)
}

type ProductValidationRepository interface {
	Create(ctx context.Context, event *model.ProductValidationEvent) error
	ExistsByDedupeKey(ctx context.Context, tenantID, eventType, dedupeKey string) (bool, error)
}

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, token *model.PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error)
	Delete(ctx context.Context, id string) error
}
