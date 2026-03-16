package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"my-asset-manager/internal/models"
	"my-asset-manager/internal/scanner"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type ScanHandler struct {
	DB *gorm.DB
}

func (h *AssetHandler) StartScan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID := vars["id"]

	var asset models.Asset
	if err := h.DB.First(&asset, "id = ?", assetID).Error; err != nil {
		http.Error(w, "Asset not found", http.StatusNotFound)
		return
	}

	var req struct {
		ScanType string `json:"scan_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job := models.ScanJob{
		ID:        uuid.New().String(),
		AssetID:   assetID,
		ScanType:  req.ScanType,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := h.DB.Create(&job).Error; err != nil {
		http.Error(w, "Could not create scan job", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[+] Đã nhận request: %s cho %s (Job ID: %s)\n", job.ScanType, asset.Name, job.ID)

	go h.runScanEngine(job, asset)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

func (h *AssetHandler) GetScanStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	var job models.ScanJob
	if err := h.DB.First(&job, "id = ?", jobID).Error; err != nil {
		http.Error(w, "Scan job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *AssetHandler) GetScanResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	var job models.ScanJob
	if err := h.DB.First(&job, "id = ?", jobID).Error; err != nil {
		http.Error(w, "Scan job not found", http.StatusNotFound)
		return
	}

	var result models.ScanResult
	if err := h.DB.Where("job_id = ?", jobID).First(&result).Error; err != nil {
		if job.Status == "completed" {
			http.Error(w, "Detailed results not found", http.StatusNotFound)
		} else {
			http.Error(w, "Scan is not completed yet", http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{
        "job_id": "%s",
        "scan_type": "%s",
        "status": "%s",
        "results": %s
    }`, job.ID, job.ScanType, job.Status, result.ResultData)
}

func (h *AssetHandler) ListScansByAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var jobs []models.ScanJob
	h.DB.Where("asset_id = ?", vars["id"]).Order("created_at DESC").Find(&jobs)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (h *AssetHandler) GetAllAssetResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var results []models.ScanResult
	h.DB.Table("scan_results").
		Select("scan_results.*").
		Joins("JOIN scan_jobs ON scan_jobs.id = scan_results.job_id").
		Where("scan_jobs.asset_id = ?", vars["id"]).
		Scan(&results)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}


func (h *AssetHandler) GetLatestWhois(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID := vars["id"]

	var results []models.ScanResult
	
	err := h.DB.Table("scan_results").
		Joins("JOIN scan_jobs ON scan_jobs.id = scan_results.job_id").
		Where("scan_jobs.asset_id = ? AND scan_results.scan_type = ?", assetID, "whois").
		Order("scan_results.created_at DESC").
		Limit(1).     
		Find(&results).Error 

	w.Header().Set("Content-Type", "application/json")
	
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if len(results) == 0 {
		w.Write([]byte(`{}`)) 
		return
	}

	fmt.Fprintf(w, "%s", results[0].ResultData)
}


func (h *AssetHandler) runScanEngine(job models.ScanJob, asset models.Asset) {
	startTime := time.Now()
	fmt.Printf("[*] Job %s đang chạy...\n", job.ID)
	h.DB.Model(&models.ScanJob{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
		"status":     "running",
		"started_at": &startTime,
	})

	var s scanner.Scanner
	switch job.ScanType {
	case "ip":
		s = &scanner.IPScanner{}
	case "port":
		s = &scanner.PortScanner{}
	case "ssl":
		s = &scanner.SSLScanner{}
	case "tech":
		s = &scanner.TechScanner{}
	case "whois":
		s = &scanner.WhoisScanner{}
	case "dns":
		s = &scanner.DNSScanner{}
	}

	if s == nil {
		fmt.Printf("[!] Loại scan %s không tồn tại\n", job.ScanType)
		h.DB.Model(&models.ScanJob{}).Where("id = ?", job.ID).Update("status", "failed")
		return
	}

	results, err := s.Run(&asset)
	endTime := time.Now()

	if err != nil {
		fmt.Printf("[x] Job %s thất bại: %v\n", job.ID, err)
		h.DB.Model(&models.ScanJob{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
			"status":   "failed",
			"error":    err.Error(),
			"ended_at": &endTime,
		})
	} else {
        fmt.Printf("[v] Job %s hoàn thành thành công\n", job.ID)
        
        h.saveScanResults(job.ID, job.ScanType, results)

        h.DB.Model(&models.ScanJob{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
            "status":   "completed",
            "results":  1,
            "ended_at": &endTime,
        })
    }
}

func (h *AssetHandler) saveScanResults(jobID, scanType string, data interface{}) {
    var finalData string

    if s, ok := data.(string); ok {
        finalData = s
    } else {
        jsonData, _ := json.Marshal(data)
        finalData = string(jsonData)
    }

    result := models.ScanResult{
        JobID:      jobID,
        ScanType:   scanType,
        ResultData: finalData, // Đảm bảo đây là chuỗi JSON sạch
        CreatedAt:  time.Now(),
    }
    h.DB.Create(&result)
}
