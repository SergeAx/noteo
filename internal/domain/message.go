package domain

// Message represents a notification message to be sent
type Message struct {
	UserID TelegramUserID
	Text   string
	Muted  bool
}
