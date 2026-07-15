package dto

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WSError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	WSTypeExpenseCreated    = "EXPENSE_CREATED"
	WSTypeExpenseUpdated    = "EXPENSE_UPDATED"
	WSTypeExpenseDeleted    = "EXPENSE_DELETED"
	WSTypeMemberUpdated     = "MEMBER_UPDATED"
	WSTypeMemberCreated     = "MEMBER_CREATED"
	WSTypeCardUpdated       = "CARD_UPDATED"
	WSTypeCardCreated       = "CARD_CREATED"
	WSTypeCardDeleted       = "CARD_DELETED"
	WSTypeInvoiceUpdated    = "INVOICE_UPDATED"
	WSTypeFixedBillCreated  = "FIXED_BILL_CREATED"
	WSTypeFixedBillUpdated  = "FIXED_BILL_UPDATED"
	WSTypeFixedBillDeleted  = "FIXED_BILL_DELETED"
	WSTypePermissionsUpdate = "PERMISSIONS_UPDATED"
)
