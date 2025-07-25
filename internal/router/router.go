// Package router defines HTTP routing, adapters, and related logic for the ecom-backend project.
package router

import (
	"net/http"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/STaninnat/ecom-backend/handlers"
	authhandlers "github.com/STaninnat/ecom-backend/handlers/auth"
	carthandlers "github.com/STaninnat/ecom-backend/handlers/cart"
	categoryhandlers "github.com/STaninnat/ecom-backend/handlers/category"
	orderhandlers "github.com/STaninnat/ecom-backend/handlers/order"
	paymenthandlers "github.com/STaninnat/ecom-backend/handlers/payment"
	producthandlers "github.com/STaninnat/ecom-backend/handlers/product"
	reviewhandlers "github.com/STaninnat/ecom-backend/handlers/review"
	uploadhandlers "github.com/STaninnat/ecom-backend/handlers/upload"
	userhandlers "github.com/STaninnat/ecom-backend/handlers/user"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/middlewares"
)

// router.go: Main API router setup, middleware configuration, and route registration.

// Config holds the configuration for setting up the API router.
type Config struct {
	*handlers.Config
}

// SetupRouter initializes and returns the main chi.Mux router for the API.
// Sets up global middleware, static file serving, and mounts the versioned API subrouter.
// Organizes resource-specific subrouters for clarity and maintainability.
func (apicfg *Config) SetupRouter(logger *logrus.Logger) *chi.Mux {
	router := chi.NewRouter()

	apicfg.setupGlobalMiddleware(router, logger)
	apicfg.setupStaticFileServer(router)

	handlerConfigs := apicfg.createHandlerConfigs()
	apicfg.setupUploadHandlers(handlerConfigs)
	apicfg.setupMongoHandlers(handlerConfigs, logger)

	cacheConfigs := apicfg.createCacheConfigs()
	v1Router := apicfg.createV1Router(handlerConfigs, cacheConfigs)

	router.Mount("/v1", v1Router)
	return router
}

func (apicfg *Config) setupGlobalMiddleware(router *chi.Mux, logger *logrus.Logger) {
	// Standard logging for all requests
	router.Use(middleware.Logger)
	// Recover from panics and return 500 errors
	router.Use(middleware.Recoverer)

	// Security headers for all responses (custom middleware)
	router.Use(middlewares.SecurityHeaders)
	// Attach a unique request ID to each request (custom middleware)
	router.Use(middlewares.RequestIDMiddleware)
	// Custom logging middleware with path-based filtering:
	// - Only logs requests to /v1 and its subpaths
	// - Skips logging for /v1/healthz and /v1/error endpoints
	router.Use(middlewares.LoggingMiddleware(
		logger,
		map[string]struct{}{"/v1": {}},
		map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
	))

	// Add distributed rate limiting middleware (100 requests per 15 minutes per IP, backed by Redis)
	router.Use(middlewares.RedisRateLimiter(apicfg.RedisClient, 100, 15*time.Minute))

	// CORS middleware: allows cross-origin requests from any HTTP/HTTPS origin.
	// - AllowedOrigins: Accepts all subdomains for both http and https (useful for dev and prod)
	// - AllowedMethods: Permits common RESTful methods
	// - AllowedHeaders: Accepts any header (for flexibility with clients)
	// - ExposedHeaders: Exposes 'Link' header to clients
	// - AllowCredentials: Disallows cookies/auth headers by default (set to true if needed)
	// - MaxAge: Caches preflight response for 5 minutes
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))
}

func (apicfg *Config) setupStaticFileServer(router *chi.Mux) {
	// --- Static File Server ---
	// Serve static files from the same directory as uploads (uploadPath).
	// This ensures that files uploaded via the local backend are accessible at /static/*.
	// If using S3, this will only serve files present in the local uploadPath directory.
	//
	// Why this matters:
	// - If uploadPath and the static file server path differ, uploaded files won't be accessible at /static/*.
	// - If using S3, files uploaded to S3 will NOT be accessible at /static/* unless you sync them to the local uploadPath or implement an S3 proxy handler.
	//
	// Example:
	//   uploadPath = "./uploads"  -->  /static/* serves uploaded files (local backend)
	//   uploadPath = "/var/data/uploads"  -->  /static/* serves from that directory
	//   S3 backend  -->  /static/* only serves files present in uploadPath, not S3
	fs := http.FileServer(http.Dir(apicfg.UploadPath))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))
}

