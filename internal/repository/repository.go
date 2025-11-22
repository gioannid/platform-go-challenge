// package repository defines the Data Access layer: main interface with possible data operations
// and specific implementations (only in memory for now)
package repository

import (
	"context"

	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/google/uuid"
)

// FavouriteRepository defines the interface for favourite storage operations
type FavouriteRepository interface {
	// Favourites management. Note that favourites are user-scoped (do not exist globally).
	// This allows implementations of FavouriteRepository to optimize storage per case
	// without having to necessarily use e.g. a relational DB schema.
	ListFavourites(ctx context.Context, userID uuid.UUID, query *domain.PageQuery) ([]*domain.Favourite, int, error) // we return also total count beyond any pagination
	GetFavourite(ctx context.Context, userID, favouriteID uuid.UUID) (*domain.Favourite, error)
	AddFavourite(ctx context.Context, favourite *domain.Favourite) error
	RemoveFavourite(ctx context.Context, userID, favouriteID uuid.UUID) error
	IsFavourite(ctx context.Context, userID, assetID uuid.UUID) (bool, error)

	// Asset management
	GetAsset(ctx context.Context, assetID uuid.UUID) (*domain.Asset, error)
	CreateAsset(ctx context.Context, asset *domain.Asset) error
	UpdateAssetDescription(ctx context.Context, assetID uuid.UUID, description string) error
	DeleteAsset(ctx context.Context, assetID uuid.UUID) error

	// Health check
	Ping(ctx context.Context) error
	Sanity(ctx context.Context) error

	// TODO: Implement user management (operations on users)
}
