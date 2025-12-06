package model

import (
	"time"

	"github.com/google/uuid"
)

type Memory struct {
	ID              uuid.UUID `json:"id"`
	PartnerID       uuid.UUID `json:"partner_id"`
	ScenarioID      uuid.UUID `json:"scenario_id"`
	GeneratedPrompt string    `json:"generated_prompt"`
	OccurredAt      time.Time `json:"occurred_at"`
}