type handlerConfigs struct {
	auth     *authhandlers.HandlersAuthConfig
	user     *userhandlers.HandlersUserConfig
	product  *producthandlers.HandlersProductConfig
	category *categoryhandlers.HandlersCategoryConfig
	upload   any
	order    *orderhandlers.HandlersOrderConfig
	payment  *paymenthandlers.HandlersPaymentConfig
	review   *reviewhandlers.HandlersReviewConfig
	cart     *carthandlers.HandlersCartConfig
}

func (apicfg *Config) createHandlerConfigs() *handlerConfigs {
	// --- Handler Configurations ---
	// Auth handler config: provides dependencies for auth-related handlers
	authHandlersConfig := &authhandlers.HandlersAuthConfig{Config: apicfg.Config}
	// User handler config: provides dependencies for user-related handlers
	userHandlersConfig := &userhandlers.HandlersUserConfig{Config: apicfg.Config}
	// Product handler config: includes DB, connection, and logger for product endpoints
	productHandlersConfig := &producthandlers.HandlersProductConfig{
		DB:     apicfg.DB,
		DBConn: apicfg.DBConn,
		Logger: apicfg.Config,
	}
	// Category handler config: for category endpoints
	categoryHandlersConfig := &categoryhandlers.HandlersCategoryConfig{Config: apicfg.Config}

	// --- Order and Payment Handler Configs ---
	orderHandlersConfig := &orderhandlers.HandlersOrderConfig{Config: apicfg.Config}
	paymentHandlersConfig := &paymenthandlers.HandlersPaymentConfig{Config: apicfg.Config}

	// Initialize MongoDB-dependent configs as nil
	var cartConfig *carthandlers.HandlersCartConfig
	var reviewConfig *reviewhandlers.HandlersReviewConfig

	return &handlerConfigs{
		auth:     authHandlersConfig,
		user:     userHandlersConfig,
		product:  productHandlersConfig,
		category: categoryHandlersConfig,
		order:    orderHandlersConfig,
		payment:  paymentHandlersConfig,
		cart:     cartConfig,
		review:   reviewConfig,
	}
}

func (apicfg *Config) setupUploadHandlers(configs *handlerConfigs) {
	// --- Upload Handler Setup ---
	// Set up the upload handler and service, supporting both S3 and local backends
	productDB := uploadhandlers.NewProductDBAdapter(apicfg.DB)
	var fileStorage uploadhandlers.FileStorage
	if apicfg.UploadBackend == "s3" {
		// Use S3 for file storage
		fileStorage = &uploadhandlers.S3FileStorage{
			S3Client:   apicfg.S3Client, // AWS S3 client
			BucketName: apicfg.S3Bucket, // S3 bucket name
		}
		// Upload service combines DB, path, and storage backend
		uploadService := uploadhandlers.NewUploadService(productDB, apicfg.UploadPath, fileStorage)
		// Upload handler config: provides dependencies for S3 upload endpoints
		configs.upload = &uploadhandlers.HandlersUploadS3Config{
			Config:     apicfg.Config,
			Logger:     apicfg.Config,
			UploadPath: apicfg.UploadPath,
			Service:    uploadService,
		}
	} else {
		// Use local filesystem for file storage
		fileStorage = &uploadhandlers.LocalFileStorage{}
		// Upload service combines DB, path, and storage backend
		uploadService := uploadhandlers.NewUploadService(productDB, apicfg.UploadPath, fileStorage)
		// Upload handler config: provides dependencies for local upload endpoints
		configs.upload = &uploadhandlers.HandlersUploadConfig{
			Config:     apicfg.Config,
			Logger:     apicfg.Config,
			UploadPath: apicfg.UploadPath,
			Service:    uploadService,
		}
	}
}

func (apicfg *Config) setupMongoHandlers(configs *handlerConfigs, logger *logrus.Logger) {
	// --- Review and Cart Service Setup ---
	if apicfg.MongoDB != nil {
		reviewMongoRepo := intmongo.NewReviewMongo(apicfg.MongoDB)
		cartMongoRepo := intmongo.NewCartMongo(apicfg.MongoDB)

		// Review handler config and service
		configs.review = &reviewhandlers.HandlersReviewConfig{
			Config: apicfg.Config,
		}
		reviewService := reviewhandlers.NewReviewService(reviewMongoRepo)
		err := configs.review.InitReviewService(reviewService)
		if err != nil {
			logger.Fatal("Failed to initialize review service:", err)
		}

		// Cart handler config and service
		configs.cart = &carthandlers.HandlersCartConfig{
			Config: apicfg.Config,
		}
		cartService := carthandlers.NewCartServiceWithDeps(
			cartMongoRepo,
			apicfg.DB,
			apicfg.DBConn,
			apicfg.RedisClient,
		)
		if err := configs.cart.InitCartService(cartService); err != nil {
			logger.Fatal("Failed to initialize cart service:", err)
		}
	}
}

