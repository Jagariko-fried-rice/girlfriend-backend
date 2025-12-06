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

type ScenarioRepository interface {
	// 指定したステージのシナリオからランダムに1つ取得
	FindRandomByStage(ctx context.Context, stage string) (*model.Scenario, error)
}
type UserRepository interface {
	// 全ユーザーと、そのパートナー情報を取得
	FindAllWithPartner(ctx context.Context) ([]*model.UserWithPartner, error)
}

func (r *partnerImageRepository) Create(ctx context.Context, img *model.PartnerImage) error {
	query := `
		INSERT INTO partner_images (partner_id, stage, generation_prompt, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, img.PartnerID, img.Stage, img.GenerationPrompt, img.Status)
	return err
}