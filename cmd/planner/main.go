package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"girlfriend-backend/internal/domain/model"
	dbRepo "girlfriend-backend/internal/infrastructure/repository"
)

func main() {
	godotenv.Load()
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// リポジトリの準備
	userRepo := dbRepo.NewUserRepository(db)
	scenarioRepo := dbRepo.NewScenarioRepository(db)
	imageRepo := dbRepo.NewPartnerImageRepository(db)

	ctx := context.Background()
	fmt.Println("--- 週次プランニングバッチ開始 ---")

	// 1. 全ユーザー取得
	users, err := userRepo.FindAllWithPartner(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("対象ユーザー: %d名\n", len(users))

	// 2. 各ユーザーにシナリオを割り当て
	for _, u := range users {
		// その子の成長段階に合ったシナリオをランダムに引く
		scenario, err := scenarioRepo.FindRandomByStage(ctx, u.CurrentStage)
		if err != nil {
			log.Printf("Skip: ユーザー %s のシナリオが見つかりません (stage: %s)", u.PartnerName, u.CurrentStage)
			continue
		}

		// 3. プロンプトの作成（変数置換）
		// CSVの {{name}} をパートナーの名前に置き換える
		prompt := strings.ReplaceAll(scenario.TemplateText, "{{name}}", u.PartnerName)
		// {{user_name}} は一旦固定値か、ユーザーテーブルから取得して置換する
		prompt = strings.ReplaceAll(prompt, "{{user_name}}", "あなた")

		// 4. DBに予約 (pending)
		newImage := &model.PartnerImage{
			PartnerID:        u.PartnerID,
			Stage:            u.CurrentStage,
			GenerationPrompt: prompt,
			Status:           model.ImageStatusPending,
		}

		if err := imageRepo.Create(ctx, newImage); err != nil {
			log.Printf("Error: %s さんの予約作成に失敗: %v", u.PartnerName, err)
		} else {
			fmt.Printf("予約完了: %s -> シナリオ「%s」\n", u.PartnerName, prompt)
		}
	}

	fmt.Println("--- プランニング完了 ---")
}