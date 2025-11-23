package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gioannid/platform-go-challenge/internal/config"
	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository is a mock implementation of FavouriteRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) ListFavourites(ctx context.Context, userID uuid.UUID, query *domain.PageQuery) ([]*domain.Favourite, int, error) {
	args := m.Called(ctx, userID, query)
	return args.Get(0).([]*domain.Favourite), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetFavourite(ctx context.Context, userID, favouriteID uuid.UUID) (*domain.Favourite, error) {
	args := m.Called(ctx, userID, favouriteID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Favourite), args.Error(1)
}

func (m *MockRepository) AddFavourite(ctx context.Context, favourite *domain.Favourite) error {
	args := m.Called(ctx, favourite)
	return args.Error(0)
}

func (m *MockRepository) RemoveFavourite(ctx context.Context, userID, favouriteID uuid.UUID) error {
	args := m.Called(ctx, userID, favouriteID)
	return args.Error(0)
}

func (m *MockRepository) IsFavourite(ctx context.Context, userID, assetID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, assetID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) GetAsset(ctx context.Context, assetID uuid.UUID) (*domain.Asset, error) {
	args := m.Called(ctx, assetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Asset), args.Error(1)
}

func (m *MockRepository) CreateAsset(ctx context.Context, asset *domain.Asset) error {
	args := m.Called(ctx, asset)
	return args.Error(0)
}

func (m *MockRepository) UpdateAssetDescription(ctx context.Context, assetID uuid.UUID, description string) error {
	args := m.Called(ctx, assetID, description)
	return args.Error(0)
}

func (m *MockRepository) DeleteAsset(ctx context.Context, assetID uuid.UUID) error {
	args := m.Called(ctx, assetID)
	return args.Error(0)
}

// ListAssets mocks the ListAssets method
func (m *MockRepository) ListAssets(ctx context.Context, query *domain.PageQuery) ([]*domain.Asset, int, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Asset), args.Int(1), args.Error(2)
}
func (m *MockRepository) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) Sanity(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Tests
func TestFavouriteService_AddFavourite(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	assetID := uuid.New()

	tests := []struct {
		name    string
		setup   func(*MockRepository)
		wantErr error
	}{
		{
			name: "success",
			setup: func(m *MockRepository) {
				asset := createTestAsset(t, domain.AssetTypeChart, assetID)
				m.On("GetAsset", ctx, assetID).Return(asset, nil)
				m.On("IsFavourite", ctx, userID, assetID).Return(false, nil)
				m.On("AddFavourite", ctx, mock.AnythingOfType("*domain.Favourite")).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "asset not found",
			setup: func(m *MockRepository) {
				m.On("GetAsset", ctx, assetID).Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "already favourited",
			setup: func(m *MockRepository) {
				asset := createTestAsset(t, domain.AssetTypeChart, assetID)
				m.On("GetAsset", ctx, assetID).Return(asset, nil)
				m.On("IsFavourite", ctx, userID, assetID).Return(true, nil)
			},
			wantErr: domain.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setup(mockRepo)

			svc := NewFavouriteService(mockRepo)
			fav, err := svc.AddFavourite(ctx, userID, assetID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, fav)
			} else {
				require.NoError(t, err)
				require.NotNil(t, fav)
				assert.Equal(t, userID, fav.UserID)
				assert.Equal(t, assetID, fav.AssetID)
				assert.NotNil(t, fav.Asset)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFavouriteService_ListFavourites(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockRepository)
	expectedFavs := []*domain.Favourite{
		{ID: uuid.New(), UserID: userID, AssetID: uuid.New()},
	}
	query := domain.NewPageQuery(100, 0, "created_at", "desc")

	mockRepo.On("ListFavourites", ctx, userID, query).Return(expectedFavs, 1, nil)

	svc := NewFavouriteService(mockRepo)
	favs, total, err := svc.ListFavourites(ctx, userID, query)

	require.NoError(t, err)
	assert.Equal(t, expectedFavs, favs)
	assert.Equal(t, 1, total)
	mockRepo.AssertExpectations(t)
}

func TestFavouriteService_ListFavourites_LimitEnforcement(t *testing.T) {
	cfg := config.Get()
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockRepository)
	query := domain.NewPageQuery(cfg.MaxPageItems*2, 0, "created_at", "desc")

	// Mock should be called with enforced limit of 1000
	mockRepo.On("ListFavourites", ctx, userID, mock.MatchedBy(func(q *domain.PageQuery) bool {
		return q.Limit == cfg.MaxPageItems
	})).Return([]*domain.Favourite{}, 0, nil)

	svc := NewFavouriteService(mockRepo)
	_, _, err := svc.ListFavourites(ctx, userID, query)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFavouriteService_CreateAsset(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		assetType   domain.AssetType
		description string
		data        interface{}
		wantErr     bool
	}{
		{
			name:        "valid chart",
			assetType:   domain.AssetTypeChart,
			description: "Sales Chart",
			data: domain.ChartData{
				Title:      "Q4 Sales",
				AxisXTitle: "Month",
				AxisYTitle: "Revenue",
				Data:       [][]float64{{1, 100}},
			},
			wantErr: false,
		},
		{
			name:        "invalid chart data",
			assetType:   domain.AssetTypeChart,
			description: "Invalid Chart",
			data: domain.ChartData{
				Title: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)

			if !tt.wantErr {
				mockRepo.On("CreateAsset", ctx, mock.AnythingOfType("*domain.Asset")).Return(nil)
			}

			svc := NewFavouriteService(mockRepo)
			asset, err := svc.CreateAsset(ctx, tt.assetType, tt.description, tt.data)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, asset)
			} else {
				require.NoError(t, err)
				require.NotNil(t, asset)
				assert.Equal(t, tt.assetType, asset.Type)
				assert.Equal(t, tt.description, asset.Description)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestFavouriteService_UpdateAssetDescription(t *testing.T) {
	ctx := context.Background()
	assetID := uuid.New()

	tests := []struct {
		name        string
		description string
		wantErr     bool
	}{
		{
			name:        "valid description",
			description: "Updated Description",
			wantErr:     false,
		},
		{
			name:        "empty description",
			description: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)

			if !tt.wantErr {
				mockRepo.On("UpdateAssetDescription", ctx, assetID, tt.description).Return(nil)
			}

			svc := NewFavouriteService(mockRepo)
			err := svc.UpdateAssetDescription(ctx, assetID, tt.description)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestFavouriteService_RemoveFavourite(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	favouriteID := uuid.New()

	mockRepo := new(MockRepository)
	mockRepo.On("RemoveFavourite", ctx, userID, favouriteID).Return(nil)

	svc := NewFavouriteService(mockRepo)
	err := svc.RemoveFavourite(ctx, userID, favouriteID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFavouriteService_DeleteAsset(t *testing.T) {
	ctx := context.Background()
	assetID := uuid.New()

	mockRepo := new(MockRepository)
	mockRepo.On("DeleteAsset", ctx, assetID).Return(nil)

	svc := NewFavouriteService(mockRepo)
	err := svc.DeleteAsset(ctx, assetID)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFavouriteService_HealthCheck(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockRepository)
	mockRepo.On("Ping", ctx).Return(nil)

	svc := NewFavouriteService(mockRepo)
	err := svc.HealthCheck(ctx)

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFavouriteService_ListAssets(t *testing.T) {
	ctx := context.Background()

	t.Run("successful listing with default pagination", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewFavouriteService(mockRepo)

		// Create test assets
		asset1, _ := domain.NewAsset(domain.AssetTypeChart, "Chart 1", domain.ChartData{
			Title:      "Sales Chart",
			AxisXTitle: "Month",
			AxisYTitle: "Revenue",
			Data:       [][]float64{{1, 100}, {2, 200}},
		})
		asset2, _ := domain.NewAsset(domain.AssetTypeInsight, "Insight 1", domain.InsightData{
			Text: "40% of users engage daily",
		})

		expectedAssets := []*domain.Asset{asset1, asset2}
		query := domain.NewPageQuery(20, 0, "created_at", "desc")

		mockRepo.On("ListAssets", ctx, query).Return(expectedAssets, 2, nil)

		assets, total, err := service.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Equal(t, 2, len(assets))
		assert.Equal(t, asset1.ID, assets[0].ID)
		assert.Equal(t, asset2.ID, assets[1].ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result set", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewFavouriteService(mockRepo)

		query := domain.NewPageQuery(20, 0, "created_at", "desc")
		mockRepo.On("ListAssets", ctx, query).Return([]*domain.Asset{}, 0, nil)

		assets, total, err := service.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Equal(t, 0, len(assets))
		mockRepo.AssertExpectations(t)
	})

	t.Run("enforces maximum page limit", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewFavouriteService(mockRepo)

		// Request more than max allowed (1000 is typically the max)
		query := domain.NewPageQuery(5000, 0, "created_at", "desc")

		// The service should enforce the max limit
		expectedQuery := domain.NewPageQuery(1000, 0, "created_at", "desc")
		mockRepo.On("ListAssets", ctx, expectedQuery).Return([]*domain.Asset{}, 0, nil)

		_, _, err := service.ListAssets(ctx, query)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error propagates", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewFavouriteService(mockRepo)

		query := domain.NewPageQuery(20, 0, "created_at", "desc")
		expectedErr := errors.New("database connection failed")
		mockRepo.On("ListAssets", ctx, query).Return(nil, 0, expectedErr)

		assets, total, err := service.ListAssets(ctx, query)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, assets)
		assert.Equal(t, 0, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("pagination with offset", func(t *testing.T) {
		mockRepo := new(MockRepository)
		service := NewFavouriteService(mockRepo)

		asset3, _ := domain.NewAsset(domain.AssetTypeAudience, "Audience 3", domain.AudienceData{
			Gender:             "Female",
			BirthCountry:       "UK",
			AgeGroups:          []string{"25-34"},
			HoursSocialDaily:   4.5,
			PurchasesLastMonth: 3,
		})

		expectedAssets := []*domain.Asset{asset3}
		query := domain.NewPageQuery(10, 20, "created_at", "desc")

		mockRepo.On("ListAssets", ctx, query).Return(expectedAssets, 25, nil)

		assets, total, err := service.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 25, total)
		assert.Equal(t, 1, len(assets))
		mockRepo.AssertExpectations(t)
	})
}

// Helper
func createTestAsset(t *testing.T, assetType domain.AssetType, id uuid.UUID) *domain.Asset {
	t.Helper()

	data := domain.ChartData{
		Title:      "Test Chart",
		AxisXTitle: "X",
		AxisYTitle: "Y",
		Data:       [][]float64{{1, 2}},
	}

	asset, err := domain.NewAsset(assetType, "Test Asset", data)
	require.NoError(t, err)
	asset.ID = id
	return asset
}
