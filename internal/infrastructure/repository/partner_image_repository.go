package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type partnerImageRepository struct {
	db *sql.DB
}

// NewPartnerImageRepository はコンストラクタです
func NewPartnerImageRepository(db *sql.DB) repository.PartnerImageRepository {
	return &partnerImageRepository{db: db}
}

// FindFirstPending: 未生成のデータを1つ取得
func (r *partnerImageRepository) FindFirstPending(ctx context.Context) (*model.PartnerImage, error) {
	// 1. SQLの準備
	// FOR UPDATE SKIP LOCKED を使うと、複数のバッチが同時に動いても同じデータを取り合わないようにできます
	query := `
		SELECT id, partner_id, stage, image_url, generation_prompt, status, error_message, created_at, updated_at
		FROM partner_images
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT 1
	`

	// 2. 実行
	row := r.db.QueryRowContext(ctx, query, model.ImageStatusPending)

	// 3. データのマッピング
	var i model.PartnerImage
	err := row.Scan(
		&i.ID, &i.PartnerID, &i.Stage, &i.ImageURL, &i.GenerationPrompt,
		&i.Status, &i.ErrorMessage, &i.CreatedAt, &i.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // データがない場合はエラーではなくnilを返すのが一般的
		}
		return nil, fmt.Errorf("failed to find pending image: %w", err)
	}

	return &i, nil
}

// Update: データ更新
func (r *partnerImageRepository) Update(ctx context.Context, img *model.PartnerImage) error {
	query := `
		UPDATE partner_images
		SET image_url = $1, status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, img.ImageURL, img.Status, img.ErrorMessage, img.ID)
	if err != nil {
		return fmt.Errorf("failed to update partner image: %w", err)
	}

	return nil
}

func (r *partnerImageRepository) Create(ctx context.Context, img *model.PartnerImage) error {
	query := `
		INSERT INTO partner_images (partner_id, stage, generation_prompt, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, img.PartnerID, img.Stage, img.GenerationPrompt, img.Status)
	return err
}

func (r *partnerImageRepository) FindByPartnerID(ctx context.Context, partnerID string) ([]*model.PartnerImage, error) {
	query := `
		SELECT id, partner_id, stage, image_url, generation_prompt, status, error_message, created_at, updated_at
		FROM partner_images
		WHERE partner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*model.PartnerImage
	for rows.Next() {
		var i model.PartnerImage
		if err := rows.Scan(
			&i.ID, &i.PartnerID, &i.Stage, &i.ImageURL, &i.GenerationPrompt,
			&i.Status, &i.ErrorMessage, &i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		images = append(images, &i)
	}
	return images, nil
}