func (apicfg *Config) createCacheConfigs() map[string]middlewares.CacheConfig {
	// --- Cache Configurations ---
	// Add caching for read-heavy endpoints
	productsCacheConfig := middlewares.CacheConfig{
		TTL:          30 * time.Minute, // Cache products for 30 minutes
		KeyPrefix:    "products",
		CacheService: apicfg.CacheService,
	}

	categoriesCacheConfig := middlewares.CacheConfig{
		TTL:          1 * time.Hour, // Cache categories for 1 hour
		KeyPrefix:    "categories",
		CacheService: apicfg.CacheService,
	}

	return map[string]middlewares.CacheConfig{
		"products":   productsCacheConfig,
		"categories": categoriesCacheConfig,
	}
}

func (apicfg *Config) createV1Router(configs *handlerConfigs, cacheConfigs map[string]middlewares.CacheConfig) *chi.Mux {
	v1Router := chi.NewRouter()

	// --- Health and Error Endpoints ---
	v1Router.Get("/readiness", Adapt(handlers.HandlerReadiness)) // Health check endpoint
	v1Router.Get("/healthz", Adapt(handlers.HandlerHealth))      // Detailed health check endpoint
	v1Router.Get("/errorz", Adapt(handlers.HandlerError))        // Error simulation endpoint

	// --- Swagger UI ---
	v1Router.Get("/swagger/*", httpSwagger.WrapHandler)

	apicfg.setupAuthRoutes(v1Router, configs.auth)
	apicfg.setupUserRoutes(v1Router, configs.user)
	apicfg.setupProductRoutes(v1Router, configs.product, configs.upload, cacheConfigs["products"])
	apicfg.setupCategoryRoutes(v1Router, configs.category, cacheConfigs["categories"])
	apicfg.setupOrderRoutes(v1Router, configs.order)
	apicfg.setupCartRoutes(v1Router, configs.cart)
	apicfg.setupPaymentRoutes(v1Router, configs.payment)
	apicfg.setupReviewRoutes(v1Router, configs.review)
	apicfg.setupAdminRoutes(v1Router, configs.user)

	return v1Router
}

func (apicfg *Config) setupAuthRoutes(v1Router *chi.Mux, authConfig *authhandlers.HandlersAuthConfig) {
	// --- Auth Subrouter ---
	authRouter := chi.NewRouter()
	authRouter.Post("/signup", middlewares.NoCacheHeaders(Adapt(authConfig.HandlerSignUp)).(http.HandlerFunc))        // User registration
	authRouter.Post("/signin", middlewares.NoCacheHeaders(Adapt(authConfig.HandlerSignIn)).(http.HandlerFunc))        // User login
	authRouter.Post("/signout", middlewares.NoCacheHeaders(Adapt(authConfig.HandlerSignOut)).(http.HandlerFunc))      // User logout
	authRouter.Post("/refresh", middlewares.NoCacheHeaders(Adapt(authConfig.HandlerRefreshToken)).(http.HandlerFunc)) // Refresh JWT tokens
	authRouter.Get("/google/signin", Adapt(authConfig.HandlerGoogleSignIn))                                           // Google OAuth2 start
	authRouter.Get("/google/callback", Adapt(authConfig.HandlerGoogleCallback))                                       // Google OAuth2 callback
	v1Router.Mount("/auth", authRouter)
}

func (apicfg *Config) setupUserRoutes(v1Router *chi.Mux, userConfig *userhandlers.HandlersUserConfig) {
	// --- User Subrouter ---
	usersRouter := chi.NewRouter()
	usersRouter.Get("/", middlewares.NoCacheHeaders(WithUser(userConfig.AuthHandlerGetUser)).(http.HandlerFunc))    // Get current user profile
	usersRouter.Put("/", middlewares.NoCacheHeaders(WithUser(userConfig.AuthHandlerUpdateUser)).(http.HandlerFunc)) // Update user profile
	v1Router.Mount("/users", usersRouter)
}

