package mailer

import (
	"fmt"
	"strconv"

	mail "github.com/wneessen/go-mail"
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
	port, err := strconv.Atoi(m.port)
	if err != nil {
		return fmt.Errorf("smtp port %q: %w", m.port, err)
	}

	msg := mail.NewMsg()
	if err := msg.From(m.from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	if err := msg.To(to); err != nil {
		return fmt.Errorf("mail to: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)

	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(m.username),
		mail.WithPassword(m.password),
	}
	if port == 465 {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(m.host, opts...)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
