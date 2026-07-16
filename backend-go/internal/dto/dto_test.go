package dto

import (
	"encoding/json"
	"testing"
)

func TestRegisterRequest_JSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    RegisterRequest
		wantErr bool
	}{
		{
			name:  "full request",
			input: `{"email":"test@example.com","nome":"User","password":"pass123"}`,
			want:  RegisterRequest{Email: "test@example.com", Nome: "User", Password: "pass123"},
		},
		{
			name:    "empty json",
			input:   `{}`,
			want:    RegisterRequest{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got RegisterRequest
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}

			data, err := json.Marshal(tt.want)
			if err != nil {
				t.Fatalf("Marshal error = %v", err)
			}
			var roundTrip RegisterRequest
			if err := json.Unmarshal(data, &roundTrip); err != nil {
				t.Fatalf("Round-trip Unmarshal error = %v", err)
			}
			if roundTrip != tt.want {
				t.Errorf("round-trip: got %+v, want %+v", roundTrip, tt.want)
			}
		})
	}
}

func TestLoginRequest_JSON(t *testing.T) {
	req := LoginRequest{Email: "a@b.com", Password: "secret"}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	var got LoginRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Email != req.Email || got.Password != req.Password {
		t.Errorf("round-trip: got %+v, want %+v", got, req)
	}
}

func TestGoogleLoginRequest_JSON(t *testing.T) {
	req := GoogleLoginRequest{Credential: "tok123"}
	data, _ := json.Marshal(req)
	var got GoogleLoginRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Credential != "tok123" {
		t.Errorf("got credential %q, want %q", got.Credential, "tok123")
	}
}

func TestForgotPasswordRequest_JSON(t *testing.T) {
	req := ForgotPasswordRequest{Email: "user@test.com"}
	data, _ := json.Marshal(req)
	var got ForgotPasswordRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Email != "user@test.com" {
		t.Errorf("got email %q, want %q", got.Email, "user@test.com")
	}
}

func TestResetPasswordRequest_JSON(t *testing.T) {
	req := ResetPasswordRequest{Token: "tok", Password: "newpass"}
	data, _ := json.Marshal(req)
	var got ResetPasswordRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Token != "tok" || got.Password != "newpass" {
		t.Errorf("round-trip: got %+v, want %+v", got, req)
	}
}

func TestAuthResponse_JSON(t *testing.T) {
	resp := AuthResponse{
		Token: "jwt.token.here",
		User:  UserProfile{ID: "1", Email: "e@e.com", Nome: "N"},
	}
	data, _ := json.Marshal(resp)
	var got AuthResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Token != resp.Token || got.User.Email != resp.User.Email {
		t.Errorf("round-trip: got %+v, want %+v", got, resp)
	}
}

func TestUserProfile_JSON(t *testing.T) {
	p := UserProfile{ID: "uuid", Email: "x@y.com", Nome: "Name"}
	data, _ := json.Marshal(p)
	var got UserProfile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != p {
		t.Errorf("round-trip: got %+v, want %+v", got, p)
	}
}

func TestCreateTenantRequest_JSON(t *testing.T) {
	req := CreateTenantRequest{Name: "My House"}
	data, _ := json.Marshal(req)
	var got CreateTenantRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "My House" {
		t.Errorf("got name %q, want %q", got.Name, "My House")
	}
}

func TestCreateMembroRequest_JSON(t *testing.T) {
	req := CreateMembroRequest{Nome: "John", Avatar: "🦊"}
	data, _ := json.Marshal(req)
	var got CreateMembroRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Nome != "John" || got.Avatar != "🦊" {
		t.Errorf("round-trip: got %+v, want %+v", got, req)
	}
}

func TestCreateCartaoRequest_JSON(t *testing.T) {
	req := CreateCartaoRequest{
		Nome:                "Nubank",
		DiaFechamento:       15,
		ResponsavelPadraoID: "membro-1",
	}
	data, _ := json.Marshal(req)
	var got CreateCartaoRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Nome != "Nubank" || got.DiaFechamento != 15 || got.ResponsavelPadraoID != "membro-1" {
		t.Errorf("round-trip: got %+v, want %+v", got, req)
	}
}

