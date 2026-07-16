package service

import (
	"errors"
	"html"
	"net/url"

	"github.com/luansilvadb/financeiro-divi/backend-go/internal/config"
	gomail "gopkg.in/gomail.v2"
)

var ErrSMTPNotConfigured = errors.New("serviço de email não configurado")

type EmailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendPasswordReset(email, token string) error {
	if s.cfg.SMTPUser == "" || s.cfg.SMTPPass == "" {
		return ErrSMTPNotConfigured
	}

	link := s.cfg.FrontendURL + "/#/reset-password?token=" + url.QueryEscape(token)

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.SMTPUser)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "DIVI - Recuperação de Senha")
	escapedLink := html.EscapeString(link)
	m.SetBody("text/html", "<p>Clique no link para redefinir sua senha:</p><p><a href=\""+escapedLink+"\">"+escapedLink+"</a></p>")

	d := gomail.NewDialer(s.cfg.SMTPHost, s.cfg.SMTPPort, s.cfg.SMTPUser, s.cfg.SMTPPass)
	d.SSL = s.cfg.SMTPUseTLS

	return d.DialAndSend(m)
}
