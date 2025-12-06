package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"database/sql"
)

func main() {
	// 1. 設定読み込み
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is required")
	}

	// 2. DB接続
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 3. CSVファイルを開く
	// ※必要に応じてここを "seeds/osananazimi_scenarios.csv" などに変えてください
	//  今回は基本の scenarios.csv を読む設定にしています
	file, err := os.Open("seeds/scenarios.csv")
	if err != nil {
		log.Fatalf("CSVファイルの読み込みに失敗: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// ヘッダー行をスキップ
	if _, err := reader.Read(); err != nil {
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

	// --- ここが変更点！ (UPSERT処理) ---
	// ON CONFLICT (stage, routes) DO UPDATE ...
	// これにより、同じ stage/routes があれば内容を上書きし、なければ新規作成します。
	query := `
		INSERT INTO scenarios (
			stage, routes, template_text, stat_effect, weight,
			condition_stat, condition_value, success_text, failure_text, success_effect, failure_effect
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
			failure_effect = EXCLUDED.failure_effect;
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for i, record := range records {
		weight, _ := strconv.Atoi(record[4])
		
		// 数値変換（空文字対策）
		condVal := 0
		if record[6] != "" {
			condVal, _ = strconv.Atoi(record[6])
		}

		// JSONバリデーション
		if !json.Valid([]byte(record[3])) {
			log.Fatalf("%d行目のJSON形式が不正です: %s", i+2, record[3])
		}

		// NULL許容カラムの処理 (Goのゼロ値対策)
		var condStat, succText, failText, succEff, failEff interface{}
		
		if record[5] == "" { condStat = nil } else { condStat = record[5] }
		if record[7] == "" { succText = nil } else { succText = record[7] }
		if record[8] == "" { failText = nil } else { failText = record[8] }
		// Effect系は空文字なら "{}" (空のJSON)を入れるのが安全
		if record[9] == "" { succEff = "{}" } else { succEff = record[9] }
		if record[10] == "" { failEff = "{}" } else { failEff = record[10] }

		_, err := stmt.Exec(
			record[0], // stage
			record[1], // routes
			record[2], // template_text
			record[3], // stat_effect
			weight,    // weight
			condStat,  // condition_stat
			condVal,   // condition_value
			succText,  // success_text
			failText,  // failure_text
			succEff,   // success_effect
			failEff,   // failure_effect
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