func (apicfg *Config) setupProductRoutes(v1Router *chi.Mux, productConfig *producthandlers.HandlersProductConfig, uploadConfig any, cacheConfig middlewares.CacheConfig) {
	// --- Product Subrouter ---
	productsRouter := chi.NewRouter()
	productsRouter.Get("/", middlewares.CacheMiddleware(cacheConfig)(WithOptionalUser(productConfig.HandlerGetAllProducts)).(http.HandlerFunc))                      // List all products (cached)
	productsRouter.Get("/filter", middlewares.CacheMiddleware(cacheConfig)(WithOptionalUser(productConfig.HandlerFilterProducts)).(http.HandlerFunc))                // Filter products (cached)
	productsRouter.Get("/{id}", WithUser(productConfig.HandlerGetProductByID))                                                                                       // Get product details (requires auth)
	productsRouter.Post("/", middlewares.InvalidateCache(apicfg.CacheService, "products:*")(WithAdmin(productConfig.HandlerCreateProduct)).(http.HandlerFunc))       // Admin: create product, invalidates cache
	productsRouter.Put("/", middlewares.InvalidateCache(apicfg.CacheService, "products:*")(WithAdmin(productConfig.HandlerUpdateProduct)).(http.HandlerFunc))        // Admin: update product, invalidates cache
	productsRouter.Delete("/{id}", middlewares.InvalidateCache(apicfg.CacheService, "products:*")(WithAdmin(productConfig.HandlerDeleteProduct)).(http.HandlerFunc)) // Admin: delete product, invalidates cache
	// Use correct upload handler based on backend
	if apicfg.UploadBackend == "s3" {
		s3UploadConfig := uploadConfig.(*uploadhandlers.HandlersUploadS3Config)
		productsRouter.Post("/upload-image", WithAdmin(s3UploadConfig.HandlerS3UploadProductImage))
		productsRouter.Post("/{id}/image", WithAdmin(s3UploadConfig.HandlerS3UpdateProductImageByID))
	} else {
		localUploadConfig := uploadConfig.(*uploadhandlers.HandlersUploadConfig)
		productsRouter.Post("/upload-image", WithAdmin(localUploadConfig.HandlerUploadProductImage))
		productsRouter.Post("/{id}/image", WithAdmin(localUploadConfig.HandlerUpdateProductImageByID))
	}
	v1Router.Mount("/products", productsRouter)
}

func (apicfg *Config) setupCategoryRoutes(v1Router *chi.Mux, categoryConfig *categoryhandlers.HandlersCategoryConfig, cacheConfig middlewares.CacheConfig) {
	// --- Category Subrouter ---
	categoriesRouter := chi.NewRouter()
	categoriesRouter.Get("/", middlewares.CacheMiddleware(cacheConfig)(WithOptionalUser(categoryConfig.HandlerGetAllCategories)).(http.HandlerFunc))                       // List all categories (cached)
	categoriesRouter.Post("/", middlewares.InvalidateCache(apicfg.CacheService, "categories:*")(WithAdmin(categoryConfig.HandlerCreateCategory)).(http.HandlerFunc))       // Admin: create category, invalidates cache
	categoriesRouter.Put("/", middlewares.InvalidateCache(apicfg.CacheService, "categories:*")(WithAdmin(categoryConfig.HandlerUpdateCategory)).(http.HandlerFunc))        // Admin: update category, invalidates cache
	categoriesRouter.Delete("/{id}", middlewares.InvalidateCache(apicfg.CacheService, "categories:*")(WithAdmin(categoryConfig.HandlerDeleteCategory)).(http.HandlerFunc)) // Admin: delete category, invalidates cache
	v1Router.Mount("/categories", categoriesRouter)
}

func (apicfg *Config) setupOrderRoutes(v1Router *chi.Mux, orderConfig *orderhandlers.HandlersOrderConfig) {
	// --- Order Subrouter ---
	ordersRouter := chi.NewRouter()
	ordersRouter.Post("/", WithUser(orderConfig.HandlerCreateOrder))                           // Create new order
	ordersRouter.Get("/user", WithUser(orderConfig.HandlerGetUserOrders))                      // Get orders for current user
	ordersRouter.Get("/items/{order_id}", WithUser(orderConfig.HandlerGetOrderItemsByOrderID)) // Get items for a specific order
	ordersRouter.Put("/{order_id}/status", WithAdmin(orderConfig.HandlerUpdateOrderStatus))    // Admin: update order status
	ordersRouter.Delete("/{order_id}", WithAdmin(orderConfig.HandlerDeleteOrder))              // Admin: delete order
	ordersRouter.Get("/", WithAdmin(orderConfig.HandlerGetAllOrders))                          // Admin: list all orders
	v1Router.Mount("/orders", ordersRouter)
}

