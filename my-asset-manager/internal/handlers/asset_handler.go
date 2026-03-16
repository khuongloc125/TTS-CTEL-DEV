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

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type AssetHandler struct {
	DB *gorm.DB
}

func (h *AssetHandler) GetStats(w http.ResponseWriter, r *http.Request) {
    var totalAssets, activeAssets, totalScans, completedScans int64

    h.DB.Model(&models.Asset{}).Count(&totalAssets)
    h.DB.Model(&models.Asset{}).Where("status = ?", "active").Count(&activeAssets)

    h.DB.Model(&models.ScanJob{}).Count(&totalScans)
    h.DB.Model(&models.ScanJob{}).Where("status = ?", "completed").Count(&completedScans)

    response := map[string]interface{}{
        "totalAssets":    totalAssets,
        "activeAssets":   activeAssets,
        "totalScans":     totalScans,
        "completedScans": completedScans,
        "byType":   h.getCountsByGroup("type"),
        "byStatus": h.getCountsByGroup("status"),
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (h *AssetHandler) getCountsByGroup(column string) map[string]int64 {
    results := make(map[string]int64)
    var queryResults []struct {
        Key   string `gorm:"column:key"`
        Count int64
    }
    
    h.DB.Model(&models.Asset{}).Select(column + " as `key`, count(*) as count").Group(column).Scan(&queryResults)
    
    for _, res := range queryResults {
        results[res.Key] = res.Count
    }
    return results
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

func (h *AssetHandler) SearchAssets(w http.ResponseWriter, r *http.Request) {

	queryParam := r.URL.Query().Get("q")
	if queryParam == "" {
		http.Error(w, "Search query 'q' is required", http.StatusBadRequest)
		return
	}

	var assets []models.Asset

	searchTerm := "%" + queryParam + "%"
	
	err := h.DB.Limit(100).
		Where("name LIKE ?", searchTerm).
		Find(&assets).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assets)
}

func (h *AssetHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
    var asset models.Asset
    
    if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
        http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
        return
    }

    asset.ID = uuid.New().String()
    asset.CreatedAt = time.Now()
    
    if asset.Status == "" {
        asset.Status = "active"
    }

    if err := h.DB.Create(&asset).Error; err != nil {
        http.Error(w, "Lỗi lưu Database: "+err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(asset)
}

func (h *AssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var asset models.Asset
	if err := h.DB.First(&asset, "id = ?", vars["id"]).Error; err != nil {
		http.Error(w, "Asset not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(asset)
}

func (h *AssetHandler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var asset models.Asset
	if err := h.DB.First(&asset, "id = ?", vars["id"]).Error; err != nil {
		http.Error(w, "Asset not found", http.StatusNotFound)
		return
	}
	json.NewDecoder(r.Body).Decode(&asset)
	h.DB.Save(&asset)
	json.NewEncoder(w).Encode(asset)
}

func (h *AssetHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.DB.Delete(&models.Asset{}, "id = ?", vars["id"])
	w.WriteHeader(http.StatusNoContent)
}

func (h *AssetHandler) GetLatestDNS(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    assetID := vars["id"]

    var results []models.ScanResult
    err := h.DB.Table("scan_results"). // Thêm .Table để chỉ định rõ
        Joins("JOIN scan_jobs ON scan_jobs.id = scan_results.job_id").
        Where("scan_jobs.asset_id = ? AND scan_results.scan_type = ?", assetID, "dns").
        Order("scan_results.created_at DESC").
        Limit(1).
        Find(&results).Error

    w.Header().Set("Content-Type", "application/json")
    if err != nil || len(results) == 0 {
        w.Write([]byte(`{}`)) // Trả về rỗng thay vì 404
        return
    }
    json.NewEncoder(w).Encode(results[0])
}

func (h *AssetHandler) GetSubdomains(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID := vars["id"]

	var result models.ScanResult
	err := h.DB.Joins("JOIN scan_jobs ON scan_jobs.id = scan_results.job_id").
		Where("scan_jobs.asset_id = ? AND scan_results.scan_type = ?", assetID, "dns").
		Order("scan_results.created_at DESC").
		First(&result).Error

	if err != nil {
		http.Error(w, "Subdomain results not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

