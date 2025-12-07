package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	_ "girlfriend-backend/docs"

	httpSwagger "github.com/swaggo/http-swagger"

	"girlfriend-backend/internal/infrastructure/repository"
	"girlfriend-backend/internal/interface/handler"
)

// ★追加: API全体の設定コメント
// @title           Girlfriend Backend API
// @version         1.0
// @description     美少女育成ゲームのバックエンドAPIです。
// @host            localhost:8080
// @BasePath        /

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found")
	}

	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 依存関係の注入
	partnerRepo := repository.NewPartnerRepository(db)
	memoryRepo := repository.NewMemoryRepository(db)
	sleepRepo := repository.NewSleepLogRepository(db)
	voiceRepo := repository.NewVoiceRepository(db)
	imageRepo := repository.NewPartnerImageRepository(db)

	partnerHandler := handler.NewPartnerHandler(partnerRepo, memoryRepo, voiceRepo, imageRepo)
	sleepHandler := handler.NewSleepHandler(sleepRepo, partnerRepo)

	// ルーティング設定
	mux := http.NewServeMux()

	// エンドポイント定義
	mux.HandleFunc("GET /partners/{id}", partnerHandler.GetStatus)
	mux.HandleFunc("GET /partners/{id}/memories", partnerHandler.GetMemories)
	mux.HandleFunc("GET /partners/{id}/images", partnerHandler.GetImages)

	// 睡眠API
	mux.HandleFunc("POST /sleep/start", sleepHandler.StartSleep)
	mux.HandleFunc("POST /sleep/end", sleepHandler.EndSleep)
	//パートナー登録API
	mux.HandleFunc("POST /partners", partnerHandler.CreatePartner)

	// Swagger UIのエンドポイント
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	mux.HandleFunc("GET /partners/{id}/voices", partnerHandler.GetVoices)

	// 生成された画像ファイルを配信する設定
	// /images/xxxx.png にアクセスすると output_images フォルダの中身を表示
	fileServer := http.FileServer(http.Dir("./output_images"))
	mux.Handle("/images/", http.StripPrefix("/images/", fileServer))

	// 音声ファイルの配信
	// /audio/xxx.wav にアクセスすると output_audio フォルダの中身を返す
	audioServer := http.FileServer(http.Dir("./output_audio"))
	mux.Handle("/audio/", http.StripPrefix("/audio/", audioServer))

	// CORS設定（ミドルウェア）
	corsMux := enableCORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("API Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, corsMux); err != nil {
		log.Fatal(err)
	}
}

// enableCORS: フロントエンドからのアクセスを許可するミドルウェア
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
