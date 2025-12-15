package composite

import (
	"event-booker/internal/domain"
	"event-booker/internal/notification/email"
	"event-booker/internal/notification/telegram"
	"fmt"
)

type CompositeNotifier struct {
	email    *email.Notifier
	telegram *telegram.Notifier
}

func NewCompositeNotifier(email *email.Notifier, telegram *telegram.Notifier) *CompositeNotifier {
	return &CompositeNotifier{email: email, telegram: telegram}
}

func (c *CompositeNotifier) NotifyCancellation(user *domain.User, booking *domain.Booking) error {
	var errs []error
	if user.Email != "" {
		if err := c.email.NotifyCancellation(user, booking); err != nil {
			errs = append(errs, err)
		}
	}
	if user.Telegram != "" {
		if err := c.telegram.NotifyCancellation(user, booking); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("notification errors: %v", errs)
	}
	return nil
}
