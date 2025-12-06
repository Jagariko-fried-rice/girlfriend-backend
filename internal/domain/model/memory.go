package model

import (
	"time"

	"github.com/google/uuid"
)

type Memory struct {
	ID              uuid.UUID
	PartnerID       uuid.UUID
	ScenarioID      uuid.UUID
	GeneratedPrompt string // 実際に使ったプロンプト（画像用ではなくテキスト用）
	OccurredAt      time.Time
}