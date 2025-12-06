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

// FindByStageAndRoute: ステージとルート名を指定してシナリオを取得
func (r *scenarioRepository) FindByStageAndRoute(ctx context.Context, stage string, route string) (*model.Scenario, error) {
	// image_prompt も含めて取得する
	query := `
		SELECT id, stage, routes, template_text, stat_effect, weight, 
		       condition_stat, condition_value, success_text, failure_text, success_effect, failure_effect,
		       image_prompt
		FROM scenarios
		WHERE stage = $1 AND routes = $2
		LIMIT 1
	`
	var s model.Scenario
	// Scan用の変数
	var statEffect, successEffect, failureEffect string
	var conditionStat, successText, failureText sql.NullString
	
	err := r.db.QueryRowContext(ctx, query, stage, route).Scan(
		&s.ID, &s.Stage, &s.Routes, &s.TemplateText, &statEffect, &s.Weight,
		&conditionStat, &s.ConditionValue, &successText, &failureText, &successEffect, &failureEffect,
		&s.ImagePrompt, // ★追加したフィールド
	)
	
	if err != nil {
		return nil, fmt.Errorf("scenario not found: %w", err)
	}

	// 取得した値を構造体にセット
	s.StatEffect = statEffect
	if conditionStat.Valid { s.ConditionStat = &conditionStat.String }
	if successText.Valid { s.SuccessText = &successText.String }
	if failureText.Valid { s.FailureText = &failureText.String }
	
	return &s, nil
}