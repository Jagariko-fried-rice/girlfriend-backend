package repository

import "context"

type ImageGenerator interface {

	GenerateImage(ctx context.Context, prompt string) (string, error)
}