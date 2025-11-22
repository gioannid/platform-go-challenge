package domain

import (
	"time"

	"github.com/google/uuid"
)

// Favourite represents a user's favourited asset
type Favourite struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	AssetID   uuid.UUID `json:"asset_id"`
	Asset     *Asset    `json:"asset,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// NewFavourite creates a new favourite entry
func NewFavourite(userID, assetID uuid.UUID) *Favourite {
	return &Favourite{
		ID:        uuid.New(),
		UserID:    userID,
		AssetID:   assetID,
		CreatedAt: time.Now(),
	}
}

// PageQuery represents pagination parameters
type PageQuery struct {
	Limit  int    // Number of results per page (max 1000)
	Offset int    // Starting position
	SortBy string // Field to sort by: "created_at", "type", "description"
	Order  string // Sort order: "asc" or "desc"
}

// NewPageQuery creates a PageQuery with defaults
func NewPageQuery(limit, offset int, sortBy, order string) *PageQuery {
	// Apply defaults
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return &PageQuery{
		Limit:  limit,
		Offset: offset,
		SortBy: sortBy,
		Order:  order,
	}
}
