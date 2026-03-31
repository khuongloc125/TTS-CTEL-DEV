package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	TypeDomain  = "domain"
	TypeIP      = "ip"
	TypeService = "service"

	StatusActive   = "active"
	StatusInactive = "inactive"
)

func IsValidType(t string) bool {
	return t == TypeDomain || t == TypeIP || t == TypeService
}

func IsValidStatus(s string) bool {
	return s == StatusActive || s == StatusInactive
}

func IsValidScanType(st ScanType) bool {
    switch st {
    case ScanTypeDNS, ScanTypeWHOIS, ScanTypeSSL, ScanTypePort, ScanTypeTech:
        return true
    }
    return false
}

type Asset struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	
	Type      string    `gorm:"type:varchar(20);not null" json:"type"`
	Status    string    `gorm:"type:varchar(20);default:'active'" json:"status"`
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

type BatchDeleteResponse struct {
	Deleted  int64 `json:"deleted"`
	NotFound int64 `json:"not_found"`
}

type HealthResponse struct {
	Status    string         `json:"status"`
	Database  DatabaseStatus `json:"database"`
	Timestamp string         `json:"timestamp"`
}

type DatabaseStatus struct {
	Status          string `json:"status"`
	OpenConnections int    `json:"open_connections"`
	InUse           int    `json:"in_use"`
	Idle            int    `json:"idle"`
	MaxOpen         int    `json:"max_open"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type AssetListResponse struct {
	Data       []Asset    `json:"data"`
	Pagination Pagination `json:"pagination"`
}
