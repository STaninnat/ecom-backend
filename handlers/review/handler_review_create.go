package reviewhandlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/STaninnat/ecom-backend/models"
	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersReviewConfig) HandlerCreateReview(w http.ResponseWriter, r *http.Request, user database.User) {
	ip, userAgent := handlers.GetRequestMetadata(r)
	ctx := r.Context()

	var input models.Review
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"create_review",
			"invalid request body",
			"Failed to parse body",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	input.UserID = user.ID

	if err := apicfg.ReviewMG.CreateReview(r.Context(), &input); err != nil {
		apicfg.HandlersConfig.LogHandlerError(
			ctx,
			"create_review",
			"create review failed",
			"Error creating review",
			ip, userAgent, err,
		)
		middlewares.RespondWithError(w, http.StatusInternalServerError, "Failed to create review")
		return
	}

	ctxWithUserID := context.WithValue(ctx, utils.ContextKeyUserID, user.ID)
	apicfg.HandlersConfig.LogHandlerSuccess(ctxWithUserID, "create_review", "Review created successfully", ip, userAgent)

	middlewares.RespondWithJSON(w, http.StatusOK, input)
}
