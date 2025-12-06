package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"girlfriend-backend/internal/domain/repository"

	"github.com/google/uuid"
	"girlfriend-backend/internal/domain/model"
)

type PartnerHandler struct {
	PartnerRepo repository.PartnerRepository
	MemoryRepo  repository.MemoryRepository
}
type CreatePartnerRequest struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Personality string `json:"personality"`
	HairColor   string `json:"hair_color"`
	VoiceType   string `json:"voice_type"`
}

func NewPartnerHandler(pRepo repository.PartnerRepository, mRepo repository.MemoryRepository) *PartnerHandler {
	return &PartnerHandler{PartnerRepo: pRepo, MemoryRepo: mRepo}
}

// GetStatus: パートナーのステータスを取得
// GET /partners/{id}
func (h *PartnerHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	partnerID := parts[len(parts)-1]

	partner, err := h.PartnerRepo.FindByID(r.Context(), partnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(partner)
}

// GetMemories: 思い出リストを取得
// GET /partners/{id}/memories
func (h *PartnerHandler) GetMemories(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	partnerID := parts[2]

	memories, err := h.MemoryRepo.FindByPartnerID(r.Context(), partnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(memories)
}

func (h *PartnerHandler) CreatePartner(w http.ResponseWriter, r *http.Request) {
	var req CreatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// UUIDの変換など
	// ※簡略化のため、エラーチェックを一部省略していますが、実際はuuid.Parseのエラー確認が必要です
	// import "github.com/google/uuid" が必要です
	
	// 初期値の設定
	newPartner := &model.Partner{
		// UserID: uuid.MustParse(req.UserID), // ※uuidのインポートとパース処理が必要
		Name:         req.Name,
		Personality:  req.Personality,
		HairColor:    req.HairColor,
		VoiceType:    req.VoiceType,
		CurrentStage: "infancy", // 最初は「乳児期」からスタート
		Stamina:      10,        // 初期ステータス
		Intelligence: 10,
		Sense:        10,
	}
    
    // 文字列のUUIDを変換 (ハンドラー内で変換ロジックを入れる例)
    if uid, err := uuid.Parse(req.UserID); err == nil {
        newPartner.UserID = uid
    } else {
        http.Error(w, "Invalid User ID", http.StatusBadRequest)
        return
    }

	if err := h.PartnerRepo.Create(r.Context(), newPartner); err != nil {
		http.Error(w, "Failed to create partner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(newPartner)
}