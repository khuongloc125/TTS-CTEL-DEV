package handlers

import (
	"encoding/json"
	"my-asset-manager/internal/models"
	"net/http"

	"gorm.io/gorm"
)

type AssetHandler struct {
	DB *gorm.DB
}

// 1.1 Thống kê Assets
func (h *AssetHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var stats models.StatsResponse
	stats.ByType = make(map[string]int64)
	stats.ByStatus = make(map[string]int64)

	// Đếm tổng số lượng
	h.DB.Model(&models.Asset{}).Count(&stats.Total)

	// Thống kê theo Type
	var typeResults []struct {
		Type  string
		Count int64
	}
	h.DB.Model(&models.Asset{}).Select("type, count(*) as count").Group("type").Scan(&typeResults)
	for _, res := range typeResults {
		stats.ByType[res.Type] = res.Count
	}

	// Thống kê theo Status
	var statusResults []struct {
		Status string
		Count  int64
	}
	h.DB.Model(&models.Asset{}).Select("status, count(*) as count").Group("status").Scan(&statusResults)
	for _, res := range statusResults {
		stats.ByStatus[res.Status] = res.Count
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 1.2 Đếm Assets có filter
func (h *AssetHandler) CountAssets(w http.ResponseWriter, r *http.Request) {
	assetType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	query := h.DB.Model(&models.Asset{})
	filters := make(map[string]string)

	if assetType != "" {
		query = query.Where("type = ?", assetType)
		filters["type"] = assetType
	}
	if status != "" {
		query = query.Where("status = ?", status)
		filters["status"] = status
	}

	var count int64
	query.Count(&count)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CountResponse{
		Count:   count,
		Filters: filters,
	})
}