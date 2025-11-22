package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAsset(t *testing.T) {
	tests := []struct {
		name        string
		assetType   AssetType
		description string
		data        interface{}
		wantErr     error
	}{
		{
			name:        "valid chart asset",
			assetType:   AssetTypeChart,
			description: "Sales Chart",
			data: ChartData{
				Title:      "Q4 Sales",
				AxisXTitle: "Month",
				AxisYTitle: "Revenue",
				Data:       [][]float64{{1, 100}, {2, 200}},
			},
			wantErr: nil,
		},
		{
			name:        "valid insight asset",
			assetType:   AssetTypeInsight,
			description: "Market Insight",
			data: InsightData{
				Text: "Market is growing",
			},
			wantErr: nil,
		},
		{
			name:        "valid audience asset",
			assetType:   AssetTypeAudience,
			description: "Target Audience",
			data: AudienceData{
				Gender:             "Female",
				BirthCountry:       "UK",
				AgeGroups:          []string{"25-34"},
				HoursSocialDaily:   4.5,
				PurchasesLastMonth: 15,
			},
			wantErr: nil,
		},
		{
			name:        "invalid chart - empty title",
			assetType:   AssetTypeChart,
			description: "Invalid Chart",
			data: ChartData{
				Title:      "",
				AxisXTitle: "X",
				AxisYTitle: "Y",
				Data:       [][]float64{{1, 2}},
			},
			wantErr: ErrInvalidChartData,
		},
		{
			name:        "invalid insight - empty text",
			assetType:   AssetTypeInsight,
			description: "Invalid Insight",
			data: InsightData{
				Text: "",
			},
			wantErr: ErrInvalidInsightData,
		},
		{
			name:        "invalid audience - empty gender",
			assetType:   AssetTypeAudience,
			description: "Invalid Audience",
			data: AudienceData{
				Gender:       "",
				BirthCountry: "USA",
			},
			wantErr: ErrInvalidAudienceData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, err := NewAsset(tt.assetType, tt.description, tt.data)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, asset)
			} else {
				t.Logf("Asset created:\n  id: %s\n  type: %s\n  description: %s\n  created_at: %s\n  updated_at: %s\n",
					asset.ID, asset.Type, asset.Description, asset.CreatedAt.Format(time.RFC3339), asset.UpdatedAt.Format(time.RFC3339))
				require.NoError(t, err)
				require.NotNil(t, asset)
				assert.NotEqual(t, uuid.Nil, asset.ID)
				assert.Equal(t, tt.assetType, asset.Type)
				assert.Equal(t, tt.description, asset.Description)
				assert.NotZero(t, asset.CreatedAt)
				assert.NotZero(t, asset.UpdatedAt)
				assert.True(t, asset.UpdatedAt.Equal(asset.CreatedAt))
			}
		})
	}
}

func TestAsset_Validate(t *testing.T) {
	tests := []struct {
		name    string
		asset   func() *Asset
		wantErr error
	}{
		{
			name: "valid chart",
			asset: func() *Asset {
				return &Asset{
					ID:          uuid.New(),
					Type:        AssetTypeChart,
					Description: "Test",
					Data: json.RawMessage(`{
						"title": "Chart",
						"axis_x_title": "X",
						"axis_y_title": "Y",
						"data": [[1,2]]
					}`),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
			},
			wantErr: nil,
		},
		{
			name: "empty type",
			asset: func() *Asset {
				return &Asset{
					ID:          uuid.New(),
					Type:        "",
					Description: "Test",
					Data:        json.RawMessage(`{"title": "Test"}`),
				}
			},
			wantErr: ErrInvalidAssetType,
		},
		{
			name: "invalid type",
			asset: func() *Asset {
				return &Asset{
					ID:          uuid.New(),
					Type:        "invalid",
					Description: "Test",
					Data:        json.RawMessage(`{"title": "Test"}`),
				}
			},
			wantErr: ErrInvalidAssetType,
		},
		{
			name: "missing data",
			asset: func() *Asset {
				return &Asset{
					ID:          uuid.New(),
					Type:        AssetTypeChart,
					Description: "Test",
					Data:        nil,
				}
			},
			wantErr: ErrMissingAssetData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset := tt.asset()
			err := asset.Validate()

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAssetType_Constants(t *testing.T) {
	assert.Equal(t, AssetType("chart"), AssetTypeChart)
	assert.Equal(t, AssetType("insight"), AssetTypeInsight)
	assert.Equal(t, AssetType("audience"), AssetTypeAudience)
}
