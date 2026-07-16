package repository

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}
	return gormDB, mock
}

// models with default:gen_random_uuid() use Query, others use Exec
func expectInsertQuery(mock sqlmock.Sqlmock, table string) {
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "` + table + `"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("new-id"))
	mock.ExpectCommit()
}

func expectInsertExec(mock sqlmock.Sqlmock, table string) {
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "` + table + `"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
}

func expectUpdate(mock sqlmock.Sqlmock, table string) {
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "` + table + `"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

func expectDelete(mock sqlmock.Sqlmock, table string) {
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "` + table + `"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
}

// ---------- Tenant Repository ----------

func TestGormTenantRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormTenantRepo(db)

	expectInsertQuery(mock, "tenants")
	err := repo.Create(context.Background(), &model.Tenant{Name: "Test", InviteCode: "ABC123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestGormTenantRepo_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormTenantRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "invite_code"}).
		AddRow("t-1", "My House", "INV123")
	mock.ExpectQuery(`SELECT .+ FROM "tenants" WHERE`).WillReturnRows(rows)

	tenant, err := repo.GetByID(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant.ID != "t-1" || tenant.Name != "My House" {
		t.Errorf("got %+v", tenant)
	}
}

func TestGormTenantRepo_GetByInviteCode(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormTenantRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "invite_code"}).
		AddRow("t-1", "House", "XYZ789")
	mock.ExpectQuery(`SELECT .+ FROM "tenants" WHERE`).WillReturnRows(rows)

	tenant, err := repo.GetByInviteCode(context.Background(), "XYZ789")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant.InviteCode != "XYZ789" {
		t.Errorf("got %s", tenant.InviteCode)
	}
}

func TestGormTenantRepo_GetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormTenantRepo(db)

	mock.ExpectQuery(`SELECT .+ FROM "tenants" WHERE`).WillReturnError(gorm.ErrRecordNotFound)

	tenant, err := repo.GetByID(context.Background(), "nonexistent")
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
	if tenant != nil {
		t.Fatal("expected nil tenant, got non-nil")
	}
}

// ---------- Usuario Repository ----------

func TestGormUsuarioRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormUsuarioRepo(db)

	expectInsertQuery(mock, "usuarios")
	err := repo.Create(context.Background(), &model.Usuario{Email: "test@test.com", Nome: "User"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormUsuarioRepo_GetByEmail(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormUsuarioRepo(db)

	rows := sqlmock.NewRows([]string{"id", "email", "nome"}).AddRow("u-1", "test@test.com", "User")
	mock.ExpectQuery(`SELECT .+ FROM "usuarios" WHERE`).WillReturnRows(rows)

	user, err := repo.GetByEmail(context.Background(), "test@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "test@test.com" {
		t.Errorf("got %s", user.Email)
	}
}

func TestGormUsuarioRepo_GetByGoogleID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormUsuarioRepo(db)

	rows := sqlmock.NewRows([]string{"id", "email", "google_id"}).AddRow("u-2", "g@test.com", "gid-123")
	mock.ExpectQuery(`SELECT .+ FROM "usuarios" WHERE`).WillReturnRows(rows)

	user, err := repo.GetByGoogleID(context.Background(), "gid-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *user.GoogleID != "gid-123" {
		t.Errorf("got %v", user.GoogleID)
	}
}

func TestGormUsuarioRepo_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormUsuarioRepo(db)

	expectUpdate(mock, "usuarios")
	err := repo.Update(context.Background(), &model.Usuario{ID: "u-1", Email: "upd@test.com", Nome: "Upd"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormUsuarioRepo_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormUsuarioRepo(db)

	rows := sqlmock.NewRows([]string{"id", "email", "nome"}).AddRow("u-1", "test@test.com", "User")
	mock.ExpectQuery(`SELECT .+ FROM "usuarios" WHERE`).WillReturnRows(rows)

	user, err := repo.GetByID(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u-1" {
		t.Errorf("got %s", user.ID)
	}
}

// ---------- Membro Repository ----------

func TestGormMembroRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormMembroRepo(db)

	expectInsertExec(mock, "membros_casa")
	err := repo.Create(context.Background(), &model.MembroCasa{
		ID: "m-1", TenantID: "t-1", Nome: "John", Avatar: "🐶",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormMembroRepo_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormMembroRepo(db)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "nome", "avatar", "ativo", "role"}).
		AddRow("m-1", "t-1", "John", "🐶", true, "MORADOR")
	mock.ExpectQuery(`SELECT .+ FROM "membros_casa" WHERE`).WillReturnRows(rows)

	membro, err := repo.GetByID(context.Background(), "m-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if membro.Nome != "John" {
		t.Errorf("got %s", membro.Nome)
	}
}

func TestGormMembroRepo_ListByTenant(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormMembroRepo(db)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "nome"}).
		AddRow("m-1", "t-1", "John").
		AddRow("m-2", "t-1", "Jane")
	mock.ExpectQuery(`SELECT .+ FROM "membros_casa" WHERE`).WillReturnRows(rows)

	list, err := repo.ListByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestGormMembroRepo_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormMembroRepo(db)

	expectUpdate(mock, "membros_casa")
	err := repo.Update(context.Background(), &model.MembroCasa{
		ID: "m-1", TenantID: "t-1", Nome: "John Updated", Avatar: "🦊",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- Cartao Repository ----------

func TestGormCartaoRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormCartaoRepo(db)

	expectInsertExec(mock, "cartoes")
	err := repo.Create(context.Background(), &model.Cartao{
		ID: "c-1", TenantID: "t-1", Nome: "Nubank", DiaFechamento: 15, ResponsavelPadraoID: "m-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormCartaoRepo_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormCartaoRepo(db)

	expectDelete(mock, "cartoes")
	err := repo.Delete(context.Background(), "c-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormCartaoRepo_ListByTenant(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormCartaoRepo(db)

	rows := sqlmock.NewRows([]string{"id", "nome", "dia_fechamento"}).
		AddRow("c-1", "Nubank", 15).
		AddRow("c-2", "Inter", 10)
	mock.ExpectQuery(`SELECT .+ FROM "cartoes" WHERE`).WillReturnRows(rows)

	list, err := repo.ListByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

// ---------- Fatura Repository ----------

func TestGormFaturaRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormFaturaRepo(db)

	expectInsertExec(mock, "faturas")
	err := repo.Create(context.Background(), &model.Fatura{
		ID: "f-1", TenantID: "t-1", CartaoID: "c-1",
		Mes: 1, Ano: 2024, ResponsavelID: "m-1", Status: "ABERTA",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormFaturaRepo_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormFaturaRepo(db)

	rows := sqlmock.NewRows([]string{"id", "cartao_id", "mes", "ano", "status"}).
		AddRow("f-1", "c-1", 1, 2024, "ABERTA")
	mock.ExpectQuery(`SELECT .+ FROM "faturas" WHERE`).WillReturnRows(rows)

	fatura, err := repo.GetByID(context.Background(), "f-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fatura.Status != "ABERTA" {
		t.Errorf("got %s", fatura.Status)
	}
}

func TestGormFaturaRepo_ListByTenant(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormFaturaRepo(db)

	rows := sqlmock.NewRows([]string{"id", "cartao_id", "mes", "ano", "status"}).
		AddRow("f-1", "c-1", 1, 2024, "ABERTA").
		AddRow("f-2", "c-1", 2, 2024, "FECHADA")
	mock.ExpectQuery(`SELECT .+ FROM "faturas" WHERE`).WillReturnRows(rows)

	list, err := repo.ListByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestGormFaturaRepo_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormFaturaRepo(db)

	expectUpdate(mock, "faturas")
	err := repo.Update(context.Background(), &model.Fatura{
		ID: "f-1", TenantID: "t-1", CartaoID: "c-1",
		Mes: 1, Ano: 2024, ResponsavelID: "m-1", Status: "FECHADA",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- Gasto Repository ----------

func TestGormGastoRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormGastoRepo(db)

	expectInsertExec(mock, "gastos")
	err := repo.Create(context.Background(), &model.Gasto{
		ID: "g-1", TenantID: "t-1", Descricao: "Mercado",
		ValorTotalCentavos: 5000, CompradorID: "m-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormGastoRepo_ListByTenant(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormGastoRepo(db)

	mock.ExpectQuery(`SELECT .+ FROM "gastos" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "descricao"}).
			AddRow("g-1", "Mercado").
			AddRow("g-2", "Uber"))
	mock.ExpectQuery(`SELECT .+ FROM "divisoes_gasto" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gasto_id", "membro_id", "valor_centavos"}))

	list, err := repo.ListByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestGormGastoRepo_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormGastoRepo(db)

	rows := sqlmock.NewRows([]string{"id", "descricao", "valor_total_centavos"}).
		AddRow("g-1", "Mercado", 5000)
	mock.ExpectQuery(`SELECT .+ FROM "gastos" WHERE`).WillReturnRows(rows)
	mock.ExpectQuery(`SELECT .+ FROM "divisoes_gasto" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gasto_id", "membro_id", "valor_centavos"}))

	gasto, err := repo.GetByID(context.Background(), "g-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gasto.Descricao != "Mercado" {
		t.Errorf("got %s", gasto.Descricao)
	}
}

func TestGormGastoRepo_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormGastoRepo(db)

	expectUpdate(mock, "gastos")
	err := repo.Update(context.Background(), &model.Gasto{
		ID: "g-1", TenantID: "t-1", Descricao: "Mercado Updated",
		ValorTotalCentavos: 6000, CompradorID: "m-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormGastoRepo_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormGastoRepo(db)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "divisoes_gasto"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`DELETE FROM "gastos"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := repo.Delete(context.Background(), "g-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- ContaFixa Repository ----------

func TestGormContaFixaRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormContaFixaRepo(db)

	expectInsertExec(mock, "contas_fixas")
	err := repo.Create(context.Background(), &model.ContaFixa{
		ID: "cf-1", TenantID: "t-1", Name: "Internet", Icon: "wifi",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormContaFixaRepo_Update(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormContaFixaRepo(db)

	expectUpdate(mock, "contas_fixas")
	err := repo.Update(context.Background(), &model.ContaFixa{
		ID: "cf-1", TenantID: "t-1", Name: "Internet Updated", Icon: "wifi",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormContaFixaRepo_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormContaFixaRepo(db)

	expectDelete(mock, "contas_fixas")
	err := repo.Delete(context.Background(), "cf-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormContaFixaRepo_ListByTenant(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormContaFixaRepo(db)

	rows := sqlmock.NewRows([]string{"id", "name", "icon"}).
		AddRow("cf-1", "Internet", "wifi").
		AddRow("cf-2", "Agua", "💧")
	mock.ExpectQuery(`SELECT .+ FROM "contas_fixas" WHERE`).WillReturnRows(rows)

	list, err := repo.ListByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

// ---------- AuditLog Repository ----------

func TestGormAuditLogRepo_ListByTenant(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormAuditLogRepo(db)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "acao"}).
		AddRow("a-1", "t-1", "MEMBRO_CREATED").
		AddRow("a-2", "t-1", "GASTO_CREATED")
	mock.ExpectQuery(`SELECT .+ FROM "audit_logs" WHERE`).WillReturnRows(rows)

	list, err := repo.ListByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}

func TestGormAuditLogRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormAuditLogRepo(db)

	expectInsertQuery(mock, "audit_logs")
	err := repo.Create(context.Background(), &model.AuditLog{
		TenantID: "t-1", MembroID: "m-1",
		Acao: "CRIAR_GASTO", Detalhes: `{"gasto_id":"g-1"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- ProductValidation Repository ----------

func TestGormProductValidationRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormProductValidationRepo(db)

	expectInsertQuery(mock, "product_validation_events")
	err := repo.Create(context.Background(), &model.ProductValidationEvent{
		TenantID:  "t-1",
		Type:      "TENANT_CREATED",
		DedupeKey: "key-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormProductValidationRepo_ExistsByDedupeKey_True(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormProductValidationRepo(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "product_validation_events" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.ExistsByDedupeKey(context.Background(), "t-1", "TENANT_CREATED", "key-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true")
	}
}

func TestGormProductValidationRepo_ExistsByDedupeKey_False(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormProductValidationRepo(db)

	mock.ExpectQuery(`SELECT count\(\*\) FROM "product_validation_events" WHERE`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	exists, err := repo.ExistsByDedupeKey(context.Background(), "t-1", "TENANT_CREATED", "key-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false")
	}
}

// ---------- PasswordResetToken Repository ----------

func TestGormPasswordResetTokenRepo_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormPasswordResetTokenRepo(db)

	expectInsertQuery(mock, "password_reset_tokens")
	err := repo.Create(context.Background(), &model.PasswordResetToken{
		Token: "reset-token", UserID: "u-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGormPasswordResetTokenRepo_GetByToken(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormPasswordResetTokenRepo(db)

	rows := sqlmock.NewRows([]string{"id", "token", "user_id"}).
		AddRow("prt-1", "reset-token", "u-1")
	mock.ExpectQuery(`SELECT .+ FROM "password_reset_tokens" WHERE`).WillReturnRows(rows)

	tok, err := repo.GetByToken(context.Background(), "reset-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.Token != "reset-token" {
		t.Errorf("got %s", tok.Token)
	}
}

func TestGormPasswordResetTokenRepo_Delete(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewGormPasswordResetTokenRepo(db)

	expectDelete(mock, "password_reset_tokens")
	err := repo.Delete(context.Background(), "prt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------- Interface compliance (compile-time checks) ----------

func TestInterfaceCompliance(t *testing.T) {
	var _ TenantRepository = (*GormTenantRepo)(nil)
	var _ UsuarioRepository = (*GormUsuarioRepo)(nil)
	var _ MembroRepository = (*GormMembroRepo)(nil)
	var _ CartaoRepository = (*GormCartaoRepo)(nil)
	var _ FaturaRepository = (*GormFaturaRepo)(nil)
	var _ GastoRepository = (*GormGastoRepo)(nil)
	var _ ContaFixaRepository = (*GormContaFixaRepo)(nil)
	var _ AuditLogRepository = (*GormAuditLogRepo)(nil)
	var _ ProductValidationRepository = (*GormProductValidationRepo)(nil)
	var _ PasswordResetTokenRepository = (*GormPasswordResetTokenRepo)(nil)
}

// ---------- Constructor tests ----------

func TestNewGormRepoConstructors(t *testing.T) {
	db, _ := setupMockDB(t)
	tests := []struct {
		name string
		repo interface{}
	}{
		{"TenantRepo", NewGormTenantRepo(db)},
		{"UsuarioRepo", NewGormUsuarioRepo(db)},
		{"MembroRepo", NewGormMembroRepo(db)},
		{"CartaoRepo", NewGormCartaoRepo(db)},
		{"FaturaRepo", NewGormFaturaRepo(db)},
		{"GastoRepo", NewGormGastoRepo(db)},
		{"ContaFixaRepo", NewGormContaFixaRepo(db)},
		{"AuditLogRepo", NewGormAuditLogRepo(db)},
		{"ProductValidationRepo", NewGormProductValidationRepo(db)},
		{"PasswordResetTokenRepo", NewGormPasswordResetTokenRepo(db)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.repo == nil {
				t.Fatal("expected non-nil repo")
			}
		})
	}
}
