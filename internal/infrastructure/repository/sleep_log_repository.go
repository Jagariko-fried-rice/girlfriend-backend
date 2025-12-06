package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type sleepLogRepository struct {
	db *sql.DB
}

func NewSleepLogRepository(db *sql.DB) repository.SleepLogRepository {
	return &sleepLogRepository{db: db}
}

func (r *sleepLogRepository) Create(ctx context.Context, log *model.SleepLog) error {
	query := `
		INSERT INTO sleep_logs (user_id, partner_id, slept_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	// 生成されたIDを取得してモデルにセット
	return r.db.QueryRowContext(ctx, query, log.UserID, log.PartnerID, log.SleptAt).Scan(&log.ID)
}

func (r *sleepLogRepository) UpdateWakeTime(ctx context.Context, id string, log model.SleepLog) error {
	query := `
		UPDATE sleep_logs
		SET wake_at = $1, sleep_minutes = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, log.WakeAt, log.SleepMinutes, id)
	return err
}

func (r *sleepLogRepository) FindUnfinished(ctx context.Context, partnerID string) (*model.SleepLog, error) {
	query := `
		SELECT id, user_id, partner_id, slept_at, wake_at, sleep_minutes
		FROM sleep_logs
		WHERE partner_id = $1 AND wake_at IS NULL
		ORDER BY slept_at DESC
		LIMIT 1
	`
	var l model.SleepLog
	err := r.db.QueryRowContext(ctx, query, partnerID).Scan(
		&l.ID, &l.UserID, &l.PartnerID, &l.SleptAt, &l.WakeAt, &l.SleepMinutes,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // 寝ていない場合はnilを返す
		}
		return nil, fmt.Errorf("failed to find sleep log: %w", err)
	}
	return &l, nil
}