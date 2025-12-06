package repository

import (
	"context"
)

type PartnerRepository interface {
	// ステータス（体力・知力・センス）を更新する
	UpdateStatus(ctx context.Context, partnerID string, stamina, intelligence, sense int) error
}