func TestCreateGastoRequest_JSON(t *testing.T) {
	faturaID := "fat-1"
	req := CreateGastoRequest{
		Descricao:          "Mercado",
		ValorTotalCentavos: 5000,
		CompradorID:        "membro-1",
		FaturaID:           &faturaID,
		Installments:       intPtr(3),
		TotalInstallments:  intPtr(3),
		SplitMode:          "EQUAL",
		Divisoes: []SplitItem{
			{MembroID: "m1", ValorCentavos: 2500},
			{MembroID: "m2", ValorCentavos: 2500},
		},
	}
	data, _ := json.Marshal(req)
	var got CreateGastoRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Descricao != req.Descricao || got.ValorTotalCentavos != req.ValorTotalCentavos {
		t.Errorf("round-trip: got %+v, want %+v", got, req)
	}
	if got.FaturaID == nil || *got.FaturaID != *req.FaturaID {
		t.Errorf("expected fatura_id %v, got %v", *req.FaturaID, *got.FaturaID)
	}
	if len(got.Divisoes) != 2 {
		t.Errorf("expected 2 divisoes, got %d", len(got.Divisoes))
	}
}

func TestCreateGastoRequest_ZeroInstallments(t *testing.T) {
	req := CreateGastoRequest{
		Descricao:          "Teste",
		ValorTotalCentavos: 1000,
		CompradorID:        "m1",
	}
	data, _ := json.Marshal(req)
	var got CreateGastoRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Installments != nil {
		t.Errorf("expected nil installments, got %v", *got.Installments)
	}
}

func TestSplitItem_JSON(t *testing.T) {
	item := SplitItem{MembroID: "m1", ValorCentavos: 1500}
	data, _ := json.Marshal(item)
	var got SplitItem
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != item {
		t.Errorf("round-trip: got %+v, want %+v", got, item)
	}
}

func TestCreateContaFixaRequest_JSON(t *testing.T) {
	val := int64(10000)
	req := CreateContaFixaRequest{
		Name:               "Internet",
		Icon:               "wifi",
		FixedValueCentavos: &val,
		DefaultSplit: []SplitItem{
			{MembroID: "m1", ValorCentavos: 5000},
			{MembroID: "m2", ValorCentavos: 5000},
		},
	}
	data, _ := json.Marshal(req)
	var got CreateContaFixaRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "Internet" || got.Icon != "wifi" {
		t.Errorf("round-trip: got %+v, want %+v", got, req)
	}
	if got.FixedValueCentavos == nil || *got.FixedValueCentavos != val {
		t.Errorf("expected fixed_value %d, got %v", val, got.FixedValueCentavos)
	}
}

func TestCreateContaFixaRequest_NilFixedValue(t *testing.T) {
	req := CreateContaFixaRequest{Name: "Agua", Icon: "💧"}
	data, _ := json.Marshal(req)
	var got CreateContaFixaRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.FixedValueCentavos != nil {
		t.Errorf("expected nil FixedValueCentavos, got %v", *got.FixedValueCentavos)
	}
}

func TestJoinTenantRequest_JSON(t *testing.T) {
	req := JoinTenantRequest{InviteCode: "ABC123"}
	data, _ := json.Marshal(req)
	var got JoinTenantRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.InviteCode != "ABC123" {
		t.Errorf("got invite_code %q, want %q", got.InviteCode, "ABC123")
	}
}

func TestValidateEventRequest_JSON(t *testing.T) {
	req := ValidateEventRequest{
		Type:      "TENANT_CREATED",
		DedupeKey: "tenant-1-2024-01",
		PeriodKey: "2024-01",
	}
	data, _ := json.Marshal(req)
	var got ValidateEventRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != req.Type || got.DedupeKey != req.DedupeKey || got.PeriodKey != req.PeriodKey {
		t.Errorf("round-trip: got %+v, want %+v", got, req)
	}
}

