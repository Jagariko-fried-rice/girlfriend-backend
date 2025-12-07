package model

import "github.com/google/uuid"

type VoiceLine struct {
	ID          uuid.UUID `json:"id"`
	Personality string    `json:"personality"`
	Situation   string    `json:"situation"`
	LineText    string    `json:"line_text"`
	AudioURL    string    `json:"audio_url"`
}