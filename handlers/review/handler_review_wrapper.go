package reviewhandlers

import (
	"github.com/STaninnat/ecom-backend/handlers"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
)

type HandlersReviewConfig struct {
	ReviewMG       *intmongo.ReviewMongo
	HandlersConfig *handlers.HandlersConfig
}
