package domain

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

// TelegramUserID represents a Telegram user ID.
// It ensures type safety and validation for Telegram user IDs.
type TelegramUserID int64

var (
	ErrInvalidTelegramUserID = errors.New("invalid telegram user ID")
)

// NewTelegramUserID creates a new TelegramUserID with validation.
// According to Telegram Bot API, user IDs are positive integers.
func NewTelegramUserID(id int64) (TelegramUserID, error) {
	if id <= 0 {
		return 0, fmt.Errorf("%w: must be positive", ErrInvalidTelegramUserID)
	}
	return TelegramUserID(id), nil
}

// MustNewTelegramUserID creates a new TelegramUserID and panics if validation fails.
// Use this only when you are sure the ID is valid.
func MustNewTelegramUserID(id int64) TelegramUserID {
	userID, err := NewTelegramUserID(id)
	if err != nil {
		panic(err)
	}
	return userID
}

// Int64 returns the underlying int64 value.
func (id TelegramUserID) Int64() int64 {
	return int64(id)
}

// String returns a string representation of the ID.
func (id TelegramUserID) String() string {
	return fmt.Sprintf("%d", id)
}

// Value implements the driver.Valuer interface for database storage.
func (id TelegramUserID) Value() (driver.Value, error) {
	return int64(id), nil
}

// Scan implements the sql.Scanner interface for database retrieval.
func (id *TelegramUserID) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("%w: cannot be nil", ErrInvalidTelegramUserID)
	}

	switch v := value.(type) {
	case int64:
		if v <= 0 {
			return fmt.Errorf("%w: must be positive", ErrInvalidTelegramUserID)
		}
		*id = TelegramUserID(v)
		return nil
	default:
		return fmt.Errorf("%w: unsupported type %T", ErrInvalidTelegramUserID, value)
	}
}

// Equal compares two TelegramUserIDs for equality.
func (id TelegramUserID) Equal(other TelegramUserID) bool {
	return id == other
}
