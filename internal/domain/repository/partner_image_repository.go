package repository

import (
	"context"
	"girlfriend-backend/internal/domain/model"
)

// PartnerImageRepository は partner_images テーブルへのアクセス方法を定めたインターフェースです
type PartnerImageRepository interface {
	// FindFirstPending は、ステータスが「pending（未生成）」のデータを1件取得します
	FindFirstPending(ctx context.Context) (*model.PartnerImage, error)

	// Update は、画像生成の結果（URLやステータス）を更新します
	Update(ctx context.Context, image *model.PartnerImage) error
}