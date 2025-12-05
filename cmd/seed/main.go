package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // DBドライバ
	"database/sql"
)

// Scenario はCSVの1行に対応する構造体
type Scenario struct {
	Stage        string
	Routes       string
	TemplateText string
	StatEffect   string // JSON形式の文字列として読み込む
	Weight       int
}

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
	file, err := os.Open("seeds/scenarios.csv")
	if err != nil {
		log.Fatalf("CSVファイルの読み込みに失敗: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// ヘッダー行（1行目）をスキップ
	if _, err := reader.Read(); err != nil {
		log.Fatal(err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%d 件のシナリオデータをインポートします...\n", len(records))

	// 4. データの挿入（トランザクション）
	// 途中でエラーが出たら全部取り消せるように「トランザクション」を使う
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO scenarios (stage, routes, template_text, stat_effect, weight)
		VALUES ($1, $2, $3, $4, $5)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for i, record := range records {
		// CSVのデータをGoの型に変換
		weight, _ := strconv.Atoi(record[4])

		// JSONのバリデーション（念のため）
		if !json.Valid([]byte(record[3])) {
			log.Fatalf("%d行目のJSON形式が不正です: %s", i+2, record[3])
		}

		_, err := stmt.Exec(
			record[0], // stage
			record[1], // routes
			record[2], // template_text
			record[3], // stat_effect (JSON string)
			weight,    // weight
		)
		if err != nil {
			tx.Rollback() // 失敗したら全部なかったことにする
			log.Fatalf("%d行目の挿入でエラー: %v", i+2, err)
		}
	}

	// 5. コミット（確定）
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("成功！全てのシナリオデータがDBに保存されました。")
}