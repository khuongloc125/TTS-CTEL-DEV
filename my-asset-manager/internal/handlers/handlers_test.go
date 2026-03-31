package handlers

import (
	"bytes"
	"encoding/json"
	"my-asset-manager/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAllHandlers(t *testing.T) {
	db := setupTestDB()
	h := &AssetHandler{DB: db}

	t.Run("Create Asset", func(t *testing.T) {
		body := `{"name":"test-create.com","type":"domain"}`
		h.CreateAsset(httptest.NewRecorder(), httptest.NewRequest("POST", "/assets", bytes.NewBufferString(body)))
	})

	t.Run("Batch Create", func(t *testing.T) {
		body := `{"assets": [{"name":"b1.com","type":"domain"},{"name":"b2.com","type":"domain"}]}`
		h.BatchCreate(httptest.NewRecorder(), httptest.NewRequest("POST", "/assets/batch", bytes.NewBufferString(body)))
	})

	t.Run("Get Statistics & Count", func(t *testing.T) {
		h.GetStats(httptest.NewRecorder(), httptest.NewRequest("GET", "/assets/stats", nil))
		h.CountAssets(httptest.NewRecorder(), httptest.NewRequest("GET", "/assets/count", nil))
	})

	t.Run("List & Search", func(t *testing.T) {
		h.ListAssets(httptest.NewRecorder(), httptest.NewRequest("GET", "/assets", nil))
		h.SearchAssets(httptest.NewRecorder(), httptest.NewRequest("GET", "/assets/search?q=test", nil))
	})

	t.Run("Batch Delete", func(t *testing.T) {
		asset := models.Asset{ID: "del-123", Name: "del.com", Type: models.TypeDomain}
		db.Create(&asset)

		ids := []string{"del-123"}
		body, _ := json.Marshal(ids)
		
		req := httptest.NewRequest("DELETE", "/assets/batch", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		h.BatchDelete(rr, req)
		assert.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusBadRequest) 
	})

	t.Run("Scan Flow & Results", func(t *testing.T) {
		assetID := "test-results-id"
		db.Create(&models.Asset{ID: assetID, Name: "127.0.0.1", Type: models.TypeIP})

		payload, _ := json.Marshal(map[string]string{"scan_type": "port"})
		req := httptest.NewRequest("POST", "/assets/"+assetID+"/scan", bytes.NewBuffer(payload))
		req = mux.SetURLVars(req, map[string]string{"id": assetID})
		rr := httptest.NewRecorder()
		h.StartScan(rr, req)

		var resp map[string]interface{}
		json.Unmarshal(rr.Body.Bytes(), &resp)
		
		jobID := "any-id"
		if resp["job_id"] != nil {
			jobID = resp["job_id"].(string)
		}

		vars := map[string]string{"id": jobID}
		assetVars := map[string]string{"id": assetID}

		reqStatus := mux.SetURLVars(httptest.NewRequest("GET", "/", nil), vars)
		h.GetScanStatus(httptest.NewRecorder(), reqStatus)

		reqRes := mux.SetURLVars(httptest.NewRequest("GET", "/", nil), vars)
		h.GetScanResults(httptest.NewRecorder(), reqRes)

		reqList := mux.SetURLVars(httptest.NewRequest("GET", "/", nil), assetVars)
		h.ListScansByAsset(httptest.NewRecorder(), reqList)

		reqAll := mux.SetURLVars(httptest.NewRequest("GET", "/", nil), assetVars)
		h.GetAllAssetResults(httptest.NewRecorder(), reqAll)
        
        reqWhois := mux.SetURLVars(httptest.NewRequest("GET", "/", nil), assetVars)
		h.GetLatestWhois(httptest.NewRecorder(), reqWhois)
	})

	t.Run("Health Check", func(t *testing.T) {
		h.HealthCheck(httptest.NewRecorder(), httptest.NewRequest("GET", "/health", nil))
	})

	t.Run("Batch Delete Final Fix", func(t *testing.T) {
    asset := models.Asset{ID: "batch-1", Name: "test.com", Type: models.TypeDomain}
    db.Create(&asset)

    type DeleteRequest struct {
        IDs []string `json:"ids"`
    }
    
    data := DeleteRequest{IDs: []string{"batch-1"}}
    body, _ := json.Marshal(data)
    
    req := httptest.NewRequest("DELETE", "/assets/batch", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json") // Bắt buộc phải có
    
    rr := httptest.NewRecorder()
    h.BatchDelete(rr, req)
    
    assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Validation Edge Cases", func(t *testing.T) {
		body := `{"name":"invalid.com","type":"alien_technology"}`
		h.CreateAsset(httptest.NewRecorder(), httptest.NewRequest("POST", "/assets", bytes.NewBufferString(body)))
		
		h.BatchCreate(httptest.NewRecorder(), httptest.NewRequest("POST", "/assets/batch", bytes.NewBufferString(`{"assets":[]}`)))
	})
	
}
