package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gioannid/platform-go-challenge/internal/domain"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ListFavouritesRequest represents pagination parameters
type ListFavouritesRequest struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
}

// ListFavouritesResponse represents paginated favourites response
type ListFavouritesResponse struct {
	Favourites []*domain.Favourite `json:"favourites"`
	Total      int                 `json:"total"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

// ListFavourites handles GET /users/{userId}/favourites
func (h *Handler) ListFavourites(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	sortBy := r.URL.Query().Get("sortBy")
	order := r.URL.Query().Get("order")

	query := domain.NewPageQuery(limit, offset, sortBy, order)

	favourites, total, err := h.service.ListFavourites(r.Context(), userID, query)
	if err != nil {
		respondError(w, mapDomainError(err), err)
		return
	}

	respondSuccess(w, http.StatusOK, ListFavouritesResponse{
		Favourites: favourites,
		Total:      total,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}, "")
}

// AddFavouriteRequest represents the request to add a favourite
type AddFavouriteRequest struct {
	AssetID uuid.UUID `json:"asset_id"`
}

// AddFavourite handles POST /users/{userId}/favourites
func (h *Handler) AddFavourite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	var req AddFavouriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	favourite, err := h.service.AddFavourite(r.Context(), userID, req.AssetID)
	if err != nil {
		respondError(w, mapDomainError(err), err)
		return
	}

	respondSuccess(w, http.StatusCreated, favourite, "Favourite added successfully")
}

// RemoveFavourite handles DELETE /users/{userId}/favourites/{favouriteId}
func (h *Handler) RemoveFavourite(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userID, err := uuid.Parse(vars["userId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	favouriteID, err := uuid.Parse(vars["favouriteId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.service.RemoveFavourite(r.Context(), userID, favouriteID); err != nil {
		respondError(w, mapDomainError(err), err)
		return
	}

	respondSuccess(w, http.StatusOK, nil, "Favourite removed successfully")
}

// UpdateAssetDescriptionRequest represents the request to update description
type UpdateAssetDescriptionRequest struct {
	Description string `json:"description"`
}

// UpdateAssetDescription handles PATCH /assets/{assetId}/description
func (h *Handler) UpdateAssetDescription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID, err := uuid.Parse(vars["assetId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	var req UpdateAssetDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.service.UpdateAssetDescription(r.Context(), assetID, req.Description); err != nil {
		respondError(w, mapDomainError(err), err)
		return
	}

	respondSuccess(w, http.StatusOK, nil, "Asset description updated successfully")
}

// CreateAssetRequest represents the request to create an asset
type CreateAssetRequest struct {
	Type        domain.AssetType `json:"type"`
	Description string           `json:"description"`
	Data        json.RawMessage  `json:"data"`
}

// CreateAsset handles POST /assets
func (h *Handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	// Parse data based on type
	var data interface{}
	switch req.Type {
	case domain.AssetTypeChart:
		var chartData domain.ChartData
		if err := json.Unmarshal(req.Data, &chartData); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}
		data = chartData
	case domain.AssetTypeInsight:
		var insightData domain.InsightData
		if err := json.Unmarshal(req.Data, &insightData); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}
		data = insightData
	case domain.AssetTypeAudience:
		var audienceData domain.AudienceData
		if err := json.Unmarshal(req.Data, &audienceData); err != nil {
			respondError(w, http.StatusBadRequest, err)
			return
		}
		data = audienceData
	default:
		respondError(w, http.StatusBadRequest, domain.ErrInvalidAssetType)
		return
	}

	asset, err := h.service.CreateAsset(r.Context(), req.Type, req.Description, data)
	if err != nil {
		respondError(w, mapDomainError(err), err)
		return
	}

	respondSuccess(w, http.StatusCreated, asset, "Asset created successfully")
}
