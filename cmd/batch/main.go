package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
	"girlfriend-backend/internal/infrastructure/external"
	dbRepo "girlfriend-backend/internal/infrastructure/repository"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("DB接続エラー: %v", err)
	}
	defer db.Close()

	var imageRepo repository.PartnerImageRepository
	imageRepo = dbRepo.NewPartnerImageRepository(db)

	// ★変更: Stable Diffusionクライアントを使用
	// ローカルのSD APIのURLを指定 (デフォルトは http://127.0.0.1:7860)
	sdAPI := os.Getenv("SD_API_URL")
	if sdAPI == "" {
		sdAPI = "http://127.0.0.1:7860"
	}

	// Supabase Storageの設定
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	bucketName := os.Getenv("SUPABASE_BUCKET")
	if bucketName == "" {
		bucketName = "images" // デフォルトバケット名
	}

	var storageClient *external.SupabaseStorageClient
	if supabaseURL != "" && supabaseKey != "" {
		storageClient = external.NewSupabaseStorageClient(supabaseURL, supabaseKey, bucketName)
		fmt.Println("Supabase Storageを使用します")
	} else {
		fmt.Println("注意: Supabase設定が見つからないため、ローカル保存モードで動作します")
	}

	var generator repository.ImageGenerator
	generator = external.NewStableDiffusionClient(sdAPI, storageClient)

	// 実行
	ctx := context.Background()
	fmt.Println("バッチ処理を開始します...")

	// 1. 未生成データの取得
	targetImage, err := imageRepo.FindFirstPending(ctx)
	if err != nil {
		log.Fatalf("データ検索エラー: %v", err)
	}

	if targetImage == nil {
		fmt.Println("現在、生成待ち(pending)の画像はありません。")
		return
	}

	fmt.Printf("生成対象を発見: Prompt=%s\n", targetImage.GenerationPrompt[:20]+"...")

	// 2. 画像生成 (Stable Diffusion)
	// SDは重いのでタイムアウトを長めに(5分)
	genCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	url, err := generator.GenerateImage(genCtx, targetImage.GenerationPrompt)
	if err != nil {
		log.Printf("画像生成失敗: %v", err)
		targetImage.Status = model.ImageStatusFailed
		errorMsg := err.Error()
		targetImage.ErrorMessage = &errorMsg
	} else {
		fmt.Println("生成成功！ Path:", url)
		targetImage.Status = model.ImageStatusCompleted
		targetImage.ImageURL = &url
	}

	// 3. 結果保存
	if err := imageRepo.Update(ctx, targetImage); err != nil {
		log.Fatalf("結果保存失敗: %v", err)
	}

	fmt.Println("DB更新完了。")
}