package dto

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

const (
	WSTypeExpenseCreated    = "EXPENSE_CREATED"
	WSTypeExpenseUpdated    = "EXPENSE_UPDATED"
	WSTypeExpenseDeleted    = "EXPENSE_DELETED"
	WSTypeMemberUpdated     = "MEMBER_UPDATED"
	WSTypeMemberCreated     = "MEMBER_CREATED"
	WSTypeCardCreated       = "CARD_CREATED"
	WSTypeCardDeleted       = "CARD_DELETED"
	WSTypeInvoiceUpdated    = "INVOICE_UPDATED"
	WSTypeFixedBillCreated  = "FIXED_BILL_CREATED"
	WSTypeFixedBillUpdated  = "FIXED_BILL_UPDATED"
	WSTypeFixedBillDeleted  = "FIXED_BILL_DELETED"
	WSTypePermissionsUpdate = "PERMISSIONS_UPDATED"
)
