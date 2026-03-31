package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	assetHandlers "my-asset-manager/internal/handlers"
	"my-asset-manager/internal/models"
	database "my-asset-manager/internal/mysql"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"github.com/gorilla/handlers"
)

func RegisterRoutes(db *gorm.DB) *mux.Router {
	h := &assetHandlers.AssetHandler{DB: db}
	r := mux.NewRouter()

	r.HandleFunc("/assets/stats", h.GetStats).Methods("GET")
	r.HandleFunc("/assets/{id}", h.GetAsset).Methods("GET")       // Mới: Khớp bộ 1
	r.HandleFunc("/assets/{id}", h.UpdateAsset).Methods("PUT")    // Mới: Khớp bộ 1
	r.HandleFunc("/assets/{id}", h.DeleteAsset).Methods("DELETE") 
	
	r.HandleFunc("/assets/{id}/dns", h.GetLatestDNS).Methods("GET")             // Mới
	r.HandleFunc("/assets/{id}/subdomains", h.GetSubdomains).Methods("GET")     // Mới
	
	r.HandleFunc("/assets", h.CreateAsset).Methods("POST")
	r.HandleFunc("/assets", h.ListAssets).Methods("GET")
	r.HandleFunc("/assets/{id}/scan", h.StartScan).Methods("POST")
	r.HandleFunc("/scan-jobs/{id}", h.GetScanStatus).Methods("GET")
	r.HandleFunc("/scan-jobs/{id}/results", h.GetScanResults).Methods("GET")
	r.HandleFunc("/assets/{id}/whois", h.GetLatestWhois).Methods("GET")
	r.HandleFunc("/assets/{id}/scans", h.ListScansByAsset).Methods("GET")
	r.HandleFunc("/assets/{id}/results", h.GetAllAssetResults).Methods("GET")
	r.HandleFunc("/assets/search", h.SearchAssets).Methods("GET")
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")
	
	r.HandleFunc("/assets/count", h.CountAssets).Methods("GET")

	r.HandleFunc("/assets/batch", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.BatchCreate(w, r)
		case http.MethodDelete:
			h.BatchDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return r
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Cảnh báo: Không tìm thấy file .env, sử dụng biến môi trường hệ thống")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), 
		os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	db, err := database.ConnectWithRetry(dsn, 5)
	if err != nil {
		log.Fatalf("❌ Kết thúc: %v", err)
	}

	db.AutoMigrate(&models.ScanJob{}, &models.ScanResult{}, &models.Asset{})

	r := RegisterRoutes(db)

    corsObj := handlers.AllowedOrigins([]string{"*"})
    headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
    methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

    serverPort := os.Getenv("SERVER_PORT")
    if serverPort == "" {
        serverPort = "8080"
    }

    log.Printf("🚀 Server đang chạy tại cổng :%s...\n", serverPort)
    
    err = http.ListenAndServe(":"+serverPort, handlers.CORS(corsObj, headersOk, methodsOk)(r))
    if err != nil {
        log.Fatalf("❌ Lỗi khởi động server: %v", err)
    }
}
