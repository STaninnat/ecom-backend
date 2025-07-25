// Package reviewhandlers provides HTTP handlers for managing product reviews, including CRUD operations and listing with filters and pagination.
package reviewhandlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// handler_review_get.go: Handlers and helpers for fetching reviews with support for pagination, filtering, sorting, and error handling.

// parsePagination extracts and validates pagination parameters from the HTTP request query string.
// Parses the 'page' and 'pageSize' query parameters, providing default values (page=1, pageSize=10)
// if they are missing or invalid. Only positive integer values are accepted.
// Parameters:
//   - r: *http.Request containing the query parameters to parse
//
// Returns:
//   - page: int representing the current page number (defaults to 1)
//   - pageSize: int representing the number of items per page (defaults to 10)
func parsePagination(r *http.Request) (page, pageSize int) {
	page = 1
	pageSize = 10
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = v
		}
	}
	return
}

// parseFilterSort parses all supported filter and sort query parameters from the HTTP request.
// Extracts rating filters, date ranges, media filters, and sort options, validating each parameter
// according to its expected format and range. Invalid parameters are ignored and return nil/default values.
// Parameters:
//   - r: *http.Request containing the query parameters to parse
//
// Returns:
//   - rating: *int for exact rating filter (1-5), nil if invalid or not provided
//   - minRating: *int for minimum rating filter (1-5), nil if invalid or not provided
//   - maxRating: *int for maximum rating filter (1-5), nil if invalid or not provided
//   - from: *time.Time for start date filter (RFC3339 format), nil if invalid or not provided
//   - to: *time.Time for end date filter (RFC3339 format), nil if invalid or not provided
//   - hasMedia: *bool for media filter (true/false/1), nil if not provided
//   - sort: string for sort option, empty string if not provided
func parseFilterSort(r *http.Request) (rating *int, minRating *int, maxRating *int, from, to *time.Time, hasMedia *bool, sort string) {
	q := r.URL.Query()

	if v := q.Get("rating"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 1 && i <= 5 {
			rating = &i
		}
	}
	if v := q.Get("min_rating"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 1 && i <= 5 {
			minRating = &i
		}
	}
	if v := q.Get("max_rating"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 1 && i <= 5 {
			maxRating = &i
		}
	}
	if v := q.Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = &t
		}
	}
	if v := q.Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = &t
		}
	}
	if v := q.Get("has_media"); v != "" {
		b := strings.ToLower(v) == "true" || v == "1"
		hasMedia = &b
	}
	sort = q.Get("sort")
	return
}

