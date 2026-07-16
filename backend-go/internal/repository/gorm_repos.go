package repository

import (
	"context"
	"errors"

	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrRecordNotFound = errors.New("record not found")

type GormTenantRepo struct{ db *gorm.DB }

func NewGormTenantRepo(db *gorm.DB) TenantRepository {
	return &GormTenantRepo{db: db}
}

func (r *GormTenantRepo) Create(ctx context.Context, t *model.Tenant) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *GormTenantRepo) GetByID(ctx context.Context, id string) (*model.Tenant, error) {
	var t model.Tenant
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &t, err
}

func (r *GormTenantRepo) GetByInviteCode(ctx context.Context, code string) (*model.Tenant, error) {
	var t model.Tenant
	err := r.db.WithContext(ctx).First(&t, "invite_code = ?", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &t, err
}

func (r *GormTenantRepo) GetByIDs(ctx context.Context, ids []string) ([]model.Tenant, error) {
	if len(ids) == 0 {
		return []model.Tenant{}, nil
	}
	var list []model.Tenant
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (r *GormTenantRepo) Update(ctx context.Context, t *model.Tenant) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *GormTenantRepo) ListAll(ctx context.Context) ([]model.Tenant, error) {
	var list []model.Tenant
	err := r.db.WithContext(ctx).Find(&list).Error
	return list, err
}

type GormUsuarioRepo struct{ db *gorm.DB }

func NewGormUsuarioRepo(db *gorm.DB) UsuarioRepository {
	return &GormUsuarioRepo{db: db}
}

func (r *GormUsuarioRepo) Create(ctx context.Context, u *model.Usuario) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *GormUsuarioRepo) GetByID(ctx context.Context, id string) (*model.Usuario, error) {
	var u model.Usuario
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &u, err
}

func (r *GormUsuarioRepo) GetByEmail(ctx context.Context, email string) (*model.Usuario, error) {
	var u model.Usuario
	err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &u, err
}

func (r *GormUsuarioRepo) GetByGoogleID(ctx context.Context, googleID string) (*model.Usuario, error) {
	var u model.Usuario
	err := r.db.WithContext(ctx).First(&u, "google_id = ?", googleID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &u, err
}

func (r *GormUsuarioRepo) Update(ctx context.Context, u *model.Usuario) error {
	return r.db.WithContext(ctx).Updates(u).Error
}

type GormMembroRepo struct{ db *gorm.DB }

func NewGormMembroRepo(db *gorm.DB) MembroRepository {
	return &GormMembroRepo{db: db}
}

func (r *GormMembroRepo) Create(ctx context.Context, m *model.MembroCasa) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *GormMembroRepo) GetByID(ctx context.Context, id, tenantID string) (*model.MembroCasa, error) {
	var m model.MembroCasa
	err := r.db.WithContext(ctx).First(&m, "id = ? AND tenant_id = ?", id, tenantID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &m, err
}

func (r *GormMembroRepo) GetByUserID(ctx context.Context, tenantID, userID string) (*model.MembroCasa, error) {
	var m model.MembroCasa
	err := r.db.WithContext(ctx).First(&m, "tenant_id = ? AND user_id = ?", tenantID, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &m, err
}

func (r *GormMembroRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.MembroCasa, error) {
	var list []model.MembroCasa
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&list).Error
	return list, err
}

func (r *GormMembroRepo) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.MembroCasa, int64, error) {
	var total int64
	var list []model.MembroCasa
	db := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if err := db.Model(&model.MembroCasa{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *GormMembroRepo) ListByUserID(ctx context.Context, userID string) ([]model.MembroCasa, error) {
	var list []model.MembroCasa
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&list).Error
	return list, err
}

func (r *GormMembroRepo) Update(ctx context.Context, m *model.MembroCasa) error {
	return r.db.WithContext(ctx).Save(m).Error
}

type GormCartaoRepo struct{ db *gorm.DB }

func NewGormCartaoRepo(db *gorm.DB) CartaoRepository {
	return &GormCartaoRepo{db: db}
}

func (r *GormCartaoRepo) Create(ctx context.Context, c *model.Cartao) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *GormCartaoRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.Cartao, error) {
	var list []model.Cartao
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&list).Error
	return list, err
}

