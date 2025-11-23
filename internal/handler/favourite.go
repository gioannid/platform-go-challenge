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

// / ListFavourites handles GET /users/{userId}/favourites
//
//	@Summary		List user favourites
//	@Description	Get paginated list of all favourites for a specific user
//	@Tags			favourites
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		string	true	"User ID (UUID)"
//	@Param			limit	query		int		false	"Number of items per page"	default(20)
//	@Param			offset	query		int		false	"Number of items to skip"	default(0)
//	@Param			sortBy	query		string	false	"Sort field"				Enums(created_at, updated_at)
//	@Param			order	query		string	false	"Sort order"				Enums(asc, desc)
//	@Success		200		{object}	Response{data=ListFavouritesResponse}
//	@Failure		400		{object}	InvalidUUIDError
//	@Failure		404		{object}	NotFoundError
//	@Failure		500		{object}	InternalServerError
//	@Router			/users/{userId}/favourites [get]
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
//
//	@Summary		Add favourite
//	@Description	Add an asset to user's favourites
//	@Tags			favourites
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		string				true	"User ID (UUID)"
//	@Param			request	body		AddFavouriteRequest	true	"Favourite details"
//	@Success		201		{object}	Response{data=domain.Favourite}
//	@Failure		400		{object}	BadRequestError
//	@Failure		404		{object}	NotFoundError
//	@Failure		409		{object}	ConflictError
//	@Failure		500		{object}	InternalServerError
//	@Router			/users/{userId}/favourites [post]
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
//
//	@Summary		Remove favourite
//	@Description	Remove an asset from user's favourites
//	@Tags			favourites
//	@Accept			json
//	@Produce		json
//	@Param			userId		path		string	true	"User ID (UUID)"
//	@Param			favouriteId	path		string	true	"Favourite ID (UUID)"
//	@Success		200			{object}	SuccessResponse
//	@Failure		400			{object}	InvalidUUIDError
//	@Failure		404			{object}	NotFoundError
//	@Failure		500			{object}	InternalServerError
//	@Router			/users/{userId}/favourites/{favouriteId} [delete]
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
//
//	@Summary		Update asset description
//	@Description	Update the description of an existing asset
//	@Tags			assets
//	@Accept			json
//	@Produce		json
//	@Param			assetId	path		string							true	"Asset ID (UUID)"
//	@Param			request	body		UpdateAssetDescriptionRequest	true	"New description"
//	@Success		200		{object}	SuccessResponse
//	@Failure		400		{object}	BadRequestError
//	@Failure		404		{object}	NotFoundError
//	@Failure		500		{object}	InternalServerError
//	@Router			/assets/{assetId}/description [patch]
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
	Data        json.RawMessage  `json:"data" swaggertype:"object"`
}

// CreateAsset handles POST /assets
//
//	@Summary		Create asset
//	@Description	Create a new asset of type chart, insight, or audience.
//	@Description
//	@Description	**Chart Example:**
//	@Description	```
//	@Description	{
//	@Description	  "type": "chart",
//	@Description	  "description": "Monthly sales data",
//	@Description	  "data": {
//	@Description	    "title": "Q4 2025 Sales",
//	@Description	    "axis_x_title": "Month",
//	@Description	    "axis_y_title": "Revenue (USD)",
//	@Description	    "data": [[100, 200], [300, 400]]
//	@Description	  }
//	@Description	}
//	@Description	```
//	@Description
//	@Description	**Insight Example:**
//	@Description	```
//	@Description	{
//	@Description	  "type": "insight",
//	@Description	  "description": "Social media usage",
//	@Description	  "data": {
//	@Description	    "text": "40% of millennials spend 3+ hours daily on social media"
//	@Description	  }
//	@Description	}
//	@Description	```
//	@Description
//	@Description	**Audience Example:**
//	@Description	```
//	@Description	{
//	@Description	  "type": "audience",
//	@Description	  "description": "Target demographic",
//	@Description	  "data": {
//	@Description	    "gender": "Male",
//	@Description	    "birth_country": "USA",
//	@Description	    "age_groups": ["24-35"],
//	@Description	    "hours_social_daily": 3.5,
//	@Description	    "purchases_last_month": 5
//	@Description	  }
//	@Description	}
//	@Description	```
//	@Tags			assets
//	@Accept			json
//	@Produce		json
//	@Param			request	body		CreateAssetRequest		true	"Asset creation request"
//	@Success		201		{object}	Response{data=domain.Asset}
//	@Failure		400		{object}	BadRequestError
//	@Failure		500		{object}	InternalServerError
//	@Router			/assets [post]
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

// DeleteAsset handles DELETE /assets/{assetId}
//
//	@Summary		Delete asset
//	@Description	Delete an existing asset
//	@Tags			assets
//	@Accept			json
//	@Produce		json
//	@Param			assetId	path		string	true	"Asset ID (UUID)"
//	@Success		200		{object}	SuccessResponse
//	@Failure		400		{object}	InvalidUUIDError
//	@Failure		404		{object}	NotFoundError
//	@Failure		500		{object}	InternalServerError
//	@Router			/assets/{assetId} [delete]
func (h *Handler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assetID, err := uuid.Parse(vars["assetId"])
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.service.DeleteAsset(r.Context(), assetID); err != nil {
		respondError(w, mapDomainError(err), err)
		return
	}

	respondSuccess(w, http.StatusOK, nil, "Asset deleted successfully")
}

// ListAssetsResponse represents paginated assets response
type ListAssetsResponse struct {
	Assets []*domain.Asset `json:"assets"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// ListAssets handles GET /assets
//
//	@Summary		List all assets
//	@Description	Get paginated list of all assets in the system
//	@Tags			assets
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int		false	"Number of items per page"	default(20)
//	@Param			offset	query		int		false	"Number of items to skip"	default(0)
//	@Param			sortBy	query		string	false	"Sort field"				Enums(created_at, updated_at, type, description)
//	@Param			order	query		string	false	"Sort order"				Enums(asc, desc)
//	@Success		200		{object}	Response{data=ListAssetsResponse}
//	@Failure		400		{object}	BadRequestError
//	@Failure		500		{object}	InternalServerError
//	@Router			/assets [get]
func (h *Handler) ListAssets(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	sortBy := r.URL.Query().Get("sortBy")
	order := r.URL.Query().Get("order")

	query := domain.NewPageQuery(limit, offset, sortBy, order)

	assets, total, err := h.service.ListAssets(r.Context(), query)
	if err != nil {
		respondError(w, mapDomainError(err), err)
		return
	}

	respondSuccess(w, http.StatusOK, ListAssetsResponse{
		Assets: assets,
		Total:  total,
		Limit:  query.Limit,
		Offset: query.Offset,
	}, "")
}
