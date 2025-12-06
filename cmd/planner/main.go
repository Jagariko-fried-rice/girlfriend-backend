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

// ★追加: 変換マップ（これを充実させると表現力が上がります）
var hairColorMap = map[string]string{
	"黒髪": "black hair",
	"金髪": "blonde hair",
	"銀髪": "silver hair",
	"茶髪": "brown hair",
	"青髪": "blue hair",
	"赤髪": "red hair",
}

var personalityMap = map[string]string{
	"元気":   "energetic, cheerful, big smile, open mouth",
	"おっとり": "gentle, relaxed, soft smile",
	"クール":  "cool, calm, sharp eyes, slight smile",
	"内気":   "shy, blushing, looking down",
}

func main() {
	// ... (DB接続などは既存のまま) ...
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := dbRepo.NewUserRepository(db)
	scenarioRepo := dbRepo.NewScenarioRepository(db)
	imageRepo := dbRepo.NewPartnerImageRepository(db)
	partnerRepo := dbRepo.NewPartnerRepository(db)
	memoryRepo := dbRepo.NewMemoryRepository(db)

	ctx := context.Background()
	fmt.Println("--- 固定イベント生成バッチ (ステータス・外見反映あり) ---")

	targetStage := "adult"
	targetRoute := "osananajimi_good_morning"

	scenario, err := scenarioRepo.FindByStageAndRoute(ctx, targetStage, targetRoute)
	if err != nil {
		log.Fatalf("シナリオが見つかりません: %v", err)
	}
	fmt.Printf("シナリオ「%s」を実行します。\n", scenario.Routes)

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
	// 1. 髪色変換
		hairPrompt, ok := hairColorMap[u.HairColor]
		if !ok {
			hairPrompt = "black hair" // デフォルト
		}
		// 2. 性格変換
		persPrompt, ok := personalityMap[u.Personality]
		if !ok {
			persPrompt = "normal expression" // デフォルト
		}

		// 3. 置換実行
		genPrompt := scenario.ImagePrompt
		genPrompt = strings.ReplaceAll(genPrompt, "{{hair_color}}", hairPrompt)
		genPrompt = strings.ReplaceAll(genPrompt, "{{personality}}", persPrompt)

		// E. DB更新: 画像予約
		newImage := &model.PartnerImage{
			PartnerID:        u.PartnerID,
			Stage:            "adult",
			GenerationPrompt: genPrompt, // ★置換後のプロンプトを使用
			Status:           model.ImageStatusPending,
		}
		if err := imageRepo.Create(ctx, newImage); err != nil {
			log.Printf("予約失敗: %v", err)
		} else {
			fmt.Println("完了! 画像予約を作成しました。")
		}
	}

	fmt.Println("--- 完了 ---")
}