func (r *GormCartaoRepo) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Cartao, int64, error) {
	var total int64
	var list []model.Cartao
	db := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if err := db.Model(&model.Cartao{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *GormCartaoRepo) Delete(ctx context.Context, id, tenantID string) error {
	return r.db.WithContext(ctx).Delete(&model.Cartao{}, "id = ? AND tenant_id = ?", id, tenantID).Error
}

type GormFaturaRepo struct{ db *gorm.DB }

func NewGormFaturaRepo(db *gorm.DB) FaturaRepository {
	return &GormFaturaRepo{db: db}
}

func (r *GormFaturaRepo) Create(ctx context.Context, f *model.Fatura) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *GormFaturaRepo) CreateOrUpdate(ctx context.Context, tx *gorm.DB, f *model.Fatura) error {
	if tx == nil {
		tx = r.db
	}
	return tx.WithContext(ctx).Clauses(clause.OnConflict{
		OnConstraint: "faturas_tenant_cartao_mes_ano_key",
		DoUpdates:    clause.AssignmentColumns([]string{"status", "responsavel_id", "data_pagamento_banco"}),
	}).Create(f).Error
}

func (r *GormFaturaRepo) GetByID(ctx context.Context, id, tenantID string) (*model.Fatura, error) {
	var f model.Fatura
	err := r.db.WithContext(ctx).First(&f, "id = ? AND tenant_id = ?", id, tenantID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &f, err
}

func (r *GormFaturaRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.Fatura, error) {
	var list []model.Fatura
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&list).Error
	return list, err
}

func (r *GormFaturaRepo) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Fatura, int64, error) {
	var total int64
	var list []model.Fatura
	db := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if err := db.Model(&model.Fatura{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *GormFaturaRepo) Update(ctx context.Context, f *model.Fatura) error {
	return r.db.WithContext(ctx).Save(f).Error
}

type GormGastoRepo struct{ db *gorm.DB }

func NewGormGastoRepo(db *gorm.DB) GastoRepository {
	return &GormGastoRepo{db: db}
}

func (r *GormGastoRepo) Create(ctx context.Context, g *model.Gasto) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *GormGastoRepo) GetByID(ctx context.Context, id, tenantID string) (*model.Gasto, error) {
	var g model.Gasto
	err := r.db.WithContext(ctx).Preload("Divisoes").First(&g, "gastos.id = ? AND gastos.tenant_id = ?", id, tenantID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &g, err
}

func (r *GormGastoRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.Gasto, error) {
	var list []model.Gasto
	err := r.db.WithContext(ctx).Preload("Divisoes").Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *GormGastoRepo) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.Gasto, int64, error) {
	var total int64
	var list []model.Gasto
	db := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if err := db.Model(&model.Gasto{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Preload("Divisoes").Offset(offset).Limit(limit).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (r *GormGastoRepo) Update(ctx context.Context, g *model.Gasto) error {
	return r.db.WithContext(ctx).Save(g).Error
}

func (r *GormGastoRepo) Delete(ctx context.Context, id, tenantID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("gasto_id = ? AND tenant_id = ?", id, tenantID).Delete(&model.DivisaoGasto{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Gasto{}, "id = ? AND tenant_id = ?", id, tenantID).Error
	})
}

func (r *GormGastoRepo) DeleteDivisoes(ctx context.Context, gastoID, tenantID string) error {
	return r.db.WithContext(ctx).
		Where("gasto_id = ? AND tenant_id = ?", gastoID, tenantID).
		Delete(&model.DivisaoGasto{}).Error
}

type GormContaFixaRepo struct{ db *gorm.DB }

func NewGormContaFixaRepo(db *gorm.DB) ContaFixaRepository {
	return &GormContaFixaRepo{db: db}
}

func (r *GormContaFixaRepo) Create(ctx context.Context, c *model.ContaFixa) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *GormContaFixaRepo) GetByID(ctx context.Context, id, tenantID string) (*model.ContaFixa, error) {
	var c model.ContaFixa
	err := r.db.WithContext(ctx).First(&c, "id = ? AND tenant_id = ?", id, tenantID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

func (r *GormContaFixaRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.ContaFixa, error) {
	var list []model.ContaFixa
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&list).Error
	return list, err
}

func (r *GormContaFixaRepo) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.ContaFixa, int64, error) {
	var total int64
	var list []model.ContaFixa
	db := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if err := db.Model(&model.ContaFixa{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *GormContaFixaRepo) Update(ctx context.Context, c *model.ContaFixa) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *GormContaFixaRepo) Delete(ctx context.Context, id, tenantID string) error {
	return r.db.WithContext(ctx).Delete(&model.ContaFixa{}, "id = ? AND tenant_id = ?", id, tenantID).Error
}

type GormAuditLogRepo struct{ db *gorm.DB }

func NewGormAuditLogRepo(db *gorm.DB) AuditLogRepository {
	return &GormAuditLogRepo{db: db}
}

func (r *GormAuditLogRepo) Create(ctx context.Context, l *model.AuditLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *GormAuditLogRepo) ListByTenant(ctx context.Context, tenantID string) ([]model.AuditLog, error) {
	var list []model.AuditLog
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (r *GormAuditLogRepo) ListByTenantPaginated(ctx context.Context, tenantID string, offset, limit int) ([]model.AuditLog, int64, error) {
	var total int64
	var list []model.AuditLog
	db := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if err := db.Model(&model.AuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

type GormProductValidationRepo struct{ db *gorm.DB }

func NewGormProductValidationRepo(db *gorm.DB) ProductValidationRepository {
	return &GormProductValidationRepo{db: db}
}

func (r *GormProductValidationRepo) Create(ctx context.Context, e *model.ProductValidationEvent) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *GormProductValidationRepo) ExistsByDedupeKey(ctx context.Context, tenantID, eventType, dedupeKey string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProductValidationEvent{}).
		Where("tenant_id = ? AND type = ? AND dedupe_key = ?", tenantID, eventType, dedupeKey).
		Count(&count).Error
	return count > 0, err
}

type GormPasswordResetTokenRepo struct{ db *gorm.DB }

func NewGormPasswordResetTokenRepo(db *gorm.DB) PasswordResetTokenRepository {
	return &GormPasswordResetTokenRepo{db: db}
}

func (r *GormPasswordResetTokenRepo) Create(ctx context.Context, t *model.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *GormPasswordResetTokenRepo) GetByToken(ctx context.Context, token string) (*model.PasswordResetToken, error) {
	var t model.PasswordResetToken
	err := r.db.WithContext(ctx).First(&t, "token = ?", token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRecordNotFound
	}
	return &t, err
}

func (r *GormPasswordResetTokenRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.PasswordResetToken{}, "id = ?", id).Error
}
