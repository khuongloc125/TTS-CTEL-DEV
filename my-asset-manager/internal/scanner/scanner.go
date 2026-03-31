package scanner

import (
	"my-asset-manager/internal/models"
)

type Scanner interface {
	Run(asset *models.Asset) (interface{}, error)
}
