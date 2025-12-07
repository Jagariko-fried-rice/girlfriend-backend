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

// ... (既存の importScenarios 関数はそのまま残してください) ...
// ※長くなるので省略しますが、元の importScenarios 関数も必ずファイル内に残してください
func importScenarios(db *sql.DB, filePath string) error {
    // (元のコードと同じ内容)
    file, err := os.Open(filePath)
    // ...
    // ...
    return tx.Commit()
}