func TestValidateEventRequest_NoPeriodKey(t *testing.T) {
	req := ValidateEventRequest{Type: "FIRST_EXPENSE_CREATED", DedupeKey: "gasto-1"}
	data, _ := json.Marshal(req)
	var got ValidateEventRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.PeriodKey != "" {
		t.Errorf("expected empty period_key, got %q", got.PeriodKey)
	}
}

func TestTenantResponse_JSON(t *testing.T) {
	resp := TenantResponse{ID: "t-1", Name: "House", InviteCode: "XYZ", CreatedAt: "2024-01-01T00:00:00Z"}
	data, _ := json.Marshal(resp)
	var got TenantResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got != resp {
		t.Errorf("round-trip: got %+v, want %+v", got, resp)
	}
}

func TestMembroResponse_JSON(t *testing.T) {
	renda := int64(500000)
	resp := MembroResponse{
		ID: "m-1", Nome: "John", Avatar: "🐶", Ativo: true,
		Role: "ADMIN", RendaCentavos: &renda, UserID: "u-1",
	}
	data, _ := json.Marshal(resp)
	var got MembroResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != resp.ID || got.Nome != resp.Nome || got.Role != resp.Role {
		t.Errorf("round-trip: got %+v, want %+v", got, resp)
	}
	if got.RendaCentavos == nil || *got.RendaCentavos != renda {
		t.Errorf("expected renda %d, got %v", renda, got.RendaCentavos)
	}
}

func TestMembroResponse_ZeroRenda(t *testing.T) {
	resp := MembroResponse{ID: "m-2", Nome: "Jane", Avatar: "🐱", Ativo: false, Role: "MORADOR"}
	data, _ := json.Marshal(resp)
	var got MembroResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.RendaCentavos != nil {
		t.Errorf("expected nil renda, got %v", *got.RendaCentavos)
	}
}

func TestGastoResponse_JSON(t *testing.T) {
	resp := GastoResponse{
		ID: "g-1", Descricao: "Compras", ValorTotalCentavos: 3000,
		CompradorID: "m-1", Installments: 1, TotalInstallments: 1,
		IsLoan: false, CreatedAt: "2024-01-01T00:00:00Z", SplitMode: "EQUAL",
		Divisoes: []SplitItem{{MembroID: "m1", ValorCentavos: 3000}},
	}
	data, _ := json.Marshal(resp)
	var got GastoResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.ID != resp.ID || got.Descricao != resp.Descricao || got.SplitMode != resp.SplitMode {
		t.Errorf("round-trip: got %+v, want %+v", got, resp)
	}
	if len(got.Divisoes) != 1 {
		t.Errorf("expected 1 divisao, got %d", len(got.Divisoes))
	}
}

func TestGastoResponse_Loan(t *testing.T) {
	resp := GastoResponse{
		ID: "g-2", Descricao: "Emprestimo", ValorTotalCentavos: 100000,
		CompradorID: "m-1", IsLoan: true, CreatedAt: "2024-01-01T00:00:00Z", SplitMode: "CUSTOM",
	}
	data, _ := json.Marshal(resp)
	var got GastoResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !got.IsLoan {
		t.Error("expected IsLoan to be true")
	}
}

func TestWSMessage_JSON(t *testing.T) {
	msg := WSMessage{
		Type:    WSTypeExpenseCreated,
		Payload: map[string]interface{}{"id": "123", "valor": 5000},
	}
	data, _ := json.Marshal(msg)
	var got WSMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != WSTypeExpenseCreated {
		t.Errorf("got type %q, want %q", got.Type, WSTypeExpenseCreated)
	}
}

func TestWSMessage_NilPayload(t *testing.T) {
	msg := WSMessage{Type: WSTypeExpenseDeleted}
	data, _ := json.Marshal(msg)
	var got WSMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Payload != nil {
		t.Errorf("expected nil payload, got %v", got.Payload)
	}
}

