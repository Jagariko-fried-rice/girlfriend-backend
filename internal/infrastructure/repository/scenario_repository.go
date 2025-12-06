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

// 共通のSELECT文（DRY原則：同じSQLを何度も書かない）
const selectScenarioColumns = `
	SELECT id, stage, routes, template_text, stat_effect, weight, 
	       condition_stat, condition_value, success_text, failure_text, success_effect, failure_effect,
	       image_prompt
	FROM scenarios
`

// ヘルパー関数：SQLの行データを構造体に変換する
func scanScenario(row *sql.Row) (*model.Scenario, error) {
	var s model.Scenario
	var statEffect, successEffect, failureEffect string
	var conditionStat, successText, failureText sql.NullString

	err := row.Scan(
		&s.ID, &s.Stage, &s.Routes, &s.TemplateText, &statEffect, &s.Weight,
		&conditionStat, &s.ConditionValue, &successText, &failureText, &successEffect, &failureEffect,
		&s.ImagePrompt,
	)
	if err != nil {
		return nil, err
	}

	// データの詰め替え
	s.StatEffect = statEffect
	s.SuccessEffect = successEffect
	s.FailureEffect = failureEffect
	if conditionStat.Valid { s.ConditionStat = &conditionStat.String }
	if successText.Valid { s.SuccessText = &successText.String }
	if failureText.Valid { s.FailureText = &failureText.String }

	return &s, nil
}

// FindRandomByStage: ランダム取得（修正版：全カラム取得）
func (r *scenarioRepository) FindRandomByStage(ctx context.Context, stage string) (*model.Scenario, error) {
	query := selectScenarioColumns + `
		WHERE stage = $1
		ORDER BY RANDOM()
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, stage)
	s, err := scanScenario(row)
	if err != nil {
		return nil, fmt.Errorf("random scenario not found: %w", err)
	}
	return s, nil
}

// FindByStageAndRoute: 指名取得
func (r *scenarioRepository) FindByStageAndRoute(ctx context.Context, stage string, route string) (*model.Scenario, error) {
	query := selectScenarioColumns + `
		WHERE stage = $1 AND routes = $2
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, stage, route)
	s, err := scanScenario(row)
	if err != nil {
		return nil, fmt.Errorf("scenario not found: %w", err)
	}
	return s, nil
}