package main

import (
	"context"
	"database/sql"
	"encoding/json"
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
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 全リポジトリの準備
	userRepo := dbRepo.NewUserRepository(db)
	scenarioRepo := dbRepo.NewScenarioRepository(db)
	imageRepo := dbRepo.NewPartnerImageRepository(db)
	partnerRepo := dbRepo.NewPartnerRepository(db) // 追加
	memoryRepo := dbRepo.NewMemoryRepository(db)   // 追加

	ctx := context.Background()
	fmt.Println("--- 固定イベント生成バッチ (ステータス変動あり) ---")

	// 1. 今回のターゲット（大人・おはよう）
	targetStage := "adult"
	targetRoute := "osananajimi_good_morning"

	scenario, err := scenarioRepo.FindByStageAndRoute(ctx, targetStage, targetRoute)
	if err != nil {
		log.Fatalf("シナリオが見つかりません: %v", err)
	}
	fmt.Printf("シナリオ「%s」を実行します。\n", scenario.Routes)

	// 2. ユーザー取得
	users, err := userRepo.FindAllWithPartner(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		fmt.Printf("User: %s (Stamina:%d, Intel:%d, Sense:%d) -> ", u.PartnerName, u.Stamina, u.Intelligence, u.Sense)

		// A. ステータスの計算 (Growth)
		// CSVに入っているJSON (例: {"stamina": 10}) を読み取って計算します
		// ※今回は強制的に「成功」扱いとして success_effect を使います
		var effects map[string]int
		if err := json.Unmarshal([]byte(scenario.StatEffect), &effects); err != nil {
			// StatEffectが空ならSuccessEffectを使うなどの分岐もここで可能
			// 今回は seeds/scenarios.csv の image_prompt ありの行の stat_effect は "{}" になっているので、
			// success_effect を使うようにします
			json.Unmarshal([]byte(scenario.SuccessEffect), &effects) // エラー無視(簡易)
		}
		// もしStatEffectも空ならSuccessEffectを見る
		if len(effects) == 0 {
			json.Unmarshal([]byte(scenario.SuccessEffect), &effects)
		}

		newStamina := u.Stamina + effects["stamina"]
		newIntel := u.Intelligence + effects["intelligence"]
		newSense := u.Sense + effects["sense"]

		// B. DB更新: パートナー (Update)
		if err := partnerRepo.UpdateStatus(ctx, u.PartnerID.String(), newStamina, newIntel, newSense); err != nil {
			log.Printf("ステータス更新失敗: %v", err)
			continue
		}

		// C. テキスト作成
		// 成功テキストがあれば使い、なければテンプレートを使う
		baseText := scenario.TemplateText
		if scenario.SuccessText != nil && *scenario.SuccessText != "" {
			baseText = *scenario.SuccessText
		}
		finalText := strings.ReplaceAll(baseText, "{{name}}", u.PartnerName)
		finalText = strings.ReplaceAll(finalText, "{{user_name}}", "あなた")

		// D. DB更新: 思い出 (Insert)
		memory := &model.Memory{
			PartnerID:       u.PartnerID,
			ScenarioID:      scenario.ID,
			GeneratedPrompt: finalText,
		}
		if err := memoryRepo.Create(ctx, memory); err != nil {
			log.Printf("思い出記録失敗: %v", err)
			// 思い出失敗しても画像生成は進める
		}

		// E. DB更新: 画像予約 (Insert)
		newImage := &model.PartnerImage{
			PartnerID:        u.PartnerID,
			Stage:            "adult", // 強制大人
			GenerationPrompt: scenario.ImagePrompt,
			Status:           model.ImageStatusPending,
		}
		if err := imageRepo.Create(ctx, newImage); err != nil {
			log.Printf("予約失敗: %v", err)
		} else {
			fmt.Printf("完了! 新ステータス(Sta:%d, Int:%d, Sen:%d)\n", newStamina, newIntel, newSense)
		}
	}

	fmt.Println("--- 完了 ---")
}