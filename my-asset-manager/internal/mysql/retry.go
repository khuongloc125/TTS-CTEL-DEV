package database

import (
	"fmt"
	"log"
	"time"

	"my-asset-manager/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectWithRetry(dsn string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 1; i <= maxRetries; i++ {
		log.Printf("🔄 Database connection attempt %d/%d...", i, maxRetries)

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		
		if err == nil {

			sqlDB, errConn := db.DB()
			if errConn == nil {
				if errPing := sqlDB.Ping(); errPing == nil {
					log.Println("✅ Database connected successfully!")

					db.AutoMigrate(&models.Asset{})
					return db, nil
				}
			}
		}

		if i < maxRetries {

			waitTime := time.Duration(1<<(uint(i)-1)) * time.Second
			
			log.Printf("⚠️  Connection failed: %v. Retrying in %v...", err, waitTime)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("could not connect to database after %d attempts", maxRetries)
}
