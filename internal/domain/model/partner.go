package model

import "github.com/google/uuid"

type Partner struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	Personality  string    `json:"personality"`
	HairColor    string    `json:"hair_color"`
	VoiceType    string    `json:"voice_type"`
	CurrentStage string    `json:"current_stage"`
	Stamina      int       `json:"stamina"`
	Intelligence int       `json:"intelligence"`
	Sense        int       `json:"sense"`
}