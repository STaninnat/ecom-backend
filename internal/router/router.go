package router

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/handlers"
	authhandlers "github.com/STaninnat/ecom-backend/handlers/auth_handler"
	orderhandlers "github.com/STaninnat/ecom-backend/handlers/order_handler"
	paymenthandlers "github.com/STaninnat/ecom-backend/handlers/payment_handler"
	producthandlers "github.com/STaninnat/ecom-backend/handlers/product_handler"
	rolehandlers "github.com/STaninnat/ecom-backend/handlers/role_handler"
	uploadawshandlers "github.com/STaninnat/ecom-backend/handlers/upload_aws_handler"
	uploadhandlers "github.com/STaninnat/ecom-backend/handlers/upload_local_handler"
	userhandlers "github.com/STaninnat/ecom-backend/handlers/user_handler"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

type RouterConfig struct {
	*handlers.HandlersConfig
}

func (apicfg *RouterConfig) SetupRouter(logger *logrus.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middlewares.RequestIDMiddleware)
	router.Use(middlewares.LoggingMiddleware(
		logger,
		map[string]struct{}{"/v1": {}},
		map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
	))

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	fs := http.FileServer(http.Dir("./uploads"))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	roleHandlersConfig := &rolehandlers.HandlersRoleConfig{HandlersConfig: apicfg.HandlersConfig}
	authHandlersConfig := &authhandlers.HandlersAuthConfig{HandlersConfig: apicfg.HandlersConfig}
	userHandlersConfig := &userhandlers.HandlersUserConfig{HandlersConfig: apicfg.HandlersConfig}
	productHandlersConfig := &producthandlers.HandlersProductConfig{HandlersConfig: apicfg.HandlersConfig}
	uploadHandlersConfig := &uploadhandlers.HandlersUploadConfig{HandlersConfig: apicfg.HandlersConfig}
	uploadAWSHandlers := &uploadawshandlers.HandlersUploadAWSConfig{HandlersConfig: apicfg.HandlersConfig}
	orderHandlersConfig := &orderhandlers.HandlersOrderConfig{HandlersConfig: apicfg.HandlersConfig}
	paymentHandlersConfig := &paymenthandlers.HandlersPaymentConfig{HandlersConfig: apicfg.HandlersConfig}

	v1Router := chi.NewRouter()

	// check resp endoint
	v1Router.Get("/healthz", handlers.HandlerReadiness)
	v1Router.Get("/error", handlers.HandlerError)

	// normal endpoint
	v1Router.Post("/auth/signup", authHandlersConfig.HandlerSignUp)
	v1Router.Post("/auth/signin", authHandlersConfig.HandlerSignIn)
	v1Router.Post("/auth/signout", authHandlersConfig.HandlerSignOut)
	v1Router.Post("/auth/refresh", authHandlersConfig.HandlerRefreshToken)

	v1Router.Get("/auth/google/signin", authHandlersConfig.HandlerGoogleSignIn)
	v1Router.Get("/auth/google/callback", authHandlersConfig.HandlerGoogleCallback)

	v1Router.Post("/payments/webhook", paymentHandlersConfig.HandlerStripeWebhook)

	v1Router.Get("/products", apicfg.HandlerOptionalMiddleware(productHandlersConfig.HandlerGetAllProducts))
	v1Router.Get("/products/filter", apicfg.HandlerOptionalMiddleware(productHandlersConfig.HandlerFilterProducts))
	v1Router.Get("/categories", apicfg.HandlerOptionalMiddleware(productHandlersConfig.HandlerGetAllCategories))

	v1Router.Get("/users", apicfg.HandlerMiddleware(userHandlersConfig.HandlerGetUser))
	v1Router.Put("/users", apicfg.HandlerMiddleware(userHandlersConfig.HandlerUpdateUser))

	v1Router.Get("/products/{id}", apicfg.HandlerMiddleware(productHandlersConfig.HandlerGetProductByID))

	v1Router.Post("/orders", apicfg.HandlerMiddleware(orderHandlersConfig.HandlerCreateOrder))
	v1Router.Get("/orders/user", apicfg.HandlerMiddleware(orderHandlersConfig.HandlerGetUserOrders))
	v1Router.Get("/orders/items/{order_id}", apicfg.HandlerMiddleware(orderHandlersConfig.HandlerGetOrderItemsByOrderID))

	v1Router.Post("/payments/intent", apicfg.HandlerMiddleware(paymentHandlersConfig.HandlerCreatePayment))
	v1Router.Post("/payments/confirm", apicfg.HandlerMiddleware(paymentHandlersConfig.HandlerConfirmPayment))
	v1Router.Get("/payments/{order_id}", apicfg.HandlerMiddleware(paymentHandlersConfig.HandlerGetPayment))
	v1Router.Get("/payments/history", apicfg.HandlerMiddleware(paymentHandlersConfig.HandlerGetPaymentHistory))
	v1Router.Post("/payments/{order_id}/refund", apicfg.HandlerMiddleware(paymentHandlersConfig.HandlerRefundPayment))

	// admin endpoint
	v1Router.Post("/admin/user/promote", apicfg.HandlerAdminOnlyMiddleware(roleHandlersConfig.PromoteUserToAdmin))

	v1Router.Post("/categories", apicfg.HandlerAdminOnlyMiddleware(productHandlersConfig.HandlerCreateCategory))
	v1Router.Put("/categories", apicfg.HandlerAdminOnlyMiddleware(productHandlersConfig.HandlerUpdateCategory))
	v1Router.Delete("/categories/{id}", apicfg.HandlerAdminOnlyMiddleware(productHandlersConfig.HandlerDeleteCategory))

	v1Router.Post("/products", apicfg.HandlerAdminOnlyMiddleware(productHandlersConfig.HandlerCreateProduct))
	v1Router.Put("/products", apicfg.HandlerAdminOnlyMiddleware(productHandlersConfig.HandlerUpdateProduct))
	v1Router.Delete("/products/{id}", apicfg.HandlerAdminOnlyMiddleware(productHandlersConfig.HandlerDeleteProduct))

	v1Router.Post("/products/upload-image", apicfg.HandlerAdminOnlyMiddleware(uploadHandlersConfig.HandlerUploadProductImage))
	v1Router.Post("/products/{id}/image", apicfg.HandlerAdminOnlyMiddleware(uploadHandlersConfig.HandlerUpdateProductImageByID))

	v1Router.Post("/products/upload-image-s3", apicfg.HandlerAdminOnlyMiddleware(uploadAWSHandlers.HandlersUploadProductImageS3))
	v1Router.Post("/products/{id}/image-s3", apicfg.HandlerAdminOnlyMiddleware(uploadAWSHandlers.HandlerUpdateProductImageS3ByID))

	v1Router.Put("/orders/{order_id}/status", apicfg.HandlerAdminOnlyMiddleware(orderHandlersConfig.HandlerUpdateOrderStatus))
	v1Router.Delete("/orders/{order_id}", apicfg.HandlerAdminOnlyMiddleware(orderHandlersConfig.HandlerDeleteOrder))
	v1Router.Get("/orders", apicfg.HandlerAdminOnlyMiddleware(orderHandlersConfig.HandlerGetAllOrders))

	v1Router.Get("/admin/payments/{status}", apicfg.HandlerAdminOnlyMiddleware(paymentHandlersConfig.HandlerAdminGetPayments))

	router.Mount("/v1", v1Router)
	return router
}
