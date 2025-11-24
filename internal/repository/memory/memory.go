// Package memory is an in-memory implementation of the repository interface, limited by the physical memory available
// of the running instance (node or container). This implementation provides:
//   - For individual Get, Add, Update, Remove operations on assets, on average O(1) time complexity.
//     In the worst-case scenario (key hash collisions) operations could degrade to O(N),
//     but this should be rare as well-randomized uuid.UUID keys are used.
//   - For AddFavourite and RemoveFavourite operations, average O(M) time complexity where M is the number of favourites
//     for a user, due to the need to check for existing favourites before adding/removing. An additional index that would
//     map (userID, assetID) to favouriteID could reduce this to O(1) at the cost of increased memory usage.
//   - For IsFavourite operations, average O(M) time complexity where M is the number of favourites for a user,
//     as it requires scanning the user's favourites. An additional index mapping (userID, assetID) to favouriteID
//     could reduce this to O(1) but would increase memory usage.
//   - For List operations (ListAssets, ListFavourites), a dominant time complexity of O(N log N) due to sorting.
//     Pagination is applied after sorting, which is O(1), so N refers to the total number of items before pagination.
//   - For DeleteAsset operations, a quite poor O(U * <F>) time complexity where U is the number of users and
//     <F> is the average number of favourites per user, since it requires scanning all users' favourites to remove hanging
//     references to the deleted asset. Depending on how rare (or needed at all) asset deletions are, this may be acceptable.
//   - Thread syncrhonization via sync.RWMutex allowing concurrent read but serializing write operations. This is generally
//     a good approach for in-memory stores. However, under high write contention, the mutex itself can also become a bottleneck.
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
		} else {
			// Return an error if the asset linked to a favourite does not exist
			return nil, 0, domain.ErrDataIntegrity
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

// ListAssets returns paginated list of all assets in the system
func (r *MemoryRepository) ListAssets(ctx context.Context, query *domain.PageQuery) ([]*domain.Asset, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Convert map to slice
	assets := make([]*domain.Asset, 0, len(r.assets))
	for _, asset := range r.assets {
		assets = append(assets, asset)
	}

	total := len(assets)

	// Sort based on query
	sort.Slice(assets, func(i, j int) bool {
		switch query.SortBy {
		case "type":
			if query.Order == "asc" {
				return assets[i].Type < assets[j].Type
			}
			return assets[i].Type > assets[j].Type
		case "description":
			if query.Order == "asc" {
				return assets[i].Description < assets[j].Description
			}
			return assets[i].Description > assets[j].Description
		case "updated_at":
			if query.Order == "asc" {
				return assets[i].UpdatedAt.Before(assets[j].UpdatedAt)
			}
			return assets[i].UpdatedAt.After(assets[j].UpdatedAt)
		default: // created_at
			if query.Order == "asc" {
				return assets[i].CreatedAt.Before(assets[j].CreatedAt)
			}
			return assets[i].CreatedAt.After(assets[j].CreatedAt)
		}
	})

	// Apply pagination
	start := query.Offset
	end := query.Offset + query.Limit

	if start > len(assets) {
		return []*domain.Asset{}, total, nil
	}
	if end > len(assets) {
		end = len(assets)
	}

	return assets[start:end], total, nil
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
