package service

import (
	"context"

	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/repository"
	"gorm.io/gorm"
)

type mockFaturaRepo struct {
	repository.FaturaRepository
	faturas map[string]*model.Fatura
}

func (m *mockFaturaRepo) Create(ctx context.Context, f *model.Fatura) error {
	if m.faturas == nil {
		m.faturas = make(map[string]*model.Fatura)
	}
	m.faturas[f.ID] = f
	return nil
}

func (m *mockFaturaRepo) CreateOrUpdate(ctx context.Context, tx *gorm.DB, f *model.Fatura) error {
	// Delegate to Create for mock simplicity — tests that need upsert semantics
	// should override this method.
	if m.faturas == nil {
		m.faturas = make(map[string]*model.Fatura)
	}
	m.faturas[f.ID] = f
	return nil
}
