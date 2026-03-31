package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegisterRoutesCoverage(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	
	r := RegisterRoutes(db)
	
	assert.NotNil(t, r)
	assert.NotNil(t, r.Get("health")) 
}