func (apicfg *Config) setupCartRoutes(v1Router *chi.Mux, cartConfig *carthandlers.HandlersCartConfig) {
	// --- Cart Subrouter ---
	// Only register cart routes if MongoDB is configured and cart config is initialized
	if apicfg.MongoDB != nil && cartConfig != nil {
		cartRouter := chi.NewRouter()
		cartRouter.Post("/items", WithUser(cartConfig.HandlerAddItemToUserCart))        // Add item to user cart
		cartRouter.Put("/items", WithUser(cartConfig.HandlerUpdateItemQuantity))        // Update item quantity in user cart
		cartRouter.Get("/items", WithUser(cartConfig.HandlerGetUserCart))               // Get current user's cart
		cartRouter.Delete("/items", WithUser(cartConfig.HandlerRemoveItemFromUserCart)) // Remove item from user cart
		cartRouter.Delete("/", WithUser(cartConfig.HandlerClearUserCart))               // Clear user cart
		cartRouter.Post("/checkout", WithUser(cartConfig.HandlerCheckoutUserCart))      // Checkout user cart
		v1Router.Mount("/cart", cartRouter)
	}
	// --- Guest Cart Subrouter ---
	// Only register guest cart routes if MongoDB is configured and cart config is initialized
	if apicfg.MongoDB != nil && cartConfig != nil {
		guestCartRouter := chi.NewRouter()
		guestCartRouter.Post("/items", Adapt(cartConfig.HandlerAddItemToGuestCart))        // Add item to guest cart (no auth)
		guestCartRouter.Get("/", Adapt(cartConfig.HandlerGetGuestCart))                    // Get guest cart (no auth)
		guestCartRouter.Put("/items", Adapt(cartConfig.HandlerUpdateGuestItemQuantity))    // Update item in guest cart (no auth)
		guestCartRouter.Delete("/items", Adapt(cartConfig.HandlerRemoveItemFromGuestCart)) // Remove item from guest cart (no auth)
		guestCartRouter.Delete("/", Adapt(cartConfig.HandlerClearGuestCart))               // Clear guest cart (no auth)
		v1Router.Mount("/guest-cart", guestCartRouter)
	}
}

func (apicfg *Config) setupPaymentRoutes(v1Router *chi.Mux, paymentConfig *paymenthandlers.HandlersPaymentConfig) {
	// --- Payment Subrouter ---
	paymentsRouter := chi.NewRouter()
	paymentsRouter.Post("/webhook", Adapt(paymentConfig.HandlerStripeWebhook))              // Stripe webhook endpoint
	paymentsRouter.Post("/intent", WithUser(paymentConfig.HandlerCreatePayment))            // Create payment intent
	paymentsRouter.Post("/confirm", WithUser(paymentConfig.HandlerConfirmPayment))          // Confirm payment
	paymentsRouter.Get("/{order_id}", WithUser(paymentConfig.HandlerGetPayment))            // Get payment for order
	paymentsRouter.Get("/history", WithUser(paymentConfig.HandlerGetPaymentHistory))        // Get payment history for user
	paymentsRouter.Post("/{order_id}/refund", WithUser(paymentConfig.HandlerRefundPayment)) // Refund payment for order
	paymentsRouter.Get("/admin/{status}", WithAdmin(paymentConfig.HandlerAdminGetPayments)) // Admin: get payments by status
	v1Router.Mount("/payments", paymentsRouter)
}

func (apicfg *Config) setupReviewRoutes(v1Router *chi.Mux, reviewConfig *reviewhandlers.HandlersReviewConfig) {
	// --- Review Subrouter ---
	if reviewConfig != nil {
		reviewsRouter := chi.NewRouter()
		reviewsRouter.Get("/product/{product_id}", Adapt(reviewConfig.HandlerGetReviewsByProductID)) // Get reviews for a product
		reviewsRouter.Get("/{id}", Adapt(reviewConfig.HandlerGetReviewByID))                         // Get review by ID
		reviewsRouter.Post("/", WithUser(reviewConfig.HandlerCreateReview))                          // Create review (auth required)
		reviewsRouter.Get("/user", WithUser(reviewConfig.HandlerGetReviewsByUserID))                 // Get reviews by user
		reviewsRouter.Put("/{id}", WithUser(reviewConfig.HandlerUpdateReviewByID))                   // Update review (auth required)
		reviewsRouter.Delete("/{id}", WithUser(reviewConfig.HandlerDeleteReviewByID))                // Delete review (auth required)
		v1Router.Mount("/reviews", reviewsRouter)
	}
}

func (apicfg *Config) setupAdminRoutes(v1Router *chi.Mux, userConfig *userhandlers.HandlersUserConfig) {
	// --- Admin Subrouter ---
	adminRouter := chi.NewRouter()
	adminRouter.Post("/user/promote", WithAdmin(userConfig.AuthHandlerPromoteUserToAdmin)) // Promote user to admin
	v1Router.Mount("/admin", adminRouter)
}
