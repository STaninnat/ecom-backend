package authhandlers

import (
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/handlers/cart"
)

type HandlersAuthConfig struct {
	*handlers.HandlersConfig
	*cart.HandlersCartConfig
}
