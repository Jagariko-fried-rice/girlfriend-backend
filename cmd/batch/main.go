package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // DBドライバ

	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
	"girlfriend-backend/internal/infrastructure/external"
	dbRepo "girlfriend-backend/internal/infrastructure/repository" // 名前が被るので別名をつける
)

func main() {
	// 1. 環境設定の読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	// 2. データベース接続
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("DB接続エラー: %v", err)
	}
	defer db.Close()

	// 3. 依存関係の組み立て (Dependency Injection)
	// 「DBリポジトリ」と「AIクライアント」を用意します
	var imageRepo repository.PartnerImageRepository
	imageRepo = dbRepo.NewPartnerImageRepository(db)

	var generator repository.ImageGenerator
	// ※APIキーは一旦ダミーですが、ここを本番のキー(os.Getenv("GOOGLE_API_KEY"))に変えれば本番化できます
	generator = external.NewImagenClient("DUMMY_API_KEY")

	// 4. バッチ処理の実行
	ctx := context.Background()
	fmt.Println("バッチ処理を開始します...")

	// ---------------------------------------------------------
	// A. 未生成のデータを1件探す
	// ---------------------------------------------------------
	targetImage, err := imageRepo.FindFirstPending(ctx)
	if err != nil {
		log.Fatalf("データ検索中にエラーが発生: %v", err)
	}

	if targetImage == nil {
		fmt.Println("現在、生成待ち(pending)の画像はありません。")
		return
	}

	fmt.Printf("生成対象を発見: ID=%s, Prompt=%s\n", targetImage.ID, targetImage.GenerationPrompt)

	// ---------------------------------------------------------
	// B. 画像生成を実行
	// ---------------------------------------------------------
	// 30秒でタイムアウトするように設定
	genCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url, err := generator.GenerateImage(genCtx, targetImage.GenerationPrompt)
	if err != nil {
		log.Printf("画像生成失敗: %v", err)
		targetImage.Status = model.ImageStatusFailed
		errorMsg := err.Error()
		targetImage.ErrorMessage = &errorMsg
	} else {
		fmt.Println("生成成功！ URL:", url)
		targetImage.Status = model.ImageStatusCompleted
		targetImage.ImageURL = &url
	}

	// ---------------------------------------------------------
	// C. 結果を保存
	// ---------------------------------------------------------
	if err := imageRepo.Update(ctx, targetImage); err != nil {
		log.Fatalf("結果の保存に失敗: %v", err)
	}

	fmt.Println("DBの更新が完了しました。")
}