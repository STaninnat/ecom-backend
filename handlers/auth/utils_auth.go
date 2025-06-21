package authhandlers

import (
	"context"
	"net/http"

	"github.com/STaninnat/ecom-backend/utils"
)

func (apicfg *HandlersAuthConfig) MergeCart(r *http.Request, ctx context.Context, userID string) {
	sessionID := utils.GetSessionIDFromRequest(r)
	if sessionID != "" {
		guestCart, err := apicfg.GetGuestCart(ctx, sessionID)
		if err == nil && len(guestCart.Items) > 0 {
			_ = apicfg.CartMG.MergeGuestCartToUser(ctx, userID, guestCart.Items)
			_ = apicfg.DeleteGuestCart(ctx, sessionID)
		}
	}
}
