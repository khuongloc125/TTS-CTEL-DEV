package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"my-asset-manager/internal/handlers"
	database "my-asset-manager/internal/mysql"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load file .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Lỗi không thể load file .env")
	}

	// 2. Lấy các giá trị từ môi trường
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	serverPort := os.Getenv("SERVER_PORT")

	// 3. Tạo chuỗi DSN cho MySQL
	// Format: user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	// 4. Kết nối Database
	db, err := database.NewMySQLDB(dsn)
	if err != nil {
		log.Fatal("Lỗi kết nối DB:", err)
	}

	// 5. Khởi tạo Handler
	h := &handlers.AssetHandler{DB: db}

	// 6. Định nghĩa Routes
	http.HandleFunc("/assets/stats", h.GetStats)
	http.HandleFunc("/assets/count", h.CountAssets)

	log.Printf("Server đang chạy tại cổng :%s...\n", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}