package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 2. seedsフォルダ内の全CSVファイルを取得
	files, err := filepath.Glob("seeds/*.csv")
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Println("seedsフォルダにCSVファイルが見つかりません。")
		return
	}

	fmt.Printf("%d 個のCSVファイルが見つかりました。\n", len(files))

	// 3. ファイルごとのループ処理
	for _, file := range files {
		fileName := filepath.Base(file)
		fmt.Printf("処理中: %s ... ", fileName)

		// ファイル名で処理を振り分け (Dispatcher)
		if strings.Contains(fileName, "scenarios") {
			if err := importScenarios(db, file); err != nil {
				log.Printf("失敗: %v\n", err)
			} else {
				fmt.Println("成功")
			}
		} else {
			// 将来他のテーブル（items.csvなど）が増えたらここに追加
			fmt.Println("スキップ (対応するインポーターがありません)")
		}
	}
	fmt.Println("全ての処理が完了しました。")
}

// importScenarios: シナリオデータのインポート処理（ロジックを分離）
func importScenarios(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil { // ヘッダー読み飛ばし
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// ImagePrompt対応済みのクエリ
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
		return err
	}
	defer stmt.Close()

	for i, record := range records {
		weight, _ := strconv.Atoi(record[4])
		
		condVal := 0
		if record[6] != "" {
			condVal, _ = strconv.Atoi(record[6])
		}

		if !json.Valid([]byte(record[3])) {
			tx.Rollback()
			return fmt.Errorf("%d行目のJSON形式が不正です: %s", i+2, record[3])
		}

		var condStat, succText, failText, succEff, failEff interface{}
		
		if record[5] == "" { condStat = nil } else { condStat = record[5] }
		if record[7] == "" { succText = nil } else { succText = record[7] }
		if record[8] == "" { failText = nil } else { failText = record[8] }
		if record[9] == "" { succEff = "{}" } else { succEff = record[9] }
		if record[10] == "" { failEff = "{}" } else { failEff = record[10] }

		imagePrompt := ""
		if len(record) > 11 {
			imagePrompt = record[11]
		}
		if imagePrompt == "" {
			imagePrompt = record[2]
		}

		_, err := stmt.Exec(
			record[0], record[1], record[2], record[3], weight,
			condStat, condVal, succText, failText, succEff, failEff,
			imagePrompt,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("%d行目でDBエラー: %w", i+2, err)
		}
	}

	return tx.Commit()
}