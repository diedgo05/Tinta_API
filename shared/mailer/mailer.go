package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer interface {
	Send(to, subject, htmlBody string) error
}

type SMTPMailer struct {
	host     string
	port     string
	user     string
	password string
	from     string
}

func NewSMTPMailer(host, port, user, password, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, user: user, password: password, from: from}
}

func (m *SMTPMailer) Send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.user, m.password, m.host)

	headers := fmt.Sprintf(
		"From: Tinta <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		m.from, to, subject,
	)
	msg := []byte(headers + htmlBody)

	if err := smtp.SendMail(addr, auth, m.from, []string{to}, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func VerificationCodeEmail(code string) (subject, htmlBody string) {
	subject = "Tu código de verificación de Tinta"
	htmlBody = fmt.Sprintf(`
<div style="font-family:Arial,sans-serif;max-width:420px;margin:0 auto;padding:24px;">
  <h2 style="color:#1C6B50;margin-bottom:4px;">Tinta</h2>
  <p style="color:#333;font-size:14px;">Usa este código para verificar tu correo dentro de la app:</p>
  <div style="background:#F2F5EF;border-radius:12px;padding:20px;text-align:center;margin:16px 0;">
    <span style="font-size:32px;font-weight:800;letter-spacing:8px;color:#1C6B50;">%s</span>
  </div>
  <p style="color:#888;font-size:12px;">Este código vence en 15 minutos. Si tú no pediste esto, puedes ignorar este correo.</p>
</div>`, code)
	return subject, htmlBody
}