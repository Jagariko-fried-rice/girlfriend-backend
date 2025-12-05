package main

import (
	"log"
	"os"

	// godotenv をインポート
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	// 1. .envファイルを読み込む
	// Load() は .env ファイルを探して読み込みます。ファイルがない場合などはエラーを返します。
	if err := godotenv.Load(); err != nil {
		log.Println("注意: .envファイルが見つかりません。環境変数から読み込みます。")
		// 本番環境（クラウド上）では.envファイルを使わず、サーバーの設定機能を使うことがあるため、
		// ここではFatal(強制終了)にせず、ログを出すだけに留めるのが一般的です。
	}

	// 2. 環境変数からDB接続URLを取得
	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		log.Fatal("エラー: 環境変数 DB_URL が設定されていません")
	}

	// 3. マイグレーションの準備
	m, err := migrate.New(
		"file://db/migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("マイグレーションの準備に失敗しました: %v", err)
	}

	// 4. マイグレーションの実行（Up）
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("マイグレーションの実行に失敗しました: %v", err)
	}

	log.Println("成功！データベースの作成が完了しました！")
}