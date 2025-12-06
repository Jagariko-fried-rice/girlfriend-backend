package model

import (
	"time"

	"github.com/google/uuid"
)

// PartnerImage は partner_images テーブルの1行に対応する構造体
type PartnerImage struct {
	ID               uuid.UUID
	PartnerID        uuid.UUID
	Stage            string
	ImageURL         *string // NULLになる可能性があるためポインタにする
	GenerationPrompt string
	Status           string
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ステータス定数（コード内で文字列を直接書かないため）
const (
	ImageStatusPending   = "pending"
	ImageStatusCompleted = "completed"
	ImageStatusFailed    = "failed"
)