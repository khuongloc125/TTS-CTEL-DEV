package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAssetModelAll(t *testing.T) {
	t.Run("Validations", func(t *testing.T) {
		assert.True(t, IsValidType(TypeDomain))
		assert.True(t, IsValidStatus(StatusActive))
		assert.True(t, IsValidScanType(ScanTypePort))
	})

	t.Run("BeforeCreate Hook", func(t *testing.T) {
		db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		db.AutoMigrate(&Asset{})
		asset := &Asset{Name: "test.com", Type: TypeDomain}
		db.Create(asset)
		assert.NotEmpty(t, asset.ID)
	})

	t.Run("Validations - False Cases", func(t *testing.T) {
		assert.False(t, IsValidType("unknown"))
		assert.False(t, IsValidStatus("expired"))
		assert.False(t, IsValidScanType("hack_satellite"))
	})
}
