package repository

import (
	"context"
	"database/sql"
	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type voiceRepository struct {
	db *sql.DB
}

func NewVoiceRepository(db *sql.DB) repository.VoiceRepository {
	return &voiceRepository{db: db}
}

func (r *voiceRepository) FindByPersonality(ctx context.Context, personality string) ([]*model.VoiceLine, error) {
	query := `
		SELECT id, personality, situation, line_text, COALESCE(audio_url, '')
		FROM voice_lines
		WHERE personality = $1
	`
	rows, err := r.db.QueryContext(ctx, query, personality)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var voices []*model.VoiceLine
	for rows.Next() {
		var v model.VoiceLine
		if err := rows.Scan(&v.ID, &v.Personality, &v.Situation, &v.LineText, &v.AudioURL); err != nil {
			return nil, err
		}
		voices = append(voices, &v)
	}
	return voices, nil
}