package repository

import (
	"context"

	"girlfriend-backend/internal/domain/model"
)

type PartnerRepository interface {
	// ステータス（体力・知力・センス）を更新する
	UpdateStatus(ctx context.Context, partnerID string, stamina, intelligence, sense int) error

	FindByID(ctx context.Context, partnerID string) (*model.Partner, error)

	Create(ctx context.Context, partner *model.Partner) error
}