package model

import (
	"time"

	"github.com/google/uuid"
)

// PartnerImage は partner_images テーブルの1行に対応する構造体
type PartnerImage struct {
	ID               uuid.UUID `json:"id"`
	PartnerID        uuid.UUID `json:"partner_id"`
	Stage            string    `json:"stage"`
	ImageURL         *string   `json:"image_url"`
	GenerationPrompt string    `json:"generation_prompt"`
	Status           string    `json:"status"`
	ErrorMessage     *string   `json:"error_message"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ステータス定数（コード内で文字列を直接書かないため）
const (
	ImageStatusPending   = "pending"
	ImageStatusCompleted = "completed"
	ImageStatusFailed    = "failed"
)
