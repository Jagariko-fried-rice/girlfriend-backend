package repository

import (
	"context"
	"database/sql"
	
	// ★以下の2行が必要です
	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type partnerRepository struct {
	db *sql.DB
}

func NewPartnerRepository(db *sql.DB) repository.PartnerRepository {
	return &partnerRepository{db: db}
}

func (r *partnerRepository) UpdateStatus(ctx context.Context, partnerID string, stamina, intelligence, sense int) error {
	query := `
		UPDATE partners
		SET stamina = $1, intelligence = $2, sense = $3
		WHERE id = $4
	`
	_, err := r.db.ExecContext(ctx, query, stamina, intelligence, sense, partnerID)
	return err
}

func (r *partnerRepository) FindByID(ctx context.Context, partnerID string) (*model.Partner, error) {
	query := `
		SELECT id, user_id, name, current_stage, stamina, intelligence, sense
		FROM partners WHERE id = $1
	`
	var p model.Partner
	err := r.db.QueryRowContext(ctx, query, partnerID).Scan(
		&p.ID, &p.UserID, &p.Name, &p.CurrentStage, &p.Stamina, &p.Intelligence, &p.Sense,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}