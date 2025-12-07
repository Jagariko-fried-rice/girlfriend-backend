package repository

import (
	"context"
	"girlfriend-backend/internal/domain/model"
)

// PartnerImageRepository は partner_images テーブルへのアクセス方法を定めたインターフェースです
type PartnerImageRepository interface {
	FindFirstPending(ctx context.Context) (*model.PartnerImage, error)
	Update(ctx context.Context, image *model.PartnerImage) error
	Create(ctx context.Context, image *model.PartnerImage) error
	FindByPartnerID(ctx context.Context, partnerID string) ([]*model.PartnerImage, error)
}
