// Package memory provides an in-memory implementation of the repository interfaces.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/google/uuid"
)

// MemoryRepository implements an in-memory storage with thread safety
type MemoryRepository struct {
	mu         sync.RWMutex
	assets     map[uuid.UUID]*domain.Asset                   // assets indexed by assetID
	favourites map[uuid.UUID]map[uuid.UUID]*domain.Favourite // favourites indexed first by userID, with each indexed by favouriteID
	// (userID -> favouriteID -> Favourite)
}

// NewRepository creates a new in-memory repository
func NewRepository() *MemoryRepository {
	return &MemoryRepository{
		assets:     make(map[uuid.UUID]*domain.Asset),
		favourites: make(map[uuid.UUID]map[uuid.UUID]*domain.Favourite),
	}
}

// ListFavourites returns paginated list of user's favourites
func (r *MemoryRepository) ListFavourites(ctx context.Context, userID uuid.UUID, query *domain.PageQuery) ([]*domain.Favourite, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userFavs, exists := r.favourites[userID]
	if !exists {
		return []*domain.Favourite{}, 0, nil
	}

	// Convert map to slice
	favs := make([]*domain.Favourite, 0, len(userFavs))
	for _, fav := range userFavs {
		// Attach asset data
		if asset, ok := r.assets[fav.AssetID]; ok {
			favCopy := *fav
			favCopy.Asset = asset
			favs = append(favs, &favCopy)
		}
	}

	total := len(favs)

	// Sort based on query
	sort.Slice(favs, func(i, j int) bool {
		switch query.SortBy {
		case "type":
			if query.Order == "asc" {
				return favs[i].Asset.Type < favs[j].Asset.Type
			}
			return favs[i].Asset.Type > favs[j].Asset.Type
		case "description":
			if query.Order == "asc" {
				return favs[i].Asset.Description < favs[j].Asset.Description
			}
			return favs[i].Asset.Description > favs[j].Asset.Description
		default: // created_at
			if query.Order == "asc" {
				return favs[i].CreatedAt.Before(favs[j].CreatedAt)
			}
			return favs[i].CreatedAt.After(favs[j].CreatedAt)
		}
	})

	// Apply pagination
	start := query.Offset
	end := query.Offset + query.Limit

	if start > len(favs) {
		return []*domain.Favourite{}, total, nil
	}
	if end > len(favs) {
		end = len(favs)
	}

	return favs[start:end], total, nil
}

// GetFavourite retrieves a specific favourite
func (r *MemoryRepository) GetFavourite(ctx context.Context, userID, favouriteID uuid.UUID) (*domain.Favourite, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userFavs, exists := r.favourites[userID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	fav, exists := userFavs[favouriteID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	// Attach asset data
	if asset, ok := r.assets[fav.AssetID]; ok {
		favCopy := *fav
		favCopy.Asset = asset
		return &favCopy, nil
	}

	return fav, nil
}

// AddFavourite adds a new favourite for a user
func (r *MemoryRepository) AddFavourite(ctx context.Context, favourite *domain.Favourite) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if asset exists
	if _, exists := r.assets[favourite.AssetID]; !exists {
		return domain.ErrNotFound
	}

	// Initialize user's favourites map if needed
	if r.favourites[favourite.UserID] == nil {
		r.favourites[favourite.UserID] = make(map[uuid.UUID]*domain.Favourite)
	}

	// Check if already favourited
	for _, fav := range r.favourites[favourite.UserID] {
		if fav.AssetID == favourite.AssetID {
			return domain.ErrAlreadyExists
		}
	}

	r.favourites[favourite.UserID][favourite.ID] = favourite
	return nil
}

// RemoveFavourite removes a favourite for a user
func (r *MemoryRepository) RemoveFavourite(ctx context.Context, userID, favouriteID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userFavs, exists := r.favourites[userID]
	if !exists {
		return domain.ErrNotFound
	}

	if _, exists := userFavs[favouriteID]; !exists {
		return domain.ErrNotFound
	}

	delete(userFavs, favouriteID)
	return nil
}

// IsFavourite checks if an asset is favourited by a user
func (r *MemoryRepository) IsFavourite(ctx context.Context, userID, assetID uuid.UUID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userFavs, exists := r.favourites[userID]
	if !exists {
		return false, nil
	}

	for _, fav := range userFavs {
		if fav.AssetID == assetID {
			return true, nil
		}
	}

	return false, nil
}

// GetAsset retrieves an asset by ID
func (r *MemoryRepository) GetAsset(ctx context.Context, assetID uuid.UUID) (*domain.Asset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	asset, exists := r.assets[assetID]
	if !exists {
		return nil, domain.ErrNotFound
	}

	return asset, nil
}

// CreateAsset stores a new asset
func (r *MemoryRepository) CreateAsset(ctx context.Context, asset *domain.Asset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.assets[asset.ID]; exists {
		return domain.ErrAlreadyExists
	}

	r.assets[asset.ID] = asset
	return nil
}

// UpdateAssetDescription updates an asset's description
func (r *MemoryRepository) UpdateAssetDescription(ctx context.Context, assetID uuid.UUID, description string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	asset, exists := r.assets[assetID]
	if !exists {
		return domain.ErrNotFound
	}

	asset.Description = description
	asset.UpdatedAt = time.Now()
	return nil
}

// DeleteAsset removes an asset
func (r *MemoryRepository) DeleteAsset(ctx context.Context, assetID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.assets[assetID]; !exists {
		return domain.ErrNotFound
	}

	// Remove asset
	delete(r.assets, assetID)

	// Remove from all user favourites
	for _, userFavs := range r.favourites {
		for favID, fav := range userFavs {
			if fav.AssetID == assetID {
				delete(userFavs, favID)
			}
		}
	}

	return nil
}

// Ping checks if the repository is accessible
func (r *MemoryRepository) Ping(ctx context.Context) error {
	return nil
}

// Sanity performs a sanity test for orphan favourites.
func (r *MemoryRepository) Sanity(ctx context.Context) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check for orphan favourites
	for userID, userFavs := range r.favourites {
		for favID, fav := range userFavs {
			if _, exists := r.assets[fav.AssetID]; !exists {
				return fmt.Errorf("sanity check failed: orphan favourite found (userID: %s, favouriteID: %s, assetID: %s)", userID, favID, fav.AssetID)
			}
		}
	}
	return nil
}
