package email

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

const smtpTimeout = 30 * time.Second

type Sender interface {
	Send(to, subject, body string) error
	SendMulti(to []string, subject, body string) error
}

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

type smtpSender struct {
	cfg Config
}

type noOpSender struct{}

func New(cfg Config) Sender {
	if cfg.Host == "" {
		return &noOpSender{}
	}
	return &smtpSender{cfg: cfg}
}

// sanitizeHeader strips CR and LF to prevent header injection.
func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}

func (s *smtpSender) send(recipients []string, toHeader, subject, body string) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))

	conn, err := net.DialTimeout("tcp", addr, smtpTimeout)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		conn.Close()
		return fmt.Errorf("smtp set deadline: %w", err)
	}

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if s.cfg.User != "" {
		auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	envelopeFrom := s.cfg.From
	fromHeader := s.cfg.From
	if parsed, err := mail.ParseAddress(s.cfg.From); err == nil {
		envelopeFrom = parsed.Address
		fromHeader = parsed.String()
	}

	msg := strings.Join([]string{
		"From: " + sanitizeHeader(fromHeader),
		"To: " + sanitizeHeader(toHeader),
		"Subject: " + mime.QEncoding.Encode("utf-8", sanitizeHeader(subject)),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n"),
	}, "\r\n")

	if err := c.Mail(envelopeFrom); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	for _, r := range recipients {
		if err := c.Rcpt(r); err != nil {
			return fmt.Errorf("smtp RCPT TO %s: %w", r, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	return c.Quit()
}

func (s *smtpSender) Send(to, subject, body string) error {
	return s.send([]string{to}, to, subject, body)
}

// SendMulti sends one message to all recipients in a single SMTP transaction.
// The To header is set to "undisclosed-recipients:;" to avoid exposing
// individual addresses to each other.
func (s *smtpSender) SendMulti(to []string, subject, body string) error {
	if len(to) == 0 {
		return nil
	}
	return s.send(to, "undisclosed-recipients:;", subject, body)
}

func (n *noOpSender) Send(_, _, _ string) error              { return nil }
func (n *noOpSender) SendMulti(_ []string, _, _ string) error { return nil }

// SendRegistrationPending sends a pending-approval notice to a newly registered user.
// The call is synchronous; callers are responsible for running it asynchronously.
func SendRegistrationPending(s Sender, to string) error {
	subject := "Your Family Journal account is pending approval"
	body := "Hello,\r\n\r\n" +
		"Your Family Journal account has been created and is pending admin approval.\r\n\r\n" +
		"You will receive another email as soon as your account has been activated.\r\n\r\n" +
		"Family Journal"
	return s.Send(to, subject, body)
}

// SendAccountActivated notifies a user that their account has been activated.
// The call is synchronous; callers are responsible for running it asynchronously.
func SendAccountActivated(s Sender, to string) error {
	subject := "Your Family Journal account has been activated"
	body := "Hello,\r\n\r\n" +
		"Your Family Journal account has been activated. You can now log in.\r\n\r\n" +
		"Family Journal"
	return s.Send(to, subject, body)
}

// SendNewUserNotification alerts all admin users that a new registration is pending.
// The call is synchronous; callers are responsible for running it asynchronously.
func SendNewUserNotification(s Sender, adminEmails []string, newUserEmail string) error {
	if len(adminEmails) == 0 {
		return nil
	}
	subject := "New user registered: " + newUserEmail
	body := "Hello,\r\n\r\n" +
		"A new user has registered with the following email address:\r\n\r\n" +
		"  " + newUserEmail + "\r\n\r\n" +
		"Please log in to the admin panel to review and activate their account.\r\n\r\n" +
		"Family Journal"
	return s.SendMulti(adminEmails, subject, body)
}

