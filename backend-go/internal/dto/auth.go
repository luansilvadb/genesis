package dto

type RegisterRequest struct {
	Email      string  `json:"email" binding:"required,email"`
	Nome       string  `json:"nome" binding:"required"`
	Password   string  `json:"password" binding:"required,min=8"`
	InviteCode *string `json:"inviteCode,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type GoogleLoginRequest struct {
	Credential string  `json:"credential" binding:"required"`
	InviteCode *string `json:"inviteCode,omitempty"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

type UserProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Nome  string `json:"nome"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type SessionTenant struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"inviteCode"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

type SessionResponse struct {
	Tenants []SessionTenant `json:"tenants"`
}

type InvitePreviewMembro struct {
	ID     string `json:"id"`
	Nome   string `json:"nome"`
	Avatar string `json:"avatar"`
}

type InvitePreviewResponse struct {
	ID                 string                `json:"id"`
	Name               string                `json:"name"`
	MembrosDisponiveis []InvitePreviewMembro `json:"membrosDisponiveis"`
}
