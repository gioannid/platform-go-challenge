package service

import (
	"context"
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
