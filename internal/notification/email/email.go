package email

import (
	"event-booker/internal/config"
	"event-booker/internal/domain"
	"fmt"
	"net/smtp"
)

type Notifier struct {
	cfg *config.Config
}

func NewNotifier(cfg *config.Config) *Notifier {
	return &Notifier{cfg: cfg}
}

func (n *Notifier) NotifyCancellation(email string, booking *domain.Booking) error {
	auth := smtp.PlainAuth("", n.cfg.EmailConfig.SMTPUser, n.cfg.EmailConfig.SMTPPassword, n.cfg.EmailConfig.SMTPHost)
	to := []string{email}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: Booking Cancelled\r\n" +
		"\r\n" +
		"Your booking for event " + booking.EventID + " has been cancelled due to expiration.\r\n")
	addr := fmt.Sprintf("%s:%d", n.cfg.EmailConfig.SMTPHost, n.cfg.EmailConfig.SMTPPort)
	return smtp.SendMail(addr, auth, n.cfg.EmailConfig.FromEmail, to, msg)
}
