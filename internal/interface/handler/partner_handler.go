package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"girlfriend-backend/internal/domain/repository"
)

type PartnerHandler struct {
	PartnerRepo repository.PartnerRepository
	MemoryRepo  repository.MemoryRepository
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
