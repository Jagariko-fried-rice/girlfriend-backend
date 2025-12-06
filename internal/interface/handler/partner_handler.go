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
}

func NewPartnerHandler(pRepo repository.PartnerRepository, mRepo repository.MemoryRepository) *PartnerHandler {
	return &PartnerHandler{PartnerRepo: pRepo, MemoryRepo: mRepo}
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