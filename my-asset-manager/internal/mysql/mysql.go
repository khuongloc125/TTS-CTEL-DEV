package database

import (
	"my-asset-manager/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMySQLDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Tự động tạo bảng (Code-First)
	err = db.AutoMigrate(&models.Asset{})
	if err != nil {
		return nil, err
	}

	return db, nil
}