package repository

import (
	"context"
	"girlfriend-backend/internal/domain/model"
)

type UserRepository interface {
	// 全ユーザーと、そのパートナー情報を取得
	FindAllWithPartner(ctx context.Context) ([]*model.UserWithPartner, error)
}