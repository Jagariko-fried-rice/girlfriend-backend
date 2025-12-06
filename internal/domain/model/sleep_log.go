package model

import (
	"time"

	"github.com/google/uuid"
)

type SleepLog struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	PartnerID    uuid.UUID  `json:"partner_id"`
	SleptAt      time.Time  `json:"slept_at"`
	WakeAt       *time.Time `json:"wake_at"`       // まだ起きてない時はNULLなのでポインタ
	SleepMinutes *int       `json:"sleep_minutes"` // 同上
}