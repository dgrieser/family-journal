package email

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Sender interface {
	Send(to, subject, body string) error
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

func (s *smtpSender) Send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	msg := strings.Join([]string{
		"From: " + s.cfg.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.cfg.User != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)
	}

	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg))
}

func (n *noOpSender) Send(_, _, _ string) error {
	return nil
}

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
		for _, admin := range adminEmails {
			if err := s.Send(admin, subject, body); err != nil {
				log.Printf("email: failed to send new user notification to admin %s: %v", admin, err)
			}
		}
	}()
}
