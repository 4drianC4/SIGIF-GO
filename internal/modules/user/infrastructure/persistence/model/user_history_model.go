package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type HistoryFieldChange struct {
	From *string `json:"from"`
	To   *string `json:"to"`
}

type HistoryChanges map[string]HistoryFieldChange

func (c HistoryChanges) Value() (driver.Value, error) {
	if c == nil {
		return "{}", nil
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}

func (c *HistoryChanges) Scan(value any) error {
	var raw []byte
	switch v := value.(type) {
	case nil:
		*c = HistoryChanges{}
		return nil
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return fmt.Errorf("unsupported history changes type %T", value)
	}
	changes := HistoryChanges{}
	if err := json.Unmarshal(raw, &changes); err != nil {
		return err
	}
	*c = changes
	return nil
}

type UserHistoryModel struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID      `gorm:"type:uuid;not null;index:idx_user_history_user_created,priority:1"`
	Action           string         `gorm:"type:varchar(40);not null"`
	Description      string         `gorm:"type:varchar(255);not null"`
	Changes          HistoryChanges `gorm:"type:jsonb;not null;default:'{}'"`
	PerformedByID    *uuid.UUID     `gorm:"type:uuid;index"`
	PerformedByName  string         `gorm:"type:varchar(161)"`
	PerformedByEmail string         `gorm:"type:varchar(160)"`
	CreatedAt        time.Time      `gorm:"not null;index:idx_user_history_user_created,priority:2,sort:desc"`
}

func (UserHistoryModel) TableName() string {
	return "user_history"
}
