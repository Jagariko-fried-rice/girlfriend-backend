package repository

import (
	"context"
	"database/sql"
	"fmt"

	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type scenarioRepository struct {
	db *sql.DB
}

func NewScenarioRepository(db *sql.DB) repository.ScenarioRepository {
	return &scenarioRepository{db: db}
}

func (r *scenarioRepository) FindRandomByStage(ctx context.Context, stage string) (*model.Scenario, error) {
	// RANDOM() でランダムに1件取得
	query := `
		SELECT id, stage, routes, template_text
		FROM scenarios
		WHERE stage = $1
		ORDER BY RANDOM()
		LIMIT 1
	`
	var s model.Scenario
	err := r.db.QueryRowContext(ctx, query, stage).Scan(&s.ID, &s.Stage, &s.Routes, &s.TemplateText)
	if err != nil {
		return nil, fmt.Errorf("scenario not found: %w", err)
	}
	return &s, nil
}