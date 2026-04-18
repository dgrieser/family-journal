package email

import (
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

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

func (s *smtpSender) send(recipients []string, toHeader, subject, body string) error {
	addr := net.JoinHostPort(s.cfg.Host, fmt.Sprintf("%d", s.cfg.Port))
	msg := strings.Join([]string{
		"From: " + s.cfg.From,
		"To: " + toHeader,
		"Subject: " + subject,
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	}

	envelopeFrom := s.cfg.From
	if parsed, err := mail.ParseAddress(s.cfg.From); err == nil {
		envelopeFrom = parsed.Address
	}

	return smtp.SendMail(addr, auth, envelopeFrom, recipients, []byte(msg))
}

func (s *smtpSender) Send(to, subject, body string) error {
	return s.send([]string{to}, to, subject, body)
}

func (s *smtpSender) SendMulti(to []string, subject, body string) error {
	if len(to) == 0 {
		return nil
	}
	return s.send(to, strings.Join(to, ", "), subject, body)
}

func (n *noOpSender) Send(_, _, _ string) error              { return nil }
func (n *noOpSender) SendMulti(_ []string, _, _ string) error { return nil }

func SendRegistrationPending(s Sender, to string) {
	go func() {
		subject := "Your Family Journal account is pending approval"
		body := "Hello,\r\n\r\n" +
			"Your Family Journal account has been created and is pending admin approval.\r\n\r\n" +
			"You will receive another email as soon as your account has been activated.\r\n\r\n" +
			"Family Journal"
		if err := s.Send(to, subject, body); err != nil {
			log.Printf("email: failed to send registration pending to %s: %v", to, err)
		}
	}()
}

func SendAccountActivated(s Sender, to string) {
	go func() {
		subject := "Your Family Journal account has been activated"
		body := "Hello,\r\n\r\n" +
			"Your Family Journal account has been activated. You can now log in.\r\n\r\n" +
			"Family Journal"
		if err := s.Send(to, subject, body); err != nil {
			log.Printf("email: failed to send activation email to %s: %v", to, err)
		}
	}()
}

func SendNewUserNotification(s Sender, adminEmails []string, newUserEmail string) {
	if len(adminEmails) == 0 {
		return
	}
	go func() {
		subject := "New user registered: " + newUserEmail
		body := "Hello,\r\n\r\n" +
			"A new user has registered with the following email address:\r\n\r\n" +
			"  " + newUserEmail + "\r\n\r\n" +
			"Please log in to the admin panel to review and activate their account.\r\n\r\n" +
			"Family Journal"
		if err := s.SendMulti(adminEmails, subject, body); err != nil {
			log.Printf("email: failed to send new user notification to admins: %v", err)
		}
	}()
}
