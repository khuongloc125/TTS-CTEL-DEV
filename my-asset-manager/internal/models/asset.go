package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Asset struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Type      string    `gorm:"type:enum('domain','ip','service');not null" json:"type"`
	Status    string    `gorm:"type:enum('active','inactive');default:'active'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *Asset) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return
}

type StatsResponse struct {
	Total    int64            `json:"total"`
	ByType   map[string]int64 `json:"by_type"`
	ByStatus map[string]int64 `json:"by_status"`
}

type CountResponse struct {
	Count   int64             `json:"count"`
	Filters map[string]string `json:"filters"`
}

type BatchCreateRequest struct {
	Assets []Asset `json:"assets"`
}

type BatchCreateResponse struct {
	Created int      `json:"created"`
	IDs     []string `json:"ids"`
}