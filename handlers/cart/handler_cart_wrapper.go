package cart

import (
	"github.com/STaninnat/ecom-backend/handlers"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
)

type HandlersCartConfig struct {
	CartMG         *intmongo.CartMongo
	HandlersConfig *handlers.HandlersConfig
}

type CartResponse struct {
	Message string `json:"message"`
	OrderID string `json:"order_id"`
}
