package repository

import (
	"context"
	"girlfriend-backend/internal/domain/model"
)

type MemoryRepository interface {
	// 思い出を記録する
	Create(ctx context.Context, memory *model.Memory) error
}