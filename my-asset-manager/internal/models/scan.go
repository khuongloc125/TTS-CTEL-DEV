package models

import (
	"time"
)

type ScanType string

const (
	ScanTypeDNS   ScanType = "dns"
	ScanTypeWHOIS ScanType = "whois"
	ScanTypeSSL   ScanType = "ssl"
	ScanTypePort  ScanType = "port"
	ScanTypeTech  ScanType = "tech"
)

type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
)

type ScanJob struct {
	ID        string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	AssetID   string     `gorm:"type:varchar(36);not null" json:"asset_id"`
	ScanType  string     `gorm:"type:varchar(20);not null" json:"scan_type"` 
	Status    string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	StartedAt *time.Time `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	Error     string     `gorm:"type:text" json:"error"`
	Results   int        `json:"results"` 
	CreatedAt time.Time  `json:"created_at"`
}

type ScanResult struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    JobID     string    `gorm:"type:varchar(36);index" json:"job_id"`
    ScanType  string    `json:"scan_type"`
    ResultData string   `gorm:"type:longtext" json:"result_data"` // Lưu JSON string ở đây
    CreatedAt time.Time `json:"created_at"`
}

