package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"girlfriend-backend/internal/domain/model"
	"girlfriend-backend/internal/domain/repository"
)

type SleepHandler struct {
	SleepRepo   repository.SleepLogRepository
	PartnerRepo repository.PartnerRepository // ユーザーID取得用
}

func NewSleepHandler(sRepo repository.SleepLogRepository, pRepo repository.PartnerRepository) *SleepHandler {
	return &SleepHandler{SleepRepo: sRepo, PartnerRepo: pRepo}
}

// リクエストボディの定義
type SleepRequest struct {
	PartnerID string `json:"partner_id"`
}

// StartSleep godoc
// @Summary      おやすみ（睡眠開始）
// @Description  指定されたパートナーの睡眠計測を開始します
// @Tags         sleep
// @Accept       json
// @Produce      json
// @Param        request body handler.SleepRequest true "パートナーID"
// @Success      200  {object}  model.SleepLog
// @Router       /sleep/start [post]
func (h *SleepHandler) StartSleep(w http.ResponseWriter, r *http.Request) {
	var req SleepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 1. 既に寝ていないかチェック
	activeLog, err := h.SleepRepo.FindUnfinished(r.Context(), req.PartnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if activeLog != nil {
		http.Error(w, "Already sleeping!", http.StatusConflict) // 409 Conflict
		return
	}

	// 2. パートナー情報からUserIDを取得（紐付けのため）
	partner, err := h.PartnerRepo.FindByID(r.Context(), req.PartnerID)
	if err != nil {
		http.Error(w, "Partner not found", http.StatusBadRequest)
		return
	}

	// 3. ログ作成
	newLog := &model.SleepLog{
		UserID:    partner.UserID,
		PartnerID: partner.ID,
		SleptAt:   time.Now(), // 現在時刻
	}

	if err := h.SleepRepo.Create(r.Context(), newLog); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newLog)
}

// EndSleep godoc
// @Summary      おはよう（起床）
// @Description  睡眠を終了し、睡眠時間を計算して記録します
// @Tags         sleep
// @Accept       json
// @Produce      json
// @Param        request body handler.SleepRequest true "パートナーID"
// @Success      200  {object}  model.SleepLog
// @Router       /sleep/end [post]
func (h *SleepHandler) EndSleep(w http.ResponseWriter, r *http.Request) {
	var req SleepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 1. 寝ているデータを探す
	activeLog, err := h.SleepRepo.FindUnfinished(r.Context(), req.PartnerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if activeLog == nil {
		http.Error(w, "Not sleeping", http.StatusBadRequest)
		return
	}

	// 2. 時間計算
	now := time.Now()
	duration := now.Sub(activeLog.SleptAt)
	minutes := int(duration.Minutes())

	// 3. 更新データ作成
	updateData := model.SleepLog{
		WakeAt:       &now,
		SleepMinutes: &minutes,
	}

	if err := h.SleepRepo.UpdateWakeTime(r.Context(), activeLog.ID.String(), updateData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// レスポンス用にデータを完成させる
	activeLog.WakeAt = &now
	activeLog.SleepMinutes = &minutes

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activeLog)
}