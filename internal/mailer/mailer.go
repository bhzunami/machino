package mailer

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Mailer sends transactional emails via SMTP.
type Mailer struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// Config holds SMTP connection settings read from environment variables.
type Config struct {
	Host     string // SMTP_HOST
	Port     string // SMTP_PORT (default: 587)
	Username string // SMTP_USERNAME
	Password string // SMTP_PASSWORD
	From     string // SMTP_FROM (falls back to Username)
}

func New(cfg Config) *Mailer {
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	from := cfg.From
	if from == "" {
		from = cfg.Username
	}
	return &Mailer{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     from,
	}
}

// Enabled reports whether SMTP is configured.
func (m *Mailer) Enabled() bool {
	return m.host != "" && m.username != ""
}

// SendPasswordReset sends a password-reset email with the given token.
func (m *Mailer) SendPasswordReset(to, token string) error {
	subject := "Mach I No – Passwort zurücksetzen"
	body := fmt.Sprintf(
		"Hallo,\r\n\r\n"+
			"du hast eine Passwort-Zurücksetzung für dein Mach I No Konto angefordert.\r\n\r\n"+
			"Dein Reset-Token:\r\n\r\n    %s\r\n\r\n"+
			"Gib diesen Token auf der Login-Seite unter \"Passwort vergessen\" ein.\r\n"+
			"Der Token ist 30 Minuten gültig.\r\n\r\n"+
			"Falls du diese Anfrage nicht gestellt hast, kannst du diese E-Mail ignorieren.\r\n",
		token,
	)
	return m.send(to, subject, body)
}

func (m *Mailer) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	msg := buildMessage(m.from, to, subject, body)
	addr := m.host + ":" + m.port
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return b.String()
}
