package telegram

import (
	"event-booker/internal/domain"
	"fmt"
	"net/http"
	"net/url"
)

type Notifier struct {
	token string
}

func NewNotifier(token string) *Notifier {
	return &Notifier{token: token}
}

func (n *Notifier) NotifyCancellation(user *domain.User, booking *domain.Booking) error {
	if user.Telegram == "" {
		return nil
	}
	text := fmt.Sprintf("Your booking for event %s has been cancelled due to expiration.", booking.EventID)
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		n.token, user.Telegram, url.QueryEscape(text))
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api error: %d", resp.StatusCode)
	}
	return nil
}
