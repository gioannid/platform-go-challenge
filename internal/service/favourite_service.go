// Package service defines the Service Layer, implementing the business logic of handling assets and
// marking assets as favourites
package service

import (
	"context"
	"fmt"

	"github.com/gioannid/platform-go-challenge/internal/config"
	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/gioannid/platform-go-challenge/internal/repository"
	"github.com/google/uuid"
)

// FavouriteService handles business logic for favourites
type FavouriteService struct {
	repo repository.FavouriteRepository
}

// NewFavouriteService creates a new service instance
func NewFavouriteService(repo repository.FavouriteRepository) *FavouriteService {
	return &FavouriteService{
		repo: repo,
	}
}

// ListFavourites returns paginated list of user's favourites
func (s *FavouriteService) ListFavourites(ctx context.Context, userID uuid.UUID, query *domain.PageQuery) ([]*domain.Favourite, int, error) {
	cfg := config.Get()
	// Business logic validation
	if query.Limit > cfg.MaxPageItems {
		query.Limit = cfg.MaxPageItems // Enforce maximum
	}

	return s.repo.ListFavourites(ctx, userID, query)
}

// AddFavourite adds an asset to user's favourites
func (s *FavouriteService) AddFavourite(ctx context.Context, userID, assetID uuid.UUID) (*domain.Favourite, error) {
	// Check if asset exists
	asset, err := s.repo.GetAsset(ctx, assetID)
	if err != nil {
		return nil, fmt.Errorf("asset not found: %w", err)
	}

	// Check if already favourited
	isFav, err := s.repo.IsFavourite(ctx, userID, assetID)
	if err != nil {
		return nil, err
	}
	if isFav {
		return nil, domain.ErrAlreadyExists
	}

	// Create favourite
	fav := domain.NewFavourite(userID, assetID)
	if err := s.repo.AddFavourite(ctx, fav); err != nil {
		return nil, err
	}

	// Return favourite with asset data attached
	fav.Asset = asset
	return fav, nil
}

// RemoveFavourite removes an asset from user's favourites
func (s *FavouriteService) RemoveFavourite(ctx context.Context, userID, favouriteID uuid.UUID) error {
	return s.repo.RemoveFavourite(ctx, userID, favouriteID)
}

// CreateAsset creates a new asset
func (s *FavouriteService) CreateAsset(ctx context.Context, assetType domain.AssetType, description string, data interface{}) (*domain.Asset, error) {
	asset, err := domain.NewAsset(assetType, description, data)
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateAsset(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

// UpdateAssetDescription updates an asset's description
func (s *FavouriteService) UpdateAssetDescription(ctx context.Context, assetID uuid.UUID, description string) error {
	// Validate description not empty
	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	return s.repo.UpdateAssetDescription(ctx, assetID, description)
}

// DeleteAsset deletes an asset
func (s *FavouriteService) DeleteAsset(ctx context.Context, assetID uuid.UUID) error {
	return s.repo.DeleteAsset(ctx, assetID)
}

// HealthCheck verifies service health
func (s *FavouriteService) HealthCheck(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
