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

func (r *memoryRepository) FindByPartnerID(ctx context.Context, partnerID string) ([]*model.Memory, error) {
	query := `
		SELECT id, partner_id, scenario_id, generated_prompt, occurred_at
		FROM memories
		WHERE partner_id = $1
		ORDER BY occurred_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, partnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*model.Memory
	for rows.Next() {
		var m model.Memory
		if err := rows.Scan(&m.ID, &m.PartnerID, &m.ScenarioID, &m.GeneratedPrompt, &m.OccurredAt); err != nil {
			return nil, err
		}
		memories = append(memories, &m)
	}
	return memories, nil
}