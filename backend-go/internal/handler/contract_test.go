package handler

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/dto"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
)

// ── Cartao Contract Tests ─────────────────────────────────────────────────

func TestContract_CartaoResponse_NoInternalFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Este teste valida o formato esperado da resposta de cartão
	// usando o DTO diretamente (sem mock do service).

	cartao := &model.Cartao{
		ID:                  "test-123",
		TenantID:            "tenant-interno",
		Nome:                "Cartão Teste",
		DiaFechamento:       15,
		ResponsavelPadraoID: "membro-456",
	}

	resp := dto.CartaoToResponse(cartao)
	body, _ := json.Marshal(resp)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("erro ao deserializar resposta: %v", err)
	}

	// Deve conter apenas os campos públicos
	if _, ok := data["tenantId"]; ok {
		t.Error("CartaoResponse não deve expor tenantId")
	}
	if _, ok := data["createdAt"]; ok {
		t.Error("CartaoResponse não deve expor createdAt")
	}

	// Deve conter os campos esperados
	required := []string{"id", "nome", "diaFechamento", "responsavelPadraoId"}
	for _, f := range required {
		if _, ok := data[f]; !ok {
			t.Errorf("CartaoResponse deve conter o campo '%s'", f)
		}
	}
}

func TestContract_FaturaResponse_NoInternalFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	fatura := &model.Fatura{
		ID:            "fat-123",
		TenantID:      "tenant-interno",
		CartaoID:      "cartao-456",
		Mes:           6,
		Ano:           2026,
		ResponsavelID: "membro-789",
		Status:        "ABERTA",
	}

	resp := dto.FaturaToResponse(fatura)
	body, _ := json.Marshal(resp)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("erro ao deserializar resposta: %v", err)
	}

	// Não deve expor campos internos
	if _, ok := data["tenantId"]; ok {
		t.Error("FaturaResponse não deve expor tenantId")
	}
	if _, ok := data["createdAt"]; ok {
		t.Error("FaturaResponse não deve expor createdAt")
	}

	// Deve conter os campos esperados
	required := []string{"id", "cartaoId", "mes", "ano", "responsavelId", "status"}
	for _, f := range required {
		if _, ok := data[f]; !ok {
			t.Errorf("FaturaResponse deve conter o campo '%s'", f)
		}
	}
}

func TestContract_CartoesToResponse_EmptyList(t *testing.T) {
	resp := dto.CartoesToResponse([]model.Cartao{})
	if resp == nil {
		t.Error("CartoesToResponse deveria retornar slice vazio, não nil")
	}
	if len(resp) != 0 {
		t.Errorf("esperado slice vazio, obtido %d itens", len(resp))
	}
}

func TestContract_FaturasToResponse_EmptyList(t *testing.T) {
	resp := dto.FaturasToResponse([]model.Fatura{})
	if resp == nil {
		t.Error("FaturasToResponse deveria retornar slice vazio, não nil")
	}
	if len(resp) != 0 {
		t.Errorf("esperado slice vazio, obtido %d itens", len(resp))
	}
}

func TestContract_FaturaToResponse_NilDataPagamento(t *testing.T) {
	fatura := &model.Fatura{
		ID:            "fat-nil",
		CartaoID:      "cart-1",
		Mes:           1,
		Ano:           2026,
		ResponsavelID: "memb-1",
		Status:        "ABERTA",
		// DataPagamentoBanco is nil
	}

	resp := dto.FaturaToResponse(fatura)
	body, _ := json.Marshal(resp)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("erro ao deserializar resposta: %v", err)
	}

	if _, ok := data["dataPagamentoBanco"]; ok {
		t.Error("dataPagamentoBanco não deve aparecer quando é nil")
	}
}

// ── Response Format Contract Tests ────────────────────────────────────────

func TestContract_ListEndpoints_ReturnDTOs(t *testing.T) {
	// Valida que todos os DTOs de lista existem e são serializáveis
	// Esta é uma validação estática do contrato

	gin.SetMode(gin.TestMode)

	// Simula o que um handler faria ao retornar lista de cartões
	cartoes := []model.Cartao{
		{ID: "c1", Nome: "Visa", DiaFechamento: 10, ResponsavelPadraoID: "m1"},
	}
	resp := dto.CartoesToResponse(cartoes)
	body, _ := json.Marshal(resp)

	// Verifica que é um array JSON
	if len(body) == 0 || body[0] != '[' {
		t.Error("CartoesToResponse deve serializar como array JSON")
	}

	// Simula resposta paginada
	paginated := dto.NewPaginatedResponse(resp, 1, dto.PaginationParams{Page: 1, PageSize: 20})
	pagBody, _ := json.Marshal(paginated)

	var pagData map[string]interface{}
	if err := json.Unmarshal(pagBody, &pagData); err != nil {
		t.Fatalf("erro ao deserializar resposta paginada: %v", err)
	}

	for _, f := range []string{"data", "total", "page", "page_size", "totalPages"} {
		if _, ok := pagData[f]; !ok {
			t.Errorf("PaginatedResponse deve conter '%s'", f)
		}
	}
}

func TestContract_GastoResponse_SplitModeValues(t *testing.T) {
	// Valida que os valores de SplitMode são os esperados pelo frontend
	// Frontend espera: EQUAL, INCOME, CUSTOM (uppercase)
	if model.SplitModeEqual != "EQUAL" {
		t.Errorf("SplitModeEqual deve ser 'EQUAL', obtido '%s'", model.SplitModeEqual)
	}
	if model.SplitModeIncome != "INCOME" {
		t.Errorf("SplitModeIncome deve ser 'INCOME', obtido '%s'", model.SplitModeIncome)
	}
	if model.SplitModeCustom != "CUSTOM" {
		t.Errorf("SplitModeCustom deve ser 'CUSTOM', obtido '%s'", model.SplitModeCustom)
	}
}

func TestContract_WSMessageTypes_AreCorrect(t *testing.T) {
	// Valida que as constantes WS são exportadas corretamente
	types := map[string]string{
		dto.WSTypeExpenseCreated:    "EXPENSE_CREATED",
		dto.WSTypeExpenseUpdated:    "EXPENSE_UPDATED",
		dto.WSTypeExpenseDeleted:    "EXPENSE_DELETED",
		dto.WSTypeCardCreated:       "CARD_CREATED",
		dto.WSTypeCardDeleted:       "CARD_DELETED",
		dto.WSTypeInvoiceUpdated:    "INVOICE_UPDATED",
		dto.WSTypeMemberCreated:     "MEMBER_CREATED",
		dto.WSTypeMemberUpdated:     "MEMBER_UPDATED",
		dto.WSTypeFixedBillCreated:  "FIXED_BILL_CREATED",
		dto.WSTypeFixedBillDeleted:  "FIXED_BILL_DELETED",
		dto.WSTypePermissionsUpdate: "PERMISSIONS_UPDATED",
	}

	for constant, expected := range types {
		if constant != expected {
			t.Errorf("constante WS '%s' deveria ser '%s', mas é '%s'", constant, expected, constant)
		}
	}
}
