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
		// ★追加: ボイス用CSVの処理
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

// ★追加: ボイスデータのインポート関数
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

	// 毎回洗い替え（全削除して入れ直し）
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

	// ファイル操作の準備
	sourceDir := "seeds/assets/voice" // 元ファイル置き場
	outputDir := "output_audio"       // 配信フォルダ
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.Mkdir(outputDir, 0755)
	}

	for _, record := range records {
		personality := record[0]
		situation := record[1]
		text := record[2]
		fileName := ""
		if len(record) > 3 {
			fileName = record[3]
		}

		dbAudioPath := "" 

		// 音声ファイルがあればコピー
		if fileName != "" {
			srcPath := filepath.Join(sourceDir, fileName)
			dstPath := filepath.Join(outputDir, fileName)

			if _, err := os.Stat(srcPath); err == nil {
				if err := copyFile(srcPath, dstPath); err != nil {
					log.Printf("警告: コピー失敗 (%s): %v\n", fileName, err)
				} else {
					dbAudioPath = "/audio/" + fileName // URLパスとして保存
				}
			} else {
				log.Printf("警告: 音声ファイルなし (%s)\n", srcPath)
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