package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gioannid/platform-go-challenge/internal/config"
	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/gioannid/platform-go-challenge/internal/handler"
	"github.com/gioannid/platform-go-challenge/internal/middleware"
	"github.com/gioannid/platform-go-challenge/internal/repository/memory"
	"github.com/gioannid/platform-go-challenge/internal/server"
	"github.com/gioannid/platform-go-challenge/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer creates a test server with all dependencies
func setupTestServer(t *testing.T) (*httptest.Server, *memory.MemoryRepository) {
	t.Helper()

	cfg := &config.Config{
		ServerAddress: ":0",
		AuthEnabled:   false,
	}

	repo := memory.NewRepository()
	svc := service.NewFavouriteService(repo)
	h := handler.NewHandler(svc)
	mw := server.NewChain(middleware.Logger())

	srv := server.New(cfg, h, mw)
	testServer := httptest.NewServer(srv.Router())

	return testServer, repo
}

func TestIntegration_CompleteWorkflow(t *testing.T) {
	ts, repo := setupTestServer(t)
	defer ts.Close()

	ctx := context.Background()
	userID := uuid.New()

	// Step 1: Create an asset
	chartData := domain.ChartData{
		Title:      "Q4 Sales",
		AxisXTitle: "Month",
		AxisYTitle: "Revenue",
		Data:       [][]float64{{1, 1000}, {2, 1500}, {3, 2000}},
	}

	createAssetReq := map[string]interface{}{
		"type":        "chart",
		"description": "Quarterly Sales Report",
		"data":        chartData,
	}

	body, _ := json.Marshal(createAssetReq)
	resp, err := http.Post(ts.URL+"/api/v1/assets", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createAssetResp handler.Response
	err = json.NewDecoder(resp.Body).Decode(&createAssetResp)
	require.NoError(t, err)
	require.True(t, createAssetResp.Success)

	assetData := createAssetResp.Data.(map[string]interface{})
	assetID := uuid.MustParse(assetData["id"].(string))

	// Step 2: Add asset to favourites
	addFavReq := map[string]interface{}{
		"asset_id": assetID.String(),
	}

	body, _ = json.Marshal(addFavReq)
	resp, err = http.Post(ts.URL+"/api/v1/users/"+userID.String()+"/favourites", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var addFavResp handler.Response
	err = json.NewDecoder(resp.Body).Decode(&addFavResp)
	require.NoError(t, err)
	require.True(t, addFavResp.Success)

	favData := addFavResp.Data.(map[string]interface{})
	favouriteID := uuid.MustParse(favData["id"].(string))

	// Step 3: List favourites
	resp, err = http.Get(ts.URL + "/api/v1/users/" + userID.String() + "/favourites")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp handler.Response
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)
	require.True(t, listResp.Success)

	listData := listResp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), listData["total"])

	// Step 4: Update asset description
	updateDescReq := map[string]interface{}{
		"description": "Updated Q4 Sales Report",
	}

	body, _ = json.Marshal(updateDescReq)
	req, _ := http.NewRequest(http.MethodPatch, ts.URL+"/api/v1/assets/"+assetID.String()+"/description", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify description was updated
	asset, err := repo.GetAsset(ctx, assetID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Q4 Sales Report", asset.Description)

	// Step 5: Remove favourite
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/users/"+userID.String()+"/favourites/"+favouriteID.String(), nil)
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify favourite was removed
	isFav, err := repo.IsFavourite(ctx, userID, assetID)
	require.NoError(t, err)
	assert.False(t, isFav)

	// Step 6: Delete asset
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/assets/"+assetID.String(), nil)
	_, err = client.Do(req)
	require.NoError(t, err)

	// Verify asset was deleted
	_, err = repo.GetAsset(ctx, assetID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_DeleteAssetRemovesFavourite(t *testing.T) {
	ts, repo := setupTestServer(t)
	defer ts.Close()

	ctx := context.Background()
	userID := uuid.New()

	// Step 1: Create an asset
	chartData := domain.ChartData{Title: "Temporary Chart"}
	createAssetReq := map[string]interface{}{
		"type":        "chart",
		"description": "Asset to be deleted",
		"data":        chartData,
	}
	body, _ := json.Marshal(createAssetReq)
	resp, err := http.Post(ts.URL+"/api/v1/assets", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createAssetResp handler.Response
	err = json.NewDecoder(resp.Body).Decode(&createAssetResp)
	require.NoError(t, err)
	assetData := createAssetResp.Data.(map[string]interface{})
	assetID := uuid.MustParse(assetData["id"].(string))

	// Step 2: Add asset to favourites
	addFavReq := map[string]interface{}{"asset_id": assetID.String()}
	body, _ = json.Marshal(addFavReq)
	resp, err = http.Post(ts.URL+"/api/v1/users/"+userID.String()+"/favourites", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Verify favourite exists
	isFavBeforeDelete, err := repo.IsFavourite(ctx, userID, assetID)
	require.NoError(t, err)
	assert.True(t, isFavBeforeDelete, "Favourite should exist before asset deletion")

	// Step 3: Delete the asset
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/assets/"+assetID.String(), nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Step 4: Verify asset was deleted
	_, err = repo.GetAsset(ctx, assetID)
	assert.ErrorIs(t, err, domain.ErrNotFound, "Asset should be deleted")

	// Step 5: Verify favourite no longer exists for the user
	isFavAfterDelete, err := repo.IsFavourite(ctx, userID, assetID)
	require.NoError(t, err)
	assert.False(t, isFavAfterDelete, "Favourite should be removed after asset deletion")

	// Step 6: Try to list favourites for the user, expecting the deleted asset not to appear
	resp, err = http.Get(ts.URL + "/api/v1/users/" + userID.String() + "/favourites")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp handler.Response
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)

	listData := listResp.Data.(map[string]interface{})
	assert.Equal(t, float64(0), listData["total"], "No favourites should be listed after asset deletion")
	assert.Empty(t, listData["favourites"], "Favourites list should be empty after asset deletion")
}

func TestIntegration_PaginationAndSorting(t *testing.T) {
	ts, repo := setupTestServer(t)
	defer ts.Close()

	ctx := context.Background()
	userID := uuid.New()

	// Create 10 assets and favourite them
	assetIDs := make([]uuid.UUID, 10)
	for i := 0; i < 10; i++ {
		asset, err := domain.NewAsset(
			domain.AssetTypeChart,
			string(rune('A'+i))+" Chart",
			domain.ChartData{
				Title:      "Chart " + string(rune('0'+i)),
				AxisXTitle: "X",
				AxisYTitle: "Y",
				Data:       [][]float64{{float64(i), float64(i * 100)}},
			},
		)
		require.NoError(t, err)
		require.NoError(t, repo.CreateAsset(ctx, asset))

		fav := domain.NewFavourite(userID, asset.ID)
		require.NoError(t, repo.AddFavourite(ctx, fav))

		assetIDs[i] = asset.ID
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Test pagination: page 1
	resp, err := http.Get(ts.URL + "/api/v1/users/" + userID.String() + "/favourites?limit=5&offset=0")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var page1 handler.Response
	json.NewDecoder(resp.Body).Decode(&page1)
	data1 := page1.Data.(map[string]interface{})
	assert.Equal(t, float64(10), data1["total"])
	assert.Equal(t, float64(5), data1["limit"])

	// Test pagination: page 2
	resp, err = http.Get(ts.URL + "/api/v1/users/" + userID.String() + "/favourites?limit=5&offset=5")
	require.NoError(t, err)

	var page2 handler.Response
	json.NewDecoder(resp.Body).Decode(&page2)
	data2 := page2.Data.(map[string]interface{})
	assert.Equal(t, float64(10), data2["total"])

	// Test sorting by description
	resp, err = http.Get(ts.URL + "/api/v1/users/" + userID.String() + "/favourites?sortBy=description&order=asc")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIntegration_ListAssets(t *testing.T) {
	ts, repo := setupTestServer(t)
	defer ts.Close()

	ctx := context.Background()

	// Create 5 assets with varying descriptions and creation times
	asset1, _ := domain.NewAsset(
		domain.AssetTypeChart,
		"Alpha Chart",
		domain.ChartData{
			Title:      "Chart 1",
			AxisXTitle: "X",
			AxisYTitle: "Y",
			Data:       [][]float64{{1, 100}},
		},
	)
	time.Sleep(10 * time.Millisecond)
	asset2, _ := domain.NewAsset(
		domain.AssetTypeInsight,
		"Beta Insight",
		domain.InsightData{Text: "This is an insight."},
	)
	time.Sleep(10 * time.Millisecond)
	asset3, _ := domain.NewAsset(
		domain.AssetTypeAudience,
		"Gamma Audience",
		domain.AudienceData{
			Gender:             "Female",
			BirthCountry:       "USA",
			AgeGroups:          []string{"25-34", "35-44"},
			HoursSocialDaily:   2.5,
			PurchasesLastMonth: 5,
		},
	)
	time.Sleep(10 * time.Millisecond)
	asset4, _ := domain.NewAsset(
		domain.AssetTypeChart,
		"Delta Chart",
		domain.ChartData{
			Title:      "Chart 4",
			AxisXTitle: "X",
			AxisYTitle: "Y",
			Data:       [][]float64{{1, 100}},
		},
	)
	time.Sleep(10 * time.Millisecond)
	asset5, _ := domain.NewAsset(
		domain.AssetTypeInsight,
		"Epsilon Insight",
		domain.InsightData{Text: "Another important insight."},
	)
	require.NoError(t, repo.CreateAsset(ctx, asset1))
	require.NoError(t, repo.CreateAsset(ctx, asset2))
	require.NoError(t, repo.CreateAsset(ctx, asset3))
	require.NoError(t, repo.CreateAsset(ctx, asset4))
	require.NoError(t, repo.CreateAsset(ctx, asset5))

	// Test default list (no params)
	resp, err := http.Get(ts.URL + "/api/v1/assets")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var listResp handler.Response
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)
	require.True(t, listResp.Success)
	listData := listResp.Data.(map[string]interface{})
	assert.Equal(t, float64(5), listData["total"])
	assets := listData["assets"].([]interface{})
	assert.Len(t, assets, 5)

	// Test pagination
	resp, err = http.Get(ts.URL + "/api/v1/assets?limit=2&offset=1")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)
	listData = listResp.Data.(map[string]interface{})
	assert.Equal(t, float64(5), listData["total"])
	assets = listData["assets"].([]interface{})
	assert.Len(t, assets, 2)

	// Test sorting by description ascending
	resp, err = http.Get(ts.URL + "/api/v1/assets?sortBy=description&order=asc")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)
	listData = listResp.Data.(map[string]interface{})
	assets = listData["assets"].([]interface{})
	assert.Equal(t, "Alpha Chart", assets[0].(map[string]interface{})["description"])
	assert.Equal(t, "Beta Insight", assets[1].(map[string]interface{})["description"])

	// Test sorting by created_at descending (newest first)
	resp, err = http.Get(ts.URL + "/api/v1/assets?sortBy=created_at&order=desc")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)
	listData = listResp.Data.(map[string]interface{})
	assets = listData["assets"].([]interface{})
	assert.Equal(t, "Epsilon Insight", assets[0].(map[string]interface{})["description"])
	assert.Equal(t, "Delta Chart", assets[1].(map[string]interface{})["description"])
}

func TestIntegration_ErrorCases(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	tests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "add favourite - asset not found",
			method:         http.MethodPost,
			url:            "/api/v1/users/" + uuid.New().String() + "/favourites",
			body:           map[string]interface{}{"asset_id": uuid.New().String()},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "add favourite - duplicate",
			method:         http.MethodPost,
			url:            "/api/v1/users/invalid-uuid/favourites",
			body:           map[string]interface{}{"asset_id": uuid.New().String()},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "create asset - invalid type",
			method:         http.MethodPost,
			url:            "/api/v1/assets",
			body:           map[string]interface{}{"type": "invalid", "description": "Test", "data": map[string]interface{}{}},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Reader
			if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				body = bytes.NewReader(b)
			}

			var req *http.Request
			var err error
			if body != nil {
				req, err = http.NewRequest(tt.method, ts.URL+tt.url, body)
			} else {
				req, err = http.NewRequest(tt.method, ts.URL+tt.url, nil)
			}
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestIntegration_HealthCheck(t *testing.T) {
	ts, _ := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResp handler.Response
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	require.NoError(t, err)
	assert.True(t, healthResp.Success)
}
