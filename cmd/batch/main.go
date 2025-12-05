package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"girlfriend-backend/internal/domain/repository"
	"girlfriend-backend/internal/infrastructure/external"
)

func main() {

	var generator repository.ImageGenerator
	generator = external.NewImagenClient("YOUR_API_KEY")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	targetPrompt := "黒髪の美少女, 学童期, 公園で遊んでいる"

	fmt.Println("バッチ処理を開始します...")

	imageURL, err := generator.GenerateImage(ctx, targetPrompt)
	if err != nil {
		log.Fatalf("画像生成に失敗しました: %v", err)
	}

	fmt.Printf("成功！ 画像URL: %s\n", imageURL)
	fmt.Println("バッチ処理が完了しました。")
}