package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type PartnerHandler struct {
	PartnerRepo repository.PartnerRepository
	MemoryRepo  repository.MemoryRepository
	VoiceRepo   repository.VoiceRepository
	ImageRepo   repository.PartnerImageRepository
}

func NewPartnerHandler(pRepo repository.PartnerRepository, mRepo repository.MemoryRepository, vRepo repository.VoiceRepository, iRepo repository.PartnerImageRepository) *PartnerHandler {
	return &PartnerHandler{PartnerRepo: pRepo, MemoryRepo: mRepo, VoiceRepo: vRepo, ImageRepo: iRepo}
}

// GetStatus godoc
// @Summary      パートナー情報の取得
// @Description  指定されたIDのパートナーのステータス情報を返します
// @Tags         partners
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Partner ID (UUID)"
// @Success      200  {object}  model.Partner
// @Router       /partners/{id} [get]
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

// GetMemories godoc
// @Summary      思い出の取得
// @Description  指定されたIDのパートナーの思い出ログ一覧を返します
// @Tags         partners
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Partner ID (UUID)"
// @Success      200  {array}   model.Memory
// @Router       /partners/{id}/memories [get]
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

// リクエストボディ受け取り用の構造体
type CreatePartnerRequest struct {
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Personality string `json:"personality"`
	HairColor   string `json:"hair_color"`
	VoiceType   string `json:"voice_type"`
}

// CreatePartner godoc
// @Summary      パートナー新規作成
// @Description  新しいパートナーを作成し、初期ステータスを登録します
// @Tags         partners
// @Accept       json
// @Produce      json
// @Param        request body handler.CreatePartnerRequest true "作成情報"
// @Success      201  {object}  model.Partner
// @Router       /partners [post]
func (h *PartnerHandler) CreatePartner(w http.ResponseWriter, r *http.Request) {
	var req CreatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid User ID format", http.StatusBadRequest)
		return
	}

	newPartner := &model.Partner{
		UserID:       uid,
		Name:         req.Name,
		Personality:  req.Personality,
		HairColor:    req.HairColor,
		VoiceType:    req.VoiceType,
		CurrentStage: "infancy",
		Stamina:      10,
		Intelligence: 10,
		Sense:        10,
	}

	if err := h.PartnerRepo.Create(r.Context(), newPartner); err != nil {
		http.Error(w, "Failed to create partner: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newPartner)
}

// GetVoices godoc
// @Summary      ボイス一覧の取得
// @Description  パートナーの性格に基づいたセリフと音声URLを取得します
// @Tags         partners
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Partner ID (UUID)"
// @Success      200  {array}   model.VoiceLine
// @Router       /partners/{id}/voices [get]
func (h *PartnerHandler) GetVoices(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	partnerID := parts[2]

	// 1. パートナーの性格を取得
	partner, err := h.PartnerRepo.FindByID(r.Context(), partnerID)
	if err != nil {
		http.Error(w, "Partner not found", http.StatusNotFound)
		return
	}

	// 2. ボイス一覧を取得
	voices, err := h.VoiceRepo.FindByPersonality(r.Context(), partner.Personality)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(voices)
}

// GetImages godoc
// @Summary      生成画像一覧の取得
// @Description  パートナーの生成された画像リストを取得します
// @Tags         partners
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Partner ID (UUID)"
// @Success      200  {array}   model.PartnerImage
// @Router       /partners/{id}/images [get]
func (h *PartnerHandler) GetImages(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	partnerID := parts[2]

	images, err := h.ImageRepo.FindByPartnerID(r.Context(), partnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(images)
}