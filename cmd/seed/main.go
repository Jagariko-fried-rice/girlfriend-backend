package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
		log.Fatal(err)
	}
	defer db.Close()

	file, err := os.Open("seeds/scenarios.csv")
	if err != nil {
		log.Fatalf("CSVファイルの読み込みに失敗: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil { // ヘッダー読み飛ばし
		log.Fatal(err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d 件のシナリオデータを処理します...\n", len(records))

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// 変更点: image_prompt ($12) を追加
	query := `
		INSERT INTO scenarios (
			stage, routes, template_text, stat_effect, weight,
			condition_stat, condition_value, success_text, failure_text, success_effect, failure_effect,
			image_prompt
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (stage, routes) 
		DO UPDATE SET
			template_text = EXCLUDED.template_text,
			stat_effect = EXCLUDED.stat_effect,
			weight = EXCLUDED.weight,
			condition_stat = EXCLUDED.condition_stat,
			condition_value = EXCLUDED.condition_value,
			success_text = EXCLUDED.success_text,
			failure_text = EXCLUDED.failure_text,
			success_effect = EXCLUDED.success_effect,
			failure_effect = EXCLUDED.failure_effect,
			image_prompt = EXCLUDED.image_prompt;
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for i, record := range records {
		weight, _ := strconv.Atoi(record[4])
		
		condVal := 0
		if record[6] != "" {
			condVal, _ = strconv.Atoi(record[6])
		}

		if !json.Valid([]byte(record[3])) {
			log.Fatalf("%d行目のJSON形式が不正です: %s", i+2, record[3])
		}

		var condStat, succText, failText, succEff, failEff interface{}
		
		if record[5] == "" { condStat = nil } else { condStat = record[5] }
		if record[7] == "" { succText = nil } else { succText = record[7] }
		if record[8] == "" { failText = nil } else { failText = record[8] }
		if record[9] == "" { succEff = "{}" } else { succEff = record[9] }
		if record[10] == "" { failEff = "{}" } else { failEff = record[10] }

		// 追加: 12列目(image_prompt)の取得処理
		imagePrompt := ""
		if len(record) > 11 {
			imagePrompt = record[11]
		}
		// もし空なら、日本語テキストをコピー（エラー回避）
		if imagePrompt == "" {
			imagePrompt = record[2]
		}

		_, err := stmt.Exec(
			record[0], record[1], record[2], record[3], weight,
			condStat, condVal, succText, failText, succEff, failEff,
			imagePrompt, // 追加 ($12)
		)
		if err != nil {
			tx.Rollback()
			log.Fatalf("%d行目の処理でエラー: %v", i+2, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("成功！データの更新・登録が完了しました。")
}