func TestWSTypeConstants(t *testing.T) {
	if WSTypeExpenseCreated != "EXPENSE_CREATED" {
		t.Errorf("WSTypeExpenseCreated = %q", WSTypeExpenseCreated)
	}
	if WSTypeExpenseUpdated != "EXPENSE_UPDATED" {
		t.Errorf("WSTypeExpenseUpdated = %q", WSTypeExpenseUpdated)
	}
	if WSTypeExpenseDeleted != "EXPENSE_DELETED" {
		t.Errorf("WSTypeExpenseDeleted = %q", WSTypeExpenseDeleted)
	}
	if WSTypeMemberUpdated != "MEMBER_UPDATED" {
		t.Errorf("WSTypeMemberUpdated = %q", WSTypeMemberUpdated)
	}
	if WSTypeInvoiceUpdated != "INVOICE_UPDATED" {
		t.Errorf("WSTypeInvoiceUpdated = %q", WSTypeInvoiceUpdated)
	}
}

func TestContaFixaResponse_JSON(t *testing.T) {
	resp := ContaFixaResponse{
		ID:                 "cf-1",
		Name:               "Internet",
		Icon:               "wifi",
		FixedValueCentavos: int64Ptr(10000),
		DefaultSplit: []SplitItem{
			{MembroID: "m1", ValorCentavos: 5000},
			{MembroID: "m2", ValorCentavos: 5000},
		},
		CreatedAt: "2024-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(resp)
	var got ContaFixaResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "Internet" || got.Icon != "wifi" {
		t.Errorf("round-trip: got %+v, want %+v", got, resp)
	}
	if len(got.DefaultSplit) != 2 {
		t.Errorf("expected 2 split items, got %d", len(got.DefaultSplit))
	}
	if got.DefaultSplit[0].MembroID != "m1" {
		t.Errorf("expected m1, got %s", got.DefaultSplit[0].MembroID)
	}
}

func TestContaFixaResponse_EmptyDefaultSplit(t *testing.T) {
	resp := ContaFixaResponse{
		ID:        "cf-2",
		Name:      "Agua",
		Icon:      "opacity",
		CreatedAt: "2024-01-01T00:00:00Z",
	}
	data, _ := json.Marshal(resp)
	var got ContaFixaResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.DefaultSplit != nil {
		t.Errorf("expected nil defaultSplit, got %v", got.DefaultSplit)
	}
}

func TestMembroResponse_WithCreatedAt(t *testing.T) {
	resp := MembroResponse{
		ID:        "m-1",
		Nome:      "João",
		Avatar:    "🦊",
		Ativo:     true,
		Role:      "ADMIN",
		CreatedAt: "2024-01-15T10:30:00Z",
	}
	data, _ := json.Marshal(resp)
	var got MembroResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.CreatedAt != "2024-01-15T10:30:00Z" {
		t.Errorf("expected CreatedAt, got %q", got.CreatedAt)
	}
}

func TestGastoResponse_CardOwnerId(t *testing.T) {
	resp := GastoResponse{
		ID:                 "g-1",
		Descricao:          "Compra",
		ValorTotalCentavos: 5000,
		CompradorID:        "m1",
		CardOwnerID:        strPtr("m2"),
		SplitMode:          "EQUAL",
	}
	data, _ := json.Marshal(resp)

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["cardOwnerId"]; !ok {
		t.Error("expected cardOwnerId key in JSON")
	}
	if _, ok := raw["cardOwner"]; ok {
		t.Error("cardOwner key should NOT exist in JSON (should be cardOwnerId)")
	}

	var got GastoResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.CardOwnerID == nil || *got.CardOwnerID != "m2" {
		t.Errorf("expected cardOwnerId m2, got %v", got.CardOwnerID)
	}
}

func int64Ptr(i int64) *int64 { return &i }
func strPtr(s string) *string { return &s }

func intPtr(i int) *int {
	return &i
}
