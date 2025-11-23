// package domain defines data models and possible errors
package domain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AssetType represents the type of asset
type AssetType string

const (
	AssetTypeChart    AssetType = "chart"
	AssetTypeInsight  AssetType = "insight"
	AssetTypeAudience AssetType = "audience"
)

// Asset represents a generic asset that can be favourited
type Asset struct {
	ID          uuid.UUID       `json:"id"`
	Type        AssetType       `json:"type"`
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data" swaggertype:"object,string" example:"{\"title\":\"Sample Chart\"}"` // Polymorphic data field
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ChartData represents chart-specific data
type ChartData struct {
	Title      string      `json:"title"`
	AxisXTitle string      `json:"axis_x_title"`
	AxisYTitle string      `json:"axis_y_title"`
	Data       [][]float64 `json:"data"` // 2D array of data points
}

// InsightData represents insight-specific data
type InsightData struct {
	Text string `json:"text"`
}

// AudienceData represents audience-specific data
type AudienceData struct {
	Gender             string   `json:"gender"` // "Male", "Female"
	BirthCountry       string   `json:"birth_country"`
	AgeGroups          []string `json:"age_groups"` // e.g., ["18-24", "25-34"]
	HoursSocialDaily   float64  `json:"hours_social_daily"`
	PurchasesLastMonth int      `json:"purchases_last_month"`
}

// Validate ensures the asset is properly formed
func (a *Asset) Validate() error {
	if a.Type == "" {
		return ErrInvalidAssetType
	}

	switch a.Type {
	case AssetTypeChart, AssetTypeInsight, AssetTypeAudience:
		// Valid types
	default:
		return ErrInvalidAssetType
	}

	if len(a.Data) == 0 {
		return ErrMissingAssetData
	}

	// Type-specific validation
	switch a.Type {
	case AssetTypeChart:
		var chartData ChartData
		if err := json.Unmarshal(a.Data, &chartData); err != nil {
			return fmt.Errorf("invalid chart data: %w", err)
		}
		if chartData.Title == "" {
			return ErrInvalidChartData
		}

	case AssetTypeInsight:
		var insightData InsightData
		if err := json.Unmarshal(a.Data, &insightData); err != nil {
			return fmt.Errorf("invalid insight data: %w", err)
		}
		if insightData.Text == "" {
			return ErrInvalidInsightData
		}

	case AssetTypeAudience:
		var audienceData AudienceData
		if err := json.Unmarshal(a.Data, &audienceData); err != nil {
			return fmt.Errorf("invalid audience data: %w", err)
		}
		if audienceData.Gender == "" {
			return ErrInvalidAudienceData
		}
	}

	return nil
}

// NewAsset creates a new asset with generated ID and timestamps
func NewAsset(assetType AssetType, description string, data interface{}) (*Asset, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal asset data: %w", err)
	}

	now := time.Now()
	asset := &Asset{
		ID:          uuid.New(),
		Type:        assetType,
		Description: description,
		Data:        dataBytes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := asset.Validate(); err != nil {
		return nil, err
	}

	return asset, nil
}
