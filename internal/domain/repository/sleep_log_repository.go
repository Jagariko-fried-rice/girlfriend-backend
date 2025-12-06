package repository

import (
	"context"
	"girlfriend-backend/internal/domain/model"
)

type SleepLogRepository interface {
	// 寝る: 新しい睡眠ログを作成
	Create(ctx context.Context, log *model.SleepLog) error
	
	// 起きる: 起床時間と睡眠時間を更新
	UpdateWakeTime(ctx context.Context, id string, wakeAt model.SleepLog) error
	
	// 確認: このパートナーの「まだ起きていない（WakeAtがNULL）」最新のログを取得
	FindUnfinished(ctx context.Context, partnerID string) (*model.SleepLog, error)
}