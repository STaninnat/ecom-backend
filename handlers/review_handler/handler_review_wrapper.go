package reviewhandlers

import (
	"github.com/STaninnat/ecom-backend/handlers"
	"github.com/STaninnat/ecom-backend/internal/mongo"
)

type HandlersReviewConfig struct {
	ReviewRepo     *mongo.ReviewMongo
	HandlersConfig *handlers.HandlersConfig
}
