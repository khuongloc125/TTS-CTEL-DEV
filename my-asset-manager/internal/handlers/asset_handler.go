package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"my-asset-manager/internal/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AssetHandler struct {
	DB *gorm.DB
}

func (h *AssetHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	var stats models.StatsResponse
	stats.ByType = make(map[string]int64)
	stats.ByStatus = make(map[string]int64)

	h.DB.Model(&models.Asset{}).Count(&stats.Total)

	var typeResults []struct {
		Type  string
		Count int64
	}
	h.DB.Model(&models.Asset{}).Select("type, count(*) as count").Group("type").Scan(&typeResults)
	for _, res := range typeResults {
		stats.ByType[res.Type] = res.Count
	}

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

func (h *AssetHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	var req models.BatchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Assets) > 100 {
		http.Error(w, "Limit exceeded: maximum 100 assets per request", http.StatusBadRequest)
		return
	}

	createdIDs := []string{}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		for _, asset := range req.Assets {
			if !isValidType(asset.Type) {
				return fmt.Errorf("invalid asset type: %s for asset: %s", asset.Type, asset.Name)
			}
            
            if asset.Name == "" {
                return fmt.Errorf("asset name cannot be empty")
            }

			if err := tx.Create(&asset).Error; err != nil {
				return err 
			}
			
			createdIDs = append(createdIDs, asset.ID)
		}
		return nil 
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.BatchCreateResponse{
		Created: len(createdIDs),
		IDs:     createdIDs,
	})
}

func isValidType(t string) bool {
	switch t {
	case "domain", "ip", "service":
		return true
	}
	return false
}

func (h *AssetHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		http.Error(w, "ids parameter is required", http.StatusBadRequest)
		return
	}

	ids := strings.Split(idsParam, ",")
	totalRequested := int64(len(ids))

	result := h.DB.Where("id IN ?", ids).Delete(&models.Asset{})
	
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	deletedCount := result.RowsAffected
	notFoundCount := totalRequested - deletedCount

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.BatchDeleteResponse{
		Deleted:  deletedCount,
		NotFound: notFoundCount,
	})
}

func (h *AssetHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Lấy đối tượng sql.DB từ GORM
	sqlDB, err := h.DB.DB()
	
	health := models.HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err != nil || sqlDB.Ping() != nil {
		health.Status = "degraded"
		health.Database.Status = "disconnected"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		stats := sqlDB.Stats()
		health.Database = models.DatabaseStatus{
			Status:          "connected",
			OpenConnections: stats.OpenConnections,
			InUse:           stats.InUse,
			Idle:            stats.Idle,
			MaxOpen:         stats.MaxOpenConnections,
		}
		w.WriteHeader(http.StatusOK) 
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 20 }
	if limit > 100 { limit = 100 } 

	assetType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")

	query := h.DB.Model(&models.Asset{})

	if assetType != "" {
		query = query.Where("type = ?", assetType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var assets []models.Asset
	offset := (page - 1) * limit

	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&assets).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	resp := models.AssetListResponse{
		Data: assets,
		Pagination: models.Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}