package repository

import (
	"context"
	"girlfriend-backend/internal/domain/model"
)

type VoiceRepository interface {
	FindByPersonality(ctx context.Context, personality string) ([]*model.VoiceLine, error)
}