// HandlerGetReviewsByProductID handles HTTP GET requests to retrieve paginated, filtered, and sorted reviews for a product.
// @Summary      Get reviews by product ID
// @Description  Retrieves paginated, filtered, and sorted reviews for a product
// @Tags         reviews
// @Produce      json
// @Param        product_id  path  string  true  "Product ID"
// @Param        page        query int     false "Page number"
// @Param        pageSize    query int     false "Page size"
// @Param        rating      query int     false "Exact rating filter"
// @Param        min_rating  query int     false "Minimum rating filter"
// @Param        max_rating  query int     false "Maximum rating filter"
// @Param        from        query string  false "Start date (RFC3339)"
// @Param        to          query string  false "End date (RFC3339)"
// @Param        has_media   query bool    false "Has media filter"
// @Param        sort        query string  false "Sort option"
// @Success      200  {object}  PaginatedReviewsResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/reviews/product/{product_id} [get]
func (cfg *HandlersReviewConfig) HandlerGetReviewsByProductID(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	productID := chi.URLParam(r, "product_id")
	if productID == "" {
		cfg.Logger.LogHandlerError(ctx, "get_reviews_by_product_id", "invalid_request", "Product ID is required", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}

	page, pageSize := parsePagination(r)
	rating, minRating, maxRating, from, to, hasMedia, sort := parseFilterSort(r)
	resultAny, err := cfg.GetReviewService().GetReviewsByProductIDPaginated(ctx, productID, page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort)
	if err != nil {
		cfg.handleReviewError(w, r, err, "get_reviews_by_product_id", ip, userAgent)
		return
	}
	result, ok := resultAny.(PaginatedReviewsResponse)
	if !ok {
		cfg.Logger.LogHandlerError(ctx, "get_reviews_by_product_id", "internal_error", "Unexpected response type", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	cfg.Logger.LogHandlerSuccess(ctx, "get_reviews_by_product_id", "Got reviews successfully", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, PaginatedReviewsResponse{
		Data:       result.Data,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
		Code:       "success",
		Message:    "Reviews fetched successfully",
	})
}

// HandlerGetReviewsByUserID handles HTTP GET requests to retrieve paginated, filtered, and sorted reviews for the authenticated user.
// @Summary      Get reviews by user
// @Description  Retrieves paginated, filtered, and sorted reviews for the authenticated user
// @Tags         reviews
// @Produce      json
// @Param        page        query int     false "Page number"
// @Param        pageSize    query int     false "Page size"
// @Param        rating      query int     false "Exact rating filter"
// @Param        min_rating  query int     false "Minimum rating filter"
// @Param        max_rating  query int     false "Maximum rating filter"
// @Param        from        query string  false "Start date (RFC3339)"
// @Param        to          query string  false "End date (RFC3339)"
// @Param        has_media   query bool    false "Has media filter"
// @Param        sort        query string  false "Sort option"
// @Success      200  {object}  PaginatedReviewsResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/reviews/user [get]
func (cfg *HandlersReviewConfig) HandlerGetReviewsByUserID(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	page, pageSize := parsePagination(r)
	rating, minRating, maxRating, from, to, hasMedia, sort := parseFilterSort(r)
	resultAny, err := cfg.GetReviewService().GetReviewsByUserIDPaginated(ctx, user.ID, page, pageSize, rating, minRating, maxRating, from, to, hasMedia, sort)
	if err != nil {
		cfg.handleReviewError(w, r, err, "get_reviews_by_user", ip, userAgent)
		return
	}
	result, ok := resultAny.(PaginatedReviewsResponse)
	if !ok {
		cfg.Logger.LogHandlerError(ctx, "get_reviews_by_user", "internal_error", "Unexpected response type", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "get_reviews_by_user", "Got user reviews successfully", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, PaginatedReviewsResponse{
		Data:       result.Data,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
		Code:       "success",
		Message:    "Reviews fetched successfully",
	})
}

// HandlerGetReviewByID handles HTTP GET requests to retrieve a single review by its ID.
// @Summary      Get review by ID
// @Description  Retrieves a single review by its ID
// @Tags         reviews
// @Produce      json
// @Param        id  path  string  true  "Review ID"
// @Success      200  {object}  handlers.APIResponse
// @Failure      400  {object}  map[string]string
// @Router       /v1/reviews/{id} [get]
func (cfg *HandlersReviewConfig) HandlerGetReviewByID(w http.ResponseWriter, r *http.Request) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	reviewID := chi.URLParam(r, "id")
	if reviewID == "" {
		cfg.Logger.LogHandlerError(ctx, "get_review_by_id", "invalid_request", "Review ID is required", ip, userAgent, nil)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Review ID is required")
		return
	}

	review, err := cfg.GetReviewService().GetReviewByID(ctx, reviewID)
	if err != nil {
		cfg.handleReviewError(w, r, err, "get_review_by_id", ip, userAgent)
		return
	}

	cfg.Logger.LogHandlerSuccess(ctx, "get_review_by_id", "Got review successfully", ip, userAgent)
	middlewares.RespondWithJSON(w, http.StatusOK, handlers.APIResponse{
		Message: "Review fetched successfully",
		Code:    "success",
		Data:    review,
	})
}
