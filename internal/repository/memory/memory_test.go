package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryRepository_CreateAsset(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	asset := createTestAsset(t, domain.AssetTypeChart)

	err := repo.CreateAsset(ctx, asset)
	require.NoError(t, err)

	// Verify asset was stored
	retrieved, err := repo.GetAsset(ctx, asset.ID)
	require.NoError(t, err)
	assert.Equal(t, asset.ID, retrieved.ID)
	assert.Equal(t, asset.Type, retrieved.Type)
}

func TestMemoryRepository_CreateAsset_Duplicate(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	asset := createTestAsset(t, domain.AssetTypeChart)

	err := repo.CreateAsset(ctx, asset)
	require.NoError(t, err)

	// Try to create same asset again
	err = repo.CreateAsset(ctx, asset)
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestMemoryRepository_GetAsset_NotFound(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	_, err := repo.GetAsset(ctx, uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMemoryRepository_UpdateAssetDescription(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	asset := createTestAsset(t, domain.AssetTypeInsight)
	require.NoError(t, repo.CreateAsset(ctx, asset))

	// sleep to ensure timestamp difference
	time.Sleep(time.Second)
	newDescription := "Updated Description"
	err := repo.UpdateAssetDescription(ctx, asset.ID, newDescription)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetAsset(ctx, asset.ID)
	require.NoError(t, err)
	assert.Equal(t, newDescription, updated.Description)
	t.Logf("Asset created:\n  id: %s\n  type: %s\n  description: %s\n  created_at: %s\n  updated_at: %s\n",
		updated.ID, updated.Type, updated.Description, updated.CreatedAt.Format(time.RFC3339), updated.UpdatedAt.Format(time.RFC3339))
	assert.True(t, updated.UpdatedAt.After(updated.CreatedAt))
}

func TestMemoryRepository_DeleteAsset(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	userID := uuid.New()

	// Create asset and favourite
	asset := createTestAsset(t, domain.AssetTypeChart)
	require.NoError(t, repo.CreateAsset(ctx, asset))

	fav := domain.NewFavourite(userID, asset.ID)
	require.NoError(t, repo.AddFavourite(ctx, fav))

	// Delete asset
	err := repo.DeleteAsset(ctx, asset.ID)
	require.NoError(t, err)

	// Verify asset is deleted
	_, err = repo.GetAsset(ctx, asset.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	// Verify favourite is also removed (cascade)
	favs, _, err := repo.ListFavourites(ctx, userID, domain.NewPageQuery(10, 0, "", ""))
	require.NoError(t, err)
	assert.Empty(t, favs)
}

func TestMemoryRepository_AddFavourite(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	userID := uuid.New()

	asset := createTestAsset(t, domain.AssetTypeAudience)
	require.NoError(t, repo.CreateAsset(ctx, asset))

	fav := domain.NewFavourite(userID, asset.ID)
	err := repo.AddFavourite(ctx, fav)
	require.NoError(t, err)

	// Verify favourite exists
	isFav, err := repo.IsFavourite(ctx, userID, asset.ID)
	require.NoError(t, err)
	assert.True(t, isFav)
}

func TestMemoryRepository_AddFavourite_AssetNotFound(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	userID := uuid.New()

	fav := domain.NewFavourite(userID, uuid.New())
	err := repo.AddFavourite(ctx, fav)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMemoryRepository_AddFavourite_AlreadyExists(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	userID := uuid.New()

	asset := createTestAsset(t, domain.AssetTypeChart)
	require.NoError(t, repo.CreateAsset(ctx, asset))

	fav := domain.NewFavourite(userID, asset.ID)
	require.NoError(t, repo.AddFavourite(ctx, fav))

	// Try to add same asset as favourite again
	fav2 := domain.NewFavourite(userID, asset.ID)
	err := repo.AddFavourite(ctx, fav2)
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestMemoryRepository_ListFavourites_Pagination(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	userID := uuid.New()

	// Create 5 assets and favourite them
	assets := make([]*domain.Asset, 5)
	for i := 0; i < 5; i++ {
		assets[i] = createTestAsset(t, domain.AssetTypeChart)
		require.NoError(t, repo.CreateAsset(ctx, assets[i]))

		fav := domain.NewFavourite(userID, assets[i].ID)
		require.NoError(t, repo.AddFavourite(ctx, fav))

		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Test pagination
	query := domain.NewPageQuery(2, 0, "created_at", "desc")
	favs, total, err := repo.ListFavourites(ctx, userID, query)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, favs, 2)

	// Get next page
	query.Offset = 2
	favs, total, err = repo.ListFavourites(ctx, userID, query)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, favs, 2)
}

func TestMemoryRepository_ListFavourites_Sorting(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	userID := uuid.New()

	// Create assets with different types
	chartAsset := createTestAsset(t, domain.AssetTypeChart)
	insightAsset := createTestAsset(t, domain.AssetTypeInsight)

	require.NoError(t, repo.CreateAsset(ctx, insightAsset))
	require.NoError(t, repo.CreateAsset(ctx, chartAsset))

	require.NoError(t, repo.AddFavourite(ctx, domain.NewFavourite(userID, insightAsset.ID)))
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, repo.AddFavourite(ctx, domain.NewFavourite(userID, chartAsset.ID)))

	// Sort by type ascending
	query := domain.NewPageQuery(10, 0, "type", "asc")
	favs, _, err := repo.ListFavourites(ctx, userID, query)
	require.NoError(t, err)
	require.Len(t, favs, 2)
	assert.Equal(t, domain.AssetTypeChart, favs[0].Asset.Type)
	assert.Equal(t, domain.AssetTypeInsight, favs[1].Asset.Type)
}

func TestMemoryRepository_RemoveFavourite(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	userID := uuid.New()

	asset := createTestAsset(t, domain.AssetTypeChart)
	require.NoError(t, repo.CreateAsset(ctx, asset))

	fav := domain.NewFavourite(userID, asset.ID)
	require.NoError(t, repo.AddFavourite(ctx, fav))

	// Remove favourite
	err := repo.RemoveFavourite(ctx, userID, fav.ID)
	require.NoError(t, err)

	// Verify removal
	isFav, err := repo.IsFavourite(ctx, userID, asset.ID)
	require.NoError(t, err)
	assert.False(t, isFav)
}

func TestMemoryRepository_RemoveFavourite_NotFound(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	err := repo.RemoveFavourite(ctx, uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestMemoryRepository_Sanity(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	// Test with clean state
	err := repo.Sanity(ctx)
	require.NoError(t, err)

	// Create orphan favourite (manually for testing)
	userID := uuid.New()
	orphanAssetID := uuid.New()
	repo.favourites[userID] = make(map[uuid.UUID]*domain.Favourite)
	repo.favourites[userID][uuid.New()] = &domain.Favourite{
		ID:        uuid.New(),
		UserID:    userID,
		AssetID:   orphanAssetID,
		CreatedAt: time.Now(),
	}

	// Sanity should detect orphan
	err = repo.Sanity(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "orphan favourite")
}

// Concurrency Tests
func TestMemoryRepository_ConcurrentWrites(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()
	const goroutines = 50

	var wg sync.WaitGroup
	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			asset := createTestAsset(t, domain.AssetTypeChart)
			if err := repo.CreateAsset(ctx, asset); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent write failed: %v", err)
	}

	// Verify all assets were created
	assert.Len(t, repo.assets, goroutines)
}

func TestMemoryRepository_ConcurrentReadWrite(t *testing.T) {
	repo := NewRepository()
	ctx := context.Background()

	asset := createTestAsset(t, domain.AssetTypeChart)
	require.NoError(t, repo.CreateAsset(ctx, asset))

	var wg sync.WaitGroup
	const readers = 20
	const writers = 10

	// Concurrent readers
	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repo.GetAsset(ctx, asset.ID)
			assert.NoError(t, err)
		}()
	}

	// Concurrent writers
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			desc := "Updated " + string(rune(idx))
			err := repo.UpdateAssetDescription(ctx, asset.ID, desc)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
}

func TestMemoryRepository_ListAssets(t *testing.T) {
	ctx := context.Background()

	t.Run("empty repository returns empty list", func(t *testing.T) {
		repo := NewRepository()
		query := domain.NewPageQuery(20, 0, "created_at", "desc")

		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 0, total)
		assert.Equal(t, 0, len(assets))
	})

	t.Run("lists all assets with default sorting", func(t *testing.T) {
		repo := NewRepository()

		// Create test assets with deliberate time delays
		asset1, _ := domain.NewAsset(domain.AssetTypeChart, "First Chart", domain.ChartData{
			Title: "Chart 1", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})
		time.Sleep(10 * time.Millisecond)

		asset2, _ := domain.NewAsset(domain.AssetTypeInsight, "Second Insight", domain.InsightData{
			Text: "Insight text",
		})
		time.Sleep(10 * time.Millisecond)

		asset3, _ := domain.NewAsset(domain.AssetTypeAudience, "Third Audience", domain.AudienceData{
			Gender: "Male", BirthCountry: "USA", AgeGroups: []string{"18-24"},
			HoursSocialDaily: 3.0, PurchasesLastMonth: 5,
		})

		repo.CreateAsset(ctx, asset1)
		repo.CreateAsset(ctx, asset2)
		repo.CreateAsset(ctx, asset3)

		query := domain.NewPageQuery(20, 0, "created_at", "desc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Equal(t, 3, len(assets))

		// Verify descending order by created_at (newest first)
		assert.Equal(t, asset3.ID, assets[0].ID)
		assert.Equal(t, asset2.ID, assets[1].ID)
		assert.Equal(t, asset1.ID, assets[2].ID)
	})

	t.Run("pagination with limit and offset", func(t *testing.T) {
		repo := NewRepository()

		// Create 5 assets
		for i := 0; i < 5; i++ {
			asset, _ := domain.NewAsset(domain.AssetTypeChart, fmt.Sprintf("Chart %d", i), domain.ChartData{
				Title: fmt.Sprintf("Title %d", i), AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
			})
			repo.CreateAsset(ctx, asset)
			time.Sleep(5 * time.Millisecond) // Ensure different timestamps
		}

		// Get page 2 (items 2-3)
		query := domain.NewPageQuery(2, 2, "created_at", "asc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 5, total)
		assert.Equal(t, 2, len(assets))
	})

	t.Run("offset beyond total returns empty list", func(t *testing.T) {
		repo := NewRepository()

		asset, _ := domain.NewAsset(domain.AssetTypeChart, "Only Chart", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})
		repo.CreateAsset(ctx, asset)

		query := domain.NewPageQuery(20, 100, "created_at", "desc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 1, total) // Total count still returned
		assert.Equal(t, 0, len(assets))
	})

	t.Run("sorts by type ascending", func(t *testing.T) {
		repo := NewRepository()

		// Create assets in reverse order
		audienceAsset, _ := domain.NewAsset(domain.AssetTypeAudience, "Audience", domain.AudienceData{
			Gender: "Female", BirthCountry: "UK", AgeGroups: []string{"25-34"},
			HoursSocialDaily: 2.5, PurchasesLastMonth: 3,
		})
		insightAsset, _ := domain.NewAsset(domain.AssetTypeInsight, "Insight", domain.InsightData{
			Text: "Some insight",
		})
		chartAsset, _ := domain.NewAsset(domain.AssetTypeChart, "Chart", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})

		repo.CreateAsset(ctx, audienceAsset)
		repo.CreateAsset(ctx, insightAsset)
		repo.CreateAsset(ctx, chartAsset)

		query := domain.NewPageQuery(20, 0, "type", "asc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Equal(t, domain.AssetTypeAudience, assets[0].Type)
		assert.Equal(t, domain.AssetTypeChart, assets[1].Type)
		assert.Equal(t, domain.AssetTypeInsight, assets[2].Type)
	})

	t.Run("sorts by type descending", func(t *testing.T) {
		repo := NewRepository()

		audienceAsset, _ := domain.NewAsset(domain.AssetTypeAudience, "Audience", domain.AudienceData{
			Gender: "Female", BirthCountry: "UK", AgeGroups: []string{"25-34"},
			HoursSocialDaily: 2.5, PurchasesLastMonth: 3,
		})
		chartAsset, _ := domain.NewAsset(domain.AssetTypeChart, "Chart", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})

		repo.CreateAsset(ctx, audienceAsset)
		repo.CreateAsset(ctx, chartAsset)

		query := domain.NewPageQuery(20, 0, "type", "desc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Equal(t, domain.AssetTypeChart, assets[0].Type)
		assert.Equal(t, domain.AssetTypeAudience, assets[1].Type)
	})

	t.Run("sorts by description ascending", func(t *testing.T) {
		repo := NewRepository()

		asset1, _ := domain.NewAsset(domain.AssetTypeChart, "Zebra Chart", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})
		asset2, _ := domain.NewAsset(domain.AssetTypeChart, "Apple Chart", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})
		asset3, _ := domain.NewAsset(domain.AssetTypeChart, "Mango Chart", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})

		repo.CreateAsset(ctx, asset1)
		repo.CreateAsset(ctx, asset2)
		repo.CreateAsset(ctx, asset3)

		query := domain.NewPageQuery(20, 0, "description", "asc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 3, total)
		assert.Equal(t, "Apple Chart", assets[0].Description)
		assert.Equal(t, "Mango Chart", assets[1].Description)
		assert.Equal(t, "Zebra Chart", assets[2].Description)
	})

	t.Run("sorts by description descending", func(t *testing.T) {
		repo := NewRepository()

		asset1, _ := domain.NewAsset(domain.AssetTypeChart, "Alpha", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})
		asset2, _ := domain.NewAsset(domain.AssetTypeChart, "Zulu", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})

		repo.CreateAsset(ctx, asset1)
		repo.CreateAsset(ctx, asset2)

		query := domain.NewPageQuery(20, 0, "description", "desc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Equal(t, "Zulu", assets[0].Description)
		assert.Equal(t, "Alpha", assets[1].Description)
	})

	t.Run("sorts by updated_at", func(t *testing.T) {
		repo := NewRepository()

		asset1, _ := domain.NewAsset(domain.AssetTypeChart, "Chart 1", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})
		repo.CreateAsset(ctx, asset1)
		time.Sleep(10 * time.Millisecond)

		asset2, _ := domain.NewAsset(domain.AssetTypeChart, "Chart 2", domain.ChartData{
			Title: "Chart", AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
		})
		repo.CreateAsset(ctx, asset2)
		time.Sleep(10 * time.Millisecond)

		// Update asset1 (making it the most recently updated)
		repo.UpdateAssetDescription(ctx, asset1.ID, "Updated Chart 1")
		time.Sleep(10 * time.Millisecond)

		query := domain.NewPageQuery(20, 0, "updated_at", "desc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Equal(t, asset1.ID, assets[0].ID) // Most recently updated
		assert.Equal(t, asset2.ID, assets[1].ID)
	})

	t.Run("concurrent read access is thread-safe", func(t *testing.T) {
		repo := NewRepository()

		// Create some assets
		for i := 0; i < 10; i++ {
			asset, _ := domain.NewAsset(domain.AssetTypeChart, fmt.Sprintf("Chart %d", i), domain.ChartData{
				Title: fmt.Sprintf("Title %d", i), AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
			})
			repo.CreateAsset(ctx, asset)
		}

		// Simulate concurrent reads
		var wg sync.WaitGroup
		errors := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				query := domain.NewPageQuery(5, 0, "created_at", "desc")
				_, _, err := repo.ListAssets(ctx, query)
				if err != nil {
					errors <- err
				}
			}()
		}

		wg.Wait()
		close(errors)

		// No errors should occur
		for err := range errors {
			t.Errorf("Concurrent access error: %v", err)
		}
	})

	t.Run("partial page at end of results", func(t *testing.T) {
		repo := NewRepository()

		// Create 7 assets
		for i := 0; i < 7; i++ {
			asset, _ := domain.NewAsset(domain.AssetTypeChart, fmt.Sprintf("Chart %d", i), domain.ChartData{
				Title: fmt.Sprintf("Title %d", i), AxisXTitle: "X", AxisYTitle: "Y", Data: [][]float64{{1, 2}},
			})
			repo.CreateAsset(ctx, asset)
		}

		// Request page size of 5, offset 5 (should return 2 items)
		query := domain.NewPageQuery(5, 5, "created_at", "desc")
		assets, total, err := repo.ListAssets(ctx, query)

		assert.NoError(t, err)
		assert.Equal(t, 7, total)
		assert.Equal(t, 2, len(assets)) // Only 2 items remaining
	})
}

// Helper function
func createTestAsset(t *testing.T, assetType domain.AssetType) *domain.Asset {
	t.Helper()

	var data interface{}
	switch assetType {
	case domain.AssetTypeChart:
		data = domain.ChartData{
			Title:      "Test Chart",
			AxisXTitle: "X",
			AxisYTitle: "Y",
			Data:       [][]float64{{1, 2}},
		}
	case domain.AssetTypeInsight:
		data = domain.InsightData{
			Text: "This is a test insight.",
		}
	case domain.AssetTypeAudience:
		data = domain.AudienceData{
			Gender:             "Male",
			BirthCountry:       "USA",
			AgeGroups:          []string{"25-34", "35-44"},
			HoursSocialDaily:   2.5,
			PurchasesLastMonth: 5,
		}
	default:
		t.Fatalf("unsupported asset type for test: %s", assetType)
	}

	asset, err := domain.NewAsset(assetType, "Test Asset", data)
	require.NoError(t, err)
	return asset
}
