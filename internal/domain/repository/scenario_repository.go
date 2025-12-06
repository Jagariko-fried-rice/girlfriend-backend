package repository

import (
	"context"
	"girlfriend-backend/internal/domain/model"
)

type ScenarioRepository interface {
	// ランダム取得
	FindRandomByStage(ctx context.Context, stage string) (*model.Scenario, error)
	// ★★★ 指名取得（ここにあるべき定義です） ★★★
	FindByStageAndRoute(ctx context.Context, stage string, route string) (*model.Scenario, error)
}