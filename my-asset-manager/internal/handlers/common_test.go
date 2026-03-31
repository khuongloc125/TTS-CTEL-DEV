package handlers

import (
	"my-asset-manager/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("Không thể kết nối DB ảo: " + err.Error())
	}

	db.AutoMigrate(&models.Asset{}, &models.ScanJob{}, &models.ScanResult{})
	return db
}
