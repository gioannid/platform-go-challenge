package domain

import (
	"testing"
	"time"

	"github.com/gioannid/platform-go-challenge/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFavourite(t *testing.T) {
	userID := uuid.New()
	assetID := uuid.New()

	fav := NewFavourite(userID, assetID)

	require.NotNil(t, fav)
	assert.NotEqual(t, uuid.Nil, fav.ID)
	assert.Equal(t, userID, fav.UserID)
	assert.Equal(t, assetID, fav.AssetID)
	assert.WithinDuration(t, time.Now(), fav.CreatedAt, time.Second)
	assert.Nil(t, fav.Asset)
}

func TestNewPageQuery(t *testing.T) {
	cfg := config.Get()
	tests := []struct {
		name           string
		inputLimit     int
		inputOffset    int
		inputSortBy    string
		inputOrder     string
		expectedLimit  int
		expectedOffset int
		expectedSortBy string
		expectedOrder  string
	}{
		{
			name:           "valid query",
			inputLimit:     cfg.MaxPageItems / 2,
			inputOffset:    10,
			inputSortBy:    "type",
			inputOrder:     "asc",
			expectedLimit:  cfg.MaxPageItems / 2,
			expectedOffset: 10,
			expectedSortBy: "type",
			expectedOrder:  "asc",
		},
		{
			name:           "defaults applied - zero limit",
			inputLimit:     0,
			inputOffset:    0,
			inputSortBy:    "",
			inputOrder:     "",
			expectedLimit:  cfg.MaxPageItems,
			expectedOffset: 0,
			expectedSortBy: "created_at",
			expectedOrder:  "desc",
		},
		{
			name:           "limit exceeds maximum",
			inputLimit:     cfg.MaxPageItems * 2,
			inputOffset:    5,
			inputSortBy:    "description",
			inputOrder:     "desc",
			expectedLimit:  cfg.MaxPageItems,
			expectedOffset: 5,
			expectedSortBy: "description",
			expectedOrder:  "desc",
		},
		{
			name:           "negative offset corrected",
			inputLimit:     cfg.MaxPageItems / 2,
			inputOffset:    -10,
			inputSortBy:    "created_at",
			inputOrder:     "asc",
			expectedLimit:  cfg.MaxPageItems / 2,
			expectedOffset: 0,
			expectedSortBy: "created_at",
			expectedOrder:  "asc",
		},
		{
			name:           "invalid order defaults to desc",
			inputLimit:     cfg.MaxPageItems / 2,
			inputOffset:    0,
			inputSortBy:    "type",
			inputOrder:     "invalid",
			expectedLimit:  cfg.MaxPageItems / 2,
			expectedOffset: 0,
			expectedSortBy: "type",
			expectedOrder:  "desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := NewPageQuery(tt.inputLimit, tt.inputOffset, tt.inputSortBy, tt.inputOrder)

			assert.Equal(t, tt.expectedLimit, query.Limit)
			assert.Equal(t, tt.expectedOffset, query.Offset)
			assert.Equal(t, tt.expectedSortBy, query.SortBy)
			assert.Equal(t, tt.expectedOrder, query.Order)
		})
	}
}
