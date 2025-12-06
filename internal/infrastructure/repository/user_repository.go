package repository

import (
	"context"
	"database/sql"
	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAllWithPartner(ctx context.Context) ([]*model.UserWithPartner, error) {
	// ユーザーとパートナーを結合して取得
	query := `
		SELECT u.uid, p.id, p.name, p.current_stage
		FROM users u
		JOIN partners p ON u.uid = p.user_id
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*model.UserWithPartner
	for rows.Next() {
		var u model.UserWithPartner
		if err := rows.Scan(&u.UserID, &u.PartnerID, &u.PartnerName, &u.CurrentStage); err != nil {
			return nil, err
		}
		users = append(users, &u)
	}
	return users, nil
}