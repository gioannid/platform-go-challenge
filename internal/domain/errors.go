package domain

import "errors"

// Domain-level errors
var (
	ErrNotFound            = errors.New("resource not found")
	ErrAlreadyExists       = errors.New("resource already exists")
	ErrInvalidAssetType    = errors.New("invalid asset type")
	ErrMissingAssetData    = errors.New("missing asset data")
	ErrInvalidChartData    = errors.New("invalid chart data")
	ErrInvalidInsightData  = errors.New("invalid insight data")
	ErrInvalidAudienceData = errors.New("invalid audience data")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
)
