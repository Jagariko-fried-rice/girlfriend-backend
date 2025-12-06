package repository

import (
	"context"
	"database/sql"
	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type memoryRepository struct {
	db *sql.DB
}

func NewMemoryRepository(db *sql.DB) repository.MemoryRepository {
	return &memoryRepository{db: db}
}

func (r *memoryRepository) Create(ctx context.Context, m *model.Memory) error {
	query := `
		INSERT INTO memories (partner_id, scenario_id, generated_prompt)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query, m.PartnerID, m.ScenarioID, m.GeneratedPrompt)
	return err
}