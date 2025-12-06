package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"girlfriend-backend/internal/infrastructure/repository"
	"girlfriend-backend/internal/interface/handler"
)

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
	partnerHandler := handler.NewPartnerHandler(partnerRepo, memoryRepo)

	// ルーティング設定
	mux := http.NewServeMux()

	// エンドポイント定義
	mux.HandleFunc("GET /partners/{id}", partnerHandler.GetStatus)
	mux.HandleFunc("GET /partners/{id}/memories", partnerHandler.GetMemories)

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