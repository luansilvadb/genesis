package validator

import (
	"errors"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/luansilvadb/financeiro-divi/backend-go/internal/model"
)

var (
	ErrPasswordTooShort    = errors.New("a senha deve ter no mínimo 8 caracteres")
	ErrPasswordNoUppercase = errors.New("a senha deve conter pelo menos uma letra maiúscula")
	ErrPasswordNoNumber    = errors.New("a senha deve conter pelo menos um número")
)

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	if len(password) > 128 {
		return errors.New("a senha deve ter no máximo 128 caracteres")
	}

	hasUpper, hasLower, hasNumber := scanCharClasses(password)

	if !hasUpper {
		return ErrPasswordNoUppercase
	}
	if !hasLower {
		return errors.New("a senha deve conter pelo menos uma letra minúscula")
	}
	if !hasNumber {
		return ErrPasswordNoNumber
	}

	return nil
}

func scanCharClasses(s string) (hasUpper, hasLower, hasNumber bool) {
	for _, r := range s {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		if unicode.IsDigit(r) {
			hasNumber = true
		}
	}
	return
}

func RegisterGinValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("role", func(fl validator.FieldLevel) bool {
			role := model.Role(fl.Field().String())
			switch role {
			case model.RoleAdmin, model.RoleMorador, model.RoleVisualizador:
				return true
			}
			return false
		})
		_ = v.RegisterValidation("split_mode", func(fl validator.FieldLevel) bool {
			mode := model.SplitMode(fl.Field().String())
			switch mode {
			case model.SplitModeEqual, model.SplitModeIncome, model.SplitModeCustom:
				return true
			}
			return false
		})
	}
}
