package paymenthandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/utils"
)

// HandlerCreatePayment handles HTTP POST requests to create a new payment intent.
// Parses the request body for payment parameters, validates them, and delegates creation to the payment service.
// On success, logs the event and responds with the client secret; on error, logs and returns the appropriate error response.
// Parameters:
//   - w: http.ResponseWriter for sending the response
//   - r: *http.Request containing the request data
//   - user: database.User representing the authenticated user
func (cfg *HandlersPaymentConfig) HandlerCreatePayment(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var req CreatePaymentIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		cfg.Logger.LogHandlerError(
			ctx,
			"create_payment",
			"invalid_request",
			"Invalid request payload",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get payment service
	paymentService := cfg.GetPaymentService()

	// Create payment using service
	result, err := paymentService.CreatePayment(ctx, CreatePaymentParams{
		OrderID:  req.OrderID,
		UserID:   user.ID,
		Currency: req.Currency,
	})

	if err != nil {
		cfg.handlePaymentError(w, r, err, "create_payment", ip, userAgent)
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	cfg.Logger.LogHandlerSuccess(ctxWithUserID, "create_payment", "Created payment successful", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusCreated, CreatePaymentIntentResponse{
		ClientSecret: result.ClientSecret,
	})
}
