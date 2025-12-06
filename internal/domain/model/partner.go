package model

import "github.com/google/uuid"

type Partner struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	CurrentStage string    `json:"current_stage"`
	Stamina      int       `json:"stamina"`
	Intelligence int       `json:"intelligence"`
	Sense        int       `json:"sense"`
}