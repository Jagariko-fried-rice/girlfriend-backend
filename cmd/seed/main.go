package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

	files, err := filepath.Glob("seeds/*.csv")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fileName := filepath.Base(file)
		fmt.Printf("処理中: %s ... ", fileName)

		if strings.Contains(fileName, "scenarios") {
			if err := importScenarios(db, file); err != nil {
				log.Printf("失敗: %v\n", err)
			} else {
				fmt.Println("成功")
			}
		} else if strings.Contains(fileName, "voice_lines") {
			if err := importVoiceLines(db, file); err != nil {
				log.Printf("失敗: %v\n", err)
			} else {
				fmt.Println("成功")
			}
		} else {
			fmt.Println("スキップ")
		}
	}
	fmt.Println("全ての処理が完了しました。")
}

// ファイルコピー用のヘルパー関数
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil { return err }
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil { return err }
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ボイスデータのインポート関数（ファイルコピー版）
func importVoiceLines(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil { return err }
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil { return err } // ヘッダー読み飛ばし

	records, err := reader.ReadAll()
	if err != nil { return err }

	tx, err := db.Begin()
	if err != nil { return err }

	_, err = tx.Exec("TRUNCATE TABLE voice_lines")
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO voice_lines (personality, situation, line_text, audio_url)
		VALUES ($1, $2, $3, $4)
	`)
	if err != nil { return err }
	defer stmt.Close()

	// ★修正ポイント: ファイルの置き場所
	// アップロードされた構造に合わせて seeds/assets 直下に設定
	// もし seeds/assets/voice フォルダに入れているなら "seeds/assets/voice" にしてください
	sourceDir := "seeds/assets/voice" 
	
	outputDir := "output_audio"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	for _, record := range records {
		// CSV: personality, situation, line_text, audio_filename
		personality := record[0]
		situation := record[1]
		text := record[2]
		fileName := ""
		if len(record) > 3 {
			fileName = record[3]
		}

		dbAudioPath := "" 

		if fileName != "" {
			srcPath := filepath.Join(sourceDir, fileName)
			dstPath := filepath.Join(outputDir, fileName)

			// ファイルが存在するか確認してからコピー
			if _, err := os.Stat(srcPath); err == nil {
				if err := copyFile(srcPath, dstPath); err != nil {
					log.Printf("警告: コピー失敗 (%s): %v\n", fileName, err)
				} else {
					dbAudioPath = "/audio/" + fileName
					fmt.Printf("登録: %s -> %s\n", personality, fileName)
				}
			} else {
				log.Printf("警告: 音声ファイルが見つかりません (%s)\n", srcPath)
			}
		}

		_, err := stmt.Exec(personality, situation, text, dbAudioPath)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("DB insert error: %w", err)
		}
	}

	return tx.Commit()
}

// シナリオデータのインポート処理
func importScenarios(db *sql.DB, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil { return err }
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil { return err }

	records, err := reader.ReadAll()
	if err != nil { return err }

	tx, err := db.Begin()
	if err != nil { return err }

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
	if err != nil { return err }
	defer stmt.Close()

	for i, record := range records {
		weight, _ := strconv.Atoi(record[4])
		
		condVal := 0
		if len(record) > 6 && record[6] != "" {
			condVal, _ = strconv.Atoi(record[6])
		}

		if len(record) > 3 && !json.Valid([]byte(record[3])) {
			tx.Rollback()
			return fmt.Errorf("%d行目のJSON形式が不正です: %s", i+2, record[3])
		}

		var condStat, succText, failText, succEff, failEff interface{}
		
		if len(record) > 5 && record[5] == "" { condStat = nil } else if len(record) > 5 { condStat = record[5] }
		if len(record) > 7 && record[7] == "" { succText = nil } else if len(record) > 7 { succText = record[7] }
		if len(record) > 8 && record[8] == "" { failText = nil } else if len(record) > 8 { failText = record[8] }
		if len(record) > 9 && record[9] == "" { succEff = "{}" } else if len(record) > 9 { succEff = record[9] }
		if len(record) > 10 && record[10] == "" { failEff = "{}" } else if len(record) > 10 { failEff = record[10] }

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